package tflint

import rego.v1

deny_not_snake_case contains issue if {
	resources := terraform.resources("*", {}, {})
	not regex.match("^[a-z][a-z0-9]*(_[a-z0-9]+)*$", resources[i].name)

	issue := tflint.issue(sprintf("%s is not snake case", [resources[i].name]), resources[i].decl_range)
}

deny_not_t2_micro contains issue if {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.unknown == true

	issue := tflint.issue("instance type is unknown", instance_type.range)
}

deny_not_t2_micro contains issue if {
	resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	instance_type := resources[_].config.instance_type

	instance_type.value != "t2.micro"

	issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}
