package tflint

warn_gp3_volume[issue] {
  resources := terraform.resources("aws_instance", {"count": "number"}, {"expand_mode": "none"})
  count := resources[_].config.count

  count.unknown == true

  issue := tflint.issue("unknown resource found", count.range)
}

warn_gp3_volume[issue] {
  resources := terraform.resources("aws_instance", {"dynamic": {"__labels": ["name"], "for_each": "any"}}, {"expand_mode": "none"})
  for_each := resources[_].config.dynamic[_].config.for_each

  for_each.unknown == true

  issue := tflint.issue("unknown block found", for_each.range)
}

warn_gp3_volume[issue] {
  resources := terraform.resources("aws_instance", {"ebs_block_device": {"volume_type": "string"}}, {})
  volume_type := resources[_].config.ebs_block_device[_].config.volume_type

  volume_type.value == "gp3"

  issue := tflint.issue("gp3 is not allowed", volume_type.range)
}
