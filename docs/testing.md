# Testing

You can write tests to continue to ensure that complex policies work as intended. See also [Policy Testing](https://www.openpolicyagent.org/docs/latest/policy-testing/).

Below is an example that tests a policy that enforces a bucket name:

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

```rego
package tflint

import rego.v1

failed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_s3_bucket" "invalid" {
  bucket = "example-corp-assets"
}`})

test_deny_invalid_s3_bucket_name_failed if {
	issues := deny_invalid_s3_bucket_name with terraform.resources as failed_resources

	count(issues) == 1
	issue := issues[_]
	issue.msg == `Bucket names should always start with "example-com-"`
}

passed_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_s3_bucket" "valid" {
  bucket = "example-com-assets"
}`})

test_deny_invalid_s3_bucket_name_passed if {
	issues := deny_invalid_s3_bucket_name with terraform.resources as passed_resources

	count(issues) == 0
}
```

Functions can be mocked with `terraform.mock_*` functions. Define a new function with the HCL file as the last argument and use `with` to replace the function.

You can run tests by setting `TFLINT_OPA_TEST=1`:

```console
# Passed
$ TFLINT_OPA_TEST=1 tflint
// No output

# Failed
$ TFLINT_OPA_TEST=1 tflint
1 issue(s) found:

Error: test failed (opa_test_deny_invalid_s3_bucket_name_failed)

  on  line 0:
   (source code not available)

Reference: .tflint.d/policies/bucket_test.rego:10

```
