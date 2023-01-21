package tflint
import future.keywords

mock_resources_gp3(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
resource "aws_instance" "main" {
  ebs_block_device {
    volume_type = "gp3"
  }
}`})

test_warn_gp3_volume_passed if {
  issues := warn_gp3_volume with terraform.resources as mock_resources_gp3

  count(issues) == 1
  issue := issues[_]
  issue.msg == "gp3 is not allowed"
}

test_warn_gp3_volume_failed if {
  issues := warn_gp3_volume with terraform.resources as mock_resources_gp3

  count(issues) == 0
}

mock_resources_unknown_dynamic(type, schema, options) := terraform.mock_resources(type, schema, options, {"main.tf": `
variable "unknown" {}

resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = var.unknown

    content {
      volume_type = ebs_block_device.value
    }
  }
}`})

test_warn_gp3_volume_unknown_dynamic_passed if {
  issues := warn_gp3_volume with terraform.resources as mock_resources_unknown_dynamic

  count(issues) == 1
  issue := issues[_]
  issue.msg == "unknown block found"
}

test_warn_gp3_volume_unknown_dynamic_failed if {
  issues := warn_gp3_volume with terraform.resources as mock_resources_unknown_dynamic

  count(issues) == 0
}
