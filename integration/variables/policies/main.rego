package tflint

deny_empty_description[issue] {
  vars := terraform.variables({"description": "string"}, {})
  description := vars[_].config.description
  
  description.value == ""

  issue := tflint.issue("empty description is not allowed", description.range)
}
