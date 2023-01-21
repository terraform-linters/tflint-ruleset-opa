package tflint
import future.keywords

mock_providers(schema, options) := terraform.mock_providers(schema, options, {"main.tf": `
provider "aws" {
  alias  = "east"
  region = "us-east-1"
}`})

test_deny_us_east_1_passed if {
  issues := deny_us_east_1 with terraform.providers as mock_providers

  count(issues) == 1
  issue := issues[_]
  issue.msg == "us-east-1 is not allowed"
}

test_deny_us_east_1_failed if {
  issues := deny_us_east_1 with terraform.providers as mock_providers

  count(issues) == 0
}
