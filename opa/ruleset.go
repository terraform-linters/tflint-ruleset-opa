package opa

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// RuleSet is the custom ruleset for OPA
type RuleSet struct {
	tflint.BuiltinRuleSet

	globalConfig *tflint.Config
	config       *Config
}

// ApplyGlobalConfig is normally not expected to be overridden,
// but since rules are defined dynamically by Rego, it's inconvenient
// to enable/disable rules here (Called in the order ApplyGlobalConfig
// -> ApplyConfig).
// So just save the config so that it can be applied after ApplyConfig.
func (r *RuleSet) ApplyGlobalConfig(config *tflint.Config) error {
	r.globalConfig = config
	return nil
}

func (r *RuleSet) ConfigSchema() *hclext.BodySchema {
	r.config = &Config{}
	return hclext.ImpliedBodySchema(r.config)
}

// ApplyConfig loads policies and generates TFLint rules.
// Run ApplyGlobalConfig after the rules are generated.
func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	diags := hclext.DecodeBody(body, nil, r.config)
	if diags.HasErrors() {
		return diags
	}

	// Load local policies first (from policy_dir or the default locations). This
	// is cheap, local-only work, so doing it before any bundle fetch lets us fail
	// fast on a misconfiguration without a network round trip.
	modules := map[string]*ast.Module{}
	var data map[string]interface{}

	policyDir, err := r.config.policyDir()
	if err != nil {
		// Only os.ErrNotExist is tolerable here. If the directory is explicitly
		// declared in config or environment variables, os.ErrNotExist will not
		// be returned, resulting in load errors later in the process.
		if !os.IsNotExist(err) {
			return err
		}
		// No local policy directory present (and none explicitly configured).
	} else {
		ret, err := loader.NewFileLoader().Filtered([]string{policyDir}, nil)
		if err != nil {
			return fmt.Errorf("failed to load policies; %w", err)
		}
		modules = ret.ParsedModules()
		data = ret.Documents
	}

	// When a bundle is configured, it is the sole source of both policies and data.
	// Every rule must live in the `tflint` package for the custom builtins to resolve,
	// so a bundle and a local policy directory would always conflict. Because local
	// policies and data are ignored in favor of the bundle, reject the combination
	// outright rather than silently dropping local content; only an empty (or absent)
	// local directory is allowed alongside a bundle.
	bundleURL := r.config.bundleURL()
	if bundleURL != "" {
		if len(modules) > 0 || len(data) > 0 {
			return errors.New("bundle_url cannot be used together with local policies or data; configure only one")
		}
		b, err := fetchBundle(context.Background(), bundleURL, bundleCacheDir())
		if err != nil {
			return fmt.Errorf("failed to fetch bundle; %w", err)
		}
		modules = map[string]*ast.Module{}
		for _, m := range b.Modules {
			modules[m.Path] = m.Parsed
		}
		data = b.Data
	}

	// Generate rules only when policies were loaded. Even when there are none, we
	// must still call ApplyGlobalConfig below so the global config is applied.
	if len(modules) > 0 {
		store := inmem.NewFromObject(map[string]interface{}{})
		if len(data) > 0 {
			store = inmem.NewFromObject(data)
		}

		engine, err := NewEngine(store, modules)
		if err != nil {
			return fmt.Errorf("failed to initialize a policy engine; %w", err)
		}

		// If TFLINT_OPA_TEST is set, only run tests, not policy checks
		var testMode bool
		test := os.Getenv("TFLINT_OPA_TEST")
		if test != "" && test != "false" && test != "0" {
			testMode = true
		}

		regoRuleNames := map[string]bool{}
		for _, module := range modules {
			for _, regoRule := range module.Rules {
				ruleName := regoRule.Head.Name.String()
				if _, exists := regoRuleNames[ruleName]; exists {
					// Supports incremental rules, simply ignoring rules with the same name.
					continue
				}
				regoRuleNames[ruleName] = true

				if testMode {
					if rule := NewTestRule(regoRule, engine); rule != nil {
						// Empty for local policies; the bundle URL for bundle-sourced rules.
						rule.source = bundleURL
						r.Rules = append(r.Rules, rule)
					}
				} else {
					if rule := NewRule(regoRule, engine); rule != nil {
						// Empty for local policies; the bundle URL for bundle-sourced rules.
						rule.source = bundleURL
						r.Rules = append(r.Rules, rule)
					}
				}
			}
		}
	}

	return r.BuiltinRuleSet.ApplyGlobalConfig(r.globalConfig)
}
