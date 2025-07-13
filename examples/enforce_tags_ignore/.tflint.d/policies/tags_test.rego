package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "invalid" {
  lifecycle {
    ignore_changes = [key_name, ami]
  }
}`})

test_deny_instance_without_tags_ignore_failed if {
	issues := deny_instance_without_tags_ignore with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must have "ignore_changes = [tags]"`
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  lifecycle {
    ignore_changes = [key_name, tags]
  }
}`})

test_deny_instance_without_tags_ignore_passed if {
	issues := deny_instance_without_tags_ignore with terraform.resources as passed_resources

	count(issues) == 0
}

without_ignore_changes_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "without_ignore_changes" {
  lifecycle {
    create_before_destroy = true
  }
}`})

test_deny_instance_without_tags_ignore_without_ignore_changes if {
	issues := deny_instance_without_tags_ignore with terraform.resources as without_ignore_changes_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must have "ignore_changes = [tags]"`
}

without_lifecycle_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "without_lifecycle" {
  instance_type = "t2.micro"
}`})

test_deny_instance_without_tags_ignore_without_lifecycle if {
	issues := deny_instance_without_tags_ignore with terraform.resources as without_lifecycle_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `instance must have "ignore_changes = [tags]"`
}

deprecated_ignore_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  lifecycle {
    ignore_changes = ["key_name", "tags"]
  }
}`})

test_deny_instance_without_tags_ignore_deprecated_ignore if {
	issues := deny_instance_without_tags_ignore with terraform.resources as deprecated_ignore_resources

	count(issues) == 0
}

ignore_all_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  lifecycle {
    ignore_changes = all
  }
}`})

test_deny_instance_without_tags_ignore_deprecated_ignore if {
	issues := deny_instance_without_tags_ignore with terraform.resources as ignore_all_resources

	count(issues) == 0
}

deprecated_ignore_all_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  lifecycle {
    ignore_changes = "all"
  }
}`})

test_deny_instance_without_tags_ignore_deprecated_ignore if {
	issues := deny_instance_without_tags_ignore with terraform.resources as deprecated_ignore_all_resources

	count(issues) == 0
}
