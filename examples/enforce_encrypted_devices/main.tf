resource "aws_instance" "invalid" {
  ebs_block_device {
    encrypted = false
  }
}

resource "aws_instance" "valid" {
  ebs_block_device {
    encrypted = true
  }
}

resource "aws_instance" "default" {
  ebs_block_device {
  }
}

resource "aws_instance" "null" {
  ebs_block_device {
    encrypted = null
  }
}

resource "aws_instance" "dynamic" {
  dynamic "ebs_block_device" {
    for_each = toset([false])

    content {
      encrypted = ebs_block_device.value
    }
  }
}

variable "unknown" {}

resource "aws_instance" "unknown" {
  ebs_block_device {
    encrypted = var.unknown
  }
}

resource "aws_instance" "unknown_count" {
  count = var.unknown

  ebs_block_device {
    encrypted = false
  }
}

resource "aws_instance" "unknown_for_each" {
  for_each = var.unknown

  ebs_block_device {
    encrypted = false
  }
}

resource "aws_instance" "unknown_dynamic" {
  dynamic "ebs_block_device" {
    for_each = var.unknown

    content {
      encrypted = false
    }
  }
}
