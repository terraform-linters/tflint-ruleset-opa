package tflint

import rego.v1

mock_module_calls(schema, options) := terraform.mock_module_calls(schema, options, {"main.tf": `
module "remote" {
  source = "github.com/hashicorp/example"
}`})

test_deny_remote_source_passed if {
	issues := deny_remote_source with terraform.module_calls as mock_module_calls

	count(issues) == 1
	issue := issues[_]
	issue.msg == "remote module is not allowed"
}

test_deny_remote_source_failed if {
	issues := deny_remote_source with terraform.module_calls as mock_module_calls

	count(issues) == 0
}
