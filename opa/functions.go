package opa

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/tester"
	"github.com/open-policy-agent/opa/types"
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
		moduleRangeFunc(runner).asOption(),
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
		moduleRangeFunc(runner).asTester(),
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
			return typedBlockFunc(dataType, schema, options, "data", runner)
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
