# TFLint OPA Ruleset vs. OPA/Conftest/Sentinel

Besides this ruleset, there are other solutions for Policy as Code. This document compares them and provides information to help you decide which solution to adopt.

## TFLint OPA Ruleset vs. OPA

OPA officially publishes [an example of applying a policy to Terraform](https://www.openpolicyagent.org/docs/latest/terraform/).

This way is reliable and stable as it depends only on Terraform's plan file structure and OPA. If you are already satisfied with this way, there may be little benefit to adopting the TFLint OPA ruleset.

On the other hand, the advantage of the TFLint OPA ruleset is that you don't need to run `terraform plan` to apply policies. So you can write policies against all files, not just diffs, and quickly check if your code violates the policies.

## TFLint OPA Ruleset vs. Conftest

[Conftest](https://www.conftest.dev/) is a popular solution developed under the Open Policy Agent organization.

Conftest has native support for HCL and supports many other formats such as Dockerfile and YAML. This is a great option if you want to enforce policies with the same tool for many other configs.

However, Conftest does not support semantics such as variables in HCL. If you want to write a policy against an evaluated configuration, you need to write the policy against a plan file.

## TFLint OPA Ruleset vs. Sentinel

[Sentinel](https://www.hashicorp.com/sentinel) is a Policy as Code solution developed by HashiCorp.

Sentinel is a commercial product and works seamlessly with Terraform and other HashiCorp products. If policy enforcement is important to your organization, Sentinel may be a good option.

On the other hand, TFLint OPA Ruleset is a good option to start if policy enforcement is not yet important or if commercial product is not available.
