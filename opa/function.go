package opa

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// Functions return custom functions as Rego options.
func Functions(runner tflint.Runner) []func(*rego.Rego) {
	return []func(*rego.Rego){
		resourcesFunc(runner),
	}
}

// terraform.resources: resources := terraform.resources(resource_type, schema)
//
// Returns Terraform resources.
//
//	resource_type (string): resource type to retrieve. "*" is a special character that returns all resources.
//	schema        (schema): Schema for attributes referenced in rules.
//
// Returns:
//
//	resources (array[resource]) Terraform resources
func resourcesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.resources",
			Decl: types.NewFunction(
				types.Args(types.S, schemaTy),
				types.NewArray(nil, resourceTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, a *ast.Term, b *ast.Term) (*ast.Term, error) {
			var resourceType string
			if err := ast.As(a.Value, &resourceType); err != nil {
				return nil, err
			}
			var schemaJSON map[string]any
			if err := ast.As(b.Value, &schemaJSON); err != nil {
				return nil, err
			}
			schema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
			if err != nil {
				return nil, err
			}

			var content *hclext.BodyContent
			// "*" is a special character that returns all resources
			if resourceType == "*" {
				content, err = runner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       schema,
						},
					},
				}, nil)
			} else {
				content, err = runner.GetResourceContent(resourceType, schema, nil)
			}
			if err != nil {
				return nil, err
			}

			resources, err := resourcesToJSON(content.Blocks, tyMap, "schema", runner)
			if err != nil {
				return nil, err
			}
			v, err := ast.InterfaceToValue(resources)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	)
}

// schema (object[string: any<string, schema>]): representation of body schema
var schemaTy = types.Named("schema", types.NewObject(
	nil,
	// Same as types.NewDynamicProperty(types.S, types.Or(types.S, schemaTy)). Recursive type is not supported.
	types.NewDynamicProperty(types.S, types.A),
)).Description("representation of body schema")

func jsonToSchema(in map[string]any, tyMap map[string]cty.Type, path string) (*hclext.BodySchema, map[string]cty.Type, error) {
	schema := &hclext.BodySchema{}

	for k, v := range in {
		key := fmt.Sprintf("%s.%s", path, k)

		switch cv := v.(type) {
		case string:
			expr, diags := hclsyntax.ParseExpression([]byte(cv), "", hcl.InitialPos)
			if diags.HasErrors() {
				return schema, tyMap, fmt.Errorf("type expr parse error in %s; %w", key, diags)
			}
			ty, diags := typeexpr.TypeConstraint(expr)
			if diags.HasErrors() {
				return schema, tyMap, fmt.Errorf("type constraint parse error in %s; %w", key, diags)
			}
			tyMap[key] = ty

			schema.Attributes = append(schema.Attributes, hclext.AttributeSchema{Name: k})

		case map[string]any:
			var inner *hclext.BodySchema
			var err error
			inner, tyMap, err = jsonToSchema(cv, tyMap, key)
			if err != nil {
				return schema, tyMap, err
			}
			schema.Blocks = append(schema.Blocks, hclext.BlockSchema{
				Type: k,
				Body: inner,
			})

		default:
			return schema, tyMap, fmt.Errorf("%s is not string or object, got %T", key, v)
		}
	}

	return schema, tyMap, nil
}

// resource (object<type: string, name: string, config: body, decl_range: range, type_range: range>): representation of "resource" blocks
var resourceTy = types.Named("resource", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("type", types.S),
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
		types.NewStaticProperty("type_range", rangeTy),
	},
	nil,
)).Description(`representation of "resource" blocks`)

func resourcesToJSON(resources hclext.Blocks, tyMap map[string]cty.Type, path string, runner tflint.Runner) ([]map[string]any, error) {
	ret := make([]map[string]any, len(resources))

	for i, block := range resources {
		body, err := bodyToJSON(block.Body, tyMap, path, runner)
		if err != nil {
			return ret, err
		}

		ret[i] = map[string]any{
			"type":       block.Labels[0],
			"name":       block.Labels[1],
			"config":     body,
			"decl_range": rangeToJSON(block.DefRange),
			"type_range": rangeToJSON(block.LabelRanges[0]),
		}
	}
	return ret, nil
}

