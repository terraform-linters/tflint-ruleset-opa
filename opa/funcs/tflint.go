package funcs

import (
	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
)

// Issue is the result of the query.
type Issue struct {
	Message string
	Range   hcl.Range
}

// issue (object<msg: string, range: range>) message and source range
var issueTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("msg", types.S),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
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
		Impl: func(_ rego.BuiltinContext, msgArg *ast.Term, rngArg *ast.Term) (*ast.Term, error) {
			return ast.ObjectTerm(
				ast.Item(ast.StringTerm("msg"), msgArg),
				ast.Item(ast.StringTerm("range"), rngArg),
			), nil
		},
	}
}

// AsIssue converts JSON to an Issue object.
func AsIssue(in any) (*Issue, error) {
	ret, err := jsonToObject(in, "issue")
	if err != nil {
		return nil, err
	}

	msg, err := jsonToString(ret["msg"], "issue.msg")
	if err != nil {
		return nil, err
	}
	rng, err := jsonToRange(ret["range"], "issue.range")
	if err != nil {
		return nil, err
	}

	return &Issue{Message: msg, Range: rng}, nil
}
