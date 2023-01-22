resource "aws_instance" "invalid" {
  instance_type = "t1.micro"
}

resource "aws_instance" "valid" {
  instance_type = "t2.micro"
}

variable "variable" {
  default = "m5.large"
}
resource "aws_instance" "variable" {
  instance_type = var.variable
}

variable "unknown" {}
resource "aws_instance" "variable" {
  instance_type = var.unknown
}

variable "sensitive" {
  default   = "m5.large"
  sensitive = true
}
resource "aws_instance" "sensitive" {
  instance_type = var.sensitive
}
