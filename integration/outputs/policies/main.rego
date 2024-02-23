package tflint

import rego.v1

deny_no_outputs contains issue if {
	outputs := terraform.outputs({}, {})
	count(outputs) == 0

	issue := tflint.issue("module must expose outputs", terraform.module_range())
}
