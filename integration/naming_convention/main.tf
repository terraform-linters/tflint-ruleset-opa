resource "aws_instance" "valid_name" {
  instance_type = "t2.micro"
}

resource "aws_s3_bucket" "valid_name" {
  bucket = "foo"
}

resource "aws_instance" "invalid-name" {
  instance_type = "t2.micro"
}

resource "aws_s3_bucket" "invalid-name" {
  bucket = "foo"
}
