package tflint
import future.keywords

mock_actions(type, schema, options) := terraform.mock_actions(type, schema, options, {"main.tf": `
action "aws_lambda_invoke" "invalid" {
  config {
    function_name = "123456789012:function:deprecated-function:1"
  }
}`})

test_deny_deprecated_function_invokes_passed if {
  issues := deny_deprecated_function_invokes with terraform.actions as mock_actions

  count(issues) == 1
  issue := issues[_]
  issue.msg == "deprecated-function is deprecated"
}

test_deny_deprecated_function_invokes_failed if {
  issues := deny_deprecated_function_invokes with terraform.actions as mock_actions

  count(issues) == 0
}
