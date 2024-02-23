package tflint

import rego.v1

mock_checks(schema, options) := terraform.mock_checks(schema, options, {"main.tf": `
check "deterministic" {
  assert {
    condition = 200 == 200
    error_message = "condition should be true"
  }
}`})

test_deny_deterministic_check_condition_passed if {
	issues := deny_deterministic_check_condition with terraform.checks as mock_checks

	count(issues) == 1
	issue := issues[_]
	issue.msg == "deterministic check condtion is not allowed"
}

test_deny_deterministic_check_condition_failed if {
	issues := deny_deterministic_check_condition with terraform.checks as mock_checks

	count(issues) == 0
}
