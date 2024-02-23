package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "invalid" {
  tags = {
    "production" = true
  }
}`})

test_deny_untagged_instance_failed if {
	issues := deny_untagged_instance with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must be tagged with the "Environment" tag`
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  tags = {
    "Environment" = "production"
  }
}`})

test_deny_untagged_instance_passed if {
	issues := deny_untagged_instance with terraform.resources as passed_resources

	count(issues) == 0
}

undef_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "undef" {
}`})

test_deny_untagged_instance_undef if {
	issues := deny_untagged_instance with terraform.resources as undef_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must be tagged with the "Environment" tag`
}

null_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "undef" {
  tags = null
}`})

test_deny_untagged_instance_null if {
	issues := deny_untagged_instance with terraform.resources as null_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must be tagged with the "Environment" tag`
}

unknown_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown" {
  tags = var.unknown
}`})

test_deny_untagged_instance_null if {
	issues := deny_untagged_instance with terraform.resources as unknown_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Dynamic value is not allowed in tags"
}
