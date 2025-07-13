package opa

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/open-policy-agent/opa/v1/types"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/zclconf/go-cty/cty"
)

// Functions return custom functions as Rego options.
func Functions(runner tflint.Runner) []func(*rego.Rego) {
	return []func(*rego.Rego){
		resourcesFunc(runner).asOption(),
		dataSourcesFunc(runner).asOption(),
		moduleCallsFunc(runner).asOption(),
		providersFunc(runner).asOption(),
		settingsFunc(runner).asOption(),
		variablesFunc(runner).asOption(),
		outputsFunc(runner).asOption(),
		localsFunc(runner).asOption(),
		movedBlocksFunc(runner).asOption(),
		importsFunc(runner).asOption(),
		checksFunc(runner).asOption(),
		removedBlocksFunc(runner).asOption(),
		ephemeralResourcesFunc(runner).asOption(),
		moduleRangeFunc(runner).asOption(),
		exprListFunc().asOption(),
		exprMapFunc().asOption(),
		exprCallFunc().asOption(),
		issueFunc().asOption(),
	}
}

// TesterFunctions return custom functions as tester.Builtin.
func TesterFunctions(runner tflint.Runner) []*tester.Builtin {
	return []*tester.Builtin{
		resourcesFunc(runner).asTester(),
		dataSourcesFunc(runner).asTester(),
		moduleCallsFunc(runner).asTester(),
		providersFunc(runner).asTester(),
		settingsFunc(runner).asTester(),
		variablesFunc(runner).asTester(),
		outputsFunc(runner).asTester(),
		localsFunc(runner).asTester(),
		movedBlocksFunc(runner).asTester(),
		importsFunc(runner).asTester(),
		checksFunc(runner).asTester(),
		removedBlocksFunc(runner).asTester(),
		ephemeralResourcesFunc(runner).asTester(),
		moduleRangeFunc(runner).asTester(),
		exprListFunc().asTester(),
		exprMapFunc().asTester(),
		exprCallFunc().asTester(),
		issueFunc().asTester(),
	}
}

// MockFunctions return mocks for custom functions as Rego options.
// Mock functions are usually not needed outside of testing,
// but are provided for compilation.
func MockFunctions() []func(*rego.Rego) {
	return []func(*rego.Rego){
		mockFunction3(resourcesFunc).asOption(),
		mockFunction3(dataSourcesFunc).asOption(),
		mockFunction2(moduleCallsFunc).asOption(),
		mockFunction2(providersFunc).asOption(),
		mockFunction2(settingsFunc).asOption(),
		mockFunction2(variablesFunc).asOption(),
		mockFunction2(outputsFunc).asOption(),
		mockFunction1(localsFunc).asOption(),
		mockFunction2(movedBlocksFunc).asOption(),
		mockFunction2(importsFunc).asOption(),
		mockFunction2(checksFunc).asOption(),
		mockFunction2(removedBlocksFunc).asOption(),
		mockFunction3(ephemeralResourcesFunc).asOption(),
	}
}

