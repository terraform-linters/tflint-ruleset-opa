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

	// All valud rules must start with deny_/violation_/warn_/notice_ (e.g. deny_test)
	var severity tflint.Severity
	if strings.HasPrefix(regoName, "deny_") || strings.HasPrefix(regoName, "violation_") {
		severity = tflint.ERROR
	} else if strings.HasPrefix(regoName, "warn_") {
		severity = tflint.WARNING
	} else if strings.HasPrefix(regoName, "notice_") {
		severity = tflint.NOTICE
	} else {
		return nil
	}

	return &Rule{
		engine: engine,
		// Add "opa_" to the rule name in TFLint (e.g. opa_deny_test)
		name:     fmt.Sprintf("opa_%s", regoName),
		regoName: regoName,
		severity: severity,
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
	issues, err := r.engine.RunQuery(r, runner)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		if err := runner.EmitIssue(r, issue.Message, issue.Range); err != nil {
			return err
		}
	}

	return nil
}

func (r *Rule) RegoName() string {
	return r.regoName
}
