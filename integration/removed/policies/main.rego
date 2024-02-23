package tflint

import rego.v1

deny_removed_blocks contains issue if {
	moved := terraform.removed_blocks({}, {})
	count(moved) > 0

	issue := tflint.issue("removed blocks are not allowed", moved[0].decl_range)
}
