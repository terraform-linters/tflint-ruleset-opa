package opa

import (
	"testing"

	"github.com/liamg/memoryfs"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/loader"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func TestNewTestRule(t *testing.T) {
	tests := []struct {
		name string
		rule *ast.Rule
		want *TestRule
	}{
		{
			name: "test rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "test_deny"}},
			want: &TestRule{name: "opa_test_deny"},
		},
		{
			name: "non-test rule",
			rule: &ast.Rule{Head: &ast.Head{Name: "deny_test"}},
			want: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rule := NewTestRule(test.rule, nil)
			if rule == nil {
				if test.want == nil {
					return
				}
				t.Fatal("rule is nil")
			}

			if test.want.name != rule.name {
				t.Fatalf("want: %s, got: %s", test.want.name, rule.name)
			}
		})
	}
}

func TestCheck_test_not_deny_t2_micro(t *testing.T) {
	fs := memoryfs.New()
	test := `
package tflint
import future.keywords

mock_resources_t1_micro(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": ` + "`" + `
resource "aws_instance" "main" {
	instance_type = "t1.micro"
}` + "`" + `})

test_not_deny_t2_micro if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_t1_micro

	count(issues) == 1
	issue := issues[_]
	issue.msg == "t2.micro is only allowed"
}`
	fs.WriteFile("main_test.rego", []byte(test), 0o644)

	tests := []struct {
		name   string
		policy string
		want   helper.Issues
	}{
		{
			name: "test failed",
			policy: `
package tflint

deny_not_t2_micro[issue] {
	resources := terraform.resources("aws_db_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}`,
			want: helper.Issues{
				{
					Rule:    &TestRule{},
					Message: "test failed",
				},
			},
		},
		{
			name: "test passed",
			policy: `
package tflint

deny_not_t2_micro[issue] {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs.WriteFile("main.rego", []byte(test.policy), 0o644)

			ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := NewEngine(ret)
			if err != nil {
				t.Fatal(err)
			}
			rule := NewTestRule(&ast.Rule{Head: &ast.Head{Name: "test_not_deny_t2_micro"}}, engine)

			runner := helper.TestRunner(t, map[string]string{})
			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}

func TestCheck_test_deny_not_snake_case(t *testing.T) {
	fs := memoryfs.New()
	test := `
package tflint
import future.keywords

mock_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": ` + "`" + `
resource "aws_instance" "main-v2" {}
` + "`" + `})

test_deny_not_snake_case if {
	issues := deny_not_snake_case
		with terraform.resources as mock_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "main-v2 is not snake case"
	issue.range.start.line == 2
}`
	fs.WriteFile("main_test.rego", []byte(test), 0o644)

	tests := []struct {
		name   string
		policy string
		want   helper.Issues
	}{
		{
			name: "test failed",
			policy: `
package tflint

deny_not_snake_case[issue] {
	resources := terraform.resources("*", {}, {})
	regex.match("^[a-z][a-z0-9]*(_[a-z0-9]+)*$", resources[i].name)

	issue := tflint.issue(sprintf("%s is not snake case", [resources[i].name]), resources[i].decl_range)
}`,
			want: helper.Issues{
				{
					Rule:    &TestRule{},
					Message: "test failed",
				},
			},
		},
		{
			name: "test passed",
			policy: `
package tflint

deny_not_snake_case[issue] {
	resources := terraform.resources("*", {}, {})
	not regex.match("^[a-z][a-z0-9]*(_[a-z0-9]+)*$", resources[i].name)

	issue := tflint.issue(sprintf("%s is not snake case", [resources[i].name]), resources[i].decl_range)
}`,
			want: helper.Issues{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs.WriteFile("main.rego", []byte(test.policy), 0o644)

			ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := NewEngine(ret)
			if err != nil {
				t.Fatal(err)
			}
			rule := NewTestRule(&ast.Rule{Head: &ast.Head{Name: "test_deny_not_snake_case"}}, engine)

			runner := helper.TestRunner(t, map[string]string{})
			if err := rule.Check(runner); err != nil {
				t.Fatal(err)
			}

			helper.AssertIssues(t, test.want, runner.Issues)
		})
	}
}
