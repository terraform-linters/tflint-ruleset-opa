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
			// "__labels" is a special character that allows you to set block labels.
			var labels []string
			if lv, exists := cv["__labels"]; exists {
				delete(cv, "__labels")
				clv, ok := lv.([]any)
				if !ok {
					return schema, tyMap, fmt.Errorf("%s.__labels is not array of string, got %T", key, lv)
				}
				for _, v := range clv {
					v, ok := v.(string)
					if !ok {
						return schema, tyMap, fmt.Errorf("%s.__labels is not array of string, got %T", key, lv)
					}
					labels = append(labels, v)
				}
			}

			var inner *hclext.BodySchema
			var err error
			inner, tyMap, err = jsonToSchema(cv, tyMap, key)
			if err != nil {
				return schema, tyMap, err
			}
			schema.Blocks = append(schema.Blocks, hclext.BlockSchema{
				Type:       k,
				LabelNames: labels,
				Body:       inner,
			})

		default:
			return schema, tyMap, fmt.Errorf("%s is not string or object, got %T", key, v)
		}
	}

	return schema, tyMap, nil
}

// options (object[expand_mode?: any<"none", "expand">]) options to change the retrieve/evaluate behavior
var optionsTy = types.Named("options", types.NewObject(
	nil,
	// Use dynamic properties as optional static properties are not supported.
	types.NewDynamicProperty(types.S, types.S),
)).Description("options to change the retrieve/evaluate behavior")

func jsonToOption(in map[string]string) (*option, error) {
	out := &option{}

	for k, v := range in {
		switch k {
		case "expand_mode":
			out.expandModeSet = true
			switch v {
			case "none":
				out.expandMode = tflint.ExpandModeNone
			case "expand":
				out.expandMode = tflint.ExpandModeExpand
			default:
				return out, fmt.Errorf("unknown expand mode: %s", v)
			}

		default:
			return out, fmt.Errorf("unknown option: %s", k)
		}
	}

	return out, nil
}

// typed_block (object<type: string, name: string, config: body, decl_range: range>) representation of a block labeled with type and name
var typedBlockTy = types.Named("typed_block", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("type", types.S),
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of a block labeled with type and name")

func typedBlocksToJSON(blocks hclext.Blocks, tyMap map[string]cty.Type, path string, runner tflint.Runner) ([]map[string]any, error) {
	ret := make([]map[string]any, len(blocks))

	for i, block := range blocks {
		body, err := bodyToJSON(block.Body, tyMap, path, runner)
		if err != nil {
			return ret, err
		}

		ret[i] = map[string]any{
			"type":       block.Labels[0],
			"name":       block.Labels[1],
			"config":     body,
			"decl_range": rangeToJSON(block.DefRange),
		}
	}
	return ret, nil
}

// named_block (object<name: string, config: body, decl_range: range>) representation of a block labeled with name
var namedBlockTy = types.Named("named_block", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of a block labeled with name")

func namedBlocksToJSON(blocks hclext.Blocks, tyMap map[string]cty.Type, path string, runner tflint.Runner) ([]map[string]any, error) {
	ret := make([]map[string]any, len(blocks))

	for i, block := range blocks {
		body, err := bodyToJSON(block.Body, tyMap, path, runner)
		if err != nil {
			return ret, err
		}

		ret[i] = map[string]any{
			"name":       block.Labels[0],
			"config":     body,
			"decl_range": rangeToJSON(block.DefRange),
		}
	}
	return ret, nil
}

// block (object<config: body, decl_range: range>) representation of an unlabeled block
var blockTy = types.Named("block", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of an unlabeled block")

func blocksToJSON(blocks hclext.Blocks, tyMap map[string]cty.Type, path string, runner tflint.Runner) ([]map[string]any, error) {
	ret := make([]map[string]any, len(blocks))

	for i, block := range blocks {
		body, err := bodyToJSON(block.Body, tyMap, path, runner)
		if err != nil {
			return ret, err
		}

		ret[i] = map[string]any{
			"config":     body,
			"decl_range": rangeToJSON(block.DefRange),
		}
	}
	return ret, nil
}

// local (object<name: string, expr: expr, decl_range: range>) representation of a local value
var localTy = types.Named("local", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("expr", exprTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of a local value")

func localsToJSON(locals hclext.Attributes, runner tflint.Runner) ([]map[string]any, error) {
	ret := []map[string]any{}

	for name, attr := range locals {
		expr, err := exprToJSON(attr.Expr, map[string]cty.Type{name: cty.DynamicPseudoType}, name, runner)
		if err != nil {
			return ret, err
		}

		ret = append(ret, map[string]any{
			"name":       name,
			"expr":       expr,
			"decl_range": rangeToJSON(attr.Range),
		})
	}
	return ret, nil
}

// body (object[string: any<expr, array[nested_block]>]) representation of config body
var bodyTy = types.Named("body", types.NewObject(
	nil,
	types.NewDynamicProperty(
		types.S,
		types.Or(exprTy, types.NewArray(nil, nestedBlockTy)),
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
		json, err := nestedBlockToJSON(block, tyMap, fmt.Sprintf("%s.%s", path, block.Type), runner)
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

// expr (object<value: any, unknown: boolean, sensitive: boolean, range: range>) representation of an expression
var exprTy = types.Named("expr", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("value", types.A),
		types.NewStaticProperty("unknown", types.B),
		types.NewStaticProperty("sensitive", types.B),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
)).Description("representation of an expression")

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

// nested_block (object<config: object[string: any<expr, array[nested_block]>], labels: array[string], decl_range: range>) representation of a nested block
var nestedBlockTy = types.Named("nested_block", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("config", types.NewObject(
			nil,
			// Same as bodyTy
			types.NewDynamicProperty(
				types.S,
				// Same as types.Or(exprTy, types.NewArray(nil, nestedBlockTy)
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
						types.NewStaticProperty("labels", types.NewArray(nil, types.S)),
						types.NewStaticProperty("decl_range", rangeTy),
					},
					nil,
				))),
			)),
		),
		types.NewStaticProperty("labels", types.NewArray(nil, types.S)),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)).Description("representation of a nested block")

func nestedBlockToJSON(block *hclext.Block, tyMap map[string]cty.Type, path string, runner tflint.Runner) (map[string]any, error) {
	body, err := bodyToJSON(block.Body, tyMap, path, runner)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"config":     body,
		"labels":     block.Labels,
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

// range (object<filename: string, start: pos, end: pos>) range of a source file
var rangeTy = types.Named("range", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("filename", types.S),
		types.NewStaticProperty("start", posTy),
		types.NewStaticProperty("end", posTy),
	},
	nil,
)).Description("range of a source file")

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

// pos (object<line: number, column: number, bytes: number>) position of a source file
var posTy = types.Named("pos", types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("line", types.N),
		types.NewStaticProperty("column", types.N),
		types.NewStaticProperty("bytes", types.N),
	},
	nil,
)).Description("position of a source file")

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
