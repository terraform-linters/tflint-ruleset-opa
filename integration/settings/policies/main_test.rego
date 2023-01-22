package tflint
import future.keywords

mock_settings(schema, options) := terraform.mock_settings(schema, options, {"main.tf": `
terraform {
  cloud {
    hostname = "app.terraform.io"
  }
}`})

test_deny_default_hostname_passed if {
  issues := deny_default_hostname with terraform.settings as mock_settings

  count(issues) == 1
  issue := issues[_]
  issue.msg == "default hostname should be omitted"
}

test_deny_default_hostname_failed if {
  issues := deny_default_hostname with terraform.settings as mock_settings

  count(issues) == 0
}
