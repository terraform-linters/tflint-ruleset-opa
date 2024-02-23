# Enforce encrypted devices

This is an example of rules for attributes in nested blocks.

## Requirements

- Disallow `encrypted = false` in EBS block devices.
- Disallow device without the `encrypted` attribute because the default is `false`.
- Disallow all unknown cases (unknown value, meta-arguments, dynamic blocks).
- Ignore if resource not created.

## Results

```console
$ tflint
8 issue(s) found:

Error: EBS block device must be encrypted (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 3:
   3:     encrypted = false

Reference: .tflint.d/policies/device.rego:29

Error: EBS block device must be encrypted (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 14:
  14:   ebs_block_device {

Reference: .tflint.d/policies/device.rego:29

Error: EBS block device must be encrypted (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 20:
  20:     encrypted = null

Reference: .tflint.d/policies/device.rego:29

Error: EBS block device must be encrypted (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 29:
  29:       encrypted = ebs_block_device.value

Reference: .tflint.d/policies/device.rego:29

Error: Dynamic value is not allowed in encrypted (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 38:
  38:     encrypted = var.unknown

Reference: .tflint.d/policies/device.rego:29

Error: Dynamic value is not allowed in count (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 43:
  43:   count = var.unknown

Reference: .tflint.d/policies/device.rego:29

Error: Dynamic value is not allowed in for_each (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 51:
  51:   for_each = var.unknown

Reference: .tflint.d/policies/device.rego:29

Error: Dynamic value is not allowed in for_each (opa_deny_unencrypted_ebs_block_device)

  on main.tf line 60:
  60:     for_each = var.unknown

Reference: .tflint.d/policies/device.rego:29

```
