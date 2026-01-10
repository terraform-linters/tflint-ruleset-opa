package funcs

import (
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/terraform-linters/tflint-ruleset-opa/opa/tester"
)

func TestResourcesFunc(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		resourceType string
		schema       map[string]any
		options      map[string]string
		want         []map[string]any
	}{
		{
			name: "resource type",
			config: `
resource "aws_instance" "main" {
	instance_type = "t2.micro"
}

resource "aws_s3_bucket" "main" {
	bucket = "foo"
}`,
			resourceType: "aws_instance",
			schema:       map[string]any{"instance_type": "string"},
			want: []map[string]any{
				{
					"type": "aws_instance",
					"name": "main",
					"config": map[string]any{
						"instance_type": map[string]any{
							"value":     "t2.micro",
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 18,
									"byte":   51,
								},
								"end": map[string]int{
									"line":   3,
									"column": 28,
									"byte":   61,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 31,
							"byte":   31,
						},
					},
				},
			},
		},
		{
			name: "wildcard",
			config: `
resource "aws_instance" "main" {
	instance_type = "t2.micro"
}

resource "aws_s3_bucket" "main" {
	bucket = "foo"
}`,
			resourceType: "*",
			schema:       map[string]any{},
			want: []map[string]any{
				{
					"type":   "aws_instance",
					"name":   "main",
					"config": map[string]any{},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 31,
							"byte":   31,
						},
					},
				},
				{
					"type":   "aws_s3_bucket",
					"name":   "main",
					"config": map[string]any{},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   6,
							"column": 1,
							"byte":   65,
						},
						"end": map[string]int{
							"line":   6,
							"column": 32,
							"byte":   96,
						},
					},
				},
			},
		},
		{
			name: "nested block",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = 30
	}
}`,
			resourceType: "aws_instance",
			schema:       map[string]any{"ebs_block_device": map[string]any{"volume_size": "number"}},
			want: []map[string]any{
				{
					"type": "aws_instance",
					"name": "main",
					"config": map[string]any{
						"ebs_block_device": []map[string]any{
							{
								"config": map[string]any{
									"volume_size": map[string]any{
										"value":     30,
										"unknown":   false,
										"sensitive": false,
										"ephemeral": false,
										"range": map[string]any{
											"filename": "main.tf",
											"start": map[string]int{
												"line":   4,
												"column": 17,
												"byte":   70,
											},
											"end": map[string]int{
												"line":   4,
												"column": 19,
												"byte":   72,
											},
										},
									},
								},
								"labels": []string(nil),
								"decl_range": map[string]any{
									"filename": "main.tf",
									"start": map[string]int{
										"line":   3,
										"column": 2,
										"byte":   35,
									},
									"end": map[string]int{
										"line":   3,
										"column": 18,
										"byte":   51,
									},
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 31,
							"byte":   31,
						},
					},
				},
			},
		},
		{
			name: "labeled block",
			config: `
resource "aws_instance" "main" {
	dynamic "ebs_block_device" {}
}`,
			resourceType: "aws_instance",
			schema:       map[string]any{"dynamic": map[string]any{"__labels": []string{"type"}}},
			want: []map[string]any{
				{
					"type": "aws_instance",
					"name": "main",
					"config": map[string]any{
						"dynamic": []map[string]any{
							{
								"config": map[string]any{},
								"labels": []string{"ebs_block_device"},
								"decl_range": map[string]any{
									"filename": "main.tf",
									"start": map[string]int{
										"line":   3,
										"column": 2,
										"byte":   35,
									},
									"end": map[string]int{
										"line":   3,
										"column": 28,
										"byte":   61,
									},
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 31,
							"byte":   31,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceType, err := ast.InterfaceToValue(test.resourceType)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ResourcesFunc(runner).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction3(ResourcesFunc).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestDataSourcesFunc(t *testing.T) {
	tests := []struct {
		name     string
		config   string
		dataType string
		schema   map[string]any
		options  map[string]string
		want     []map[string]any
	}{
		{
			name: "data source",
			config: `
