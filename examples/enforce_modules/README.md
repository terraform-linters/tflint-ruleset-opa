# Enforce modules

This is an example of a policy that disallows declaring resources directly and enforces you to always use modules instead.

## Requirements

- Disallow all `resource` declarations.

## Results

```console
$ tflint
1 issue(s) found:

Error: Declaring resources is not allowed. Use modules instead. (opa_deny_resource_declarations)

  on main.tf line 1:
   1: resource "aws_instance" "main" {

Reference: .tflint.d/policies/module.rego:5

```
