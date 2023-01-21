package tflint

deny_us_east_1[issue] {
  providers := terraform.providers({"region": "string"}, {})
  region := providers[_].config.region

  region.value == "us-east-1"

  issue := tflint.issue("us-east-1 is not allowed", region.range)
}

deny_provider_ref[issue] {
  resources := terraform.resources("*", {"provider": "any"}, {})
  provider := resources[_].config.provider

  issue := tflint.issue("provider reference is not allowed", provider.range)
}
