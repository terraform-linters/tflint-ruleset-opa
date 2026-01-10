package funcs

import (
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

type option struct {
	ExpandMode    tflint.ExpandMode
	ExpandModeSet bool
}

func (o *option) AsGetModuleContentOptions() *tflint.GetModuleContentOption {
	if o.ExpandModeSet {
		return &tflint.GetModuleContentOption{ExpandMode: o.ExpandMode}
	}
	return nil
}

// terraform.resources: resources := terraform.resources(resource_type, schema, options)
//
// Returns Terraform resources.
//
//	resource_type (string)  resource type to retrieve. "*" is a special character that returns all resources.
//	schema        (schema)  schema for attributes referenced in rules.
//	options       (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	resources (array[typed_block]) Terraform "resource" blocks
func ResourcesFunc(runner tflint.Runner) *Function3 {
	return &Function3{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.resources",
				Decl: types.NewFunction(
					types.Args(types.S, schemaTy, optionsTy),
					types.NewArray(nil, typedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(resourceType, schema, options, "resource", runner)
		},
	}
}

// terraform.data_sources: data_sources := terraform.data_sources(data_type, schema, options)
//
// Returns Terraform data sources.
//
//	data_type (string)  data type to retrieve. "*" is a special character that returns all data sources.
//	schema    (schema)  schema for attributes referenced in rules.
//	options   (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	data_sources (array[typed_block]) Terraform "data" blocks
func DataSourcesFunc(runner tflint.Runner) *Function3 {
	return &Function3{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.data_sources",
				Decl: types.NewFunction(
					types.Args(types.S, schemaTy, optionsTy),
					types.NewArray(nil, typedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, dataType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			var typeName string
			if err := ast.As(dataType.Value, &typeName); err != nil {
				return nil, err
			}
			var schemaJSON map[string]any
			if err := ast.As(schema.Value, &schemaJSON); err != nil {
				return nil, err
			}
			innerSchema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
			if err != nil {
				return nil, err
			}
			var optionJSON map[string]string
			if err := ast.As(options.Value, &optionJSON); err != nil {
				return nil, err
			}
			option, err := jsonToOption(optionJSON)
			if err != nil {
				return nil, err
			}

			content, err := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type:       "data",
						LabelNames: []string{"type", "name"},
						Body:       innerSchema,
					},
					{
						Type:       "check",
						LabelNames: []string{"name"},
						Body: &hclext.BodySchema{
							Blocks: []hclext.BlockSchema{
								{
									Type:       "data",
									LabelNames: []string{"type", "name"},
									Body:       innerSchema,
								},
							},
						},
					},
				},
			}, option.AsGetModuleContentOptions())
			if err != nil {
				return nil, err
			}

			blocks := []*hclext.Block{}
			for _, block := range content.Blocks {
				switch block.Type {
				case "data":
					// "*" is a special character that returns all blocks
					if typeName == block.Labels[0] || typeName == "*" {
						blocks = append(blocks, block)
					}
				case "check":
					for _, inner := range block.Body.Blocks {
						if typeName == inner.Labels[0] || typeName == "*" {
							blocks = append(blocks, inner)
						}
					}
				}
			}

			out, err := typedBlocksToJSON(blocks, tyMap, "schema", runner)
			if err != nil {
				return nil, err
			}
			v, err := ast.InterfaceToValue(out)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	}
}

// terraform.module_calls: modules := terraform.module_calls(schema, options)
//
// Returns Terraform module calls.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	modules (array[named_block]) Terraform "module" blocks
func ModuleCallsFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.module_calls",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, namedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "module", runner)
		},
	}
}

// terraform.providers: providers := terraform.providers(schema, options)
//
// Returns Terraform providers.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	providers (array[named_block]) Terraform "provider" blocks
func ProvidersFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.providers",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, namedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "provider", runner)
		},
	}
}

// terraform.settings: settings := terraform.settings(schema, options)
//
// Returns Terraform settings.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	settings (array[block]) Terraform "terraform" blocks
func SettingsFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.settings",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, blockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "terraform", runner)
		},
	}
}

// terraform.variables: variables := terraform.variables(schema, options)
//
// Returns Terraform variables.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	variables (array[named_block]) Terraform "variable" blocks
func VariablesFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.variables",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, namedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "variable", runner)
		},
	}
}

