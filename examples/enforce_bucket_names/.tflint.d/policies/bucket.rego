package tflint
import future.keywords.contains
import future.keywords.if

# Set `expand_mode: none` to check names even if they are not created
s3_buckets := terraform.resources("aws_s3_bucket", {"bucket": "string"}, {"expand_mode": "none"})

s3_bucket_names contains name if {
  name := s3_buckets[_].config.bucket
}

# Rules for unknown values
deny_invalid_s3_bucket_name[issue] {
  s3_bucket_names[i].unknown

  issue := tflint.issue("Dynamic value is not allowed in bucket", s3_bucket_names[i].range)
}
# Rules for invalid names
deny_invalid_s3_bucket_name[issue] {
  not startswith(s3_bucket_names[i].value, "example-com-")

  issue := tflint.issue(`Bucket names should always start with "example-com-"`, s3_bucket_names[i].range)
}
