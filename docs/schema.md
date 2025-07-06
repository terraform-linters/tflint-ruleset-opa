# Terraform Schema

Some functions take a Terraform schema as an argument. This document describes the details of the schema.

Schema is an object that defines an internal body structure. TFLint decodes the body based on the schema, so the schema is always required to access attributes. Values not set in the schema are not included in the return value.

For example, a schema required to decode the top-level `instance_type` is:

```rego
{"instance_type": "string"}
```

The object's key is the attribute name and the value represents the type. The type syntax is essentially the same as [Terraform's type constraints](https://developer.hashicorp.com/terraform/language/expressions/type-constraints).

## `any` Type

TFLint implicitly converts values according to their type, which is useful when working with numbers.

```rego
{"size": "number"}
```

```hcl
resource "aws_ebs_volume" "number" {
  size = 50
}

resource "aws_ebs_volume" "string" {
  size = "50" # => convert to number in JSON
}
```

If you don't know the attribute type, you can use `any`. In this case no conversion is done, but the raw value from the config file is available. In the above example, the JSON will contain 50 and "50".

```rego
{"size": "any"}
```

## Nested Blocks

A schema for decoding nested blocks is:

```rego
{"ebs_block_device": {"volume_size": "string"}}
```

You can use objects instead of types as values. The objects represent nested schemas.

A more special case is the labeled block schema. For example, here is a schema available to retrieve a dynamic block like:

```hcl
resource "aws_instance" "main" {
  dynamic "ebs_block_device" {
    for_each = var.devices
  }
}
```

```rego
{"dynamic": {"__labels": ["type"], "for_each": "any"}}
```

The `__labels` is a special key that sets labels. The value defines the label name in an array, not the type. Label names are basically meaningless.

## `expr` Type

The `expr` type can be used as a special type. Attributes specified as `expr` type are not evaluated immediately, but the structure of the expression is included in the value.

```hcl
variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "main" {
  instance_type = var.instance_type
}
```

```rego
{"instance_type": "string"}
```

```json
{
  "value": "t2.micro",
  "unknown": false,
  "sensitive": false,
  "ephemeral": false,
  "range": {...}
}
```

```rego
{"instance_type": "expr"}
```

```json
{
  "value": "var.instance_type",
  "range": {...}
}
```

This is useful for writing policies over expression structures. For example, the `expr` type is the only way to handle meta-arguments such as `ignore_changes` that cannot be evaluated in the normal way.

The value obtained with the `expr` type is called `raw_expr` type and can be passed to HCL static analysis functions such as [`hcl.expr_list`](./functions.md#hclexpr_list), [`hcl.expr_map`](./functions.md#hclexpr_map), and [`hcl.expr_call`](./functions.md#hclexpr_call).
