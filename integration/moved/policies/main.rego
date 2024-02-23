package tflint

import rego.v1

deny_moved_blocks contains issue if {
	moved := terraform.moved_blocks({}, {})
	count(moved) > 0

	issue := tflint.issue("moved blocks are not allowed", moved[0].decl_range)
}
