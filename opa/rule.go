package opa

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Rule is a container for rules defined by Rego to satisfy tflint.Rule
type Rule struct {
	tflint.DefaultRule

	engine *Engine

	name     string
	regoName string
	severity tflint.Severity
}

var _ tflint.Rule = (*Rule)(nil)

// NewRule returns a tflint.Rule from rule name in Rego.
// Note that the rule names in TFLint and in Rego are different.
func NewRule(regoName string, engine *Engine) *Rule {
	// All valid rules must start with "deny_" (e.g. deny_test)
	if !strings.HasPrefix(regoName, "deny_") {
		return nil
	}

	return &Rule{
		engine: engine,
		// Add "opa_" to the rule name in TFLint (e.g. opa_deny_test)
		name:     fmt.Sprintf("opa_%s", regoName),
		regoName: regoName,
	}
}

func (r *Rule) Name() string {
	return r.name
}

func (r *Rule) Enabled() bool {
	return true
}

func (r *Rule) Severity() tflint.Severity {
	return r.severity
}

func (r *Rule) Check(runner tflint.Runner) error {
	results, err := r.engine.RunQuery(r)
	if err != nil {
		return err
	}

	for _, ret := range results {
		if err := runner.EmitIssue(r.WithSeverity(ret.severity), ret.message, hcl.Range{}); err != nil {
			return err
		}
	}

	return nil
}

func (r *Rule) RegoName() string {
	return r.regoName
}

func (r *Rule) WithSeverity(severity tflint.Severity) *Rule {
	r.severity = severity
	return r
}
