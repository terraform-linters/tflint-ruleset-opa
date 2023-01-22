package tflint
import future.keywords

mock_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main-v2" {}
`})

test_deny_not_snake_case_passed if {
  issues := deny_not_snake_case
    with terraform.resources as mock_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == "main-v2 is not snake case"
  issue.range.start.line == 2
}

test_deny_not_snake_case_failed if {
  issues := deny_not_snake_case
    with terraform.resources as mock_resources

  count(issues) == 0
}
