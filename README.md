# TFLint Ruleset powered by Open Policy Agent (OPA)
[![Build Status](https://github.com/terraform-linters/tflint-ruleset-opa/workflows/build/badge.svg?branch=main)](https://github.com/terraform-linters/tflint-ruleset-opa/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint-ruleset-opa.svg)](https://github.com/terraform-linters/tflint-ruleset-opa/releases/latest)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)

TFLint ruleset plugin for writing custom rules in [Rego](https://www.openpolicyagent.org/docs/latest/policy-language/).

NOTE: This plugin is working in progress. Not intended for general use.

## Requirements

- TFLint v0.40+
- Go v1.19

## Installation

This plugin is working in progress. There is no way to install it.

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
