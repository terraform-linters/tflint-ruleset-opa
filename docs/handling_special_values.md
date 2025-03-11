# Handling unknown/null/undefined values

There are three special values to be aware of when writing policies. Unknown value, null, undefined value. This document describes when these can occur and how the policy should handle them.

## Unknown values

Not all values can be determined statically in Terraform. Imagine a config like the one below where you have variables that are not given actual values in the CI:

```hcl
# This value is provided with `TF_VAR_bucket_name=[NAME] terraform apply`.
variable "bucket_name" {
  type = string
}

resource "aws_s3_bucket" "unknown" {
  bucket = var.bucket_name # => unknown value
}
```

Ideally, you should also set `TF_VAR_bucket_name` in CI, but if it's not available, you need to consider what to do with these unknown values.

Cases that return unknown values are:

- Variables without values
- Variables marked with `sensitive = true` or `ephemeral = true`
- Resource attributes (e.g. `aws_instance.web.arn`)
- Data attributes (e.g. `data.aws_ami.web.id`)
- Module outputs (e.g. `module.vpc.vpc_id`)
- `self`
- Local values that resolves to unknown values

In this case the returned JSON looks like this:

```json
[
  {
    "type": "aws_s3_bucket",
    "name": "unknown",
    "config": {
      "bucket": {
        "unknown": true,
        "sensitive": false,
        "ephemeral": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

Notice that `value` does not exist and `unknown` is true. Neither of the following policies are violated for unknown values, because OPA halts evaluating when it hits an undefined value.

```rego
package tflint

import rego.v1

bucket_names contains name if {
	buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {})
	name := buckets[_].config.bucket
}

deny_invalid_s3_bucket_name contains issue if {
	not startswith(bucket_names[i].value, "example-com-")

	issue := tflint.issue(`Bucket names should always start with "example-com-"`, bucket_names[i].range)
}

deny_valid_s3_bucket_name contains issue if {
	startswith(bucket_names[i].value, "example-com-")

	issue := tflint.issue(`Bucket names should not always start with "example-com-"`, bucket_names[i].range)
}
```

This behavior is useful for detecting erroneous values, but inconvenient if you want to ensure policy enforcement. In such cases, you can add a policy to warn for unknown values:

```rego
package tflint

import rego.v1

bucket_names contains name if {
	buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {})
	name := buckets[_].config.bucket
}

deny_invalid_s3_bucket_name contains issue if {
	bucket_names[i].unknown

	issue := tflint.issue(`Dynamic value is not allowed in bucket name`, bucket_names[i].range)
}

deny_invalid_s3_bucket_name contains issue if {
	not startswith(bucket_names[i].value, "example-com-")

	issue := tflint.issue(`Bucket names should always start with "example-com-"`, bucket_names[i].range)
}
```

```console
$ tflint
1 issue(s) found:

Error: Dynamic value is not allowed in bucket name (opa_deny_invalid_s3_bucket_name)

  on main.tf line 7:
   4:   bucket = var.bucket_name # => unknown value

Reference: .tflint.d/policies/main.rego:10

```

### Unknown values in meta-arguments

Another example where the policy may not apply is when meta-arguments are unknown. Imagine a config like this:

```hcl
# This value is provided with `TF_VAR_bucket_count=[COUNT] terraform apply`.
variable "bucket_count" {
  type = number
}

resource "aws_s3_bucket" "unknown" {
  count = var.bucket_count # => unknown value

  bucket = "example-org-${count.index}"
}
```

In this case, the bucket may or may not be created, so TFLint conservatively treats it as never created. In other words, `terraform.resources` returns an empty array, so even if the bucket name violates the policy, it will not be detected.

To find this out, add a policy like the following:

```rego
package tflint

import rego.v1

deny_invalid_s3_bucket_name contains issue if {
	buckets := terraform.resources("aws_s3_bucket", {"count": "number"}, {"expand_mode": "none"})
	count := buckets[_].config.count
	count.unknown

	issue := tflint.issue(`Dynamic value is not allowed in count`, count.range)
}

deny_invalid_s3_bucket_name contains issue if {
	buckets := terraform.resources("aws_s3_bucket", {"for_each": "any"}, {"expand_mode": "none"})
	for_each := buckets[_].config.for_each
	for_each.unknown

	issue := tflint.issue(`Dynamic value is not allowed in for_each`, for_each.range)
}

deny_invalid_s3_bucket_name contains issue if {
	buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {})
	bucket := buckets[_].config.bucket
	not startswith(bucket.value, "example-com-")

	issue := tflint.issue(`Bucket names should always start with "example-com-"`, bucket.range)
}
```

```console
$ tflint
1 issue(s) found:

