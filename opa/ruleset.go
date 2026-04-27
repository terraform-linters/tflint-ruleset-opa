package opa

import (
	"context"
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/loader"
	"github.com/open-policy-agent/opa/v1/storage"
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

	bundleURL := r.config.bundleURL()
	bundleModules := map[string]*ast.Module{}
	localModules := map[string]*ast.Module{}
	var store storage.Store

	if bundleURL != "" {
		b, err := fetchBundle(context.Background(), bundleURL, bundleCacheDir())
		if err != nil {
			return fmt.Errorf("failed to fetch bundle; %w", err)
		}
		for _, m := range b.Modules {
			bundleModules[m.Path] = m.Parsed
		}
		if b.Data != nil {
			store = inmem.NewFromObject(b.Data)
		}
	}

	policyDir, err := r.config.policyDir()
	if err != nil {
		// Only os.ErrNotExist is tolerable here. If the directory is explicitly
		// declared in config or environment variables, os.ErrNotExist will not
		// be returned, resulting in load errors later in the process.
		if !os.IsNotExist(err) {
			return err
		}
		// No local policies found and no bundle configured — nothing to do.
		if bundleURL == "" {
			return nil
		}
	} else {
		ret, err := loader.NewFileLoader().Filtered([]string{policyDir}, nil)
		if err != nil {
			return fmt.Errorf("failed to load policies; %w", err)
		}
		for k, m := range ret.ParsedModules() {
			localModules[k] = m
		}
		s, err := ret.Store()
		if err != nil {
			return fmt.Errorf("failed to create policy store; %w", err)
		}
		// When local data files exist, they take precedence over bundle data.
		if len(ret.Documents) > 0 {
			store = s
		}
	}

	// Merge modules: local policies override bundle policies with the same package path
	localPackages := map[string]bool{}
	for _, m := range localModules {
		localPackages[m.Package.Path.String()] = true
	}

	modules := map[string]*ast.Module{}
	for k, m := range bundleModules {
		if !localPackages[m.Package.Path.String()] {
			modules[k] = m
		}
	}
	for k, m := range localModules {
		modules[k] = m
	}

	if len(modules) == 0 {
		return nil
	}

	if store == nil {
		store = inmem.NewFromObject(map[string]interface{}{})
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
					r.Rules = append(r.Rules, rule)
				}
			} else {
				if rule := NewRule(regoRule, engine); rule != nil {
					r.Rules = append(r.Rules, rule)
				}
			}
		}
	}

	return r.BuiltinRuleSet.ApplyGlobalConfig(r.globalConfig)
}
