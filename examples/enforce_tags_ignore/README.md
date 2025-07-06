# Enforce tags ignore

This is an example of a policy that disallows declaring AWS instances without `ignore_changes = [tags]`.

## Requirements

- Disallow AWS instances that do not ignore `tags`.
- Allow `ignore_changes = all`.
- Allow deprecated `ignore_changes` syntax (e.g. `"tags"`, `"all"`).
- Always warn even if the instance is not created.

## Results

```console
$ tflint
3 issue(s) found:

Error: instance must have "ignore_changes = [tags]" (opa_deny_instance_without_tags_ignore)

  on main.tf line 7:
   7: resource "aws_instance" "invalid" {

Reference: .tflint.d/policies/tags.rego:37

Error: instance must have "ignore_changes = [tags]" (opa_deny_instance_without_tags_ignore)

  on main.tf line 13:
  13: resource "aws_instance" "without_ignore_changes" {

Reference: .tflint.d/policies/tags.rego:37

Error: instance must have "ignore_changes = [tags]" (opa_deny_instance_without_tags_ignore)

  on main.tf line 19:
  19: resource "aws_instance" "without_lifecycle" {

Reference: .tflint.d/policies/tags.rego:37

```