// terraform.outputs: outputs := terraform.outputs(schema, options)
//
// Returns Terraform outputs.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	outputs (array[named_block]) Terraform "output" blocks
func OutputsFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.outputs",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, namedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "output", runner)
		},
	}
}

// terraform.locals: locals := terraform.locals(options)
//
// Returns Terraform local values.
//
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	locals (array[local]) Terraform local values
func LocalsFunc(runner tflint.Runner) *Function1 {
	return &Function1{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.locals",
				Decl: types.NewFunction(
					types.Args(optionsTy),
					types.NewArray(nil, localTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, optionArg *ast.Term) (*ast.Term, error) {
			var optionJSON map[string]string
			if err := ast.As(optionArg.Value, &optionJSON); err != nil {
				return nil, err
			}
			option, err := jsonToOption(optionJSON)
			if err != nil {
				return nil, err
			}

			content, err := runner.GetModuleContent(&hclext.BodySchema{
				Blocks: []hclext.BlockSchema{
					{
						Type: "locals",
						Body: &hclext.BodySchema{Mode: hclext.SchemaJustAttributesMode},
					},
				},
			}, option.AsGetModuleContentOptions())
			if err != nil {
				return nil, err
			}

			locals := []map[string]any{}
			for _, block := range content.Blocks {
				out, err := localsToJSON(block.Body.Attributes, runner)
				if err != nil {
					return nil, err
				}
				locals = append(locals, out...)
			}

			v, err := ast.InterfaceToValue(locals)
			if err != nil {
				return nil, err
			}

			return ast.NewTerm(v), nil
		},
	}
}

// terraform.moved_blocks: blocks := terraform.moved_blocks(schema, options)
//
// Returns Terraform moved blocks.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	blocks (array[block]) Terraform "moved" blocks
func MovedBlocksFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.moved_blocks",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, blockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "moved", runner)
		},
	}
}

// terraform.imports: blocks := terraform.imports(schema, options)
//
// Returns Terraform import blocks.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	blocks (array[block]) Terraform "import" blocks
func ImportsFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.imports",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, blockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "import", runner)
		},
	}
}

// terraform.checks: blocks := terraform.checks(schema, options)
//
// Returns Terraform check blocks.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	blocks (array[block]) Terraform "check" blocks
func ChecksFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.checks",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, blockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "check", runner)
		},
	}
}

// terraform.removed_blocks: blocks := terraform.removed_blocks(schema, options)
//
// Returns Terraform removed blocks.
//
//	schema  (schema)  schema for attributes referenced in rules.
//	options (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	blocks (array[block]) Terraform "removed" blocks
func RemovedBlocksFunc(runner tflint.Runner) *Function2 {
	return &Function2{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.removed_blocks",
				Decl: types.NewFunction(
					types.Args(schemaTy, optionsTy),
					types.NewArray(nil, blockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "removed", runner)
		},
	}
}

