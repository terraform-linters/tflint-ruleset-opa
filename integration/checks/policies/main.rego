package tflint

import rego.v1

deny_deterministic_check_condition contains issue if {
	checks := terraform.checks({"assert": {"condition": "bool"}}, {})
	condition = checks[_].config.assert[_].config.condition
	condition.unknown == false

	issue := tflint.issue("deterministic check condtion is not allowed", condition.range)
}
