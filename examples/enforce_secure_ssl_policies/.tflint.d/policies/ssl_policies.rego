package tflint

import rego.v1

aws_lb_listeners := terraform.resources(
	"aws_lb_listener",
	{"ssl_policy": "string"},
	{"expand_mode": "none"},
)

insecure_ssl_policies := {
	"ELBSecurityPolicy-2016-08",
	"ELBSecurityPolicy-TLS-1-1-2017-01",
	"ELBSecurityPolicy-TLS13-1-0-2021-06",
	"ELBSecurityPolicy-TLS13-1-0-PQ-2025-09",
	"ELBSecurityPolicy-TLS13-1-1-2021-06",
}

deny_insecure_ssl_policy contains issue if {
	listener := aws_lb_listeners[_]
	ssl_policy := listener.config.ssl_policy

	not ssl_policy.unknown
	is_string(ssl_policy.value)
	ssl_policy.value in insecure_ssl_policies

	issue := tflint.issue(
		sprintf(
			"Insecure SSL policy detected: %s",
			[ssl_policy.value],
		),
		ssl_policy.range,
	)
}

