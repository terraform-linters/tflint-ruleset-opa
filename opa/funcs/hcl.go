package funcs

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
)

// hcl.expr_list: exprs := hcl.expr_list(expr)
// Extract a list of expressions from the given static list expression.
// This is equivalent to hcl.ExprList in hashicorp/hcl.
//
//	expr (raw_expr) static list expression which is retrieved as an expr type.
//
// Returns:
//
//	exprs (array[raw_expr]) expressions as elements of the list.
func ExprListFunc() *Function1 {
	return &Function1{
		Function: Function{
			Decl: &rego.Function{
				Name:    "hcl.expr_list",
				Decl:    types.NewFunction(types.Args(rawExprTy), types.NewArray(nil, rawExprTy)),
				Memoize: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			exprs, diags := hcl.ExprList(expr)
			if diags.HasErrors() {
				return nil, diags
			}
			ret := make([]map[string]any, len(exprs))
			for i, e := range exprs {
				ret[i] = rawExprToJSON(e, []byte(src))
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
}

// key_value (object<key: raw_expr, value: raw_expr>) representation of a key value pair
var keyValueTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("key", rawExprTy),
		types.NewStaticProperty("value", rawExprTy),
	},
	nil,
)

// hcl.expr_map: pairs := hcl.expr_map(expr)
//
// Extract a list of key value pairs from the given static map expression.
// This is equivalent to hcl.ExprMap in hashicorp/hcl.
//
//	expr (raw_expr) static map expression which is retrieved as an expr type.
//
// Returns:
//
//	pairs (array[key_value]) key value pairs of the map as expressions.
func ExprMapFunc() *Function1 {
	return &Function1{
		Function: Function{
			Decl: &rego.Function{
				Name:    "hcl.expr_map",
				Decl:    types.NewFunction(types.Args(rawExprTy), types.NewArray(nil, keyValueTy)),
				Memoize: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			pairs, diags := hcl.ExprMap(expr)
			if diags.HasErrors() {
				return nil, diags
			}
			ret := make([]map[string]map[string]any, len(pairs))
			for i, pair := range pairs {
				ret[i] = map[string]map[string]any{
					"key":   rawExprToJSON(pair.Key, []byte(src)),
					"value": rawExprToJSON(pair.Value, []byte(src)),
				}
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
}

// call (object<name: string, name_range: range, arguments: array[raw_expr], args_range: range>) representation of a static function call
var callTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("name_range", rangeTy),
		types.NewStaticProperty("arguments", types.NewArray(nil, rawExprTy)),
		types.NewStaticProperty("args_range", rangeTy),
	},
	nil,
)

// hcl.expr_call: call := hcl.expr_call(expr)
//
// Extract the function name and arguments from the given function call expression.
// This is equivalent to hcl.ExprCall in hashicorp/hcl.
//
//	expr (raw_expr) function call expression which is retrieved as an expr type.
//
// Returns:
//
//	call (call) function call object
func ExprCallFunc() *Function1 {
	return &Function1{
		Function: Function{
			Decl: &rego.Function{
				Name:    "hcl.expr_call",
				Decl:    types.NewFunction(types.Args(rawExprTy), callTy),
				Memoize: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			call, diags := hcl.ExprCall(expr)
			if diags.HasErrors() {
				return nil, diags
			}

			arguments := make([]map[string]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				arguments[i] = rawExprToJSON(arg, []byte(src))
			}
			ret := map[string]any{
				"name":       call.Name,
				"name_range": rangeToJSON(call.NameRange),
				"arguments":  arguments,
				"args_range": rangeToJSON(call.ArgsRange),
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
}

func astAsExpr(v *ast.Term) (hcl.Expression, string, error) {
	var exprMap map[string]any
	if err := ast.As(v.Value, &exprMap); err != nil {
		return nil, "", err
	}

	value, ok := exprMap["value"]
	if !ok {
		return nil, "", fmt.Errorf("expr must have a 'value' key")
	}
	valueStr, ok := value.(string)
	if !ok {
		return nil, "", fmt.Errorf("expr value must be a string")
	}

	rngArg, ok := exprMap["range"]
	if !ok {
		return nil, "", fmt.Errorf("expr must have a 'range' key")
	}
	rng, err := jsonToRange(rngArg, "expr.range")
	if err != nil {
		return nil, "", err
	}

	expr, diags := hclext.ParseExpression([]byte(valueStr), rng.Filename, rng.Start)
	if diags.HasErrors() {
		return nil, "", diags
	}
	return expr, valueStr, nil
}

// prependZeroPadding to the source code so that the position of the expression
// is correct. This is necessary because the position is relative to the start of
// the source code.
func prependZeroPadding(src string, pos hcl.Pos) string {
	return strings.Repeat("0", pos.Byte) + src
}
