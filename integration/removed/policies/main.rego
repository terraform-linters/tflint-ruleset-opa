package tflint

deny_removed_blocks[issue] {
  moved := terraform.removed_blocks({}, {})
  count(moved) > 0

  issue := tflint.issue("removed blocks are not allowed", moved[0].decl_range)
}
