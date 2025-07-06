resource "aws_instance" "valid" {
  ami = get_ami_id("service1", "v1")

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "aws_instance" "invalid" {
  ami = get_ami_id("service1", "v0.9")

  lifecycle {
    ignore_changes = [ami]
  }
}

module "valid" {
  providers = {
    aws = aws.usw1
  }
}

module "invalid" {
  providers = {
    aws = aws.usw2
  }
}
