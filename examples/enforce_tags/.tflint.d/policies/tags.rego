package tflint

import rego.v1

aws_instances := terraform.resources("aws_instance", {"tags": "map(string)"}, {"expand_mode": "none"})

# "tags" is null
is_untagged(config) if {
	is_null(config.tags.value)
}

# "tags" is defined, but "Environment" not found
is_untagged(config) if {
	not is_null(config.tags.value)
	not "Environment" in object.keys(config.tags.value)
}

# "tags" is not defined
is_untagged(config) if {
	not "tags" in object.keys(config)
}

# Rules for unknown tags
deny_untagged_instance contains issue if {
	aws_instances[i].config.tags.unknown

	issue := tflint.issue("Dynamic value is not allowed in tags", aws_instances[i].config.tags.range)
}

# Rules for invalid tags
deny_untagged_instance contains issue if {
	is_untagged(aws_instances[i].config)

	issue := tflint.issue(`instance must be tagged with the "Environment" tag`, aws_instances[i].decl_range)
}
