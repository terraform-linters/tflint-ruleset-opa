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

type Function1 struct {
	Function

	Func rego.Builtin1
}

func (f *Function1) Rego() func(*rego.Rego) {
	return rego.Function1(f.Decl, f.Func)
}

func (f *Function1) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

func MockFunction1(base func(tflint.Runner) *Function1) *Function2 {
	return &Function2{
		Function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a)
		},
	}
}

type Function2 struct {
	Function

	Func rego.Builtin2
}

func (f *Function2) Rego() func(*rego.Rego) {
	return rego.Function2(f.Decl, f.Func)
}

func (f *Function2) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

func MockFunction2(base func(tflint.Runner) *Function2) *Function3 {
	return &Function3{
		Function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a, b)
		},
	}
}

type Function3 struct {
	Function

	Func rego.Builtin3
}

func (f *Function3) Rego() func(*rego.Rego) {
	return rego.Function3(f.Decl, f.Func)
}

func (f *Function3) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

func MockFunction3(base func(tflint.Runner) *Function3) *Function4 {
	return &Function4{
		Function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, c *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := tftester.NewRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a, b, c)
		},
	}
}

type Function4 struct {
	Function

	Func rego.Builtin4
}

func (f *Function4) Rego() func(*rego.Rego) {
	return rego.Function4(f.Decl, f.Func)
}

func (f *Function4) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}

type FunctionDyn struct {
	Function

	Func rego.BuiltinDyn
}

func (f *FunctionDyn) Rego() func(*rego.Rego) {
	return rego.FunctionDyn(f.Decl, f.Func)
}

func (f *FunctionDyn) Tester() *tester.Builtin {
	return f.Function.asTester(f.Rego())
}
