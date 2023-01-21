package tflint

deny_default_hostname[issue] {
  settings := terraform.settings({"cloud": {"hostname": "string"}}, {})
  hostname := settings[_].config.cloud[_].config.hostname

  hostname.value == "app.terraform.io"

  issue := tflint.issue("default hostname should be omitted", hostname.range)
}
