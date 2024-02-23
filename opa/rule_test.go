package opa

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/liamg/memoryfs"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func TestNewRule(t *testing.T) {
	tests := []struct {
		name string
		rule *ast.Rule
		want *Rule
	}{
		{
			name: "deny rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "deny_test"}},
			want: &Rule{name: "opa_deny_test", severity: tflint.ERROR},
		},
		{
			name: "violation rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "violation_test"}},
			want: &Rule{name: "opa_violation_test", severity: tflint.ERROR},
		},
		{
			name: "warn rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "warn_test"}},
			want: &Rule{name: "opa_warn_test", severity: tflint.WARNING},
		},
		{
			name: "notice rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "notice_test"}},
			want: &Rule{name: "opa_notice_test", severity: tflint.NOTICE},
		},
		{
			name: "invalid rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "other_rule"}},
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule := NewRule(test.rule, nil)
			if rule == nil {
				if test.want == nil {
					return
				}
				t.Fatal("rule is nil")
			}

			if test.want.name != rule.name {
				t.Fatalf("want: %s, got: %s", test.want.name, rule.name)
			}
			if test.want.severity != rule.severity {
				t.Fatalf("want: %s, got: %s", test.want.severity, rule.severity)
			}
		})
	}
}

func TestCheck_deny_instance_type(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_instance_type contains issue if {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_instance_type"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "main" {
	instance_type = "t1.micro"
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "t2.micro is only allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 18}, End: hcl.Pos{Line: 3, Column: 28}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_non_snake_case(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_not_snake_case contains issue if {
	resources := terraform.resources("*", {}, {})
	not regex.match("^[a-z][a-z0-9]*(_[a-z0-9]+)*$", resources[i].name)

	issue := tflint.issue(sprintf("%s is not snake case", [resources[i].name]), resources[i].decl_range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_not_snake_case"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "foo-bar" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "foo-bar is not snake case",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 34}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "foo_bar" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_warn_standard_volume(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

warn_standard_volume contains issue if {
	resources := terraform.resources("aws_instance", {"ebs_block_device": {"volume_type": "string"}}, {})
	volume_type := resources[_].config.ebs_block_device[_].config.volume_type
	volume_type.value == "standard"

	issue := tflint.issue("standard is not allowed", volume_type.range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "warn_standard_volume"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_type = "standard"
	}
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "standard is not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 17}, End: hcl.Pos{Line: 4, Column: 27}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_type = "gp3"
	}
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_large_volume(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_large_volume contains issue if {
	resources := terraform.resources("aws_instance", {"ebs_block_device": {"volume_size": "number"}}, {})
	volume_size := resources[_].config.ebs_block_device[_].config.volume_size
	volume_size.value > 30

	issue := tflint.issue("Allowed volume size under 30 GB", volume_size.range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_large_volume"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = 50
	}
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "Allowed volume size under 30 GB",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 17}, End: hcl.Pos{Line: 4, Column: 19}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = 30
	}
}`,
			want: helper.Issues{},
		},
		{
			name: "allowed (string)",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = "30"
	}
}`,
			want: helper.Issues{},
		},
		{
			name: "not allowed (float)",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = 30.5
	}
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "Allowed volume size under 30 GB",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 4, Column: 17}, End: hcl.Pos{Line: 4, Column: 21}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_untagged_instance(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

is_not_tagged(config) if {
	not "Environment" in object.keys(config.tags.value)
}
is_not_tagged(config) if {
	not "tags" in object.keys(config)
}

deny_untagged_instance contains issue if {
	resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
	resource := resources[_]

	is_not_tagged(resource.config)

	issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_untagged_instance"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed (no tags)",
			config: `
resource "aws_instance" "main" {
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "instance should be tagged with Environment",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 31}},
				},
			},
		},
		{
			name: "not allowed (no Environment tags)",
			config: `
resource "aws_instance" "main" {
	tags = {
		Production = true
	}
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "instance should be tagged with Environment",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 31}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	tags = {
		Environment = "production"
	}
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_resource(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_resource contains issue if {
	resources := terraform.resources("*", {}, {})
	count(resources) > 0

	issue := tflint.issue("resource is not allowed", resources[0].decl_range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_resource"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "main" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "resource is not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 2, Column: 1}, End: hcl.Pos{Line: 2, Column: 31}},
				},
			},
		},
		{
			name: "allowed",
			config: `
module "secure_instance" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_no_resource(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_no_resource contains issue if {
	resources := terraform.resources("*", {}, {})
	count(resources) == 0

	issue := tflint.issue("resource must be defined", terraform.module_range())
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_no_resource"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "resource must be defined",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 1, Column: 1}, End: hcl.Pos{Line: 1, Column: 1}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	instance_type = "t2.micro"
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_deny_dynamic_block(t *testing.T) {
	fs := memoryfs.New()
	policy := `
package tflint

import rego.v1

deny_dynamic_block contains issue if {
	resources := terraform.resources("*", {"dynamic": {"__labels": ["name"]}}, {"expand_mode": "none"})
	dynamic := resources[_].config.dynamic[_]
  
	issue := tflint.issue("dynamic block is not allowed", dynamic.decl_range)
}`
	fs.WriteFile("main.rego", []byte(policy), 0o644)

	ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
	if err != nil {
		t.Fatal(err)
	}
	engine, err := NewEngine(ret)
	if err != nil {
		t.Fatal(err)
	}
	rule := NewRule(&ast.Rule{Head: &ast.Head{Name: "deny_dynamic_block"}}, engine)

	tests := []struct {
		name   string
		config string
		want   helper.Issues
	}{
		{
			name: "not allowed",
			config: `
resource "aws_instance" "main" {
	dynamic "ebs_block_device" {
		for_each = var.devices

		content {
			volume_size = ebs_block_device.value["size"]
		}
	}
}`,
			want: helper.Issues{
				{
					Rule:    rule,
					Message: "dynamic block is not allowed",
					Range:   hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3, Column: 2}, End: hcl.Pos{Line: 3, Column: 28}},
				},
			},
		},
		{
			name: "allowed",
			config: `
resource "aws_instance" "main" {
	ebs_block_device {
		volume_size = 30
	}
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"main.tf": test.config})

			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}
