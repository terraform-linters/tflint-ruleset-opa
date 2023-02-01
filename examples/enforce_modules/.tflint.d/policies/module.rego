package tflint

deny_resource_declarations[issue] {
  resources := terraform.resources("*", {}, {"expand_mode": "none"})
  count(resources) > 0

  issue := tflint.issue("Declaring resources is not allowed. Use modules instead.", resources[0].decl_range)
}
