package funcs

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-opa/opa/tester"
	"github.com/zclconf/go-cty/cty"
)

var emptyRange = map[string]any{
	"filename": "",
	"start":    map[string]int{"line": 0, "column": 0, "byte": 0},
	"end":      map[string]int{"line": 0, "column": 0, "byte": 0},
}

func TestJSONToSchema(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
		want  *hclext.BodySchema
		tyMap map[string]cty.Type
		err   string
	}{
		{
			name:  "attribute schema",
			input: map[string]any{"instance_type": "string"},
			want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
			},
			tyMap: map[string]cty.Type{"schema.instance_type": cty.String},
		},
		{
			name:  "block schema",
			input: map[string]any{"ebs_block_device": map[string]any{"volume_size": "number"}},
			want: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "ebs_block_device",
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "volume_size"}},
						},
					},
				},
			},
			tyMap: map[string]cty.Type{"schema.ebs_block_device.volume_size": cty.Number},
		},
		{
			name:  "labeled block schema",
			input: map[string]any{"dynamic": map[string]any{"__labels": []any{"type"}}},
			want: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "dynamic",
						LabelNames: []string{"type"},
						Body:       &hclext.BodySchema{},
					},
				},
			},
			tyMap: map[string]cty.Type{},
		},
		{
			name:  "expr type",
			input: map[string]any{"instance_type": "expr"},
			want: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
			},
			tyMap: map[string]cty.Type{"schema.instance_type": exprCty},
		},
		{
			name:  "invalid schema type",
			input: map[string]any{"nested": map[string]any{"number": 1}},
			err:   "schema.nested.number is not string or object, got int",
		},
		{
			name:  "invalid type string",
			input: map[string]any{"nested": map[string]any{"number": "unknown"}},
			err:   `type constraint parse error in schema.nested.number; Invalid type specification; The keyword "unknown" is not a valid type specification.`,
		},
		{
			name:  "invalid labels",
			input: map[string]any{"dynamic": map[string]any{"__labels": "type"}},
			err:   "schema.dynamic.__labels is not array of string, got string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, tyMap, err := jsonToSchema(test.input, map[string]cty.Type{}, "schema")
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			if diff := cmp.Diff(test.want, schema); diff != "" {
				t.Error(diff)
			}

			opts := cmp.Options{
				cmp.Comparer(func(x, y cty.Type) bool {
					return x.GoString() == y.GoString()
				}),
			}
			if diff := cmp.Diff(test.tyMap, tyMap, opts); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestJSONToOption(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]string
		want  *option
		err   string
	}{
		{
			name: "empty",
			want: &option{},
		},
		{
			name:  "expand_mode = none",
			input: map[string]string{"expand_mode": "none"},
			want:  &option{ExpandMode: tflint.ExpandModeNone, ExpandModeSet: true},
		},
		{
			name:  "expand_mode = expand",
			input: map[string]string{"expand_mode": "expand"},
			want:  &option{ExpandMode: tflint.ExpandModeExpand, ExpandModeSet: true},
		},
		{
			name:  "unknown option",
			input: map[string]string{"unknown": "option"},
			err:   "unknown option: unknown",
		},
		{
			name:  "unknown expand_mode",
			input: map[string]string{"expand_mode": "unknown"},
			err:   "unknown expand mode: unknown",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := jsonToOption(test.input)
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestTypedBlocksToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input hclext.Blocks
		want  []map[string]any
	}{
		{
			name: "typed block",
			input: hclext.Blocks{
				{
					Type:   "resource",
					Labels: []string{"aws_instance", "main"},
					Body:   &hclext.BodyContent{},
				},
			},
			want: []map[string]any{
				{
					"type":       "aws_instance",
					"name":       "main",
					"config":     map[string]any{},
					"decl_range": emptyRange,
				},
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := typedBlocksToJSON(test.input, map[string]cty.Type{}, "", runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestNamedBlocksToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input hclext.Blocks
		want  []map[string]any
	}{
		{
			name: "named block",
			input: hclext.Blocks{
				{
					Type:   "provider",
					Labels: []string{"aws"},
					Body:   &hclext.BodyContent{},
				},
			},
			want: []map[string]any{
				{
					"name":       "aws",
					"config":     map[string]any{},
					"decl_range": emptyRange,
				},
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := namedBlocksToJSON(test.input, map[string]cty.Type{}, "", runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestBlocksToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input hclext.Blocks
		want  []map[string]any
	}{
		{
			name: "block",
			input: hclext.Blocks{
				{
					Type: "ebs_block_device",
					Body: &hclext.BodyContent{},
				},
			},
			want: []map[string]any{
				{
					"config":     map[string]any{},
					"decl_range": emptyRange,
				},
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := blocksToJSON(test.input, map[string]cty.Type{}, "", runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestLocalsToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input hclext.Attributes
		want  []map[string]any
	}{
		{
			name: "locals",
			input: hclext.Attributes{
				"foo": {Name: "foo", Expr: hcl.StaticExpr(cty.StringVal("bar"), hcl.Range{})},
			},
			want: []map[string]any{
				{
					"name": "foo",
					"expr": map[string]any{
						"value":     "bar",
						"unknown":   false,
						"sensitive": false,
						"ephemeral": false,
						"range":     emptyRange,
					},
					"decl_range": emptyRange,
				},
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := localsToJSON(test.input, runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestBodyToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input *hclext.BodyContent
		tyMap map[string]cty.Type
		want  map[string]any
	}{
		{
			name: "body",
			input: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"foo": &hclext.Attribute{Name: "foo", Expr: hcl.StaticExpr(cty.StringVal("bar"), hcl.Range{})},
				},
				Blocks: hclext.Blocks{
					{
						Type: "baz",
						Body: &hclext.BodyContent{},
					},
					{
						Type: "baz",
						Body: &hclext.BodyContent{},
					},
				},
			},
			tyMap: map[string]cty.Type{"schema.foo": cty.String},
			want: map[string]any{
				"foo": map[string]any{
					"value":     "bar",
					"unknown":   false,
					"sensitive": false,
					"ephemeral": false,
					"range":     emptyRange,
				},
				"baz": []map[string]any{
					{
						"config":     map[string]any{},
						"labels":     []string(nil),
						"decl_range": emptyRange,
					},
					{
						"config":     map[string]any{},
						"labels":     []string(nil),
						"decl_range": emptyRange,
					},
				},
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := bodyToJSON(test.input, test.tyMap, "schema", runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestExprToJSON(t *testing.T) {
	parseWithPos := func(src string, start hcl.Pos) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", start)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}
	parse := func(src string) hcl.Expression {
		return parseWithPos(src, hcl.InitialPos)
	}

	tests := []struct {
		name   string
		input  hcl.Expression
		ty     cty.Type
		want   map[string]any
		source string
		err    string
	}{
		{
			name:  "string literal",
			input: hcl.StaticExpr(cty.StringVal("foo"), hcl.Range{}),
			ty:    cty.String,
			want: map[string]any{
				"value":     "foo",
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "number literal",
			input: hcl.StaticExpr(cty.NumberIntVal(1), hcl.Range{}),
			ty:    cty.Number,
			want: map[string]any{
				"value":     float64(1),
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "bool literal",
			input: hcl.StaticExpr(cty.BoolVal(true), hcl.Range{}),
			ty:    cty.Bool,
			want: map[string]any{
				"value":     true,
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "list literal",
			input: hcl.StaticExpr(cty.ListVal([]cty.Value{cty.StringVal("foo"), cty.StringVal("bar")}), hcl.Range{}),
			ty:    cty.List(cty.String),
			want: map[string]any{
				"value":     []any{"foo", "bar"},
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "object literal",
			input: hcl.StaticExpr(cty.ObjectVal(map[string]cty.Value{"foo": cty.StringVal("bar")}), hcl.Range{}),
			ty:    cty.Object(map[string]cty.Type{"foo": cty.String}),
			want: map[string]any{
				"value":     map[string]any{"foo": "bar"},
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "null literal",
			input: hcl.StaticExpr(cty.NullVal(cty.String), hcl.Range{}),
			ty:    cty.String,
			want: map[string]any{
				"value":     nil,
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "type conversion",
			input: hcl.StaticExpr(cty.StringVal("1"), hcl.Range{}),
			ty:    cty.Number,
			want: map[string]any{
				"value":     float64(1),
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "dynamic type",
			input: hcl.StaticExpr(cty.StringVal("1"), hcl.Range{}),
			ty:    cty.DynamicPseudoType,
			want: map[string]any{
				"value":     "1",
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "nested dynamic type",
			input: hcl.StaticExpr(cty.ListVal([]cty.Value{cty.StringVal("1")}), hcl.Range{}),
			ty:    cty.List(cty.DynamicPseudoType),
			want: map[string]any{
				"value":     []any{"1"},
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range":     emptyRange,
			},
		},
		{
			name:  "variable",
			input: parse("var.foo"),
			ty:    cty.String,
			want: map[string]any{
				"value":     "bar",
				"unknown":   false,
				"sensitive": false,
				"ephemeral": false,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 8, "byte": 7},
				},
			},
			source: `variable "foo" { default = "bar" }`,
		},
		{
			name:  "unknown",
			input: parse("var.foo"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": false,
				"ephemeral": false,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 8, "byte": 7},
				},
			},
			source: `variable "foo" {}`,
		},
		{
			name:  "composite unknown",
			input: parse("[var.foo]"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": false,
				"ephemeral": false,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 10, "byte": 9},
				},
			},
			source: `variable "foo" {}`,
		},
		{
			name:  "sensitive",
			input: parse("var.foo"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": true,
				"ephemeral": false,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 8, "byte": 7},
				},
			},
			source: `variable "foo" { sensitive = true }`,
		},
		{
			name:  "composite sensitive",
			input: parse("[var.foo]"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": true,
				"ephemeral": false,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 10, "byte": 9},
				},
			},
			source: `variable "foo" { sensitive = true }`,
		},
		{
			name:  "ephemeral",
			input: parse("var.foo"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": false,
				"ephemeral": true,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 8, "byte": 7},
				},
			},
			source: `variable "foo" { ephemeral = true }`,
		},
		{
			name:  "composite ephemeral",
			input: parse("[var.foo]"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": false,
				"ephemeral": true,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 10, "byte": 9},
				},
			},
			source: `variable "foo" { ephemeral = true }`,
		},
		{
			name:  "sensitive + ephemeral",
			input: parse("var.foo"),
			ty:    cty.String,
			want: map[string]any{
				"unknown":   true,
				"sensitive": true,
				"ephemeral": true,
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 1, "column": 1, "byte": 0},
					"end":      map[string]int{"line": 1, "column": 8, "byte": 7},
				},
			},
			source: `
variable "foo" {
  sensitive = true
  ephemeral = true
}`,
		},
		{
			name:  "expr type",
			input: parseWithPos("instance_type", hcl.Pos{Line: 3, Column: 21, Byte: 54}),
			ty:    exprCty,
			want: map[string]any{
				"value": "instance_type",
				"range": map[string]any{
					"filename": "main.tf",
					"start":    map[string]int{"line": 3, "column": 21, "byte": 54},
					"end":      map[string]int{"line": 3, "column": 34, "byte": 67},
				},
			},
			source: `
resource "aws_instance" "main" {
  ignore_changes = [instance_type]
}`,
		},
		{
			name:  "invalid type",
			input: hcl.StaticExpr(cty.StringVal("foo"), hcl.Range{Filename: "main.tf", Start: hcl.InitialPos, End: hcl.InitialPos}),
			ty:    cty.Number,
			err:   "type error in main.tf:1,1-1; a number is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.source})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			got, err := exprToJSON(test.input, map[string]cty.Type{"expr": test.ty}, "expr", runner)
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestNestedBlockToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input *hclext.Block
		want  map[string]any
	}{
		{
			name:  "unlabeled block",
			input: &hclext.Block{Body: &hclext.BodyContent{}},
			want: map[string]any{
				"config":     map[string]any{},
				"labels":     []string(nil),
				"decl_range": emptyRange,
			},
		},
		{
			name:  "labeled block",
			input: &hclext.Block{Labels: []string{"type"}, Body: &hclext.BodyContent{}},
			want: map[string]any{
				"config":     map[string]any{},
				"labels":     []string{"type"},
				"decl_range": emptyRange,
			},
		},
	}

	runner, diags := tester.NewRunner(map[string]string{})
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := nestedBlockToJSON(test.input, map[string]cty.Type{}, "", runner)
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRangeToJSON(t *testing.T) {
	tests := []struct {
		name  string
		input hcl.Range
		want  map[string]any
	}{
		{
			name: "valid range",
			input: hcl.Range{
				Filename: "main.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   1,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 31,
					Byte:   31,
				},
			},
			want: map[string]any{
				"filename": "main.tf",
				"start": map[string]int{
					"line":   1,
					"column": 1,
					"byte":   1,
				},
				"end": map[string]int{
					"line":   1,
					"column": 31,
					"byte":   31,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := rangeToJSON(test.input)

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestJSONToRange(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  hcl.Range
		err   string
	}{
		{
			name: "valid range",
			input: map[string]any{
				"filename": "main.tf",
				"start":    map[string]any{"line": json.Number("1"), "column": json.Number("1"), "byte": json.Number("1")},
				"end":      map[string]any{"line": json.Number("1"), "column": json.Number("31"), "byte": json.Number("31")},
			},
			want: hcl.Range{
				Filename: "main.tf",
				Start: hcl.Pos{
					Line:   1,
					Column: 1,
					Byte:   1,
				},
				End: hcl.Pos{
					Line:   1,
					Column: 31,
					Byte:   31,
				},
			},
		},
		{
			name:  "invalid type",
			input: "",
			err:   "range is not object, got string",
		},
		{
			name: "invalid position type",
			input: map[string]any{
				"filename": "main.tf",
				"start":    "1,1",
				"end":      "1,31",
			},
			err: "range.start is not object, got string",
		},
		{
			name: "invalid line type",
			input: map[string]any{
				"filename": "main.tf",
				"start":    map[string]any{"line": "", "column": json.Number("1"), "byte": json.Number("1")},
				"end":      map[string]any{"line": json.Number("1"), "column": json.Number("31"), "byte": json.Number("31")},
			},
			err: "range.start.line is not a number, got string",
		},
		{
			name: "invalid column type",
			input: map[string]any{
				"filename": "main.tf",
				"start":    map[string]any{"line": json.Number("1"), "column": "", "byte": json.Number("1")},
				"end":      map[string]any{"line": json.Number("1"), "column": json.Number("31"), "byte": json.Number("31")},
			},
			err: "range.start.column is not a number, got string",
		},
		{
			name: "invalid byte type",
			input: map[string]any{
				"filename": "main.tf",
				"start":    map[string]any{"line": json.Number("1"), "column": json.Number("1"), "byte": ""},
				"end":      map[string]any{"line": json.Number("1"), "column": json.Number("31"), "byte": json.Number("31")},
			},
			err: "range.start.byte is not a number, got string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := jsonToRange(test.input, "range")
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
