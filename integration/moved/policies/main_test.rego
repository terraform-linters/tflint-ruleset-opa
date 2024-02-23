package tflint

import rego.v1

mock_moved_blocks(schema, options) := terraform.mock_moved_blocks(schema, options, {"main.tf": `
moved {
  from = aws_instance.foo
  to   = aws_instance.bar
}`})

test_deny_moved_blocks_passed if {
	issues := deny_moved_blocks with terraform.moved_blocks as mock_moved_blocks

	count(issues) == 1
	issue := issues[_]
	issue.msg == "moved blocks are not allowed"
}

test_deny_moved_blocks_failed if {
	issues := deny_moved_blocks with terraform.moved_blocks as mock_moved_blocks

	count(issues) == 0
}
