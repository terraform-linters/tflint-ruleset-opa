package tflint

import rego.v1

deny_remote_source contains issue if {
	modules := terraform.module_calls({"source": "string"}, {})
	source := modules[_].config.source

	not startswith(source.value, "./")

	issue := tflint.issue("remote module is not allowed", source.range)
}
