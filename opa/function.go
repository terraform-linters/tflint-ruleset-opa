package opa

import (
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/logger"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

// Functions return custom functions as Rego options.
func Functions(runner tflint.Runner) []func(*rego.Rego) {
	return []func(*rego.Rego){
		resourcesFunc(runner),
	}
}

// terraform.resources: resources := terraform.resources(resource_type)
//
// Returns Terraform resources.
//
//	resource_type (string): resource type to retrieve. "*" is a special character that returns all resources.
//
// Returns:
//
//	resources (array[object<type: string, name: string, config: object[string: any], decl_range: rangeTy, type_range: rangeTy>]) Terraform resources
func resourcesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function1(
		&rego.Function{
			Name: "terraform.resources",
			Decl: types.NewFunction(
				types.Args(types.S),
				types.NewArray(
					nil,
					types.NewObject(
						[]*types.StaticProperty{
							types.NewStaticProperty("type", types.S),
							types.NewStaticProperty("name", types.S),
							types.NewStaticProperty("config", types.NewObject(nil, types.NewDynamicProperty(types.S, types.A))),
							types.NewStaticProperty("decl_range", rangeTy),
							types.NewStaticProperty("type_range", rangeTy),
						},
						nil,
					),
				),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, a *ast.Term) (*ast.Term, error) {
			var resourceType string
			if err := ast.As(a.Value, &resourceType); err != nil {
				return nil, err
			}

			var content *hclext.BodyContent
			var err error
			// "*" is a special character that returns all resources
			if resourceType == "*" {
				content, err = runner.GetModuleContent(&hclext.BodySchema{
					Blocks: []hclext.BlockSchema{
						{
							Type:       "resource",
							LabelNames: []string{"type", "name"},
						},
					},
				}, nil)
			} else {
				content, err = runner.GetResourceContent(resourceType, nil, nil)
			}
			if err != nil {
				return nil, err
			}

			v, err := ast.InterfaceToValue(resourcesToJSON(content.Blocks))
			if err != nil {
				return nil, err
			}
			logger.Debug(fmt.Sprintf(`terraform.resoures("%s"): %s`, resourceType, v))

			return ast.NewTerm(v), nil
		},
	)
}

func resourcesToJSON(in hclext.Blocks) []map[string]any {
	ret := make([]map[string]any, len(in))

	for i, block := range in {
		ret[i] = map[string]any{
			"type":       block.Labels[0],
			"name":       block.Labels[1],
			"config":     bodyToJSON(block.Body),
			"decl_range": rangeToJSON(block.DefRange),
			"type_range": rangeToJSON(block.LabelRanges[0]),
		}
	}
	return ret
}

func bodyToJSON(in *hclext.BodyContent) map[string]any {
	ret := map[string]any{}

	// TODO: attributes

	for _, block := range in.Blocks {
		switch r := ret[block.Type].(type) {
		case nil:
			ret[block.Type] = []map[string]any{blockToJSON(block)}
		case []map[string]any:
			ret[block.Type] = append(r, blockToJSON(block))
		default:
			panic(fmt.Sprintf("unknown type: %T", ret[block.Type]))
		}
	}

	return ret
}

func blockToJSON(in *hclext.Block) map[string]any {
	return map[string]any{
		"type":   in.Type,
		"labels": in.Labels,
		"body":   bodyToJSON(in.Body),
	}
}
