package tflint

import rego.v1

deny_resource_declarations contains issue if {
	resources := terraform.resources("*", {}, {"expand_mode": "none"})
	count(resources) > 0

	issue := tflint.issue("Declaring resources is not allowed. Use modules instead.", resources[0].decl_range)
}
