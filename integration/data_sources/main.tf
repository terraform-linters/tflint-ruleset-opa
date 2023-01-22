data "aws_ami" "valid" {
  owners = ["self"]
}

data "aws_ami" "invalid" {
  owners = ["amazon"]
}
