package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "invalid" {
  ami             = "ami-12345"
  instance_type = "t2.micro"

  provisioner "local-exec" {
    command = "echo hello"
  }
}`})

test_deny_local_exec_provisioner_failed if {
	issues := deny_local_exec_provisioner with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `local-exec provisioner is not allowed (on aws_instance.invalid)`
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "valid" {
  ami             = "ami-12345"
  instance_type = "t2.micro"

  provisioner "remote-exec" {
    host = "example.com"
  }
}`})

test_deny_local_exec_provisioner_passed if {
	issues := deny_local_exec_provisioner with terraform.resources as passed_resources

	count(issues) == 0
}

default_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "default" {
  ami             = "ami-12345"
  instance_type = "t2.micro"
}`})

test_deny_local_exec_provisioner_default if {
	issues := deny_local_exec_provisioner with terraform.resources as default_resources

	count(issues) == 0
}

null_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "null" {
  ami             = "ami-12345"
  instance_type = "t2.micro"

  provisioner "local-exec" {
    command = null
  }
}`})

test_deny_local_exec_provisioner_null if {
	issues := deny_local_exec_provisioner with terraform.resources as null_resources

	count(issues) == 0
}

unknown_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "provisioner_type" {
  default = "local-exec"
}

resource "aws_instance" "unknown" {
  ami             = "ami-12345"
  instance_type = "t2.micro"

  provisioner var.provisioner_type {
    command = "echo hello"
  }
}`})

test_deny_local_exec_provisioner_unknown if {
	issues := deny_local_exec_provisioner with terraform.resources as unknown_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `local-exec provisioner is not allowed (on aws_instance.unknown)`
}

unknown_count_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_count" {
  count = var.unknown
}`})

test_deny_local_exec_provisioner_unknown_count if {
	issues := deny_local_exec_provisioner with terraform.resources as unknown_count_resources

	count(issues) == 0
}

unknown_for_each_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_for_each" {
  for_each = var.unknown
}`})

test_deny_local_exec_provisioner_unknown_for_each if {
	issues := deny_local_exec_provisioner with terraform.resources as unknown_for_each_resources

	count(issues) == 0
}

unknown_dynamic_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "unknown_dynamic" {
  ami             = "ami-12345"
  instance_type = "t2.micro"

  dynamic "provisioner" {
    for_each = var.unknown
    content {
      "local-exec" {
        command = "echo hello"
      }
    }
  }
}`})

test_deny_local_exec_provisioner_unknown_dynamic if {
	issues := deny_local_exec_provisioner with terraform.resources as unknown_dynamic_resources

	count(issues) == 0
}
