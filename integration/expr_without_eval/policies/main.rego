package tflint

import rego.v1

deny_require_lifecycle_ignore_tags contains issue if {
  res := terraform.resources("*",{"lifecycle": {"ignore_changes": "expr"}},{"expand_mode": "none"})[_]
  not has_tags_ignore(res)

  issue := tflint.issue(
    sprintf("Resource %s.%s must have lifecycle.ignore_changes include \"tags\"", [res.type, res.name]),
    res.decl_range
  )
}

# helper succeeds only if there's a lifecycle.ignore_changes list and one of its elements == "tags"
has_tags_ignore(res) if {
  lifeblock := res.config.lifecycle[_]

  ic := lifeblock.config.ignore_changes

  some i
  hcl.expr_list(ic)[i].value == "tags"
}

deny_deprecated_regions contains issue if {
  call := terraform.module_calls({"providers": "expr"}, {})[_]
  has_deprecated_region(call)

  issue := tflint.issue(
    "deprecated region reference found",
    call.config.providers.range,
  )
}

has_deprecated_region(call) if {
  providers := call.config.providers

  some i
  hcl.expr_map(providers)[i].value.value == "aws.usw2"
}

deny_deprecated_ami_version contains issue if {
  instance := terraform.resources("aws_instance", {"ami": "expr"}, {})[_]
  is_function_call_with_deprecated_version(instance)

  issue := tflint.issue(
    "get_ami_id function should be called with v1",
    instance.decl_range
  )
}

is_function_call_with_deprecated_version(instance) if {
  ami := instance.config.ami

  call := hcl.expr_call(ami)
  call.name == "get_ami_id"
  call.arguments[1].value != `"v1"`
}
