package opa

import (
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-opa/opa/funcs"
)

// Functions return custom functions as Rego options.
func Functions(runner tflint.Runner) []func(*rego.Rego) {
	return []func(*rego.Rego){
		funcs.ResourcesFunc(runner).Rego(),
		funcs.DataSourcesFunc(runner).Rego(),
		funcs.ModuleCallsFunc(runner).Rego(),
		funcs.ProvidersFunc(runner).Rego(),
		funcs.SettingsFunc(runner).Rego(),
		funcs.VariablesFunc(runner).Rego(),
		funcs.OutputsFunc(runner).Rego(),
		funcs.LocalsFunc(runner).Rego(),
		funcs.MovedBlocksFunc(runner).Rego(),
		funcs.ImportsFunc(runner).Rego(),
		funcs.ChecksFunc(runner).Rego(),
		funcs.RemovedBlocksFunc(runner).Rego(),
		funcs.EphemeralResourcesFunc(runner).Rego(),
		funcs.ActionsFunc(runner).Rego(),
		funcs.ModuleRangeFunc(runner).Rego(),
		funcs.ExprListFunc().Rego(),
		funcs.ExprMapFunc().Rego(),
		funcs.ExprCallFunc().Rego(),
		funcs.IssueFunc().Rego(),
	}
}

// TesterFunctions return custom functions as tester.Builtin.
func TesterFunctions(runner tflint.Runner) []*tester.Builtin {
	return []*tester.Builtin{
		funcs.ResourcesFunc(runner).Tester(),
		funcs.DataSourcesFunc(runner).Tester(),
		funcs.ModuleCallsFunc(runner).Tester(),
		funcs.ProvidersFunc(runner).Tester(),
		funcs.SettingsFunc(runner).Tester(),
		funcs.VariablesFunc(runner).Tester(),
		funcs.OutputsFunc(runner).Tester(),
		funcs.LocalsFunc(runner).Tester(),
		funcs.MovedBlocksFunc(runner).Tester(),
		funcs.ImportsFunc(runner).Tester(),
		funcs.ChecksFunc(runner).Tester(),
		funcs.RemovedBlocksFunc(runner).Tester(),
		funcs.EphemeralResourcesFunc(runner).Tester(),
		funcs.ActionsFunc(runner).Tester(),
		funcs.ModuleRangeFunc(runner).Tester(),
		funcs.ExprListFunc().Tester(),
		funcs.ExprMapFunc().Tester(),
		funcs.ExprCallFunc().Tester(),
		funcs.IssueFunc().Tester(),
	}
}

// MockFunctions return mocks for custom functions as Rego options.
// Mock functions are usually not needed outside of testing,
// but are provided for compilation.
func MockFunctions() []func(*rego.Rego) {
	return []func(*rego.Rego){
		funcs.MockFunction3(funcs.ResourcesFunc).Rego(),
		funcs.MockFunction3(funcs.DataSourcesFunc).Rego(),
		funcs.MockFunction2(funcs.ModuleCallsFunc).Rego(),
		funcs.MockFunction2(funcs.ProvidersFunc).Rego(),
		funcs.MockFunction2(funcs.SettingsFunc).Rego(),
		funcs.MockFunction2(funcs.VariablesFunc).Rego(),
		funcs.MockFunction2(funcs.OutputsFunc).Rego(),
		funcs.MockFunction1(funcs.LocalsFunc).Rego(),
		funcs.MockFunction2(funcs.MovedBlocksFunc).Rego(),
		funcs.MockFunction2(funcs.ImportsFunc).Rego(),
		funcs.MockFunction2(funcs.ChecksFunc).Rego(),
		funcs.MockFunction2(funcs.RemovedBlocksFunc).Rego(),
		funcs.MockFunction3(funcs.EphemeralResourcesFunc).Rego(),
		funcs.MockFunction3(funcs.ActionsFunc).Rego(),
	}
}

// TesterMockFunctions return mocks for custom functions.
func TesterMockFunctions() []*tester.Builtin {
	return []*tester.Builtin{
		funcs.MockFunction3(funcs.ResourcesFunc).Tester(),
		funcs.MockFunction3(funcs.DataSourcesFunc).Tester(),
		funcs.MockFunction2(funcs.ModuleCallsFunc).Tester(),
		funcs.MockFunction2(funcs.ProvidersFunc).Tester(),
		funcs.MockFunction2(funcs.SettingsFunc).Tester(),
		funcs.MockFunction2(funcs.VariablesFunc).Tester(),
		funcs.MockFunction2(funcs.OutputsFunc).Tester(),
		funcs.MockFunction1(funcs.LocalsFunc).Tester(),
		funcs.MockFunction2(funcs.MovedBlocksFunc).Tester(),
		funcs.MockFunction2(funcs.ImportsFunc).Tester(),
		funcs.MockFunction2(funcs.ChecksFunc).Tester(),
		funcs.MockFunction2(funcs.RemovedBlocksFunc).Tester(),
		funcs.MockFunction3(funcs.EphemeralResourcesFunc).Tester(),
		funcs.MockFunction3(funcs.ActionsFunc).Tester(),
	}
}
