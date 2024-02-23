package tflint

import rego.v1

mock_variables(schema, options) := terraform.mock_variables(schema, options, {"main.tf": `
variable "foo" {
  description = ""
}`})

test_deny_empty_description_passed if {
	issues := deny_empty_description with terraform.variables as mock_variables

	count(issues) == 1
	issue := issues[_]
	issue.msg == "empty description is not allowed"
}

test_deny_empty_description_failed if {
	issues := deny_empty_description with terraform.variables as mock_variables

	count(issues) == 0
}
