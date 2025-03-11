# Introduction

This plugin is backed by [Open Policy Agent (OPA)](https://www.openpolicyagent.org/docs/latest/) and allows you to write custom rules for TFLint in the policy language (Rego). This document will guide you through the step-by-step process of getting started writing policies.

First, refer to the official documentation for [OPA concepts](https://www.openpolicyagent.org/docs/latest/) and [Policy Language](https://www.openpolicyagent.org/docs/latest/policy-language/). The documentation that follows assumes familiarity with these concepts.

As an example, create the following policy as `.tflint.d/policies/bucket.rego`:

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

Suppose you apply a policy to the following files:

```hcl
resource "aws_s3_bucket" "invalid" {
  bucket = "example-corp-assets"
}

resource "aws_s3_bucket" "valid" {
  bucket = "example-com-assets"
}
```

Let's go through it line by line.

```rego
package tflint
```

The first line is the package declaration. All valid policies must be described under the `tflint` package.

```rego
import rego.v1
```

This declaration ensures compatibility with future OPA v1 syntax. See [The `rego.v1` Import](https://www.openpolicyagent.org/docs/latest/policy-language/#the-regov1-import) for details.

```rego
deny_invalid_s3_bucket_name contains issue if {
```

The next line is the rule declaration. A valid rule name must start with `deny_`, `violation_`, `warn_` or `notice_`. The rule name in TFLint is the rule name with "opa_" prefix (e.g. `opa_deny_invalid_s3_bucket_name`), and the severity is error for `deny_` or `violation_`, warning for `warn_`, and notice for `notice_`.

The rule should return a set of issue objects, not a boolean. An issue is created on the last line when all conditions are met.

```rego
buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {})
```

The next line is to retrieve `aws_s3_bucket` resources. The policy language written in this plugin primarily uses JSON retrieved by custom functions rather than input data. The `terraform.resources` is a custom function to retrieve `resource` blocks in Terraform configs. For more custom functions, see [Functions](./functions.md).

Note that the schema must be declared when referencing inside a resource block. The example above declares that the `bucket` attribute exists as a string. See [Terraform Schema](./schema.md) for details.

The return value of this function will be the following JSON:

```json
[
  {
    "type": "aws_s3_bucket",
    "name": "invalid",
    "config": {
      "bucket": {
        "value": "example-corp-assets",
        "unknown": false,
        "sensitive": false,
        "ephemeral": false,
        "range": {
          "filename": "main.tf",
          "start": { "line": 2, "column": 12, "byte": 48 },
          "end": { "line": 2, "column": 33, "byte": 69 }
        }
      }
    },
    "decl_range": {
      "filename": "main.tf",
      "start": { "line": 1, "column": 1, "byte": 0 },
      "end": { "line": 1, "column": 35, "byte": 34 }
    }
  },
  {
    "type": "aws_s3_bucket",
    "name": "valid",
    "config": {
      "bucket": {
        "value": "example-com-assets",
        "unknown": false,
        "sensitive": false,
        "ephemeral": false,
        "range": {
          "filename": "main.tf",
          "start": { "line": 6, "column": 12, "byte": 119 },
          "end": { "line": 6, "column": 32, "byte": 139 }
        }
      }
    },
    "decl_range": {
      "filename": "main.tf",
      "start": { "line": 5, "column": 1, "byte": 73 },
      "end": { "line": 5, "column": 33, "byte": 105 }
    }
  }
]
```

Attributes set in the schema are included under the `config` if they actually exist.

```rego
name := buckets[_].config.bucket
not startswith(name.value, "example-com-")
```

The next line is to get the `bucket` attributes. Note that the value is `bucket.value` and `bucket` is an object.

```rego
issue := tflint.issue(`Bucket names should always start with "example-com-"`, name.range)
```

The last line is to generate an issue. Use the `tflint.issue` function to specify the message and range.

This allows you to raise an issue for config that violates your policy:

```console
$ tflint
1 issue(s) found:

Error: Bucket names should always start with "example-com-" (opa_deny_invalid_s3_bucket_name)

  on main.tf line 2:
   2:   bucket = "example-corp-assets"

Reference: .tflint.d/policies/main.rego:5

```

Note that this policy cannot enforce policy in all cases. See [Handling unknown/null/undefined values](./handling_special_values.md) for details.
