package opa

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/lang/marks"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

// schema (object[string: any<string, schema>]) representation of body schema
var schemaTy = types.NewObject(
	nil,
	// Same as types.NewDynamicProperty(types.S, types.Or(types.S, schemaTy)). Recursive type is not supported.
	types.NewDynamicProperty(types.S, types.A),
)

func jsonToSchema(in map[string]any, tyMap map[string]cty.Type, path string) (*hclext.BodySchema, map[string]cty.Type, error) {
	schema := &hclext.BodySchema{}

	for k, v := range in {
		key := fmt.Sprintf("%s.%s", path, k)

		switch cv := v.(type) {
		case string:
			expr, diags := hclsyntax.ParseExpression([]byte(cv), "", hcl.InitialPos)
			if diags.HasErrors() {
				return schema, tyMap, fmt.Errorf("type expr parse error in %s; %s", key, withoutSubject(diags))
			}
			ty, diags := typeexpr.TypeConstraint(expr)
			if diags.HasErrors() {
				return schema, tyMap, fmt.Errorf("type constraint parse error in %s; %s", key, withoutSubject(diags))
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
var optionsTy = types.NewObject(
	nil,
	// Use dynamic properties as optional static properties are not supported.
	types.NewDynamicProperty(types.S, types.S),
)

func jsonToOption(in map[string]string) (*option, error) {
	out := &option{}

	for k, v := range in {
		switch k {
		case "expand_mode":
			out.ExpandModeSet = true
			switch v {
			case "none":
				out.ExpandMode = tflint.ExpandModeNone
			case "expand":
				out.ExpandMode = tflint.ExpandModeExpand
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
var typedBlockTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("type", types.S),
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)

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
var namedBlockTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)

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
var blockTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("config", bodyTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)

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
var localTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("expr", exprTy),
		types.NewStaticProperty("decl_range", rangeTy),
	},
	nil,
)

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
var bodyTy = types.NewObject(
	nil,
	types.NewDynamicProperty(
		types.S,
		types.Or(exprTy, types.NewArray(nil, nestedBlockTy)),
	),
)

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

// expr (object<value: any, unknown: boolean, sensitive: boolean, ephemeral: boolean, range: range>) representation of an expression
var exprTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("value", types.A),
		types.NewStaticProperty("unknown", types.B),
		types.NewStaticProperty("sensitive", types.B),
		types.NewStaticProperty("ephemeral", types.B),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
)

func exprToJSON(expr hcl.Expression, tyMap map[string]cty.Type, path string, runner tflint.Runner) (map[string]any, error) {
	ret := map[string]any{
		"unknown":   false,
		"sensitive": false,
		"ephemeral": false,
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
	if value.ContainsMarked() {
		ret["unknown"] = true
		if marks.Contains(value, marks.Sensitive) {
			ret["sensitive"] = true
		}
		if marks.Contains(value, marks.Ephemeral) {
			ret["ephemeral"] = true
		}
		return ret, nil
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

	value, err = convert.Convert(value, ty)
	if err != nil {
		return ret, fmt.Errorf("type error in %s; %w", expr.Range(), err)
	}

	// Convert cty.Value to JSON representation and unmarshal as any type.
	// This allows values of any type to be valid JSON values.
	out, err := ctyjson.Marshal(value, ty)
	if err != nil {
		return ret, fmt.Errorf("internal marshal error in %s; %w", expr.Range(), err)
	}
	var val any
	if err := json.Unmarshal(out, &val); err != nil {
		return ret, fmt.Errorf("internal unmarshal error in %s; %w", expr.Range(), err)
	}
	ret["value"] = val

	return ret, nil
}

// nested_block (object<config: object[string: any<expr, array[nested_block]>], labels: array[string], decl_range: range>) representation of a nested block
var nestedBlockTy = types.NewObject(
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
)

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
var issueTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("msg", types.S),
		types.NewStaticProperty("range", rangeTy),
	},
	nil,
)

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

	return &Issue{Message: msg, Range: rng}, nil
}

// range (object<filename: string, start: pos, end: pos>) range of a source file
var rangeTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("filename", types.S),
		types.NewStaticProperty("start", posTy),
		types.NewStaticProperty("end", posTy),
	},
	nil,
)

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

// pos (object<line: number, column: number, byte: number>) position of a source file
var posTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("line", types.N),
		types.NewStaticProperty("column", types.N),
		types.NewStaticProperty("byte", types.N),
	},
	nil,
)

func posToJSON(pos hcl.Pos) map[string]int {
	return map[string]int{
		"line":   pos.Line,
		"column": pos.Column,
		"byte":   pos.Byte,
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
	by, err := jsonToInt(pos["byte"], fmt.Sprintf("%s.byte", path))
	if err != nil {
		return hcl.Pos{}, err
	}

	return hcl.Pos{Line: line, Column: column, Byte: by}, nil
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

func withoutSubject(diags hcl.Diagnostics) string {
	count := len(diags)
	switch {
	case count == 0:
		return "no diagnostics"
	case count == 1:
		return fmt.Sprintf("%s; %s", diags[0].Summary, diags[0].Detail)
	default:
		return fmt.Sprintf("%s; %s, and %d other diagnostic(s)", diags[0].Summary, diags[0].Detail, count-1)
	}
}
