package opa

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/loader"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// RuleSet is the custom ruleset for OPA
type RuleSet struct {
	tflint.BuiltinRuleSet

	config *tflint.Config
	engine *Engine
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
// and initialize a policy engine.
// The engine will generate rules from the compiled policies
// and it run ApplyGlobalConfig.
func (r *RuleSet) ApplyConfig(body *hclext.BodyContent) error {
	r.engine = EmptyEmgine()

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

	r.engine, err = NewEngine(ret)
	if err != nil {
		return fmt.Errorf("failed to initialize a policy engine; %w", err)
	}

	r.Rules = r.engine.rules

	return r.BuiltinRuleSet.ApplyGlobalConfig(r.config)
}

// Check runs queries and emits issues by results
func (r *RuleSet) Check(runner tflint.Runner) error {
	for _, rule := range r.EnabledRules {
		rule := rule.(*Rule)

		results, err := r.engine.RunQuery(rule)
		if err != nil {
			return fmt.Errorf(`failed to check "%s" rule; %w`, rule.Name(), err)
		}

		for _, ret := range results {
			if err := runner.EmitIssue(rule.WithSeverity(ret.severity), ret.message, hcl.Range{}); err != nil {
				return fmt.Errorf(`failed to emit "%s" rule issue; %w`, rule.Name(), err)
			}
		}
	}

	return nil
}
