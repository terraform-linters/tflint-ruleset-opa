package tflint

import rego.v1

deny_weak_password contains issue if {
	passwords := terraform.ephemeral_resources("random_password", {"length": "number"}, {})
	length := passwords[_].config.length
	length.value < 32

	issue := tflint.issue("Password must be at least 32 characters long", length.range)
}
