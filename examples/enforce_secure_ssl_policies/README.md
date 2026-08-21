# Enforce secure SSL policies

This is an example of enforcing that insecure SSL policies are not allowed on AWS Load Balancer listeners in Terraform configurations.

## Requirements

- Disallow insecure SSL policies (e.g., `ELBSecurityPolicy-2016-08`, `ELBSecurityPolicy-TLS-1-1-2017-01`, `ELBSecurityPolicy-TLS13-1-0-2021-06`, `ELBSecurityPolicy-TLS13-1-0-PQ-2025-09`, `ELBSecurityPolicy-TLS13-1-1-2021-06`).
- Allow secure SSL policies.
- Report issues only when an insecure SSL policy is explicitly set.
- Ignore when no SSL policy is configured or when the policy is unknown.

## Results

```console
$ tflint
1 issue(s) found:

Error: Insecure SSL policy detected: ELBSecurityPolicy-2016-08 (on aws_lb_listener.example) (opa_deny_insecure_ssl_policy)

  on main.tf line 2:
 107:   ssl_policy = "ELBSecurityPolicy-2016-08"

Reference: .tflint.d/policies/ssl_policies.rego:27

```