// TesterMockFunctions return mocks for custom functions.
func TesterMockFunctions() []*tester.Builtin {
	return []*tester.Builtin{
		mockFunction3(resourcesFunc).asTester(),
		mockFunction3(dataSourcesFunc).asTester(),
		mockFunction2(moduleCallsFunc).asTester(),
		mockFunction2(providersFunc).asTester(),
		mockFunction2(settingsFunc).asTester(),
		mockFunction2(variablesFunc).asTester(),
		mockFunction2(outputsFunc).asTester(),
		mockFunction1(localsFunc).asTester(),
		mockFunction2(movedBlocksFunc).asTester(),
		mockFunction2(importsFunc).asTester(),
		mockFunction2(checksFunc).asTester(),
		mockFunction2(removedBlocksFunc).asTester(),
		mockFunction3(ephemeralResourcesFunc).asTester(),
	}
}

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
func resourcesFunc(runner tflint.Runner) *function3 {
	return &function3{
		function: function{
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
		Func: func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func dataSourcesFunc(runner tflint.Runner) *function3 {
	return &function3{
		function: function{
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
		Func: func(_ rego.BuiltinContext, dataType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func moduleCallsFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func providersFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func settingsFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func variablesFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func outputsFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func localsFunc(runner tflint.Runner) *function1 {
	return &function1{
		function: function{
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
		Func: func(_ rego.BuiltinContext, optionArg *ast.Term) (*ast.Term, error) {
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
func movedBlocksFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func importsFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func checksFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func removedBlocksFunc(runner tflint.Runner) *function2 {
	return &function2{
		function: function{
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
		Func: func(_ rego.BuiltinContext, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
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
func ephemeralResourcesFunc(runner tflint.Runner) *function3 {
	return &function3{
		function: function{
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
		Func: func(_ rego.BuiltinContext, resourceType *ast.Term, schema *ast.Term, options *ast.Term) (*ast.Term, error) {
			return typedBlockFunc(resourceType, schema, options, "ephemeral", runner)
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
func moduleRangeFunc(runner tflint.Runner) *functionDyn {
	return &functionDyn{
		function: function{
			Decl: &rego.Function{
				Name:             "terraform.module_range",
				Decl:             types.NewFunction(types.Args(), rangeTy),
				Memoize:          true,
				Nondeterministic: true,
			},
		},
		Func: func(_ rego.BuiltinContext, _ []*ast.Term) (*ast.Term, error) {
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

// hcl.expr_list: exprs := hcl.expr_list(expr)
//
// Extract a list of expressions from the given static list expression.
// This is equivalent to hcl.ExprList in hashicorp/hcl.
//
//	expr (raw_expr) static list expression which is retrieved as an expr type.
//
// Returns:
//
//	exprs (array[raw_expr]) expressions as elements of the list.
func exprListFunc() *function1 {
	return &function1{
		function: function{
			Decl: &rego.Function{
				Name:    "hcl.expr_list",
				Decl:    types.NewFunction(types.Args(rawExprTy), types.NewArray(nil, rawExprTy)),
				Memoize: true,
			},
		},
		Func: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			exprs, diags := hcl.ExprList(expr)
			if diags.HasErrors() {
				return nil, diags
			}
			ret := make([]map[string]any, len(exprs))
			for i, e := range exprs {
				ret[i] = rawExprToJSON(e, []byte(src))
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
}

// key_value (object<key: raw_expr, value: raw_expr>) representation of a key value pair
var keyValueTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("key", rawExprTy),
		types.NewStaticProperty("value", rawExprTy),
	},
	nil,
)

// hcl.expr_map: pairs := hcl.expr_map(expr)
//
// Extract a list of key value pairs from the given static map expression.
// This is equivalent to hcl.ExprMap in hashicorp/hcl.
//
//	expr (raw_expr) static map expression which is retrieved as an expr type.
//
// Returns:
//
//	pairs (array[key_value]) key value pairs of the map as expressions.
func exprMapFunc() *function1 {
	return &function1{
		function: function{
			Decl: &rego.Function{
				Name:    "hcl.expr_map",
				Decl:    types.NewFunction(types.Args(rawExprTy), types.NewArray(nil, keyValueTy)),
				Memoize: true,
			},
		},
		Func: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			pairs, diags := hcl.ExprMap(expr)
			if diags.HasErrors() {
				return nil, diags
			}
			ret := make([]map[string]map[string]any, len(pairs))
			for i, pair := range pairs {
				ret[i] = map[string]map[string]any{
					"key":   rawExprToJSON(pair.Key, []byte(src)),
					"value": rawExprToJSON(pair.Value, []byte(src)),
				}
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
}

// call (object<name: string, name_range: range, arguments: array[raw_expr], args_range: range>) representation of a static function call
var callTy = types.NewObject(
	[]*types.StaticProperty{
		types.NewStaticProperty("name", types.S),
		types.NewStaticProperty("name_range", rangeTy),
		types.NewStaticProperty("arguments", types.NewArray(nil, rawExprTy)),
		types.NewStaticProperty("args_range", rangeTy),
	},
	nil,
)

// hcl.expr_call: call := hcl.expr_call(expr)
//
// Extract the function name and arguments from the given function call expression.
// This is equivalent to hcl.ExprCall in hashicorp/hcl.
//
//	expr (raw_expr) function call expression which is retrieved as an expr type.
//
// Returns:
//
//	call (call) function call object
func exprCallFunc() *function1 {
	return &function1{
		function: function{
			Decl: &rego.Function{
				Name:    "hcl.expr_call",
				Decl:    types.NewFunction(types.Args(rawExprTy), callTy),
				Memoize: true,
			},
		},
		Func: func(_ rego.BuiltinContext, exprArg *ast.Term) (*ast.Term, error) {
			expr, src, err := astAsExpr(exprArg)
			if err != nil {
				return nil, err
			}
			src = prependZeroPadding(src, expr.Range().Start)

			call, diags := hcl.ExprCall(expr)
			if diags.HasErrors() {
				return nil, diags
			}

			arguments := make([]map[string]any, len(call.Arguments))
			for i, arg := range call.Arguments {
				arguments[i] = rawExprToJSON(arg, []byte(src))
			}
			ret := map[string]any{
				"name":       call.Name,
				"name_range": rangeToJSON(call.NameRange),
				"arguments":  arguments,
				"args_range": rangeToJSON(call.ArgsRange),
			}

			v, err := ast.InterfaceToValue(ret)
			if err != nil {
				return nil, err
			}
			return ast.NewTerm(v), nil
		},
	}
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
func issueFunc() *function2 {
	return &function2{
		function: function{
			Decl: &rego.Function{
				Name:    "tflint.issue",
				Decl:    types.NewFunction(types.Args(types.S, rangeTy), issueTy),
				Memoize: true,
			},
		},
		Func: func(_ rego.BuiltinContext, msgArg *ast.Term, rngArg *ast.Term) (*ast.Term, error) {
			return ast.ObjectTerm(
				ast.Item(ast.StringTerm("msg"), msgArg),
				ast.Item(ast.StringTerm("range"), rngArg),
			), nil
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

func astAsExpr(v *ast.Term) (hcl.Expression, string, error) {
	var exprMap map[string]any
	if err := ast.As(v.Value, &exprMap); err != nil {
		return nil, "", err
	}

	value, ok := exprMap["value"]
	if !ok {
		return nil, "", fmt.Errorf("expr must have a 'value' key")
	}
	valueStr, ok := value.(string)
	if !ok {
		return nil, "", fmt.Errorf("expr value must be a string")
	}

	rngArg, ok := exprMap["range"]
	if !ok {
		return nil, "", fmt.Errorf("expr must have a 'range' key")
	}
	rng, err := jsonToRange(rngArg, "expr.range")
	if err != nil {
		return nil, "", err
	}

	expr, diags := hclext.ParseExpression([]byte(valueStr), rng.Filename, rng.Start)
	if diags.HasErrors() {
		return nil, "", diags
	}
	return expr, valueStr, nil
}

// prependZeroPadding to the source code so that the position of the expression
// is correct. This is necessary because the position is relative to the start of
// the source code.
func prependZeroPadding(src string, pos hcl.Pos) string {
	return strings.Repeat("0", pos.Byte) + src
}

type function struct {
	Decl *rego.Function
}

func (f *function) asTester(impl func(*rego.Rego)) *tester.Builtin {
	return &tester.Builtin{
		Decl: &ast.Builtin{
			Name:             f.Decl.Name,
			Decl:             f.Decl.Decl,
			Nondeterministic: f.Decl.Nondeterministic,
		},
		Func: impl,
	}
}

func (f *function) mockDecl() function {
	split := strings.Split(f.Decl.Name, ".")
	if len(split) != 2 {
		panic(`function should be named with "<namespace>.<name>"`)
	}
	namespace := split[0]
	name := split[1]

	// Mock function takes test inputs as its last argument.
	// e.g. terraform.mock_resources(resourceType, schema, options, `{"main.tf": "foo = 1"}`)
	args := f.Decl.Decl.FuncArgs().Args
	args = append(args, types.NewObject(nil, types.NewDynamicProperty(types.S, types.S)))

	return function{
		Decl: &rego.Function{
			// Mock function names are prefixed with "mock_"
			// e.g. terraform.resources -> terraform.mock_resources
			Name: fmt.Sprintf("%s.mock_%s", namespace, name),
			Decl: types.NewFunction(
				types.Args(args...),
				f.Decl.Decl.Result(),
			),
			Memoize:          f.Decl.Memoize,
			Nondeterministic: f.Decl.Nondeterministic,
		},
	}
}

type function1 struct {
	function

	Func rego.Builtin1
}

func (f *function1) asOption() func(*rego.Rego) {
	return rego.Function1(f.Decl, f.Func)
}

func (f *function1) asTester() *tester.Builtin {
	return f.function.asTester(f.asOption())
}

func mockFunction1(base func(tflint.Runner) *function1) *function2 {
	return &function2{
		function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := NewTestRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a)
		},
	}
}

type function2 struct {
	function

	Func rego.Builtin2
}

func (f *function2) asOption() func(*rego.Rego) {
	return rego.Function2(f.Decl, f.Func)
}

func (f *function2) asTester() *tester.Builtin {
	return f.function.asTester(f.asOption())
}

func mockFunction2(base func(tflint.Runner) *function2) *function3 {
	return &function3{
		function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := NewTestRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a, b)
		},
	}
}

type function3 struct {
	function

	Func rego.Builtin3
}

func (f *function3) asOption() func(*rego.Rego) {
	return rego.Function3(f.Decl, f.Func)
}

func (f *function3) asTester() *tester.Builtin {
	return f.function.asTester(f.asOption())
}

func mockFunction3(base func(tflint.Runner) *function3) *function4 {
	return &function4{
		function: base(nil).mockDecl(),
		Func: func(ctx rego.BuiltinContext, a *ast.Term, b *ast.Term, c *ast.Term, sourcesArg *ast.Term) (*ast.Term, error) {
			var sources map[string]string
			if err := ast.As(sourcesArg.Value, &sources); err != nil {
				return nil, err
			}
			runner, diags := NewTestRunner(sources)
			if diags.HasErrors() {
				return nil, diags
			}
			return base(runner).Func(ctx, a, b, c)
		},
	}
}

type function4 struct {
	function

	Func rego.Builtin4
}

func (f *function4) asOption() func(*rego.Rego) {
	return rego.Function4(f.Decl, f.Func)
}

func (f *function4) asTester() *tester.Builtin {
	return f.function.asTester(f.asOption())
}

type functionDyn struct {
	function

	Func rego.BuiltinDyn
}

func (f *functionDyn) asOption() func(*rego.Rego) {
	return rego.FunctionDyn(f.Decl, f.Func)
}

func (f *functionDyn) asTester() *tester.Builtin {
	return f.function.asTester(f.asOption())
}
