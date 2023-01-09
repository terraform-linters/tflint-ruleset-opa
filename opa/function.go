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
		dataSourcesFunc(runner),
		moduleCallsFunc(runner),
		providersFunc(runner),
		settingsFunc(runner),
		variablesFunc(runner),
		outputsFunc(runner),
		localsFunc(runner),
		movedBlocksFunc(runner),
		moduleRangeFunc(runner),
		issueFunc(),
	}
}

type option struct {
	expandMode    tflint.ExpandMode
	expandModeSet bool
}

func (o *option) AsGetModuleContentOptions() *tflint.GetModuleContentOption {
	if o.expandModeSet {
		return &tflint.GetModuleContentOption{ExpandMode: o.expandMode}
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
func resourcesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function3(
		&rego.Function{
			Name: "terraform.resources",
			Decl: types.NewFunction(
				types.Args(types.S, schemaTy, optionsTy),
				types.NewArray(nil, typedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(resourceType, schema, options, "resource", runner)
		},
	)
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
func dataSourcesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function3(
		&rego.Function{
			Name: "terraform.data_sources",
			Decl: types.NewFunction(
				types.Args(types.S, schemaTy, optionsTy),
				types.NewArray(nil, typedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, dataType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(dataType, schema, options, "data", runner)
		},
	)
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
func moduleCallsFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.module_calls",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, namedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "module", runner)
		},
	)
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
func providersFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.providers",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, namedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "provider", runner)
		},
	)
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
func settingsFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.settings",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, blockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "terraform", runner)
		},
	)
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
func variablesFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.variables",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, namedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "variable", runner)
		},
	)
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
func outputsFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.outputs",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, namedBlockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return namedBlockFunc(schema, options, "output", runner)
		},
	)
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
func localsFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function1(
		&rego.Function{
			Name: "terraform.locals",
			Decl: types.NewFunction(
				types.Args(optionsTy),
				types.NewArray(nil, localTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, optionArg *ast.Term) (*ast.Term, error) {
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
	)
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
func movedBlocksFunc(runner tflint.Runner) func(*rego.Rego) {
	return rego.Function2(
		&rego.Function{
			Name: "terraform.moved_blocks",
			Decl: types.NewFunction(
				types.Args(schemaTy, optionsTy),
				types.NewArray(nil, blockTy),
			),
			Memoize:          true,
			Nondeterministic: true,
		},
		func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return blockFunc(schema, options, "moved", runner)
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
		func(_ rego.BuiltinContext, msgArg *ast.Term, rngArg *ast.Term) (*ast.Term, error) {
			var rng map[string]any
			if err := ast.As(rngArg.Value, &rng); err != nil {
				return nil, err
			}
			// type checking only
			if _, err := jsonToRange(rng, "range"); err != nil {
				return nil, err
			}

			return ast.ObjectTerm(
				ast.Item(ast.StringTerm("msg"), msgArg),
				ast.Item(ast.StringTerm("range"), rngArg),
			), nil
		},
	)
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
