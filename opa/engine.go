package opa

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/topdown/print"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Engine evaluates policies and returns the results.
// In other words, this is a wrapper of rego.New(...).Eval().
type Engine struct {
	store   storage.Store
	modules map[string]*ast.Module
}

// NewEngine returns a new engine based on the policies loaded
func NewEngine(ret *loader.Result) (*Engine, error) {
	store, err := ret.Store()
	if err != nil {
		return nil, err
	}

	return &Engine{
		store:   store,
		modules: ret.ParsedModules(),
	}, nil
}

// Issue is the result of the query.
type Issue struct {
	message  string
	location hcl.Range
}

// RunQuery executes a query referencing a rule and returns the generated
// Set document as Result.
// rego.ResultSet is parsed according to the following conventions:
//
// - All rules should be under the "tflint" package
// - Rule should return a tflint.issue()
//
// Example:
//
// ```
//
//	deny_test[issue] {
//	  [condition]
//
//	  issue := tflint.issue("not allowed", resource.decl_range)
//	}
//
// ```
func (e *Engine) RunQuery(rule *Rule, runner tflint.Runner) ([]*Issue, error) {
	regoOpts := []func(*rego.Rego){
		// All rules should be under the "tflint" package
		rego.Query(fmt.Sprintf("data.tflint.%s", rule.RegoName())),
		// Makes it possible to refer to the loaded YAML/JSON as the "data" document
		rego.Store(e.store),
		// Enable strict-builtin-errors to return custom function errors immediately
		rego.StrictBuiltinErrors(true),
		// Enable print() to invoke logger.Debug()
		rego.EnablePrintStatements(true),
		rego.PrintHook(&PrintHook{}),
	}

	for _, m := range e.modules {
		regoOpts = append(regoOpts, rego.ParsedModule(m))
	}

	regoOpts = append(regoOpts, Functions(runner)...)

	rs, err := rego.New(regoOpts...).Eval(context.Background())
	if err != nil {
		return nil, err
	}

	var issues []*Issue
	for _, result := range rs {
		for _, expr := range result.Expressions {
			values, ok := expr.Value.([]any)
			if !ok {
				return nil, fmt.Errorf("issue is not set, got %T", expr.Value)
			}

			for _, value := range values {
				ret, err := jsonToIssue(value, "issue")
				if err != nil {
					return nil, err
				}
				issues = append(issues, ret)
			}
		}
	}

	return issues, err
}

type PrintHook struct{}

var _ print.Hook = (*PrintHook)(nil)

func (h *PrintHook) Print(ctx print.Context, msg string) error {
	logger.Debug(msg)
	return nil
}
