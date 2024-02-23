package tflint

import rego.v1

deny_not_snake_case contains issue if {
	resources := terraform.resources("*", {}, {})
	not regex.match("^[a-z][a-z0-9]*(_[a-z0-9]+)*$", resources[i].name)

	issue := tflint.issue(sprintf("%s is not snake case", [resources[i].name]), resources[i].decl_range)
}
