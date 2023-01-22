resource "aws_instance" "valid" {
  tags = {
    "Environment" = "production"
  }
}

resource "aws_instance" "invalid" {
  tags = {
    "production" = true
  }
}

resource "aws_instance" "not_tagged" {
  instance_type = "t2.micro"
}

resource "aws_instance" "null" {
  tags = null
}
