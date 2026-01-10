## What's Changed

Support for Cosign signatures has been removed from this release. The `checksums.txt.keyless.sig` and `checksums.txt.pem` will not be included in the release.
These files are not used in normal use cases, so in most cases this will not affect you, but if you are affected, you can use Artifact Attestations instead.

### Breaking Changes
* Bump github.com/terraform-linters/tflint-plugin-sdk from 0.22.0 to 0.23.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/185
  * Requires TFLint v0.46+

### Enhancements
* Bump github.com/open-policy-agent/opa from 1.6.0 to 1.7.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/172
* Bump github.com/open-policy-agent/opa from 1.7.1 to 1.8.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/179
* Bump github.com/open-policy-agent/opa from 1.8.0 to 1.9.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/182
* Bump github.com/open-policy-agent/opa from 1.9.0 to 1.10.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/186
* Bump github.com/open-policy-agent/opa from 1.10.0 to 1.10.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/188
* Bump github.com/open-policy-agent/opa from 1.10.1 to 1.11.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/193
* Bump github.com/open-policy-agent/opa from 1.11.0 to 1.12.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/196
* funcs: Add `terraform.actions` function by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/198

### Chores
* Extract the functions into the `funcs` package by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/168
* Bump actions/checkout from 4.2.2 to 5.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/173
* Bump goreleaser/goreleaser-action from 6.3.0 to 6.4.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/174
* Bump github.com/zclconf/go-cty from 1.16.3 to 1.16.4 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/175
* dependabot: allow actions writes by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/176
* Bump actions/attest-build-provenance from 2.4.0 to 3.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/178
* Bump github.com/zclconf/go-cty from 1.16.4 to 1.17.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/180
* Bump actions/setup-go from 5.5.0 to 6.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/177
* Bump sigstore/cosign-installer from 3.9.2 to 3.10.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/181
* Bump sigstore/cosign-installer from 3.10.0 to 4.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/184
* Bump golang.org/x/crypto from 0.42.0 to 0.45.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/190
* Bump actions/checkout from 5.0.0 to 6.0.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/192
* Bump actions/setup-go from 6.0.0 to 6.1.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/191
* Bump actions/checkout from 6.0.0 to 6.0.1 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/194
* Bump actions/attest-build-provenance from 3.0.0 to 3.1.0 by @dependabot[bot] in https://github.com/terraform-linters/tflint-ruleset-opa/pull/195
* deps: Bump Go version to 1.25.5 by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/197
* Drop support for Cosign signatures by @wata727 in https://github.com/terraform-linters/tflint-ruleset-opa/pull/199


**Full Changelog**: https://github.com/terraform-linters/tflint-ruleset-opa/compare/v0.9.0...v0.10.0
