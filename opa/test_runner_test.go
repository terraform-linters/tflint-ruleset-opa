package opa

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestGetModuleContent(t *testing.T) {
	tests := []struct {
		name   string
		config string
		schema *hclext.BodySchema
		want   *hclext.BodyContent
	}{
		{
			name: "attribute",
			config: `
resource "aws_instance" "foo" {
	ami           = "ami-123456"
	instance_type = "t2.micro"
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Attributes: []hclext.AttributeSchema{{Name: "instance_type"}},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{
								"instance_type": {
									Name: "instance_type",
									Expr: &hclsyntax.TemplateExpr{
										Parts: []hclsyntax.Expression{
											&hclsyntax.LiteralValueExpr{
												SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 19}, End: hcl.Pos{Line: 4, Column: 27}},
											},
										},
										SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 18}, End: hcl.Pos{Line: 4, Column: 28}},
									},
									Range:     hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 28}},
									NameRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 15}},
								},
							},
							Blocks: hclext.Blocks{},
						},
						DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
						LabelRanges: []hcl.Range{
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 25}, End: hcl.Pos{Line: 2, Column: 30}},
						},
					},
				},
			},
		},
		{
			name: "block",
			config: `
resource "aws_instance" "foo" {
	ami = "ami-123456"
	ebs_block_device {
		volume_size = 16
	}
}`,
			schema: &hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "resource",
						LabelNames: []string{"type", "name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{Type: "ebs_block_device", Body: &hclext.BodySchema{Attributes: []hclext.AttributeSchema{{Name: "volume_size"}}}},
							},
						},
					},
				},
			},
			want: &hclext.BodyContent{
				Blocks: hclext.Blocks{
					{
						Type:   "resource",
						Labels: []string{"aws_instance", "foo"},
						Body: &hclext.BodyContent{
							Attributes: hclext.Attributes{},
							Blocks: hclext.Blocks{
								{
									Type: "ebs_block_device",
									Body: &hclext.BodyContent{
										Attributes: hclext.Attributes{
											"volume_size": {
												Name: "volume_size",
												Expr: &hclsyntax.LiteralValueExpr{
													SrcRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 17}, End: hcl.Pos{Line: 5, Column: 19}},
												},
												Range:     hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 3}, End: hcl.Pos{Line: 5, Column: 19}},
												NameRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 5, Column: 3}, End: hcl.Pos{Line: 5, Column: 14}},
											},
										},
										Blocks: hclext.Blocks{},
									},
									DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 18}},
									TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 2}, End: hcl.Pos{Line: 4, Column: 18}},
								},
							},
						},
						DefRange:  hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 30}},
						TypeRange: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 9}},
						LabelRanges: []hcl.Range{
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 10}, End: hcl.Pos{Line: 2, Column: 24}},
							{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 25}, End: hcl.Pos{Line: 2, Column: 30}},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := NewTestRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			got, err := runner.GetModuleContent(test.schema, nil)
			if err != nil {
				t.Fatal(err)
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclsyntax.LiteralValueExpr{}, "Val"),
				cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestGetModuleContent_json(t *testing.T) {
	files := map[string]string{
		"main.tf.json": `{"variable": {"foo": {"type": "string"}}}`,
	}

	runner, diags := NewTestRunner(files)
	if diags.HasErrors() {
		t.Fatal(diags)
	}

	schema := &hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: "variable",
				Body: &hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "type",
							LabelNames: []string{"name"},
							Body:       &hclext.BodySchema{},
						},
					},
				},
			},
		},
	}
	got, err := runner.GetModuleContent(schema, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(got.Blocks) != 1 {
		t.Errorf("got %d blocks, but 1 block is expected", len(got.Blocks))
	}
}

func TestEvaluateExpr(t *testing.T) {
	parse := func(src string) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "main.tf", hcl.InitialPos)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		return expr
	}

	tests := []struct {
		name   string
		config string
		expr   hcl.Expression
		want   string
		err    error
	}{
		{
			name: "literal",
			expr: parse(`"t2.micro"`),
			want: `cty.StringVal("t2.micro")`,
		},
		{
			name: "variable",
			config: `
variable "instance_type" {
	default = "t2.micro"
}`,
			expr: parse("var.instance_type"),
			want: `cty.StringVal("t2.micro")`,
		},
		{
			name:   "variable without default",
			config: `variable "instance_type" {}`,
			expr:   parse("var.instance_type"),
			want:   `cty.DynamicVal`,
		},
		{
			name: "sensitive variable",
			config: `
variable "instance_type" {
	default   = "t2.micro"
	sensitive = true
}`,
			expr: parse("var.instance_type"),
			err:  tflint.ErrSensitive,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner, diags := NewTestRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			var got cty.Value
			err := runner.EvaluateExpr(test.expr, &got, nil)
			if err != nil {
				if !errors.Is(err, test.err) {
					t.Fatal(err)
				}
				return
			}
			if err == nil && test.err != nil {
				t.Fatal("should return an error, but it does not")
			}

			if test.want != got.GoString() {
				t.Fatalf("want: %s, got: %s", test.want, got.GoString())
			}
		})
	}
}
