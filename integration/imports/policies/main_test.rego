package tflint

import rego.v1

mock_imports(schema, options) := terraform.mock_imports(schema, options, {"main.tf": `
import {
  to = aws_instance.example
  id = "i-abcd1234"
}`})

test_deny_import_blocks_passed if {
	issues := deny_import_blocks with terraform.imports as mock_imports

	count(issues) == 1
	issue := issues[_]
	issue.msg == "import blocks are not allowed"
}

test_deny_import_blocks_failed if {
	issues := deny_import_blocks with terraform.imports as mock_imports

	count(issues) == 0
}
