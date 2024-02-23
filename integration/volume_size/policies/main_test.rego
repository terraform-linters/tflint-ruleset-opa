package tflint

import rego.v1

mock_resources_50gb(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  ebs_block_device {
    volume_size = 50
  }
}`})

test_deny_large_volume_passed if {
	issues := deny_large_volume with terraform.resources as mock_resources_50gb

	count(issues) == 1
	issue := issues[_]
	issue.msg == "volume size should be 30GB or less"
}

test_deny_large_volume_failed if {
	issues := deny_large_volume with terraform.resources as mock_resources_50gb

	count(issues) == 0
}

mock_resources_50gb_string(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  ebs_block_device {
    volume_size = "50"
  }
}`})

test_deny_large_volume_string_passed if {
	issues := deny_large_volume with terraform.resources as mock_resources_50gb_string

	count(issues) == 1
	issue := issues[_]
	issue.msg == "volume size should be 30GB or less"
}

test_deny_large_volume_string_failed if {
	issues := deny_large_volume with terraform.resources as mock_resources_50gb_string

	count(issues) == 0
}

mock_resources_30_5gb(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  ebs_block_device {
    volume_size = 30.5
  }
}`})

test_deny_large_volume_float_passed if {
	issues := deny_large_volume with terraform.resources as mock_resources_30_5gb

	count(issues) == 1
	issue := issues[_]
	issue.msg == "volume size should be 30GB or less"
}

test_deny_large_volume_float_failed if {
	issues := deny_large_volume with terraform.resources as mock_resources_30_5gb

	count(issues) == 0
}
