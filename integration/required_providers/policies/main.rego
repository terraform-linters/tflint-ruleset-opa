package tflint

import rego.v1

allowed_sources := {
	"hashicorp/aws",
	"hashicorp/google",
	"hashicorp/azurerm",
}

deny_unapproved_source contains issue if {
	required_providers := terraform.required_providers({})
	required_provider := required_providers[_]
	provider_name := object.keys(required_provider.config)[_]
	provider := required_provider.config[provider_name]
	source := provider.source
	not source.value in allowed_sources

	issue := tflint.issue(sprintf("provider source %q is not allowed", [source.value]), source.range)
}
