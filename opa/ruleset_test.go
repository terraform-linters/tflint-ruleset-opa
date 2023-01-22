package opa

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

func TestApplyConfig(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		config *hclext.BodyContent
		root   string
		env    map[string]string
		want   []string
		err    bool
	}{
		{
			name: "rules exists",
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"policy_dir": &hclext.Attribute{
						Name: "policy_dir",
						Expr: hcl.StaticExpr(cty.StringVal(filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies")), hcl.Range{}),
					},
				},
			},
			want: []string{"opa_deny_not_snake_case", "opa_deny_not_t2_micro"},
		},
		{
			name: "tests exists",
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"policy_dir": &hclext.Attribute{
						Name: "policy_dir",
						Expr: hcl.StaticExpr(cty.StringVal(filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies")), hcl.Range{}),
					},
				},
			},
			env: map[string]string{
				"TFLINT_OPA_TEST": "true",
			},
			want: []string{"opa_test_deny_not_snake_case", "opa_test_not_deny_t2_micro"},
		},
		{
			name: "policy dir not exists, but the dir is default",
			root: filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies"),
			want: []string{},
		},
		{
			name: "policy dir does not exists",
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"policy_dir": &hclext.Attribute{
						Name: "policy_dir",
						Expr: hcl.StaticExpr(cty.StringVal(filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies")), hcl.Range{}),
					},
				},
			},
			err: true,
		},
	}

	original := policyRoot
	policyRoot = filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies")
	defer func() { policyRoot = original }()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.root != "" {
				original := policyRoot
				policyRoot = test.root
				defer func() { policyRoot = original }()
			}
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			ruleset := &RuleSet{config: &Config{}, globalConfig: &tflint.Config{}}
			err := ruleset.ApplyConfig(test.config)
			if err != nil {
				if test.err {
					return
				}
				t.Fatal(err)
			}
			if err == nil && test.err {
				t.Fatal("should return an error, but it does not")
			}

			got := make([]string, len(ruleset.Rules))
			for i, r := range ruleset.Rules {
				got[i] = r.Name()
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
