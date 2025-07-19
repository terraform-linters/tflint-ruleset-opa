package funcs

import (
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
)

// tflint.issue: issue := tflint.issue(msg, range)
//
// Returns issue object
//
//	msg   (string) message
//	range (range)  source range
//
// Returns:
//
//	issue (issue) issue object
func IssueFunc() *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name:    "tflint.issue",
				Decl:    types.NewFunction(types.Args(types.S, rangeTy), issueTy),
				Memoize: true,
			},
		},
		Func: func(_ rego.BuiltinContext, msgArg *ast.Term, rngArg *ast.Term) (*ast.Term, error) {
			return ast.ObjectTerm(
				ast.Item(ast.StringTerm("msg"), msgArg),
				ast.Item(ast.StringTerm("range"), rngArg),
			), nil
		},
	}
}
