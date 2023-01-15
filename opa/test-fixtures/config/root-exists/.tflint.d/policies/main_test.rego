package tflint
import future.keywords

mock_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main-v2" {}
`})

test_deny_not_snake_case if {
  issues := deny_not_snake_case
    with terraform.resources as mock_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == "main-v2 is not snake case"
  issue.range.start.line == 2
}

mock_resources_t1_micro(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  instance_type = "t1.micro"
}`})

test_not_deny_t2_micro if {
  issues := deny_not_t2_micro with terraform.resources as mock_resources_t1_micro

  count(issues) == 1
  issue := issues[_]
  issue.msg == "t2.micro is only allowed"
  issue.range.start.line == 2
}
