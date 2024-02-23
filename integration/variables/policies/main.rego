package tflint

import rego.v1

deny_empty_description contains issue if {
	vars := terraform.variables({"description": "string"}, {})
	description := vars[_].config.description

	description.value == ""

	issue := tflint.issue("empty description is not allowed", description.range)
}
