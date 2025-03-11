package tflint
import future.keywords

mock_ephemeral_resources(type, schema, options) := terraform.mock_ephemeral_resources(type, schema, options, {"main.tf": `
ephemeral "random_password" "db_password" {
  length           = 16
  override_special = "!#$%&*()-_=+[]{}<>:?"
}`})

test_deny_weak_password_passed if {
  issues := deny_weak_password with terraform.ephemeral_resources as mock_ephemeral_resources

  count(issues) == 1
  issue := issues[_]
  issue.msg == "Password must be at least 32 characters long"
}

test_deny_weak_password_failed if {
  issues := deny_weak_password with terraform.ephemeral_resources as mock_ephemeral_resources

  count(issues) == 0
}
