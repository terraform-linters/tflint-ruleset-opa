# Enforce local-exec provisioner

This is an example of enforcing that `local-exec` provisioners are not allowed in Terraform configurations.

## Requirements

- Disallow `local-exec` provisioners.
- Allow other provisioner types (e.g., `remote-exec`, `file`).
- Always warn even if the provisioner is not created.
- Ignore if no provisioner is set.

## Results

```console
$ tflint
1 issue(s) found:

Error: local-exec provisioner is not allowed (on aws_instance.example) (opa_deny_local_exec_provisioner)

  on main.tf line 2:
 107:   provisioner "local-exec" {

Reference: .tflint.d/policies/deny_local_exec_provisioner.rego:15

```
