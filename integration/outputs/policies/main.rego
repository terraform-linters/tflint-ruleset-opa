package tflint

deny_no_outputs[issue] {
  outputs := terraform.outputs({}, {})
  count(outputs) == 0

  issue := tflint.issue("module must expose outputs", terraform.module_range())
}
