package tflint

deny_import_blocks[issue] {
  imports := terraform.imports({}, {})
  count(imports) > 0

  issue := tflint.issue("import blocks are not allowed", imports[0].decl_range)
}
