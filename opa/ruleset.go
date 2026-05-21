package opa

import (
	"fmt"
	"os"

	"github.com/open-policy-agent/opa/v1/loader"
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

	policyDirs, err := r.config.policyDirs()
	if err != nil {
		// If you declare the directory in config or environment variables,
		// os.ErrNotExist will not be returned, resulting in load errors
		// later in the process.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	ret, err := loader.NewFileLoader().Filtered(policyDirs, nil)
	if err != nil {
		return fmt.Errorf("failed to load policies; %w", err)
	}

	engine, err := NewEngine(ret)
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
	for _, module := range ret.ParsedModules() {
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
