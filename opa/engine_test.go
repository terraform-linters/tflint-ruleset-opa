package opa

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/hcl/v2"
	"github.com/liamg/memoryfs"
	"github.com/open-policy-agent/opa/loader"
	"github.com/open-policy-agent/opa/version"
)

func TestRunQuery(t *testing.T) {
	tests := []struct {
		name     string
		policies map[string]string
		config   map[string]string
		want     []*Issue
		err      string
	}{
		{
			name: "simple policy",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	issue := tflint.issue("example issue", terraform.module_range())
}`,
			},
			want: []*Issue{{Message: "example issue", Range: hcl.Range{Filename: "main.tf", Start: hcl.InitialPos, End: hcl.InitialPos}}},
		},
		{
			name: "store data",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	issue := tflint.issue(sprintf("data.foo is %s", [data.foo]), terraform.module_range())
}`,
				"data.yaml": `foo: bar`,
			},
			want: []*Issue{{Message: "data.foo is bar", Range: hcl.Range{Filename: "main.tf", Start: hcl.InitialPos, End: hcl.InitialPos}}},
		},
		{
			name: "terraform functions",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}`,
			},
			config: map[string]string{
				"main.tf": `
resource "aws_instance" "main" {
	instance_type = "t1.micro"
}`,
			},
			want: []*Issue{{Message: "t2.micro is only allowed", Range: hcl.Range{Filename: "main.tf", Start: hcl.Pos{Line: 3}, End: hcl.Pos{Line: 3}}}},
		},
		{
			name: "runtime",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	issue := tflint.issue(sprintf("OPA version: %s", [opa.runtime().version]), terraform.module_range())
}`,
			},
			want: []*Issue{{Message: fmt.Sprintf("OPA version: %s", version.Version), Range: hcl.Range{Filename: "main.tf", Start: hcl.InitialPos, End: hcl.InitialPos}}},
		},
		{
			name: "strict mode",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	unused := "foo"
	issue := tflint.issue("example issue", terraform.module_range())
}`,
			},
			err: "1 error occurred: main.rego:5: rego_compile_error: assigned var unused unused",
		},
		{
			name: "builtin errors",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	issue := tflint.issue("example issue", "main.tf:1,1-1")
}`,
			},
			err: "main.rego:5: eval_builtin_error: tflint.issue: json: cannot unmarshal string into Go value of type map[string]interface {}",
		},
		{
			name: "invalid issue",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test {
	"foo" == "foo"
}`,
			},
			err: "issue is not set, got bool",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := memoryfs.New()
			for path, content := range test.policies {
				fs.WriteFile(path, []byte(content), 0o644)
			}

			ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
			if err != nil {
				t.Fatal(err)
			}

			engine, err := NewEngine(ret)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := NewTestRunner(test.config)
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			got, err := engine.RunQuery(&Rule{regoName: "deny_test"}, runner)
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if err == nil && test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			opt := cmpopts.IgnoreFields(hcl.Pos{}, "Column", "Byte")
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestRunTest(t *testing.T) {
	tests := []struct {
		name     string
		policies map[string]string
		want     []*Issue
		err      string
	}{
		{
			name: "simple policy",
			policies: map[string]string{
				"main_test.rego": `
package tflint

test_deny {
	"foo" == "bar"
}`,
			},
			want: []*Issue{{Message: "test failed"}},
		},
		{
			name: "store data",
			policies: map[string]string{
				"main_test.rego": `
package tflint

test_deny {
	"foo" == data.foo
}`,
				"data.yaml": `foo: bar`,
			},
			want: []*Issue{{Message: "test failed"}},
		},
		{
			name: "terraform functions",
			policies: map[string]string{
				"main.rego": `
package tflint

deny_test[issue] {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}`,
				"main_test.rego": `
package tflint

mock_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": ` + "`" + `
resource "aws_instance" "main" {
	instance_type = "t1.micro"
}` + "`" + `})

test_deny {
	count(deny_test) == 0 with terraform.resources as mock_resources
}
				`,
			},
			want: []*Issue{{Message: "test failed"}},
		},
		{
			name: "runtime",
			policies: map[string]string{
				"main_test.rego": `
package tflint

test_deny {
	"foo" == opa.runtime().version
}`,
			},
			want: []*Issue{{Message: "test failed"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := memoryfs.New()
			for path, content := range test.policies {
				fs.WriteFile(path, []byte(content), 0o644)
			}

			ret, err := loader.NewFileLoader().WithFS(fs).Filtered([]string{"."}, nil)
			if err != nil {
				t.Fatal(err)
			}

			engine, err := NewEngine(ret)
			if err != nil {
				t.Fatal(err)
			}

			runner, diags := NewTestRunner(map[string]string{})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			got, err := engine.RunTest(&TestRule{regoName: "test_deny"}, runner)
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if err == nil && test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			opt := cmpopts.IgnoreFields(hcl.Pos{}, "Column", "Byte")
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Error(diff)
			}
		})
	}
}