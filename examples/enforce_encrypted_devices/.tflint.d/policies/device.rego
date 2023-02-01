package tflint
import future.keywords.contains
import future.keywords.if

# Unexpanded resources: unknown check
aws_instances_unexpanded := terraform.resources("aws_instance", {"count": "number", "for_each": "any", "dynamic": {"__labels": ["type"], "for_each": "any"}}, {"expand_mode": "none"})

aws_instance_counts contains count if {
  count := aws_instances_unexpanded[_].config.count
}
aws_instance_for_eachs contains for_each if {
  for_each := aws_instances_unexpanded[_].config.for_each
}
dynamic_ebs_block_devices contains device if {
  device := aws_instances_unexpanded[_].config.dynamic[_]
  device.labels[0] == "ebs_block_device"
}

# Expanded resources: encrypted flag check
aws_instances := terraform.resources("aws_instance", {"ebs_block_device": {"encrypted": "bool"}}, {})

ebs_block_devices contains device if {
  device := aws_instances[_].config.ebs_block_device[_]
}

# Rules for unknown
deny_unencrypted_ebs_block_device[issue] {
  aws_instance_counts[i].unknown

  issue := tflint.issue("Dynamic value is not allowed in count", aws_instance_counts[i].range)
}
deny_unencrypted_ebs_block_device[issue] {
  aws_instance_for_eachs[i].unknown

  issue := tflint.issue("Dynamic value is not allowed in for_each", aws_instance_for_eachs[i].range)
}
deny_unencrypted_ebs_block_device[issue] {
  dynamic_ebs_block_devices[i].config.for_each.unknown

  issue := tflint.issue("Dynamic value is not allowed in for_each",  dynamic_ebs_block_devices[i].config.for_each.range)
}
deny_unencrypted_ebs_block_device[issue] {
  ebs_block_devices[i].config.encrypted.unknown

  issue := tflint.issue("Dynamic value is not allowed in encrypted", ebs_block_devices[i].config.encrypted.range)
}
# Rules for undefined
deny_unencrypted_ebs_block_device[issue] {
  ebs_block_devices[i].config == {}

  issue := tflint.issue("EBS block device must be encrypted", ebs_block_devices[i].decl_range)
}
# Rules for invaid value and null
deny_unencrypted_ebs_block_device[issue] {
  ebs_block_devices[i].config.encrypted.value != true

  issue := tflint.issue("EBS block device must be encrypted", ebs_block_devices[i].config.encrypted.range)
}
