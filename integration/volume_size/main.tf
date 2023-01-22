resource "aws_instance" "valid" {
  ebs_block_device {
    volume_size = 20
  }
}

resource "aws_instance" "invalid" {
  ebs_block_device {
    volume_size = 50
  }
}

resource "aws_instance" "valid_string" {
  ebs_block_device {
    volume_size = "20"
  }
}

resource "aws_instance" "invalid_string" {
  ebs_block_device {
    volume_size = "50"
  }
}

resource "aws_instance" "valid_float" {
  ebs_block_device {
    volume_size = 20.5
  }
}

resource "aws_instance" "invalid_float" {
  ebs_block_device {
    volume_size = 30.5
  }
}
