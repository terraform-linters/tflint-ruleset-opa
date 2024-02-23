## 0.6.0 (2024-02-23)

### Enhancements

- [#83](https://github.com/terraform-linters/tflint-ruleset-opa/pull/83): Bump github.com/open-policy-agent/opa from 0.60.0 to 0.61.0
- [#87](https://github.com/terraform-linters/tflint-ruleset-opa/pull/87): Add `terraform.removed_blocks` function

### Chores

- [#82](https://github.com/terraform-linters/tflint-ruleset-opa/pull/82): Bump github.com/zclconf/go-cty from 1.14.1 to 1.14.2
- [#86](https://github.com/terraform-linters/tflint-ruleset-opa/pull/86): deps: Go 1.22
- [#88](https://github.com/terraform-linters/tflint-ruleset-opa/pull/88): Rewrite policies with `import rego.v1`

## 0.5.0 (2023-12-27)

### Enhancements

- [#67](https://github.com/terraform-linters/tflint-ruleset-opa/pull/67): Add support for scoped data sources
- [#69](https://github.com/terraform-linters/tflint-ruleset-opa/pull/69): Add `terraform.imports` and `terraform.checks` functions
- [#71](https://github.com/terraform-linters/tflint-ruleset-opa/pull/71) [#74](https://github.com/terraform-linters/tflint-ruleset-opa/pull/74) [#75](https://github.com/terraform-linters/tflint-ruleset-opa/pull/75) [#79](https://github.com/terraform-linters/tflint-ruleset-opa/pull/79): Bump github.com/open-policy-agent/opa from 0.57.0 to 0.60.0

### Chores

- [#64](https://github.com/terraform-linters/tflint-ruleset-opa/pull/64) [#72](https://github.com/terraform-linters/tflint-ruleset-opa/pull/72): Bump github.com/hashicorp/hcl/v2 from 2.18.0 to 2.19.1
- [#65](https://github.com/terraform-linters/tflint-ruleset-opa/pull/65): Bump github.com/zclconf/go-cty from 1.14.0 to 1.14.1
- [#66](https://github.com/terraform-linters/tflint-ruleset-opa/pull/66): Bump golang.org/x/net from 0.15.0 to 0.17.0
- [#68](https://github.com/terraform-linters/tflint-ruleset-opa/pull/68): Fix incorrect examples of `terraform.resources`
- [#70](https://github.com/terraform-linters/tflint-ruleset-opa/pull/70): Bump github.com/google/go-cmp from 0.5.9 to 0.6.0
- [#73](https://github.com/terraform-linters/tflint-ruleset-opa/pull/73): Bump google.golang.org/grpc from 1.58.2 to 1.58.3
- [#76](https://github.com/terraform-linters/tflint-ruleset-opa/pull/76): Bump actions/setup-go from 4 to 5
- [#77](https://github.com/terraform-linters/tflint-ruleset-opa/pull/77) [#78](https://github.com/terraform-linters/tflint-ruleset-opa/pull/78): Bump github.com/hashicorp/go-hclog from 1.5.0 to 1.6.2
- [#80](https://github.com/terraform-linters/tflint-ruleset-opa/pull/80): Fix E2E tests failing with TFLint v0.50

## 0.4.0 (2023-10-09)

### Enhancements

- [#53](https://github.com/terraform-linters/tflint-ruleset-opa/pull/53) [#59](https://github.com/terraform-linters/tflint-ruleset-opa/pull/59) [#63](https://github.com/terraform-linters/tflint-ruleset-opa/pull/63): Bump github.com/open-policy-agent/opa from 0.54.0 to 0.57.0

### Chores

- [#54](https://github.com/terraform-linters/tflint-ruleset-opa/pull/54): Bump github.com/terraform-linters/tflint-plugin-sdk from 0.17.0 to 0.18.0
- [#55](https://github.com/terraform-linters/tflint-ruleset-opa/pull/55): Add raw binary entries to checksums.txt
- [#56](https://github.com/terraform-linters/tflint-ruleset-opa/pull/56) [#58](https://github.com/terraform-linters/tflint-ruleset-opa/pull/58): Bump github.com/zclconf/go-cty from 1.13.2 to 1.14.0
- [#57](https://github.com/terraform-linters/tflint-ruleset-opa/pull/57): Bump actions/checkout from 3 to 4
- [#60](https://github.com/terraform-linters/tflint-ruleset-opa/pull/60): Bump github.com/hashicorp/hcl/v2 from 2.17.0 to 2.18.0
- [#61](https://github.com/terraform-linters/tflint-ruleset-opa/pull/61): deps: Go 1.21
- [#62](https://github.com/terraform-linters/tflint-ruleset-opa/pull/62): Bump goreleaser/goreleaser-action from 4 to 5

## 0.3.0 (2023-07-19)

### Enhancements

- [#42](https://github.com/terraform-linters/tflint-ruleset-opa/pull/42) [#51](https://github.com/terraform-linters/tflint-ruleset-opa/pull/51): Bump github.com/terraform-linters/tflint-plugin-sdk from 0.16.0 to 0.17.0
- [#45](https://github.com/terraform-linters/tflint-ruleset-opa/pull/45) [#47](https://github.com/terraform-linters/tflint-ruleset-opa/pull/47) [#49](https://github.com/terraform-linters/tflint-ruleset-opa/pull/49) [#52](https://github.com/terraform-linters/tflint-ruleset-opa/pull/52): Bump github.com/open-policy-agent/opa from 0.51.0 to 0.54.0

### Chores

- [#44](https://github.com/terraform-linters/tflint-ruleset-opa/pull/44): docs: Clarify what `policy_dir` is relative to
- [#46](https://github.com/terraform-linters/tflint-ruleset-opa/pull/46): Bump github.com/zclconf/go-cty from 1.13.1 to 1.13.2
- [#48](https://github.com/terraform-linters/tflint-ruleset-opa/pull/48): Bump github.com/hashicorp/hcl/v2 from 2.16.2 to 2.17.0

## 0.2.0 (2023-04-10)

### Enhancements

- [#26](https://github.com/terraform-linters/tflint-ruleset-opa/pull/26) [#29](https://github.com/terraform-linters/tflint-ruleset-opa/pull/29) [#32](https://github.com/terraform-linters/tflint-ruleset-opa/pull/32) [#33](https://github.com/terraform-linters/tflint-ruleset-opa/pull/33) [#37](https://github.com/terraform-linters/tflint-ruleset-opa/pull/37) [#39](https://github.com/terraform-linters/tflint-ruleset-opa/pull/39): Bump github.com/open-policy-agent/opa from 0.48.0 to 0.51.0

### BugFixes

- [#40](https://github.com/terraform-linters/tflint-ruleset-opa/pull/40): Fix internal marshal error of sensitive value

### Chores

- [#24](https://github.com/terraform-linters/tflint-ruleset-opa/pull/24) [#25](https://github.com/terraform-linters/tflint-ruleset-opa/pull/25) [#31](https://github.com/terraform-linters/tflint-ruleset-opa/pull/31): Bump github.com/hashicorp/hcl/v2 from 2.15.0 to 2.16.2
- [#27](https://github.com/terraform-linters/tflint-ruleset-opa/pull/27): Bump golang.org/x/net from 0.5.0 to 0.7.0
- [#28](https://github.com/terraform-linters/tflint-ruleset-opa/pull/28) [#35](https://github.com/terraform-linters/tflint-ruleset-opa/pull/35): Bump github.com/zclconf/go-cty from 1.12.1 to 1.13.1
- [#30](https://github.com/terraform-linters/tflint-ruleset-opa/pull/30): Bump sigstore/cosign-installer from 2 to 3
- [#34](https://github.com/terraform-linters/tflint-ruleset-opa/pull/34): Bump actions/setup-go from 3 to 4
- [#36](https://github.com/terraform-linters/tflint-ruleset-opa/pull/36): Bump github.com/hashicorp/go-hclog from 1.4.0 to 1.5.0
- [#38](https://github.com/terraform-linters/tflint-ruleset-opa/pull/38): Bump github.com/terraform-linters/tflint-plugin-sdk from 0.15.0 to 0.16.0
- [#41](https://github.com/terraform-linters/tflint-ruleset-opa/pull/41): deps: Go 1.20

## 0.1.0 (2023-02-02)

Initial release 🎉
