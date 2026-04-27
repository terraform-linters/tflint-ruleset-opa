package tflint

import rego.v1

deny_not_t2_micro contains issue if {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type
	instance_type.value != "t2.micro"
	issue := tflint.issue("BUNDLE: t2.micro is only allowed", instance_type.range)
}