data "aws_ami" "main" {
	owners = ["self"]
}`,
			dataType: "aws_ami",
			schema:   map[string]any{"owners": "list(string)"},
			want: []map[string]any{
				{
					"type": "aws_ami",
					"name": "main",
					"config": map[string]any{
						"owners": map[string]any{
							"value":     []string{"self"},
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 11,
									"byte":   35,
								},
								"end": map[string]int{
									"line":   3,
									"column": 19,
									"byte":   43,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 22,
							"byte":   22,
						},
					},
				},
			},
		},
		{
			name: "scoped data source",
			config: `
check "scoped" {
	data "aws_ami" "main" {
		owners = ["self"]
	}
}`,
			dataType: "aws_ami",
			schema:   map[string]any{"owners": "list(string)"},
			want: []map[string]any{
				{
					"type": "aws_ami",
					"name": "main",
					"config": map[string]any{
						"owners": map[string]any{
							"value":     []string{"self"},
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   4,
									"column": 12,
									"byte":   54,
								},
								"end": map[string]int{
									"line":   4,
									"column": 20,
									"byte":   62,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   3,
							"column": 2,
							"byte":   19,
						},
						"end": map[string]int{
							"line":   3,
							"column": 23,
							"byte":   40,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dataType, err := ast.InterfaceToValue(test.dataType)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := DataSourcesFunc(runner).Impl(ctx, ast.NewTerm(dataType), ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction3(DataSourcesFunc).Impl(ctx, ast.NewTerm(dataType), ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestModuleCallsFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "module",
			config: `
module "aws_instance" {
	instance_type = "t2.micro"
}`,
			schema: map[string]any{"instance_type": "string"},
			want: []map[string]any{
				{
					"name": "aws_instance",
					"config": map[string]any{
						"instance_type": map[string]any{
							"value":     "t2.micro",
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 18,
									"byte":   42,
								},
								"end": map[string]int{
									"line":   3,
									"column": 28,
									"byte":   52,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 22,
							"byte":   22,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ModuleCallsFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(ModuleCallsFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestProvidersFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "provider",
			config: `
provider "aws" {
	region = "us-east-1"
}`,
			schema: map[string]any{"region": "string"},
			want: []map[string]any{
				{
					"name": "aws",
					"config": map[string]any{
						"region": map[string]any{
							"value":     "us-east-1",
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 11,
									"byte":   28,
								},
								"end": map[string]int{
									"line":   3,
									"column": 22,
									"byte":   39,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 15,
							"byte":   15,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ProvidersFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(ProvidersFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestSettingsFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "setting",
			config: `
terraform {
	required_providers {
		aws = {
			source  = "hashicorp/aws"
			version = "~> 4.0"
		}
	}
}`,
			schema: map[string]any{"required_providers": map[string]any{"aws": "map(string)"}},
			want: []map[string]any{
				{
					"config": map[string]any{
						"required_providers": []map[string]any{
							{
								"config": map[string]any{
									"aws": map[string]any{
										"value": map[string]string{
											"source":  "hashicorp/aws",
											"version": "~> 4.0",
										},
										"unknown":   false,
										"sensitive": false,
										"ephemeral": false,
										"range": map[string]any{
											"filename": "main.tf",
											"start": map[string]int{
												"line":   4,
												"column": 9,
												"byte":   43,
											},
											"end": map[string]int{
												"line":   7,
												"column": 4,
												"byte":   99,
											},
										},
									},
								},
								"labels": []string(nil),
								"decl_range": map[string]any{
									"filename": "main.tf",
									"start": map[string]int{
										"line":   3,
										"column": 2,
										"byte":   14,
									},
									"end": map[string]int{
										"line":   3,
										"column": 20,
										"byte":   32,
									},
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 10,
							"byte":   10,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := SettingsFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(SettingsFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestVariablesFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "variable",
			config: `
variable "foo" {
	nullable = true
}`,
			schema: map[string]any{"nullable": "bool"},
			want: []map[string]any{
				{
					"name": "foo",
					"config": map[string]any{
						"nullable": map[string]any{
							"value":     true,
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 13,
									"byte":   30,
								},
								"end": map[string]int{
									"line":   3,
									"column": 17,
									"byte":   34,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 15,
							"byte":   15,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := VariablesFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(VariablesFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestOutputsFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "output",
			config: `
