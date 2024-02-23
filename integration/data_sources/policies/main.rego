package tflint

import rego.v1

deny_other_ami_owners contains issue if {
	sources := terraform.data_sources("aws_ami", {"owners": "list(string)"}, {})
	owners := sources[_].config.owners

	owners.value[_] != "self"

	issue := tflint.issue("third-party AMI is not allowed", owners.range)
}
