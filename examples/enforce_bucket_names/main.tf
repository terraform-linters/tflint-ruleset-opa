resource "aws_s3_bucket" "invalid" {
  bucket = "example-corp-assets"
}

resource "aws_s3_bucket" "valid" {
  bucket = "example-com-assets"
}

variable "unknown" {}

resource "aws_s3_bucket" "unknown_value" {
  bucket = var.unknown
}

resource "aws_s3_bucket" "unknown_count" {
  count = var.unknown

  bucket = "example-corp-assets"
}

resource "aws_s3_bucket" "unknown_for_each" {
  for_each = var.unknown

  bucket = "example-corp-assets"
}
