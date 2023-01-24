# Configuration

This plugin can take advantage of additional features by configure the plugin block. Currently, this configuration is only available for customizing a policy directory.

Here's an example:

```hcl
plugin "opa" {
  // Plugin common attributes

  policy_dir = "./policies"
}
```

## `policy_dir`

Default: `./.tflint.d/policies`, `~/.tflint.d/policies`

Change the directory from which policies are loaded. The priority is as follows:

1. `policy_dir` in the config
2. `TFLINT_OPA_POLICY_DIR` environment variable
3. `./.tflint.d/policies`
4. `~/.tflint.d/policies`
