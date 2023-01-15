package opa

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/ast/location"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// TestRule is a container for tests defined by Rego to satisfy tflint.Rule
type TestRule struct {
	tflint.DefaultRule

	engine *Engine

	name     string
	regoName string
	location *location.Location
}

var _ tflint.Rule = (*TestRule)(nil)

// NewTestRule returns a tflint.Rule from a Rego rule.
// Note that the rule names in TFLint and in Rego are different.
func NewTestRule(regoRule *ast.Rule, engine *Engine) *TestRule {
	regoName := regoRule.Head.Name.String()

	// All valid tests must start with "test_" (e.g. test_deny)
	if !strings.HasPrefix(regoName, "test_") {
		return nil
	}

	return &TestRule{
		engine: engine,
		// Add "opa_" to the rule name in TFLint (e.g. opa_test_deny)
		name:     fmt.Sprintf("opa_%s", regoName),
		regoName: regoName,
		location: regoRule.Location,
	}
}

func (r *TestRule) Name() string {
	return r.name
}

func (r *TestRule) Enabled() bool {
	return true
}

func (r *TestRule) Severity() tflint.Severity {
	// Severity is always error
	return tflint.ERROR
}

func (r *TestRule) Link() string {
	return r.location.String()
}

func (r *TestRule) Check(runner tflint.Runner) error {
	issues, err := r.engine.RunTest(r, runner)
	if err != nil {
		return err
	}

	for _, issue := range issues {
		if err := runner.EmitIssue(r, issue.message, issue.location); err != nil {
			return err
		}
	}

	return nil
}

func (r *TestRule) RegoName() string {
	return r.regoName
}
