provider "aws" {
  alias  = "east"
  region = "us-east-1"
}

resource "aws_instance" "main" {
  provider = aws.east
}
