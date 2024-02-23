package tflint

import rego.v1

mock_resources_not_tagged(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  tags = {
    "production" = true
  }
}`})

test_deny_not_tagged_instance_passed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_not_tagged

	count(issues) == 1
	issue := issues[_]
	issue.msg == "instance should be tagged with Environment"
}

test_deny_not_tagged_instance_failed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_not_tagged

	count(issues) == 0
}

mock_resources_no_tags(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
}`})

test_deny_not_tagged_instance_without_tags_passed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_no_tags

	count(issues) == 1
	issue := issues[_]
	issue.msg == "instance should be tagged with Environment"
}

test_deny_not_tagged_instance_without_tags_failed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_no_tags

	count(issues) == 0
}

mock_resources_null_tags(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  tags = null
}`})

test_deny_not_tagged_instance_null_tags_passed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_null_tags

	count(issues) == 1
	issue := issues[_]
	issue.msg == "instance should be tagged with Environment"
}

test_deny_not_tagged_instance_null_tags_failed if {
	issues := deny_not_tagged_instance with terraform.resources as mock_resources_null_tags

	count(issues) == 0
}
