package opa

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/ast/location"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Rule is a container for rules defined by Rego to satisfy tflint.Rule
type Rule struct {
	tflint.DefaultRule

	engine *Engine

	name     string
	regoName string
	severity tflint.Severity
	location *location.Location
}

var _ tflint.Rule = (*Rule)(nil)

// NewRule returns a tflint.Rule from a Rego rule.
// Note that the rule names in TFLint and in Rego are different.
func NewRule(regoRule *ast.Rule, engine *Engine) *Rule {
	regoName := regoRule.Head.Name.String()

	// All valid rules must start with "deny_" (e.g. deny_test)
	if !strings.HasPrefix(regoName, "deny_") {
		return nil
	}

	return &Rule{
		engine: engine,
		// Add "opa_" to the rule name in TFLint (e.g. opa_deny_test)
		name:     fmt.Sprintf("opa_%s", regoName),
		regoName: regoName,
		location: regoRule.Location,
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

func (r *Rule) Link() string {
	return r.location.String()
}

func (r *Rule) Check(runner tflint.Runner) error {
	results, err := r.engine.RunQuery(r, runner)
	if err != nil {
		return err
	}

	for _, ret := range results {
		if err := runner.EmitIssue(r.WithSeverity(ret.severity), ret.message, ret.location); err != nil {
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
