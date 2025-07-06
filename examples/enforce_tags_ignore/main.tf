resource "aws_instance" "valid" {
  lifecycle {
    ignore_changes = [key_name, tags]
  }
}

resource "aws_instance" "invalid" {
  lifecycle {
    ignore_changes = [key_name, ami]
  }
}

resource "aws_instance" "without_ignore_changes" {
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_instance" "without_lifecycle" {
  instance_type = "t2.micro"
}

resource "aws_instance" "all" {
  lifecycle {
    ignore_changes = all
  }
}

resource "aws_instance" "deprecated_ignore" {
  lifecycle {
    ignore_changes = ["tags"]
  }
}

resource "aws_instance" "deprecated_all" {
  lifecycle {
    ignore_changes = "all"
  }
}
