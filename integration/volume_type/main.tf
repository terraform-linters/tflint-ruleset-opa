resource "aws_instance" "valid" {
  ebs_block_device {
    volume_type = "gp2"
  }
}

resource "aws_instance" "invalid" {
  ebs_block_device {
    volume_type = "gp3"
  }
}

resource "aws_instance" "dynamic_invalid" {
  dynamic "ebs_block_device" {
    for_each = ["gp3"]

    content {
      volume_type = ebs_block_device.value
    }
  }
}

resource "aws_instance" "not_created" {
  count = 0

  ebs_block_device {
    volume_type = "gp3"
  }
}

variable "unknown" {}

resource "aws_instance" "dynamic_unknown" {
  dynamic "ebs_block_device" {
    for_each = var.unknown

    content {
      volume_type = ebs_block_device.value
    }
  }
}

resource "aws_instance" "unknown" {
  count = var.unknown

  ebs_block_device {
    volume_type = "gp3"
  }
}
