# Debugging

If your policy doesn't work as intended, you can use the debugging functions provided by OPA to help troubleshoot.

## `print`

The `print` function can output arbitrary values to the log. You can check the value by setting `TFLINT_LOG=debug`.

```hcl
resource "aws_instance" "main" {
  instance_type = "t2.micro"
}
```

```rego
package tflint

import rego.v1

deny_invalid_instance_type contains issue if {
	instances := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	print(instances)
	instances[_].config.type.value == "t2.micro" # typo: type -> instance_type

	issue := tflint.issue("t2.micro is not allowed", instances[_].config.instance_type.range)
}
```

```console
$ TFLINT_LOG=debug tflint
...
16:47:48 [DEBUG] host2plugin/client.go:124: starting host-side gRPC server
16:47:48 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:47:48 [DEBUG] topdown/print.go:48: [{"config": {"instance_type": {"range": {"end": {"byte": 61, "column": 29, "line": 2}, "filename": "main.tf", "start": {"byte": 51, "column": 19, "line": 2}}, "sensitive": false, "unknown": false, "value": "t2.micro"}}, "decl_range": {"end": {"byte": 30, "column": 31, "line": 1}, "filename": "main.tf", "start": {"byte": 0, "column": 1, "line": 1}}, "name": "main", "type": "aws_instance"}]
...
```

## `trace`

The `trace` function prints a `note` to the trace. Tracing can be enabled by setting `TFLINT_OPA_TRACE=1`. Traces are printed to the log.

```rego
package tflint

import rego.v1

deny_invalid_instance_type contains issue if {
	instances := terraform.resources("aws_instance", {"instance_type": "string"}, {})
	trace("after fetch")
	instances[_].config.type.value == "t2.micro"

	issue := tflint.issue("t2.micro is not allowed", instances[_].config.instance_type.range)
}
```

```console
$ TFLINT_LOG=debug TFLINT_OPA_TRACE=1 tflint
...
16:55:12 [DEBUG] host2plugin/client.go:124: starting host-side gRPC server
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: Enter data.tflint.deny_invalid_instance_type = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Eval data.tflint.deny_invalid_instance_type = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Unify data.tflint.deny_invalid_instance_type = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Index data.tflint.deny_invalid_instance_type (matched 1 rule)
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Enter data.tflint.deny_invalid_instance_type
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Eval terraform.resources("aws_instance", {"instance_type": "string"}, {}, __local2__)
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Unify __local2__ = [{"config": {"instance_type": {"range": {"end": {"byte": 61, "column": 29, "line": 2}, "filename": "main.tf", "start": {"byte": 51, "column": 19, "line": 2}}, "sensitive": false, "unknown": false, "value": "t2.micro"}}, "decl_range": {"end": {"byte": 30, "column": 31, "line": 1}, "filename": "main.tf", "start": {"byte": 0, "column": 1, "line": 1}}, "name": "main", "type": "aws_instance"}]
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Eval instances = __local2__
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Unify instances = [{"config": {"instance_type": {"range": {"end": {"byte": 61, "column": 29, "line": 2}, "filename": "main.tf", "start": {"byte": 51, "column": 19, "line": 2}}, "sensitive": false, "unknown": false, "value": "t2.micro"}}, "decl_range": {"end": {"byte": 30, "column": 31, "line": 1}, "filename": "main.tf", "start": {"byte": 0, "column": 1, "line": 1}}, "name": "main", "type": "aws_instance"}]
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Eval trace("after fetch")
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Note "after fetch"
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Eval instances[_].config.type.value = "t2.micro"
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Unify instances[_].config.type.value = "t2.micro"
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Unify 0 = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Fail instances[_].config.type.value = "t2.micro"
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Redo trace("after fetch")
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Redo instances = __local2__
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | | Redo terraform.resources("aws_instance", {"instance_type": "string"}, {}, __local2__)
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Unify set() = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Exit data.tflint.deny_invalid_instance_type = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: Redo data.tflint.deny_invalid_instance_type = _
16:55:12 [DEBUG] go-plugin@v1.4.8/client.go:1045: tflint-ruleset-opa: 16:55:12 [DEBUG] topdown/trace.go:239: | Redo data.tflint.deny_invalid_instance_type = _
...
```
