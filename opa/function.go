package opa

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// Functions return custom functions as Rego options.
func Functions(runner tflint.Runner) []func(*rego.Rego) {
	return []func(*rego.Rego){
		resourcesFunc(runner),
		moduleRangeFunc(runner),
		issueFunc(),
	}
}

// terraform.resources: resources := terraform.resources(resource_type, schema)
//
// Returns Terraform resources.
//
//	resource_type (string) resource type to retrieve. "*" is a special character that returns all resources.
//	schema        (schema) Schema for attributes referenced in rules.
//
// Returns:
//
//	resources (array[resource]) Terraform resources
func resourcesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.resources",
			Decl: types.NewFunction(
				types.Args(types.S, schemaTy),
				types.NewArray(nil, resourceTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, a *ast.Term, b *ast.Term) (*ast.Term, error) {
			var resourceType string
			if err := ast.As(a.Value, &resourceType); err != nil {
				return nil, err
			}
			var schemaJSON map[string]any
			if err := ast.As(b.Value, &schemaJSON); err != nil {
				return nil, err
			}
			schema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
			if err != nil {
				return nil, err
			}

			var content *hclext.BodyContent
			// "*" is a special character that returns all resources
			if resourceType == "*" {
				content, err = runner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
							Body:       schema,
						},
					},
				}, nil)
			} else {
				content, err = runner.GetResourceContent(resourceType, schema, nil)
			}
			if err != nil {
				return nil, err
			}

			resources, err := resourcesToJSON(content.Blocks, tyMap, "schema", runner)
			if err != nil {
				return nil, err
			}
			v, err := ast.InterfaceToValue(resources)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	)
}

// terraform.module_range: range := terraform.module_range()
//
// Returns a range for the current Terraform module.
// This is useful in rules that check for non-existence.
//
// Returns:
//
//	range (range) a range for <dir>/main.tf:1:1
func moduleRangeFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.FunctionDyn(
		&rego.Function{
			Name:             "terraform.module_range",
			Decl:             types.NewFunction(types.Args(), rangeTy),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, _ []*ast.Term) (*ast.Term, error) {
			files, err := runner.GetFiles()
			if err != nil {
				return nil, err
			}

			// If there is no file, the current directory is assumed.
			var dir string
			for path := range files {
				dir = filepath.Dir(path)
				break
			}

			rng := hcl.Range{
				Filename: filepath.Join(dir, "main.tf"),
				Start:    hcl.InitialPos,
				End:      hcl.InitialPos,
			}
			v, err := ast.InterfaceToValue(rangeToJSON(rng))
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	)
}

// tflint.issue: issue := tflint.issue(msg, range)
//
// Returns issue object
//
//	msg   (string) message
//	range (range)  source range
//
// Returns:
//
//	issue (issue) issue object
func issueFunc() func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "tflint.issue",
			// FIXME: types.A should be range, but panic by an OPA bug.
			// @see https://github.com/open-policy-agent/opa/blob/v0.47.4/ast/check.go#L829
			Decl:    types.NewFunction(types.Args(types.S, types.A), issueTy),
			Memoize: true,
		},
		func(_ rego.BuiltinContext, a *ast.Term, b *ast.Term) (*ast.Term, error) {
			var rng map[string]any
			if err := ast.As(b.Value, &rng); err != nil {
				return nil, err
			}
			// type checking only
			if _, err := jsonToRange(rng, "range"); err != nil {
				return nil, err
			}

			return ast.ObjectTerm(
				ast.Item(ast.StringTerm("msg"), a),
				ast.Item(ast.StringTerm("range"), b),
			), nil
		},
	)
}
