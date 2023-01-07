package opa

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/open-policy-agent/opa/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// schema (object[string: any<string, schema>]) representation of body schema
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

// resource (object<type: string, name: string, config: body, decl_range: range, type_range: range>) representation of "resource" blocks
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

// body (object[string: any<expr, array[block]>]) representation of config body
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

// expr (object<value: any, unknown: boolean, sensitive: boolean, range: range>) representation of expressions
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

// block (object<config: object[string: any<expr, array[block]>], decl_range: range>) representation of nested blocks
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

// issue (object<msg: string, range: range>) message and source range
var issueTy = types.Named("issue", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("msg", types.S),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
)).Description("message and source range")

func jsonToIssue(in any, path string) (*Issue, error) {
	ret, err := jsonToObject(in, path)
	if err != nil {
		return nil, err
	}

	msg, err := jsonToString(ret["msg"], fmt.Sprintf("%s.msg", path))
	if err != nil {
		return nil, err
	}
	rng, err := jsonToRange(ret["range"], fmt.Sprintf("%s.range", path))
	if err != nil {
		return nil, err
	}

	return &Issue{message: msg, location: rng}, nil
}

// range (object<filename: string, start: pos, end: pos>) range of a source file in HCL
var rangeTy = types.Named("range", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("filename", types.S),
		types.NewStaticProperty("start", posTy),
		types.NewStaticProperty("end", posTy),
	},
	nil,
)).Description("range of a source file in HCL")

func rangeToJSON(rng hcl.Range) map[string]any {
	return map[string]any{
		"filename": rng.Filename,
		"start":    posToJSON(rng.Start),
		"end":      posToJSON(rng.End),
	}
}

func jsonToRange(in any, path string) (hcl.Range, error) {
	rng, err := jsonToObject(in, path)
	if err != nil {
		return hcl.Range{}, err
	}

	filename, err := jsonToString(rng["filename"], fmt.Sprintf("%s.filename", path))
	if err != nil {
		return hcl.Range{}, err
	}
	start, err := jsonToPos(rng["start"], fmt.Sprintf("%s.start", path))
	if err != nil {
		return hcl.Range{}, err
	}
	end, err := jsonToPos(rng["end"], fmt.Sprintf("%s.end", path))
	if err != nil {
		return hcl.Range{}, err
	}

	return hcl.Range{Filename: filename, Start: start, End: end}, nil
}

// pos (object<line: number, column: number, bytes: number>) position of a source file in HCL
var posTy = types.Named("pos", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("line", types.N),
		types.NewStaticProperty("column", types.N),
		types.NewStaticProperty("bytes", types.N),
	},
	nil,
)).Description("position of a source file in HCL")

func posToJSON(pos hcl.Pos) map[string]int {
	return map[string]int{
		"line":   pos.Line,
		"column": pos.Column,
		"bytes":  pos.Byte,
	}
}

func jsonToPos(in any, path string) (hcl.Pos, error) {
	pos, err := jsonToObject(in, path)
	if err != nil {
		return hcl.Pos{}, err
	}

	line, err := jsonToInt(pos["line"], fmt.Sprintf("%s.line", path))
	if err != nil {
		return hcl.Pos{}, err
	}
	column, err := jsonToInt(pos["column"], fmt.Sprintf("%s.column", path))
	if err != nil {
		return hcl.Pos{}, err
	}
	bytes, err := jsonToInt(pos["bytes"], fmt.Sprintf("%s.bytes", path))
	if err != nil {
		return hcl.Pos{}, err
	}

	return hcl.Pos{Line: line, Column: column, Byte: bytes}, nil
}

func jsonToObject(in any, path string) (map[string]any, error) {
	out, ok := in.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s is not object, got %T", path, in)
	}
	return out, nil
}

func jsonToString(in any, path string) (string, error) {
	out, ok := in.(string)
	if !ok {
		return "", fmt.Errorf("%s is not string, got %T", path, in)
	}
	return out, nil
}

func jsonToInt(in any, path string) (int, error) {
	jn, ok := in.(json.Number)
	if !ok {
		return 0, fmt.Errorf("%s is not a number, got %T", path, in)
	}
	num, err := jn.Int64()
	if err != nil {
		return 0, err
	}
	return int(num), nil
}
