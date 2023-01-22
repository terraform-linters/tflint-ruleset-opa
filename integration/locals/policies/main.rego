package tflint

deny_too_many_locals[issue] {
  locals := terraform.locals({})
  count(locals) > 5

  issue := tflint.issue("too many local values", terraform.module_range())
}