output "bar" {
	description = null
}`,
			schema: map[string]any{"description": "string"},
			want: []map[string]any{
				{
					"name": "bar",
					"config": map[string]any{
						"description": map[string]any{
							"value":     nil,
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 16,
									"byte":   31,
								},
								"end": map[string]int{
									"line":   3,
									"column": 20,
									"byte":   35,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 13,
							"byte":   13,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := OutputsFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(OutputsFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestLocalsFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "locals",
			config: `
locals {
	foo = "bar"
}`,
			want: []map[string]any{
				{
					"name": "foo",
					"expr": map[string]any{
						"value":     "bar",
						"unknown":   false,
						"sensitive": false,
						"ephemeral": false,
						"range": map[string]any{
							"filename": "main.tf",
							"start": map[string]int{
								"line":   3,
								"column": 8,
								"byte":   17,
							},
							"end": map[string]int{
								"line":   3,
								"column": 13,
								"byte":   22,
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   3,
							"column": 2,
							"byte":   11,
						},
						"end": map[string]int{
							"line":   3,
							"column": 13,
							"byte":   22,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := LocalsFunc(runner).Impl(ctx, ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction1(LocalsFunc).Impl(ctx, ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMovedBlocksFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "moved block",
			config: `
moved {
	from = var.foo
}

variable "foo" {}`,
			schema: map[string]any{"from": "any"},
			want: []map[string]any{
				{
					"config": map[string]any{
						"from": map[string]any{
							"unknown":   true,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 9,
									"byte":   17,
								},
								"end": map[string]int{
									"line":   3,
									"column": 16,
									"byte":   24,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 6,
							"byte":   6,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := MovedBlocksFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(MovedBlocksFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestImportsFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "imports",
			config: `
import {
	to = aws_instance.example
	id = "i-abcd1234"
}`,
			schema: map[string]any{"id": "string"},
			want: []map[string]any{
				{
					"config": map[string]any{
						"id": map[string]any{
							"value":     "i-abcd1234",
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   4,
									"column": 7,
									"byte":   43,
								},
								"end": map[string]int{
									"line":   4,
									"column": 19,
									"byte":   55,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 7,
							"byte":   7,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ImportsFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(ImportsFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestChecksFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "checks",
			config: `
check "health_check" {
	data "http" "terraform_io" {
		url = "https://www.terraform.io"
	}

	assert {
		condition = 200 == 200
		error_message = "${data.http.terraform_io.url} returned an unhealthy status code"
	}
}`,
			schema: map[string]any{"assert": map[string]any{"condition": "bool"}},
			want: []map[string]any{
				{
					"name": "health_check",
					"config": map[string]any{
						"assert": []map[string]any{
							{
								"config": map[string]any{
									"condition": map[string]any{
										"value":     true,
										"unknown":   false,
										"sensitive": false,
										"ephemeral": false,
										"range": map[string]any{
											"filename": "main.tf",
											"start": map[string]int{
												"line":   8,
												"column": 15,
												"byte":   117,
											},
											"end": map[string]int{
												"line":   8,
												"column": 25,
												"byte":   127,
											},
										},
									},
								},
								"labels": []string(nil),
								"decl_range": map[string]any{
									"filename": "main.tf",
									"start": map[string]int{
										"line":   7,
										"column": 2,
										"byte":   94,
									},
									"end": map[string]int{
										"line":   7,
										"column": 8,
										"byte":   100,
									},
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 21,
							"byte":   21,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ChecksFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(ChecksFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRemovedBlocksFunc(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		schema  map[string]any
		options map[string]string
		want    []map[string]any
	}{
		{
			name: "removed block",
			config: `
removed {
	from = var.foo
}

variable "foo" {}`,
			schema: map[string]any{"from": "any"},
			want: []map[string]any{
				{
					"config": map[string]any{
						"from": map[string]any{
							"unknown":   true,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 9,
									"byte":   19,
								},
								"end": map[string]int{
									"line":   3,
									"column": 16,
									"byte":   26,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 8,
							"byte":   8,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := RemovedBlocksFunc(runner).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction2(RemovedBlocksFunc).Impl(ctx, ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestEphemeralResourcesFunc(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		resourceType string
		schema       map[string]any
		options      map[string]string
		want         []map[string]any
	}{
		{
			name: "ephemeral resource",
			config: `
