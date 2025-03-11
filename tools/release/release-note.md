## What's Changed

In the OPA ruleset v0.8, we upgraded the embedded OPA version from v0.70 to v1.2. This means that some deprecated features will no longer be available and policies will need to be rewritten. See also https://www.openpolicyagent.org/docs/v1.2.0/v0-upgrade

If you use v0 syntax (without `if` and `contains` keywords in rule head declarations), it is recommended to use `opa fmt --write --v0-v1` to automatically rewrite your policy files. See also https://www.openpolicyagent.org/docs/v1.2.0/v0-upgrade/#upgrading-rego

Another new feature worth mentioning is support for [ephemeral resources](https://developer.hashicorp.com/terraform/language/resources/ephemeral), which was added in Terraform v1.10. You can get "ephemeral" blocks by using the `terraform.ephemeral_resources` function. Also, because `ephemeral` attribute has been added in an expression, you can write policies such as "passwords must be ephemeral".

### Breaking Changes
* Promote OPA 1.0 by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/136

### Enhancements
* Bump github.com/terraform-linters/tflint-plugin-sdk from 0.20.0 to 0.22.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/125
* Add support for ephemeral mark by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/133
* Add `terraform.ephemeral_resources` function by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/135

### Chores
* release: Introduce Artifact Attestations by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/106
* Bump goreleaser/goreleaser-action from 5 to 6 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/108
* Bump github.com/hashicorp/hcl/v2 from 2.20.1 to 2.21.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/109
* Bump github.com/open-policy-agent/opa from 0.64.1 to 0.65.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/107
* Bump github.com/open-policy-agent/opa from 0.65.0 to 0.66.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/110
* Bump github.com/open-policy-agent/opa from 0.66.0 to 0.69.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/118
* Bump github.com/open-policy-agent/opa from 0.69.0 to 0.70.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/119
* Bump github.com/hashicorp/hcl/v2 from 2.21.0 to 2.23.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/120
* Bump actions/attest-build-provenance from 1 to 2 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/122
* Bump github.com/zclconf/go-cty from 1.14.4 to 1.16.2 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/127
* deps: Go 1.24 by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/130
* Bump golang.org/x/net from 0.30.0 to 0.33.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/129
* Bump github.com/open-policy-agent/opa from 0.70.0 to 1.2.0 by @dependabot in https://github.com/terraform-linters/tflint-ruleset-opa/pull/131
* Enable Dependabot auto-merge by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/132
* Add make release for release automation by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/137
* Bump GoReleaser to v2 by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/138


**Full Changelog**: https://github.com/terraform-linters/tflint-ruleset-opa/compare/v0.7.0...v0.8.0
