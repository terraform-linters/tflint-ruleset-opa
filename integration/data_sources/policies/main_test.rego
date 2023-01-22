package tflint
import future.keywords

mock_data_sources(type, schema, options) := terraform.mock_data_sources(type, schema, options, {"main.tf": `
data "aws_ami" "main" {
  owners = ["amazon"]
}`})

test_deny_other_ami_owners_passed if {
  issues := deny_other_ami_owners with terraform.data_sources as mock_data_sources

  count(issues) == 1
  issue := issues[_]
  issue.msg == "third-party AMI is not allowed"
}

test_deny_other_ami_owners_failed if {
  issues := deny_other_ami_owners with terraform.data_sources as mock_data_sources

  count(issues) == 0
}
