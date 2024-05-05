# TFLint Ruleset powered by Open Policy Agent (OPA)
[![Build Status](https://github.com/terraform-linters/tflint-ruleset-opa/workflows/build/badge.svg?branch=main)](https://github.com/terraform-linters/tflint-ruleset-opa/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint-ruleset-opa.svg)](https://github.com/terraform-linters/tflint-ruleset-opa/releases/latest)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

TFLint ruleset plugin for writing custom rules in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/).

NOTE: This plugin is experimental. This means frequent breaking changes.

## Requirements

- TFLint v0.43+
- Go v1.22

## Installation

You can install the plugin by adding a config to `.tflint.hcl` and running `tflint --init`:

```hcl
plugin "opa" {
  enabled = true
  version = "0.7.0"
  source  = "github.com/terraform-linters/tflint-ruleset-opa"
}
```

Policy files are placed under `~/.tflint.d/policies` or `./.tflint.d/policies`. First create a directory:

```console
$ mkdir -p .tflint.d/policies
```

For more configuration about the plugin, see [Plugin Configuration](./docs/configuration.md).

## Getting Started

TFLint plugin system allows you to add custom rules, but plugins can be a pain to maintain when applying a few simple organization policies. This ruleset plugin provides the ability to write policies in Rego, instead of building plugins in Go.

For example, your organization wants to enforce S3 bucket names to always start with `example-com-*`. You can write the following policy as `./.tflint.d/policies/bucket.rego`:

```rego
package tflint

import rego.v1

deny_invalid_s3_bucket_name contains issue if {
	buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {})
	name := buckets[_].config.bucket
	not startswith(name.value, "example-com-")

	issue := tflint.issue(`Bucket names should always start with "example-com-"`, name.range)
}
```

This allows you to issue errors for Terraform configs such as:

```hcl
resource "aws_s3_bucket" "invalid" {
  bucket = "example-corp-assets"
}

resource "aws_s3_bucket" "valid" {
  bucket = "example-com-assets"
}
```

```console
$ tflint
1 issue(s) found:

Error: Bucket names should always start with "example-com-" (opa_deny_invalid_s3_bucket_name)

  on main.tf line 2:
   2:   bucket = "example-corp-assets"

Reference: .tflint.d/policies/bucket.rego:5

```

See [the documentation](./docs/) and [examples](./examples/) for details.

NOTE: This policy cannot be enforced in all cases. See [Handling unknown/null/undefined values](./docs/handling_special_values.md) for details.

## OPA Ruleset vs. Custom Ruleset

There are two options for providing custom rules: the OPA ruleset and a custom ruleset. Which is better?

If you want to enforce a small number of rules for a small team, or if your don't have a dedicated team to maintain your plugin, starting with the OPA ruleset is probably a good option.

On the other hand, building and maintaining a custom ruleset plugin is better when enforcing many complex rules or distributing to large teams.

## Building the plugin

Clone the repository locally and run the following command:

```
$ make
```

You can easily install the built plugin with the following:

```
$ make install
```

You can run the built plugin like the following:

```
$ cat << EOS > .tflint.hcl
plugin "opa" {
  enabled = true
}
EOS
$ tflint
```
