package tflint

deny_moved_blocks[issue] {
  moved := terraform.moved_blocks({}, {})
  count(moved) > 0

  issue := tflint.issue("moved blocks are not allowed", moved[0].decl_range)
}
