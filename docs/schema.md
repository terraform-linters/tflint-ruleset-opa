# Terraform Schema

Some functions take a Terraform schema as an argument. This document describes the details of the schema.

Schema is an object that defines an internal body structure. TFLint decodes the body based on the schema, so the schema is always required to access attributes. Values not set in the schema are not included in the return value.

For example, a schema required to decode the top-level `instance_type` is:

```rego
{"instance_type": "string"}
```

The object's key is the attribute name and the value represents the type. The type syntax is the same as [Terraform's type constraints](https://developer.hashicorp.com/terraform/language/expressions/type-constraints).

TFLint implicitly converts values according to their type, which is useful when working with numbers.

```hcl
resource "aws_instance" "number" {
  ebs_block_device {
    volume_size = 50
  }
}

resource "aws_instance" "string" {
  ebs_block_device {
    volume_size = "50" # => convert to number in JSON
  }
}
```

If you don't know the attribute type, you can use `any`. In this case no conversion is done, but the raw value from the config file is available.

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
