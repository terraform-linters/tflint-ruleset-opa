package opa

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/open-policy-agent/opa/loader"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// RuleSet is the custom ruleset for OPA
type RuleSet struct {
	tflint.BuiltinRuleSet

	config *tflint.Config
}

// ApplyGlobalConfig is normally not expected to be overridden,
// but since rules are defined dynamically by Rego, it's inconvenient
// to enable/disable rules here (Called in the order ApplyGlobalConfig
// -> ApplyConfig).
// So just save the config so that it can be applied after ApplyConfig.
func (r *RuleSet) ApplyGlobalConfig(config *tflint.Config) error {
	r.config = config
	return nil
}

// ApplyConfig loads policies from ~/.tflint.d/policies
// and generates TFLint rules.
// Run ApplyGlobalConfig after the rules are generated.
func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	policyDir := filepath.Join(homedir, ".tflint.d", "policies")

	info, err := os.Stat(policyDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	ret, err := loader.NewFileLoader().Filtered([]string{policyDir}, nil)
	if err != nil {
		return fmt.Errorf("failed to load policies; %w", err)
	}

	engine, err := NewEngine(ret)
	if err != nil {
		return fmt.Errorf("failed to initialize a policy engine; %w", err)
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

			if rule := NewRule(regoRule, engine); rule != nil {
				r.Rules = append(r.Rules, rule)
			}
		}
	}

	return r.BuiltinRuleSet.ApplyGlobalConfig(r.config)
}
