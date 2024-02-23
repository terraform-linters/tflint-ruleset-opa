package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "invalid" {
  ebs_block_device {
    encrypted = false
  }
}`})

test_deny_unencrypted_ebs_block_device_failed if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "EBS block device must be encrypted"
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  ebs_block_device {
    encrypted = true
  }
}`})

test_deny_unencrypted_ebs_block_device_passed if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as passed_resources

	count(issues) == 0
}

default_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "default" {
  ebs_block_device {
  }
}`})

test_deny_unencrypted_ebs_block_device_default if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as default_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "EBS block device must be encrypted"
}

null_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "null" {
  ebs_block_device {
    encrypted = null
  }
}`})

test_deny_unencrypted_ebs_block_device_default if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as null_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "EBS block device must be encrypted"
}

unknown_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown" {
  ebs_block_device {
    encrypted = var.unknown
  }
}`})

test_deny_unencrypted_ebs_block_device_unknown if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as unknown_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Dynamic value is not allowed in encrypted"
}

unknown_count_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_count" {
  count = var.unknown
}`})

test_deny_unencrypted_ebs_block_device_unknown_count if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as unknown_count_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Dynamic value is not allowed in count"
}

unknown_for_each_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_for_each" {
  for_each = var.unknown
}`})

test_deny_unencrypted_ebs_block_device_unknown_for_each if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as unknown_for_each_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Dynamic value is not allowed in for_each"
}

unknown_dynamic_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_dynamic" {
  dynamic "ebs_block_device" {
    for_each = var.unknown
  }
}`})

test_deny_unencrypted_ebs_block_device_unknown_dynamic if {
	issues := deny_unencrypted_ebs_block_device with terraform.resources as unknown_dynamic_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "Dynamic value is not allowed in for_each"
}
