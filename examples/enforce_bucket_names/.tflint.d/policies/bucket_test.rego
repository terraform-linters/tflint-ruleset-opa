package tflint
import future.keywords.if

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

unknown_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_s3_bucket" "invalid" {
  bucket = var.unknown
}`})
test_deny_invalid_s3_bucket_name_unknown if {
  issues := deny_invalid_s3_bucket_name with terraform.resources as unknown_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == "Dynamic value is not allowed in bucket"
}

unknown_count_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_s3_bucket" "invalid" {
  count  = var.unknown
  bucket = "example-corp-assets"
}`})
test_deny_invalid_s3_bucket_name_unknown_count if {
  issues := deny_invalid_s3_bucket_name with terraform.resources as failed_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == `Bucket names should always start with "example-com-"`
}

unknown_for_each_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_s3_bucket" "invalid" {
  for_each = var.unknown
  bucket   = "example-corp-assets"
}`})
test_deny_invalid_s3_bucket_name_unknown_for_each if {
  issues := deny_invalid_s3_bucket_name with terraform.resources as failed_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == `Bucket names should always start with "example-com-"`
}
