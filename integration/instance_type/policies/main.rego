package tflint

deny_not_t2_micro[issue] {
  resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
  instance_type := resources[_].config.instance_type

  instance_type.unknown == true

  issue := tflint.issue("instance type is unknown", instance_type.range)
}

deny_not_t2_micro[issue] {
  resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
  instance_type := resources[_].config.instance_type

  instance_type.sensitive == true

  issue := tflint.issue("instance type is sensitive", instance_type.range)
}

deny_not_t2_micro[issue] {
  resources := terraform.resources("aws_instance", {"instance_type": "string"}, {})
  instance_type := resources[_].config.instance_type

  instance_type.value != "t2.micro"

  issue := tflint.issue("t2.micro is only allowed", instance_type.range)
}
