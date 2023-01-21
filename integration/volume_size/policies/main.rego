package tflint

deny_large_volume[issue] {
  resources := terraform.resources("aws_instance", {"ebs_block_device": {"volume_size": "number"}}, {})
  volume_size := resources[_].config.ebs_block_device[_].config.volume_size
  volume_size.value > 30

  issue := tflint.issue("volume size should be 30GB or less", volume_size.range)
}
