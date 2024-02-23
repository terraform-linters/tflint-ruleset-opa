package tflint

import rego.v1

deny_too_many_locals contains issue if {
	locals := terraform.locals({})
	count(locals) > 5

	issue := tflint.issue("too many local values", terraform.module_range())
}
