package tflint
import future.keywords

mock_locals(options) := terraform.mock_locals(options, {})

test_deny_too_many_locals_passed if {
  issues := deny_too_many_locals with terraform.locals as mock_locals

  count(issues) == 1
  issue := issues[_]
  issue.msg == "module must expose outputs"
}

test_deny_too_many_locals_failed if {
  issues := deny_too_many_locals with terraform.locals as mock_locals

  count(issues) == 0
}
