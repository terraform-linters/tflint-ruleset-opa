package tflint

contains(array, elem) {
  array[_] = elem
}
is_not_tagged(config) {
  is_null(config.tags.value)
}
is_not_tagged(config) {
  not is_null(config.tags.value)
  not contains(object.keys(config.tags.value), "Environment")
}
is_not_tagged(config) {
  not contains(object.keys(config), "tags")
}

deny_not_tagged_instance[issue] {
  resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
  resource := resources[_]

  is_not_tagged(resource.config)

  issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}
