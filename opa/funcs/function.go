package funcs

import (
	"fmt"
	"strings"

	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	tftester "github.com/terraform-linters/tflint-ruleset-opa/opa/tester"
)

// Function represents a custom OPA function declaration that can be used in policies.
// It wraps the rego.Function and provides methods to create testable builtins.
type Function struct {
	Decl *rego.Function
}

func (f *Function) asTester(impl func(*rego.Rego)) *tester.Builtin {
	return &tester.Builtin{
		Decl: &ast.Builtin{
			Name:             f.Decl.Name,
			Decl:             f.Decl.Decl,
			Nondeterministic: f.Decl.Nondeterministic,
		},
		Func: impl,
	}
}

func (f *Function) mockDecl() Function {
	split := strings.Split(f.Decl.Name, ".")
	if len(split) != 2 {
		panic(`function should be named with "<namespace>.<name>"`)
	}
	namespace := split[0]
	name := split[1]

	// Mock function takes test inputs as its last argument.
	// e.g. terraform.mock_resources(resourceType, schema, options, `{"main.tf": "foo = 1"}`)
	args := f.Decl.Decl.FuncArgs().Args
	args = append(args, types.NewObject(nil, types.NewDynamicProperty(types.S, types.S)))

	return Function{
		Decl: &rego.Function{
			// Mock function names are prefixed with "mock_"
			// e.g. terraform.resources -> terraform.mock_resources
			Name: fmt.Sprintf("%s.mock_%s", namespace, name),
			Decl: types.NewFunction(
				types.Args(args...),
				f.Decl.Decl.Result(),
			),
			Memoize:          f.Decl.Memoize,
			Nondeterministic: f.Decl.Nondeterministic,
		},
	}
}

// Function1 represents a custom OPA function with 1 argument.
type Function1 struct {
	Function

	Impl rego.Builtin1
}

// Rego returns a rego.Rego option that can be used in policy evaluators.
func (f *Function1) Rego() func(*rego.Rego) {
	return rego.Function1(f.Decl, f.Impl)
}

// Tester returns a tester.Builtin that can be used in Rego test runners.
func (f *Function1) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

// MockFunction1 creates a mock function for Function1.
func MockFunction1(base func(tflint.Runner) *Function1) *Function2 {
	return &Function2{
		Function: base(nil).mockDecl(),
		Impl: func(ctx rego.BuiltinContext, a *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Impl(ctx, a)
		},
	}
}

// Function2 represents a custom OPA function with 2 arguments.
type Function2 struct {
	Function

	Impl rego.Builtin2
}

// Rego returns a rego.Rego option that can be used in policy evaluators.
func (f *Function2) Rego() func(*rego.Rego) {
	return rego.Function2(f.Decl, f.Impl)
}

// Tester returns a tester.Builtin that can be used in Rego test runners.
func (f *Function2) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

// MockFunction2 creates a mock function for Function2.
func MockFunction2(base func(tflint.Runner) *Function2) *Function3 {
	return &Function3{
		Function: base(nil).mockDecl(),
		Impl: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Impl(ctx, a, b)
		},
	}
}

// Function3 represents a custom OPA function with 3 arguments.
type Function3 struct {
	Function

	Impl rego.Builtin3
}

// Rego returns a rego.Rego option that can be used in policy evaluators.
func (f *Function3) Rego() func(*rego.Rego) {
	return rego.Function3(f.Decl, f.Impl)
}

// Tester returns a tester.Builtin that can be used in Rego test runners.
func (f *Function3) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

// MockFunction3 creates a mock function for Function3.
func MockFunction3(base func(tflint.Runner) *Function3) *Function4 {
	return &Function4{
		Function: base(nil).mockDecl(),
		Impl: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, c *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Impl(ctx, a, b, c)
		},
	}
}

// Function4 represents a custom OPA function with 4 arguments.
type Function4 struct {
	Function

	Impl rego.Builtin4
}

// Rego returns a rego.Rego option that can be used in policy evaluators.
func (f *Function4) Rego() func(*rego.Rego) {
	return rego.Function4(f.Decl, f.Impl)
}

// Tester returns a tester.Builtin that can be used in Rego test runners.
func (f *Function4) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

// FunctionDyn represents a custom OPA function with dynamic arguments.
type FunctionDyn struct {
	Function

	Impl rego.BuiltinDyn
}

// Rego returns a rego.Rego option that can be used in policy evaluators.
func (f *FunctionDyn) Rego() func(*rego.Rego) {
	return rego.FunctionDyn(f.Decl, f.Impl)
}

// Tester returns a tester.Builtin that can be used in Rego test runners.
func (f *FunctionDyn) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}
