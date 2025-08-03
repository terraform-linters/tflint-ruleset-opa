package funcs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/terraform-linters/tflint-ruleset-opa/opa/tester"
	"github.com/zclconf/go-cty/cty"
)

func TestExprListFunc(t *testing.T) {
	parse := func(src string, start hcl.Pos) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", start)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}

	tests := []struct {
		name   string
		expr   hcl.Expression
		want   []map[string]any
		source string
	}{
		{
			name: "static list",
			expr: parse(`[foo, bar]`, hcl.Pos{Line: 1, Column: 8, Byte: 7}),
			want: []map[string]any{
				{
					"value": "foo",
					"range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   1,
							"column": 9,
							"byte":   8,
						},
						"end": map[string]int{
							"line":   1,
							"column": 12,
							"byte":   11,
						},
					},
				},
				{
					"value": "bar",
					"range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   1,
							"column": 14,
							"byte":   13,
						},
						"end": map[string]int{
							"line":   1,
							"column": 17,
							"byte":   16,
						},
					},
				},
			},
			source: "attr = [foo, bar]",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.source})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			exprJSON, err := exprToJSON(test.expr, map[string]cty.Type{"expr": exprCty}, "expr", runner)
			if err != nil {
				t.Fatal(err)
			}
			input, err := ast.InterfaceToValue(exprJSON)
			if err != nil {
				t.Fatal(err)
			}

			ctx := rego.BuiltinContext{}
			got, err := ExprListFunc().Impl(ctx, ast.NewTerm(input))
			if err != nil {
				t.Fatal(err)
			}

			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestExprMapFunc(t *testing.T) {
	parse := func(src string, start hcl.Pos) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", start)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}

	tests := []struct {
		name   string
		expr   hcl.Expression
		want   []map[string]any
		source string
	}{
		{
			name: "static map",
			expr: parse(`{ foo = 1, bar = 2 }`, hcl.Pos{Line: 1, Column: 8, Byte: 7}),
			want: []map[string]any{
				{
					"key": map[string]any{
						"value": "foo",
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 10,
								"byte":   9,
							},
							"end": map[string]int{
								"line":   1,
								"column": 13,
								"byte":   12,
							},
						},
					},
					"value": map[string]any{
						"value": "1",
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 16,
								"byte":   15,
							},
							"end": map[string]int{
								"line":   1,
								"column": 17,
								"byte":   16,
							},
						},
					},
				},
				{
					"key": map[string]any{
						"value": "bar",
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 19,
								"byte":   18,
							},
							"end": map[string]int{
								"line":   1,
								"column": 22,
								"byte":   21,
							},
						},
					},
					"value": map[string]any{
						"value": "2",
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 25,
								"byte":   24,
							},
							"end": map[string]int{
								"line":   1,
								"column": 26,
								"byte":   25,
							},
						},
					},
				},
			},
			source: "attr = { foo = 1, bar = 2 }",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.source})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			exprJSON, err := exprToJSON(test.expr, map[string]cty.Type{"expr": exprCty}, "expr", runner)
			if err != nil {
				t.Fatal(err)
			}
			input, err := ast.InterfaceToValue(exprJSON)
			if err != nil {
				t.Fatal(err)
			}

			ctx := rego.BuiltinContext{}
			got, err := ExprMapFunc().Impl(ctx, ast.NewTerm(input))
			if err != nil {
				t.Fatal(err)
			}

			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestExprCallFunc(t *testing.T) {
	parse := func(src string, start hcl.Pos) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", start)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}

	tests := []struct {
		name   string
		expr   hcl.Expression
		want   map[string]any
		source string
	}{
		{
			name: "static call",
			expr: parse(`foo("bar", "baz")`, hcl.Pos{Line: 1, Column: 8, Byte: 7}),
			want: map[string]any{
				"name": "foo",
				"name_range": map[string]any{
					"filename": "main.tf",
					"start": map[string]int{
						"line":   1,
						"column": 8,
						"byte":   7,
					},
					"end": map[string]int{
						"line":   1,
						"column": 11,
						"byte":   10,
					},
				},
				"arguments": []map[string]any{
					{
						"value": `"bar"`,
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 12,
								"byte":   11,
							},
							"end": map[string]int{
								"line":   1,
								"column": 17,
								"byte":   16,
							},
						},
					},
					{
						"value": `"baz"`,
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   1,
								"column": 19,
								"byte":   18,
							},
							"end": map[string]int{
								"line":   1,
								"column": 24,
								"byte":   23,
							},
						},
					},
				},
				"args_range": map[string]any{
					"filename": "main.tf",
					"start": map[string]int{
						"line":   1,
						"column": 11,
						"byte":   10,
					},
					"end": map[string]int{
						"line":   1,
						"column": 25,
						"byte":   24,
					},
				},
			},
			source: `attr = foo("bar", "baz")`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.source})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			exprJSON, err := exprToJSON(test.expr, map[string]cty.Type{"expr": exprCty}, "expr", runner)
			if err != nil {
				t.Fatal(err)
			}
			input, err := ast.InterfaceToValue(exprJSON)
			if err != nil {
				t.Fatal(err)
			}

			ctx := rego.BuiltinContext{}
			got, err := ExprCallFunc().Impl(ctx, ast.NewTerm(input))
			if err != nil {
				t.Fatal(err)
			}

			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
