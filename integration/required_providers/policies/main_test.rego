package tflint

import rego.v1

mock_required_providers(options) := terraform.mock_required_providers(options, {"main.tf": `
terraform {
  required_providers {
    azurerm = {
      source                = "badguy/azurerm"
      version               = "~> 4.0"
      configuration_aliases = ["azurerm.foo"]
    }
  }
}`})

test_deny_unapproved_source_passed if {
	issues := deny_unapproved_source with terraform.required_providers as mock_required_providers

	count(issues) == 1
	issue := issues[_]
	issue.msg == `provider source "badguy/azurerm" is not allowed`
}

test_deny_unapproved_source_failed if {
	issues := deny_unapproved_source with terraform.required_providers as mock_required_providers

	count(issues) == 0
}
