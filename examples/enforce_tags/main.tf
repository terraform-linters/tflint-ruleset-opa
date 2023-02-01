resource "aws_instance" "invalid" {
  tags = {
    "production" = true
  }
}

resource "aws_instance" "valid" {
  tags = {
    "Environment" = "production"
  }
}

resource "aws_instance" "undefined" {
}

resource "aws_instance" "null" {
  tags = null
}

variable "unknown" {}

resource "aws_instance" "unknown" {
  tags = var.unknown
}
