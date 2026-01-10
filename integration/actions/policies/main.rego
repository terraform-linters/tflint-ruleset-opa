package tflint

import rego.v1

deny_deprecated_function_invokes contains issue if {
	actions := terraform.actions("aws_lambda_invoke", {"config": {"function_name": "string"}}, {})
	function_name := actions[_].config.config[_].config.function_name
	contains(function_name.value, "deprecated-function")

	issue := tflint.issue("deprecated-function is deprecated", function_name.range)
}