// terraform.ephemeral_resources: resources := terraform.ephemeral_resources(resource_type, schema, options)
//
// Returns Terraform ephemeral resources.
//
//	resource_type (string)  resource type to retrieve. "*" is a special character that returns all resources.
//	schema        (schema)  schema for attributes referenced in rules.
//	options       (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	resources (array[typed_block]) Terraform "ephemeral" blocks
func EphemeralResourcesFunc(runner tflint.Runner) *Function3 {
	return &Function3{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.ephemeral_resources",
				Decl: types.NewFunction(
					types.Args(types.S, schemaTy, optionsTy),
					types.NewArray(nil, typedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(resourceType, schema, options, "ephemeral", runner)
		},
	}
}

// terraform.actions: actions := terraform.actions(action_type, schema, options)
//
// Returns Terraform action blocks.
//
//	action_type (string)  action type to retrieve. "*" is a special character that returns all actions.
//	schema      (schema)  schema for attributes referenced in rules.
//	options     (options) options to change the retrieve/evaluate behavior.
//
// Returns:
//
//	actions (array[typed_block]) Terraform "action" blocks
func ActionsFunc(runner tflint.Runner) *Function3 {
	return &Function3{
		Function: Function{
			Decl: &rego.Function{
				Name: "terraform.actions",
				Decl: types.NewFunction(
					types.Args(types.S, schemaTy, optionsTy),
					types.NewArray(nil, typedBlockTy),
				),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(resourceType, schema, options, "action", runner)
		},
	}
}

// terraform.module_range: range := terraform.module_range()
//
// Returns a range for the current Terraform module.
// This is useful in rules that check for non-existence.
//
// Returns:
//
//	range (range) a range for <dir>/main.tf:1:1
func ModuleRangeFunc(runner tflint.Runner) *FunctionDyn {
	return &FunctionDyn{
		Function: Function{
			Decl: &rego.Function{
				Name:             "terraform.module_range",
				Decl:             types.NewFunction(types.Args(), rangeTy),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Impl: func(_ rego.BuiltinContext, _ []*ast.Term) (*ast.Term, error) {
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
	}
}

func typedBlockFunc(typeArg *ast.Term, schemaArg *ast.Term, optionArg *ast.Term, blockType string, runner tflint.Runner) (*ast.Term, error) {
	var typeName string
	if err := ast.As(typeArg.Value, &typeName); err != nil {
		return nil, err
	}
	var schemaJSON map[string]any
	if err := ast.As(schemaArg.Value, &schemaJSON); err != nil {
		return nil, err
	}
	schema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
	if err != nil {
		return nil, err
	}
	var optionJSON map[string]string
	if err := ast.As(optionArg.Value, &optionJSON); err != nil {
		return nil, err
	}
	option, err := jsonToOption(optionJSON)
	if err != nil {
		return nil, err
	}

	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       blockType,
				LabelNames: []string{"type", "name"},
				Body:       schema,
			},
		},
	}, option.AsGetModuleContentOptions())
	if err != nil {
		return nil, err
	}

	blocks := []*hclext.Block{}
	for _, block := range content.Blocks {
		// "*" is a special character that returns all blocks
		if typeName == block.Labels[0] || typeName == "*" {
			blocks = append(blocks, block)
		}
	}

	out, err := typedBlocksToJSON(blocks, tyMap, "schema", runner)
	if err != nil {
		return nil, err
	}
	v, err := ast.InterfaceToValue(out)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(v), nil
}

func namedBlockFunc(schemaArg *ast.Term, optionArg *ast.Term, blockType string, runner tflint.Runner) (*ast.Term, error) {
	var schemaJSON map[string]any
	if err := ast.As(schemaArg.Value, &schemaJSON); err != nil {
		return nil, err
	}
	schema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
	if err != nil {
		return nil, err
	}
	var optionJSON map[string]string
	if err := ast.As(optionArg.Value, &optionJSON); err != nil {
		return nil, err
	}
	option, err := jsonToOption(optionJSON)
	if err != nil {
		return nil, err
	}

	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type:       blockType,
				LabelNames: []string{"name"},
				Body:       schema,
			},
		},
	}, option.AsGetModuleContentOptions())
	if err != nil {
		return nil, err
	}

	out, err := namedBlocksToJSON(content.Blocks, tyMap, "schema", runner)
	if err != nil {
		return nil, err
	}
	v, err := ast.InterfaceToValue(out)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(v), nil
}

func blockFunc(schemaArg *ast.Term, optionArg *ast.Term, blockType string, runner tflint.Runner) (*ast.Term, error) {
	var schemaJSON map[string]any
	if err := ast.As(schemaArg.Value, &schemaJSON); err != nil {
		return nil, err
	}
	schema, tyMap, err := jsonToSchema(schemaJSON, map[string]cty.Type{}, "schema")
	if err != nil {
		return nil, err
	}
	var optionJSON map[string]string
	if err := ast.As(optionArg.Value, &optionJSON); err != nil {
		return nil, err
	}
	option, err := jsonToOption(optionJSON)
	if err != nil {
		return nil, err
	}

	content, err := runner.GetModuleContent(&hclext.BodySchema{
		Blocks: []hclext.BlockSchema{
			{
				Type: blockType,
				Body: schema,
			},
		},
	}, option.AsGetModuleContentOptions())
	if err != nil {
		return nil, err
	}

	out, err := blocksToJSON(content.Blocks, tyMap, "schema", runner)
	if err != nil {
		return nil, err
	}
	v, err := ast.InterfaceToValue(out)
	if err != nil {
		return nil, err
	}

	return ast.NewTerm(v), nil
}
