package tflint

deny_remote_source[issue] {
  modules := terraform.module_calls({"source": "string"}, {})
  source := modules[_].config.source

  not startswith(source.value, "./")

  issue := tflint.issue("remote module is not allowed", source.range)
}
