package tflint

import rego.v1

provisioners := terraform.resources(
	"*",
	{"provisioner": {"__labels": ["type"], "command": "string"}},
	{"expand_mode": "none"},
)

deny_local_exec_provisioner contains issue if {
	some resource in provisioners
	some provisioner in resource.config.provisioner
	provisioner.labels[0] == "local-exec"

	issue := tflint.issue(
		sprintf(`local-exec provisioner is not allowed (on %s.%s)`, [resource.type, resource.name]),
		provisioner.decl_range,
	)
}

