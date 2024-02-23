# Enforce tags

This is an example policy that enforces that all instances are tagged with the "Environment" tag.

## Requirements

- Disallow AWS instances that are untagged with the "Environment" tag.
- Disallow unknown tags.
- Always warn even if the instance is not created.

## Results

```console
$ tflint
4 issue(s) found:

Error: instance must be tagged with the "Environment" tag (opa_deny_untagged_instance)

  on main.tf line 1:
   1: resource "aws_instance" "invalid" {

Reference: .tflint.d/policies/tags.rego:24

Error: instance must be tagged with the "Environment" tag (opa_deny_untagged_instance)

  on main.tf line 13:
  13: resource "aws_instance" "undefined" {

Reference: .tflint.d/policies/tags.rego:24

Error: instance must be tagged with the "Environment" tag (opa_deny_untagged_instance)

  on main.tf line 16:
  16: resource "aws_instance" "null" {

Reference: .tflint.d/policies/tags.rego:24

Error: Dynamic value is not allowed in tags (opa_deny_untagged_instance)

  on main.tf line 23:
  23:   tags = var.unknown

Reference: .tflint.d/policies/tags.rego:24

```
