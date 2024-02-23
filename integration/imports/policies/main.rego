package tflint

import rego.v1

deny_import_blocks contains issue if {
	imports := terraform.imports({}, {})
	count(imports) > 0

	issue := tflint.issue("import blocks are not allowed", imports[0].decl_range)
}
