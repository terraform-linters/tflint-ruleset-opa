package tflint

import rego.v1

mock_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {}
`})

test_deny_all_resources_passed if {
	issues := deny_all_resources with terraform.resources as mock_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "resource is not allowed"
}

test_deny_all_resources_failed if {
	issues := deny_all_resources with terraform.resources as mock_resources

	count(issues) == 0
}

mock_no_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {})

test_deny_no_resources_passed if {
	issues := deny_no_resources with terraform.resources as mock_no_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == "resources should be declared"
}

test_deny_no_resources_failed if {
	issues := deny_no_resources with terraform.resources as mock_no_resources

	count(issues) == 0
}
