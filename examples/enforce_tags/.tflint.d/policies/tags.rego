package tflint

aws_instances := terraform.resources("aws_instance", {"tags": "map(string)"}, {"expand_mode": "none"})

contains(array, elem) {
  array[_] = elem
}
# "tags" is null
is_untagged(config) {
  is_null(config.tags.value)
}
# "tags" is defined, but "Environment" not found
is_untagged(config) {
  not is_null(config.tags.value)
  not contains(object.keys(config.tags.value), "Environment")
}
# "tags" is not defined
is_untagged(config) {
  not contains(object.keys(config), "tags")
}

# Rules for unknown tags
deny_untagged_instance[issue] {
  aws_instances[i].config.tags.unknown

  issue := tflint.issue("Dynamic value is not allowed in tags", aws_instances[i].config.tags.range)
}
# Rules for invalid tags
deny_untagged_instance[issue] {
  is_untagged(aws_instances[i].config)

  issue := tflint.issue(`instance must be tagged with the "Environment" tag`, aws_instances[i].decl_range)
}
