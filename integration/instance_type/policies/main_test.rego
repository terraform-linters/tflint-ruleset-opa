package tflint

import rego.v1

mock_resources_t1_micro(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  instance_type = "t1.micro"
}`})

test_not_deny_t2_micro_passed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_t1_micro

	count(issues) == 1
	issue := issues[_]
	issue.msg == "t2.micro is only allowed"
}

test_not_deny_t2_micro_failed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_t1_micro

	count(issues) == 0
}

mock_resources_unknown_instance_type(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}
resource "aws_instance" "main" {
  instance_type = var.unknown
}`})

test_not_deny_t2_micro_unknown_passed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_unknown_instance_type

	count(issues) == 1
	issue := issues[_]
	issue.msg == "instance type is unknown"
}

test_not_deny_t2_micro_unknown_failed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_unknown_instance_type

	count(issues) == 0
}

mock_resources_sensitive_instance_type(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "sensitive" {
  default = "t2.micro"
  sensitive = true
}
resource "aws_instance" "main" {
  instance_type = var.sensitive
}`})

test_not_deny_t2_micro_sensitive_passed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_sensitive_instance_type

	count(issues) == 2
}

test_not_deny_t2_micro_sensitive_failed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_sensitive_instance_type

	count(issues) == 0
}

mock_resources_ephemeral_instance_type(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "ephemeral" {
  default = "t2.micro"
  ephemeral = true
}
resource "aws_instance" "main" {
  instance_type = var.ephemeral
}`})

test_not_deny_t2_micro_ephemeral_passed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_ephemeral_instance_type

	count(issues) == 2
}

test_not_deny_t2_micro_ephemeral_failed if {
	issues := deny_not_t2_micro with terraform.resources as mock_resources_ephemeral_instance_type

	count(issues) == 0
}
