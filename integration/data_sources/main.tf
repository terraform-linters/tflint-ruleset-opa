data "aws_ami" "valid" {
  owners = ["self"]
}

data "aws_ami" "invalid" {
  owners = ["amazon"]
}

check "scoped" {
  data "aws_ami" "scoped_valid" {
    owners = ["self"]
  }

  data "aws_ami" "scoped_invalid" {
    owners = ["amazon"]
  }
}
