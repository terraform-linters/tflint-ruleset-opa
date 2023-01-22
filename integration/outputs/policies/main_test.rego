package tflint
import future.keywords

mock_outputs(schema, options) := terraform.mock_outputs(schema, options, {})

test_deny_no_outputs_passed if {
  issues := deny_no_outputs with terraform.outputs as mock_outputs

  count(issues) == 1
  issue := issues[_]
  issue.msg == "module must expose outputs"
  issue.range.start.line == 1
}

test_deny_no_outputs_failed if {
  issues := deny_no_outputs with terraform.outputs as mock_outputs

  count(issues) == 0
}