// body (object[string: any<expr, array[block]>]): representation of config body
var bodyTy = types.Named("body", types.NewObject(
	nil,
	types.NewDynamicProperty(
		types.S,
		types.Or(exprTy, types.NewArray(nil, blockTy)),
	)),
).Description("representation of config body")

func bodyToJSON(body *hclext.BodyContent, tyMap map[string]cty.Type, path string, runner tflint.Runner) (map[string]any, error) {
	ret := map[string]any{}

	for k, attr := range body.Attributes {
		value, err := exprToJSON(attr.Expr, tyMap, fmt.Sprintf("%s.%s", path, k), runner)
		if err != nil {
			return ret, err
		}

		ret[attr.Name] = value
	}

	for _, block := range body.Blocks {
		json, err := blockToJSON(block, tyMap, fmt.Sprintf("%s.%s", path, block.Type), runner)
		if err != nil {
			return ret, err
		}

		switch r := ret[block.Type].(type) {
		case nil:
			ret[block.Type] = []map[string]any{json}
		case []map[string]any:
			ret[block.Type] = append(r, json)
		default:
			panic(fmt.Sprintf("unknown type: %T", ret[block.Type]))
		}
	}

	return ret, nil
}

// expr (object<value: any, unknown: boolean, sensitive: boolean, range: range>): representation of expressions
var exprTy = types.Named("expr", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("value", types.A),
		types.NewStaticProperty("unknown", types.B),
		types.NewStaticProperty("sensitive", types.B),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
)).Description("representation of expressions")

func exprToJSON(expr hcl.Expression, tyMap map[string]cty.Type, path string, runner tflint.Runner) (map[string]any, error) {
	ret := map[string]any{
		"unknown":   false,
		"sensitive": false,
		"range":     rangeToJSON(expr.Range()),
	}

	var value cty.Value
	err := runner.EvaluateExpr(expr, &value, nil)
	if err != nil {
		if errors.Is(err, tflint.ErrSensitive) {
			// "value" is undefined to halt evaluation if the value is unknown
			ret["unknown"] = true
			ret["sensitive"] = true
			return ret, nil
		}
		return ret, err
	}
	if !value.IsWhollyKnown() {
		ret["unknown"] = true
		return ret, nil
	}

	ty, exists := tyMap[path]
	if !exists {
		// should never happen
		panic(fmt.Sprintf("cannot get type of %s", path))
	}
	if ty.HasDynamicTypes() {
		// If a type has "any", it will be converted to JSON as a dynamic type, (e.g. {"value": 1, "type": "number"})
		// so it will take advantage of the inferred type.
		ty = value.Type()
	}

	// Convert cty.Value to JSON representation and unmarshal as any type.
	// This allows values of any type to be valid JSON values.
	out, err := ctyjson.Marshal(value, ty)
	if err != nil {
		return ret, fmt.Errorf("internal marshal error: %w", err)
	}
	var val any
	if err := json.Unmarshal(out, &val); err != nil {
		return ret, fmt.Errorf("internal unmarshal error: %w", err)
	}
	ret["value"] = val

	return ret, nil
}

// block (object<config: object[string: any<expr, array[block]>], decl_range: range>): representation of nested blocks
var blockTy = types.Named("block", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("config", types.NewObject(
			nil,
			// Same as bodyTy
			types.NewDynamicProperty(
				types.S,
				// Same as types.Or(exprTy, types.NewArray(nil, blockTy)
				types.Or(exprTy, types.NewArray(nil, types.NewObject(
					[]*types.StaticProperty{
						types.NewStaticProperty("config", types.NewObject(
							nil,
							types.NewDynamicProperty(
								types.S,
								// block deeper than 3 levels should be any, as recursive type is not supported
								types.Or(exprTy, types.NewArray(nil, types.A)),
							)),
						),
						types.NewStaticProperty("decl_range", rangeTy),
					},
					nil,
				))),
			)),
		),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of nested blocks")

func blockToJSON(block *hclext.Block, tyMap map[string]cty.Type, path string, runner tflint.Runner) (map[string]any, error) {
	body, err := bodyToJSON(block.Body, tyMap, path, runner)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"config":     body,
		"decl_range": rangeToJSON(block.DefRange),
	}, nil
}
