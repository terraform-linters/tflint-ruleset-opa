package tflint

import rego.v1

deny_all_resources contains issue if {
	resources := terraform.resources("*", {}, {})
	count(resources) > 0

	issue := tflint.issue("resource is not allowed", resources[0].decl_range)
}

deny_no_resources contains issue if {
	resources := terraform.resources("*", {}, {})
	count(resources) == 0

	issue := tflint.issue("resources should be declared", terraform.module_range())
}
