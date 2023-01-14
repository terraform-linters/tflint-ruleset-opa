package opa

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/terraform/addrs"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// testRunner is a pseudo runner used in policy testing.
// This can be used to inspect Terraform config files written within tests.
// Since it is different from a real gRPC client, some features are limited.
type testRunner struct {
	files     map[string]*hcl.File
	variables map[string]*variable
}

type variable struct {
	Name      string
	Default   cty.Value
	Sensitive bool
	DeclRange hcl.Range
}

var _ tflint.Runner = (*testRunner)(nil)

func NewTestRunner(files map[string]string) (*testRunner, hcl.Diagnostics) {
	runner := &testRunner{
		files:     map[string]*hcl.File{},
		variables: map[string]*variable{},
	}
	parser := hclparse.NewParser()

	for name, src := range files {
		var file *hcl.File
		var diags hcl.Diagnostics
		if strings.HasSuffix(name, ".json") {
			file, diags = parser.ParseJSON([]byte(src), name)
		} else {
			file, diags = parser.ParseHCL([]byte(src), name)
		}
		if diags.HasErrors() {
			return runner, diags
		}

		runner.files[name] = file
	}

	for _, file := range runner.files {
		content, _, diags := file.Body.PartialContent(configFileSchema)
		if diags.HasErrors() {
			return runner, diags
		}

		for _, block := range content.Blocks {
			switch block.Type {
			case "variable":
				// Only "variable" blocks are interpreted
				variable, diags := decodeVariableBlock(block)
				if diags.HasErrors() {
					return runner, diags
				}
				runner.variables[variable.Name] = variable
			default:
				continue
			}
		}
	}

	return runner, nil
}

// GetModuleContent gets a content of the module.
// dynamic blocks, meta-arguments and overrides are not considered
func (r *testRunner) GetModuleContent(schema *hclext.BodySchema, _ *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	content := &hclext.BodyContent{}
	diags := hcl.Diagnostics{}

	for _, f := range r.files {
		c, d := hclext.PartialContent(f.Body, schema)
		diags = diags.Extend(d)
		for name, attr := range c.Attributes {
			content.Attributes[name] = attr
		}
		content.Blocks = append(content.Blocks, c.Blocks...)
	}

	if diags.HasErrors() {
		return nil, diags
	}
	return content, nil
}

var sensitiveMark = cty.NewValueMarks("sensitive")

// EvaluateExpr returns a value of the passed expression.
// Not expected to reflect anything other than cty.Value.
// It is an error to evaluate anything other than a variable.
// Functions are also not supported.
func (r *testRunner) EvaluateExpr(expr hcl.Expression, ret interface{}, _ *tflint.EvaluateExprOption) error {
	variables := map[string]cty.Value{}
	for _, variable := range r.variables {
		val := variable.Default
		if val == cty.NilVal {
			val = cty.DynamicVal
		}
		if variable.Sensitive {
			val = val.WithMarks(sensitiveMark)
		}
		variables[variable.Name] = val
	}

	val, diags := expr.Value(&hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var": cty.ObjectVal(variables),
		},
	})
	if diags.HasErrors() {
		return diags
	}
	if val.IsMarked() {
		return tflint.ErrSensitive
	}

	return gocty.FromCtyValue(val, ret)
}

// GetFile returns the hcl.File object
func (r *testRunner) GetFile(filename string) (*hcl.File, error) {
	return r.files[filename], nil
}

// GetFiles returns all hcl.File
func (r *testRunner) GetFiles() (map[string]*hcl.File, error) {
	return r.files, nil
}

func (r *testRunner) DecodeRuleConfig(name string, ret interface{}) error {
	panic("Not implemented in test runner")
}

func (r *testRunner) EmitIssue(rule tflint.Rule, message string, location hcl.Range) error {
	panic("Not implemented in test runner")
}

func (r *testRunner) EnsureNoError(err error, proc func() error) error {
	panic("Not implemented in test runner")
}

func (r *testRunner) GetModulePath() (addrs.Module, error) {
	panic("Not implemented in test runner")
}

func (r *testRunner) GetOriginalwd() (string, error) {
	panic("Not implemented in test runner")
}

func (r *testRunner) GetProviderContent(name string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	panic("Not implemented in test runner")
}

func (r *testRunner) GetResourceContent(name string, schema *hclext.BodySchema, opts *tflint.GetModuleContentOption) (*hclext.BodyContent, error) {
	panic("Not implemented in test runner")
}

func (r *testRunner) WalkExpressions(walker tflint.ExprWalker) hcl.Diagnostics {
	panic("Not implemented in test runner")
}

func decodeVariableBlock(block *hcl.Block) (*variable, hcl.Diagnostics) {
	v := &variable{
		Name:      block.Labels[0],
		DeclRange: block.DefRange,
	}

	content, _, diags := block.Body.PartialContent(&hcl.BodySchema{
		// Only supports "default" and "sensitive"
		Attributes: []hcl.AttributeSchema{
			{
				Name: "default",
			},
			{
				Name: "sensitive",
			},
		},
	})
	if diags.HasErrors() {
		return v, diags
	}

	if attr, exists := content.Attributes["default"]; exists {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			return v, diags
		}

		v.Default = val
	}

	if attr, exists := content.Attributes["sensitive"]; exists {
		diags := gohcl.DecodeExpression(attr.Expr, nil, &v.Sensitive)
		if diags.HasErrors() {
			return v, diags
		}
	}

	return v, nil
}

var configFileSchema = &hcl.BodySchema{
	Blocks: []hcl.BlockHeaderSchema{
		{
			Type:       "variable",
			LabelNames: []string{"name"},
		},
	},
}
