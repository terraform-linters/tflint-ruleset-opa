package tflint

import rego.v1

is_not_tagged(config) if {
	is_null(config.tags.value)
}

is_not_tagged(config) if {
	not is_null(config.tags.value)
	not "Environment" in object.keys(config.tags.value)
}

is_not_tagged(config) if {
	not "tags" in object.keys(config)
}

deny_not_tagged_instance contains issue if {
	resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
	resource := resources[_]

	is_not_tagged(resource.config)

	issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}