ephemeral "aws_secretsmanager_secret_version" "db_password" {
  secret_id = "secret_id"
}`,
			resourceType: "aws_secretsmanager_secret_version",
			schema:       map[string]any{"secret_id": "string"},
			want: []map[string]any{
				{
					"type": "aws_secretsmanager_secret_version",
					"name": "db_password",
					"config": map[string]any{
						"secret_id": map[string]any{
							"value":     "secret_id",
							"unknown":   false,
							"sensitive": false,
							"ephemeral": false,
							"range": map[string]any{
								"filename": "main.tf",
								"start": map[string]int{
									"line":   3,
									"column": 15,
									"byte":   77,
								},
								"end": map[string]int{
									"line":   3,
									"column": 26,
									"byte":   88,
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 60,
							"byte":   60,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceType, err := ast.InterfaceToValue(test.resourceType)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := EphemeralResourcesFunc(runner).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction3(EphemeralResourcesFunc).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestActionsFunc(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		resourceType string
		schema       map[string]any
		options      map[string]string
		want         []map[string]any
	}{
		{
			name: "action",
			config: `
action "aws_lambda_invoke" "example" {
  config {
    function_name = "123456789012:function:my-function:1"
  }
}`,
			resourceType: "aws_lambda_invoke",
			schema:       map[string]any{"config": map[string]any{"function_name": "string"}},
			want: []map[string]any{
				{
					"type": "aws_lambda_invoke",
					"name": "example",
					"config": map[string]any{
						"config": []map[string]any{
							{
								"config": map[string]any{
									"function_name": map[string]any{
										"value":     "123456789012:function:my-function:1",
										"unknown":   false,
										"sensitive": false,
										"ephemeral": false,
										"range": map[string]any{
											"filename": "main.tf",
											"start": map[string]int{
												"line":   4,
												"column": 21,
												"byte":   71,
											},
											"end": map[string]int{
												"line":   4,
												"column": 58,
												"byte":   108,
											},
										},
									},
								},
								"labels": []string(nil),
								"decl_range": map[string]any{
									"filename": "main.tf",
									"start": map[string]int{
										"line":   3,
										"column": 3,
										"byte":   42,
									},
									"end": map[string]int{
										"line":   3,
										"column": 9,
										"byte":   48,
									},
								},
							},
						},
					},
					"decl_range": map[string]any{
						"filename": "main.tf",
						"start": map[string]int{
							"line":   2,
							"column": 1,
							"byte":   1,
						},
						"end": map[string]int{
							"line":   2,
							"column": 37,
							"byte":   37,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resourceType, err := ast.InterfaceToValue(test.resourceType)
			if err != nil {
				t.Fatal(err)
			}
			schema, err := ast.InterfaceToValue(test.schema)
			if err != nil {
				t.Fatal(err)
			}
			options, err := ast.InterfaceToValue(test.options)
			if err != nil {
				t.Fatal(err)
			}
			config, err := ast.InterfaceToValue(map[string]string{"main.tf": test.config})
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(map[string]string{"main.tf": test.config})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ActionsFunc(runner).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}

			ctx = rego.BuiltinContext{}
			got, err = MockFunction3(ActionsFunc).Impl(ctx, ast.NewTerm(resourceType), ast.NewTerm(schema), ast.NewTerm(options), ast.NewTerm(config))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestModuleRangeFunc(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]string
		want   map[string]any
	}{
		{
			name:   "dir",
			config: map[string]string{filepath.Join("dir", "main.tf"): ""},
			want: map[string]any{
				"filename": filepath.Join("dir", "main.tf"),
				"start": map[string]int{
					"line":   1,
					"column": 1,
					"byte":   0,
				},
				"end": map[string]int{
					"line":   1,
					"column": 1,
					"byte":   0,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := tester.NewRunner(test.config)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			ctx := rego.BuiltinContext{}
			got, err := ModuleRangeFunc(runner).Impl(ctx, []*ast.Term{})
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
