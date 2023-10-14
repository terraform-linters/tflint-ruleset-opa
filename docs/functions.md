# Functions

These functions are a list of available custom functions in addition to OPA's default built-in functions.

## `terraform.resources`

```rego
resources := terraform.resources(resource_type, schema, options)
```

Returns Terraform resources.

- `resource_type` (string): resource type to retrieve. "*" is a special character that returns all resources.
- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `resources` (array[object<type: string, name: string, config: body, decl_range: range>]): Terraform "resource" blocks.

Types:

|Name|Type|
|---|---|
|`schema`|`object[string: any<string, schema>]`|
|`body`|`object[string: any<expr, array[nested_block]>]`|
|`expr`|`object<value: any, unknown: boolean, sensitive: boolean, range: range>`|
|`nested_block`|`object<config: object[string: any<expr, array[nested_block]>], labels: array[string], decl_range: range>`|
|`range`|`object<filename: string, start: pos, end: pos>`|
|`pos`|`object<line: number, column: number, byte: number>`|

See also [Terraform Schema](./schema.md) for more information on `schema` type.

The `options` object parameter may contain the following fields:

|Field|Required|Type|Description|
|---|---|---|---|
|`expand_mode`|no|`string`|Whether to expand resources and dynamic blocks. Valid values are `none` and `expand`(default).|

Examples:

Top level attributes

```hcl
resource "aws_instance" "main" {
  instance_type = "t2.micro"
}
```

```rego
terraform.resources("aws_instance", {"instance_type": "string"}, {})
```

```json
[
  {
    "type": "aws_instance",
    "name": "main",
    "config": {
      "bucket": {
        "value": "t2.micro",
        "unknown": false,
        "sensitive": false,
        "range": {
          "filename": "main.tf",
          "start": { "line": 2, "column": 19, "byte": 51 },
          "end": { "line": 2, "column": 29, "byte": 61 }
        }
      }
    },
    "decl_range": {...}
  }
]
```

Nested blocks

```hcl
resource "aws_instance" "main" {
  ebs_block_device {
    volume_size = 50
  }
}
```

```rego
terraform.resources("aws_instance", {"ebs_block_device": {"volume_size": "number"}}, {})
```

```json
[
  {
    "type": "aws_instance",
    "name": "main",
    "config": {
      "ebs_block_device": [
        {
          "config": {
            "volume_size": {
              "value": 50,
              "unknown": false,
              "sensitive": false,
              "range": {...}
            }
          },
          "labels": null,
          "decl_range": {...}
        }
      ]
    },
    "decl_range": {...}
  }
]
```

Expand mode

```hcl
resource "aws_instance" "count" {
  count = 0
}

resource "aws_instance" "for_each" {
  for_each = toset([])
}

resource "aws_instance" "dynamic" {
  dynamic "ebs_block_device" {
    for_each = toset([])
  }
}
```

Expand mode: expand (default)

```rego
terraform.resources("aws_instance", {"dynamic": {"__labels": ["type"]}}, {"expand_mode": "expand"})
```

```json
[
  {
    "type": "aws_instance",
    "name": "dynamic",
    "config": {},
    "decl_range": {...}
  }
]
```

Expan mode: none

```rego
terraform.resources("aws_instance", {"dynamic": {"__labels": ["type"]}}, {"expand_mode": "none"})
```

```json
[
  {
    "type": "aws_instance",
    "name": "count",
    "config": {},
    "decl_range": {...}
  }
  {
    "type": "aws_instance",
    "name": "for_each",
    "config": {},
    "decl_range": {...}
  }
  {
    "type": "aws_instance",
    "name": "dynamic",
    "config": {
      "dynamic": [
        {
          "config": {},
          "labels": ["ebs_block_device"],
          "decl_range": {...}
        }
      ]
    },
    "decl_range": {...}
  }
]
```

## `terraform.data_sources`

```rego
data_sources := terraform.data_sources(data_type, schema, options)
```

Returns Terraform data sources.

