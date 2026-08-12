# Enforce bucket names

This is an example of applying naming conventions to top-level attributes.

## Requirements

- Disallow S3 bucket names starting with anything other than "example-com-".
- Disallow unknown bucket name.
- Always warn even if the bucket is not created.
- Ignore if bucket name is not set.

## Results

```console
$ tflint
4 issue(s) found:

Error: Bucket names should always start with "example-com-" (opa_deny_invalid_s3_bucket_name)

  on main.tf line 2:
   2:   bucket = "example-corp-assets"

Reference: .tflint.d/policies/bucket.rego:13

Error: Dynamic value is not allowed in bucket (opa_deny_invalid_s3_bucket_name)

  on main.tf line 12:
  12:   bucket = var.unknown

Reference: .tflint.d/policies/bucket.rego:13

Error: Bucket names should always start with "example-com-" (opa_deny_invalid_s3_bucket_name)

  on main.tf line 18:
  18:   bucket = "example-corp-assets"

Reference: .tflint.d/policies/bucket.rego:13

Error: Bucket names should always start with "example-com-" (opa_deny_invalid_s3_bucket_name)

  on main.tf line 24:
  24:   bucket = "example-corp-assets"

Reference: .tflint.d/policies/bucket.rego:13

```
