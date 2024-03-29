package tflint

import rego.v1

mock_data_sources(type, schema, options) := terraform.mock_data_sources(type, schema, options, {"main.tf": `
data "aws_ami" "main" {
  owners = ["amazon"]
}

check "scope" {
  data "aws_ami" "scoped" {
    owners = ["amazon"]
  }
}`})

test_deny_other_ami_owners_passed if {
	issues := deny_other_ami_owners with terraform.data_sources as mock_data_sources

	count(issues) == 2
	issue := issues[_]
	issue.msg == "third-party AMI is not allowed"
}

test_deny_other_ami_owners_failed if {
	issues := deny_other_ami_owners with terraform.data_sources as mock_data_sources

	count(issues) == 0
}