- `data_type` (string): data type to retrieve. "*" is a special character that returns all data sources.
- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `data_sources` (array[object<type: string, name: string, config: body, decl_range: range>]): Terraform "data" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
data "aws_ami" "main" {
  owners = ["self"]
}
```

```rego
terraform.data_sources("aws_ami", {"owners": "list(string)"}, {})
```

```json
[
  {
    "type": "aws_ami",
    "name": "main",
    "config": {
      "owners": {
        "value": ["self"],
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.module_calls`

```rego
modules := terraform.module_calls(schema, options)
```

Returns Terraform module calls.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `modules` (array[object<name: string, config: body, decl_range: range>]): Terraform "module" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
module "aws_instance" {
  instance_type = "t2.micro"
}
```

```rego
terraform.module_calls({"instance_type": "string"}, {})
```

```json
[
  {
    "name": "aws_instance",
    "config": {
      "instance_type": {
        "value": "t2.micro",
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.providers`

```rego
providers := terraform.providers(schema, options)
```

Returns Terraform providers.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `providers` (array[object<name: string, config: body, decl_range: range>]): Terraform "provider" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
provider "aws" {
  region = "us-east-1"
}
```

```rego
terraform.providers({"region": "string"}, {})
```

```json
[
  {
    "name": "aws",
    "config": {
      "region": {
        "value": "us-east-1",
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.settings`

```rego
settings := terraform.settings(schema, options)
```

Returns Terraform settings.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `settings` (array[object<config: body, decl_range: range>]): Terraform "terraform" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}
```

```rego
terraform.settings({"required_providers": {"aws": "map(string)"}}, {})
```

```json
[
  {
    "config": {
      "required_providers": [
        {
          "config": {
            "aws": {
              "value": {
                "source": "hashicorp/aws",
                "version": "~> 4.0"
              },
              "unknown": false,
              "sensitive": false,
              "range": {...}
            }
          },
          "labels": null,
          "decl_range": {...}
        }
      ]
    },
    "decl_range": {...}
  }
]
```

## `terraform.variables`

```rego
variables := terraform.variables(schema, options)
```

Returns Terraform variables.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `variables` (array[object<name: string, config: body, decl_range: range>]): Terraform "variable" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
variable "foo" {
  nullable = true
}
```

```rego
terraform.variables({"nullable": "bool"}, {})
```

```json
[
  {
    "name": "foo",
    "config": {
      "nullable": {
        "value": true,
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.outputs`

```rego
outputs := terraform.outputs(schema, options)
```

Returns Terraform outputs.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `outputs` (array[object<name: string, config: body, decl_range: range>]): Terraform "output" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
output "bar" {
  description = null
}
```

```rego
terraform.outputs({"description": "string"}, {})
```

```json
[
  {
    "name": "bar",
    "config": {
      "description": {
        "value": null,
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.locals`

```rego
locals := terraform.locals(options)
```

Returns Terraform local values.

- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `locals` (array[object<name: string, expr: expr, decl_range: range>]): Terraform local values.

The `options` is equivalent to the argument of the `terraform.resources` function.

Examples:

```hcl
locals {
  foo = "bar"
}
```

```rego
terraform.locals({})
```

```json
[
  {
    "name": "foo",
    "expr": {
      "value": "bar",
      "unknown": false,
      "sensitive": false,
      "range": {...}
    },
    "decl_range": {...}
  }
]
```

## `terraform.moved_blocks`

```rego
blocks := terraform.moved_blocks(schema, options)
```

Returns Terraform moved blocks.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `blocks` (array[object<config: body, decl_range: range>]): Terraform "moved" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
moved {
  from = aws_instance.foo
  to   = aws_instance.bar
}
```

```rego
terraform.moved_blocks({"from": "any"}, {})
```

```json
[
  {
    "config": {
      "from": {
        "unknown": true,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.imports`

```rego
blocks := terraform.imports(schema, options)
```

Returns Terraform imports blocks.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `blocks` (array[object<config: body, decl_range: range>]): Terraform "import" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
import {
  to = aws_instance.example
  id = "i-abcd1234"
}
```

```rego
terraform.imports({"id": "string"}, {})
```

```json
[
  {
    "config": {
      "id": {
        "value": "i-abcd1234",
        "unknown": false,
        "sensitive": false,
        "range": {...}
      }
    },
    "decl_range": {...}
  }
]
```

## `terraform.checks`

```rego
blocks := terraform.checks(schema, options)
```

Returns Terraform check blocks.

- `schema` (schema): schema for attributes referenced in rules.
- `options` (object[string: string]): options to change the retrieve/evaluate behavior.

Returns:

- `blocks` (array[object<config: body, decl_range: range>]): Terraform "check" blocks.

The `schema` and `options` are equivalent to the arguments of the `terraform.resources` function.

Examples:

```hcl
check "health_check" {
  data "http" "terraform_io" {
    url = "https://www.terraform.io"
  }

  assert {
    condition = data.http.terraform_io.status_code == 200
    error_message = "${data.http.terraform_io.url} returned an unhealthy status code"
  }
}
```

```rego
terraform.checks({"assert": {"condition": "bool"}}, {})
```

```json
[
  {
    "config": {
      "assert": [
        {
          "config": {
            "condition": {
              "unknown": true,
              "sensitive": false,
              "range": {...}
            }
          },
          "labels": null,
          "decl_range": {...}
        }
      ]
    },
    "decl_range": {...}
  }
]
```

## `terraform.module_range`

```rego
range := terraform.module_range()
```

Returns a range for the current Terraform module.
This is useful in rules that check for non-existence.

Returns:

- `range` (range): a range for [DIR]/main.tf:1:1

## `tflint.issue`

```rego
issue := tflint.issue(msg, range)
```

Returns issue object.

Returns:

- `issue` (object<msg: string, range: range>): issue object.
