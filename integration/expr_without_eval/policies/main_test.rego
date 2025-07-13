package tflint

import rego.v1

invalid_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {
  "main.tf": `
resource "aws_instance" "invalid" {
  ami = get_ami_id("service1", "v0.9")

  lifecycle {
    ignore_changes = [ami]
  }
}
`,
})

valid_resources(type, schema, options) := terraform.mock_resources(type, schema, options, {
  "main.tf": `
resource "aws_instance" "valid" {
  ami = get_ami_id("service1", "v1")

  lifecycle {
    ignore_changes = [tags]
  }
}
`,
})

test_wrong_ignore_passed if {
  issues := deny_require_lifecycle_ignore_tags with terraform.resources as invalid_resources

  count(issues) == 1
  issues[_].msg == "Resource aws_instance.invalid must have lifecycle.ignore_changes include \"tags\""
}

test_wrong_ignore_failed if {
  issues := deny_require_lifecycle_ignore_tags with terraform.resources as invalid_resources

  count(issues) == 0
}

test_correct_ignore_passed if {
  issues := deny_require_lifecycle_ignore_tags with terraform.resources as valid_resources

  count(issues) == 0
}

test_correct_ignore_failed if {
  issues := deny_require_lifecycle_ignore_tags with terraform.resources as valid_resources

  count(issues) == 1
}

test_wrong_ami_passed if {
  issues := deny_deprecated_ami_version with terraform.resources as invalid_resources

  count(issues) == 1
  issues[_].msg == "get_ami_id function should be called with v1"
}

test_wrong_ami_failed if {
  issues := deny_deprecated_ami_version with terraform.resources as invalid_resources

  count(issues) == 0
}

test_correct_ami_passed if {
  issues := deny_deprecated_ami_version with terraform.resources as valid_resources

  count(issues) == 0
}

test_correct_ami_failed if {
  issues := deny_deprecated_ami_version with terraform.resources as valid_resources

  count(issues) == 1
}

invalid_calls(schema, options) := terraform.mock_module_calls(schema, options, {
  "main.tf": `
module "invalid" {
  providers = {
    aws = aws.usw2
  }
}
`,
})

valid_calls(schema, options) := terraform.mock_module_calls(schema, options, {
  "main.tf": `
module "valid" {
  providers = {
    aws = aws.usw1
  }
}
`,
})

test_wrong_region_passed if {
  issues := deny_deprecated_regions with terraform.module_calls as invalid_calls

  count(issues) == 1
  issues[_].msg == "deprecated region reference found"
}

test_wrong_region_failed if {
  issues := deny_deprecated_regions with terraform.module_calls as invalid_calls

  count(issues) == 0
}

test_correct_region_passed if {
  issues := deny_deprecated_regions with terraform.module_calls as valid_calls

  count(issues) == 0
}

test_correct_region_failed if {
  issues := deny_deprecated_regions with terraform.module_calls as valid_calls

  count(issues) == 1
}
