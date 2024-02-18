package tflint
import future.keywords

mock_removed_blocks(schema, options) := terraform.mock_removed_blocks(schema, options, {"main.tf": `
removed {
  from = aws_instance.example

  lifecycle {
    destroy = false
  }
}`})

test_deny_removed_blocks_passed if {
  issues := deny_removed_blocks with terraform.removed_blocks as mock_removed_blocks

  count(issues) == 1
  issue := issues[_]
  issue.msg == "removed blocks are not allowed"
}

test_deny_removed_blocks_failed if {
  issues := deny_removed_blocks with terraform.removed_blocks as mock_removed_blocks

  count(issues) == 0
}
