package tflint

import rego.v1

deny_us_east_1 contains issue if {
	providers := terraform.providers({"region": "string"}, {})
	region := providers[_].config.region

	region.value == "us-east-1"

	issue := tflint.issue("us-east-1 is not allowed", region.range)
}

deny_provider_ref contains issue if {
	resources := terraform.resources("*", {"provider": "any"}, {})
	provider := resources[_].config.provider

	issue := tflint.issue("provider reference is not allowed", provider.range)
}
