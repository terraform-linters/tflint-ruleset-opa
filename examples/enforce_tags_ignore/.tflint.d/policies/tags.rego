package tflint

import rego.v1

aws_instances := terraform.resources("aws_instance", {"lifecycle": {"ignore_changes": "expr"}}, {"expand_mode": "none"})

# ignore_changes = [tags]
has_tags_ignore(instance) if {
	ignore_changes := instance.config.lifecycle[_].config.ignore_changes

    startswith(ignore_changes.value, "[")
	some i
	hcl.expr_list(ignore_changes)[i].value == "tags"
}

# ignore_changes = ["tags"]
has_tags_ignore(instance) if {
	ignore_changes := instance.config.lifecycle[_].config.ignore_changes

    startswith(ignore_changes.value, "[")
	some i
	hcl.expr_list(ignore_changes)[i].value == `"tags"`
}

# ignore_changes = all
has_tags_ignore(instance) if {
	ignore_changes := instance.config.lifecycle[_].config.ignore_changes
    ignore_changes.value == "all"
}

# ignore_changes = "all"
has_tags_ignore(instance) if {
	ignore_changes := instance.config.lifecycle[_].config.ignore_changes
    ignore_changes.value == `"all"`
}

deny_instance_without_tags_ignore contains issue if {
	not has_tags_ignore(aws_instances[i])

	issue := tflint.issue(`instance must have "ignore_changes = [tags]"`, aws_instances[i].decl_range)
}
