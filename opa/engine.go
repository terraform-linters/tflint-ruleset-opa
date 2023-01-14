package opa

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/storage"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/topdown"
	"github.com/open-policy-agent/opa/topdown/print"
	"github.com/open-policy-agent/opa/version"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Engine evaluates policies and returns issues.
// In other words, this is a wrapper of rego.New(...).Eval().
type Engine struct {
	store       storage.Store
	modules     map[string]*ast.Module
	print       print.Hook
	traceWriter io.Writer
	runtime     *ast.Term
}

// NewEngine returns a new engine based on the policies loaded
func NewEngine(ret *loader.Result) (*Engine, error) {
	store, err := ret.Store()
	if err != nil {
		return nil, err
	}

	logWriter := logger.Logger().StandardWriter(&hclog.StandardLoggerOptions{ForceLevel: hclog.Debug})
	printer := topdown.NewPrintHook(logWriter)

	// If TFLINT_OPA_TRACE is set, print traces to the debug log.
	var traceWriter io.Writer
	trace := os.Getenv("TFLINT_OPA_TRACE")
	if trace != "" && trace != "false" && trace != "0" {
		traceWriter = logWriter
	}

	return &Engine{
		store:       store,
		modules:     ret.ParsedModules(),
		print:       printer,
		traceWriter: traceWriter,
		runtime:     runtime(),
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
	traceEnabled := e.traceWriter != nil

	options := []func(*rego.Rego){
		// All rules should be under the "tflint" package
		rego.Query(fmt.Sprintf("data.tflint.%s", rule.RegoName())),
		// Makes it possible to refer to the loaded YAML/JSON as the "data" document
		rego.Store(e.store),
		// Enable strict mode
		rego.Strict(true),
		// Enable strict-builtin-errors to return custom function errors immediately
		rego.StrictBuiltinErrors(true),
		// Enable print() to invoke logger.Debug()
		rego.EnablePrintStatements(true),
		rego.PrintHook(e.print),
		// Enable trace() if TFLINT_OPA_TRACE=true
		rego.Trace(traceEnabled),
		// Enable opa.runtime().env/version/commit
		rego.Runtime(e.runtime),
	}
	// Set policies
	for _, m := range e.modules {
		options = append(options, rego.ParsedModule(m))
	}
	// Enable custom functions (e.g. terraform.resources)
	// Mock functions are usually not needed outside of testing,
	// but are provided for compilation.
	options = append(options, Functions(runner)...)
	options = append(options, MockFunctions()...)

	instance := rego.New(options...)
	rs, err := instance.Eval(context.Background())
	if err != nil {
		return nil, err
	}

	if traceEnabled {
		rego.PrintTrace(e.traceWriter, instance)
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

// RunTest runs a policy test. The details are hidden inside open-policy-agent/opa/tester
// and this is a wrapper of it. Test results are emitted as issues if failed or errored.
//
// A runner is provided, but in many cases the runner is never actually used,
// as test runners are generated inside mock functions. See TesterMockFunctions for details.
func (e *Engine) RunTest(rule *TestRule, runner tflint.Runner) ([]*Issue, error) {
	traceEnabled := e.traceWriter != nil

	testRunner := tester.NewRunner().
		SetStore(e.store).
		CapturePrintOutput(true).
		EnableTracing(traceEnabled).
		SetRuntime(e.runtime).
		SetModules(e.modules).
		AddCustomBuiltins(append(TesterFunctions(runner), TesterMockFunctions()...)).
		Filter(rule.RegoName())

	ch, err := testRunner.RunTests(context.Background(), nil)
	if err != nil {
		return nil, err
	}

	var issues []*Issue
	for ret := range ch {
		if ret.Error != nil {
			// Location is not included as it is not an issue for HCL.
			issues = append(issues, &Issue{
				message: fmt.Sprintf("test errored: %s", ret.Error),
			})
			continue
		}

		if ret.Output != nil {
			logger.Debug(string(ret.Output))
		}
		if traceEnabled {
			topdown.PrettyTrace(e.traceWriter, ret.Trace)
		}

		if ret.Fail {
			issues = append(issues, &Issue{
				message: "test failed",
			})
		}
	}

	return issues, nil
}

func runtime() *ast.Term {
	env := ast.NewObject()
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.NullTerm())
		} else if len(parts) > 1 {
			env.Insert(ast.StringTerm(parts[0]), ast.StringTerm(parts[1]))
		}
	}

	obj := ast.NewObject()
	obj.Insert(ast.StringTerm("env"), ast.NewTerm(env))
	obj.Insert(ast.StringTerm("version"), ast.StringTerm(version.Version))
	obj.Insert(ast.StringTerm("commit"), ast.StringTerm(version.Vcs))

	return ast.NewTerm(obj)
}