Error: Dynamic value is not allowed in count (opa_deny_invalid_s3_bucket_name)

  on main.tf line 7:
   7:   count = var.bucket_count # => unknown value

Reference: .tflint.d/policies/main.rego:5

```

Note that you should set `{"expaned_mode": "none"}` when retrieving meta-arguments. If you don't set it, you can't retrieve a bucket, so you can't reference unknown meta-arguments.

### Unknown values in dynamic blocks

Similarly, you should also be careful with unknown values in dynamic blocks. Imagine a config like this:

```hcl
variable "block_devices" {}

resource "aws_instance" "unknown" {
  dynamic "ebs_block_device" {
    for_each = var.block_devices # => unknown value

    content {
      volume_size = 50
    }
  }
}
```

Even in the above case, it is unknown how many dynamic blocks will be expanded, so it is conservatively determined that they will not be expanded.

To find this out, add a policy like the following:

```rego
package tflint

import rego.v1

deny_large_volume contains issue if {
	instances := terraform.resources("aws_instance", {"dynamic": {"__labels": ["type"], "for_each": "any"}}, {"expand_mode": "none"})
	for_each := instances[_].config.dynamic[_].config.for_each
	for_each.unknown

	issue := tflint.issue("Dynamic value is not allowed in for_each", for_each.range)
}

deny_large_volume contains issue if {
	instances := terraform.resources("aws_instance", {"ebs_block_device": {"volume_size": "number"}}, {})
	size := instances[_].config.ebs_block_device[_].config.volume_size
	size.value > 30

	issue := tflint.issue("Volume size must be 30GB or less", size.range)
}
```

```console
$ tflint
1 issue(s) found:

Error: Dynamic value is not allowed in for_each (opa_deny_large_volume)

  on main.tf line 5:
   5:     for_each = var.block_devices # => unknown value

Reference: .tflint.d/policies/main.rego:5

```

## Null values

Note that in Terraform all values can be null. Terraform treats null as not set. For example, the following config is the same as when `tags` is not set:

```hcl
resource "aws_instance" "main" {
  tags = null
}
```

In this case the returned JSON looks like this:

```json
[
  {
    "type": "aws_instance",
    "name": "main",
    "config": {
      "tags": {
        "value": null,
        "unknown": false,
        "sensitive": false,
        "ephemeral": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

Imagine a policy that detects resources that don't have a tag like this:

```rego
package tflint

import rego.v1

deny_not_tagged_instance contains issue if {
	resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
	resource := resources[_]
	not "Environment" in object.keys(resource.config.tags.value)

	issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}
```

This works as expected for resources that have `tags` defined:

```hcl
resource "aws_instance" "main" {
  tags = {}
}
```

```console
$ tflint
1 issue(s) found:

Error: instance should be tagged with Environment (opa_deny_not_tagged_instance)

  on main.tf line 1:
   1: resource "aws_instance" "main" {

Reference: .tflint.d/policies/main.rego:5

```

But it doesn't work for null:

```console
$ tflint
Failed to check ruleset; Failed to check `opa_deny_not_tagged_instance` rule: .tflint.d/policies/main.rego:8: eval_type_error: object.keys: operand 1 must be object but got null
```

Notice that `object.keys` returns an error in the example above, but it may be ignored. To find this out, fix the policy like the following:

```rego
package tflint

import rego.v1

is_not_tagged(tags) if {
	is_null(tags)
}

is_not_tagged(tags) if {
	not is_null(tags)
	not "Environment" in object.keys(tags)
}

deny_not_tagged_instance contains issue if {
	resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
	resource := resources[_]
	is_not_tagged(resource.config.tags.value)

	issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}
```

## Undefined values

As with the above example, you also need to consider the case where `tags` is undefined. Imagine a config like this:

```hcl
resource "aws_instance" "main" {
}
```

In this case the returned JSON looks like this:

```json
[
  {
    "type": "aws_instance",
    "name": "main",
    "config": {},
    "decl_range": {...}
  }
]
```

An empty `config` makes `resource.config.tags` undefined and halts policy evaluation, so it cannot be detected by the above policy.

To find this out, fix the policy like the following:

```rego
package tflint

import rego.v1

is_not_tagged(config) if {
	is_null(config.tags.value)
}

is_not_tagged(config) if {
	not is_null(config.tags.value)
	not "Environment" in object.keys(config.tags.value)
}

is_not_tagged(config) if {
	not "tags" in object.keys(config)
}

deny_not_tagged_instance contains issue if {
	resources := terraform.resources("aws_instance", {"tags": "map(string)"}, {})
	resource := resources[_]
	is_not_tagged(resource.config)

	issue := tflint.issue("instance should be tagged with Environment", resource.decl_range)
}
```
