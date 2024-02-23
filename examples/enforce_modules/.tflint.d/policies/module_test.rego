package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  instance_type = "t2.micro"
}`})

test_deny_resource_declarations_failed if {
	issues := deny_resource_declarations with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Declaring resources is not allowed. Use modules instead."
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
module "aws_instance" {
  source = "../modules/aws_instance"

  instance_type = "t2.micro"
}`})

test_deny_resource_declarations_passed if {
	issues := deny_resource_declarations with terraform.resources as passed_resources

	count(issues) == 0
}
