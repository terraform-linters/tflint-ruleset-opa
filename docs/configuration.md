# Configuration

This plugin can take advantage of additional features by configure the plugin block. Currently, this configuration is only available for customizing the directories to load policies.

Here's an example:

```hcl
plugin "opa" {
  // Plugin common attributes

  policy_dirs = ["./policies", "./other-policies"]
}
```

## `policy_dirs`

Default: `./.tflint.d/policies`, `~/.tflint.d/policies`

Change the directories from which policies are loaded. You can specify multiple directories to load policies from different locations. The priority is as follows:

1. `policy_dirs` in the config
2. `TFLINT_OPA_POLICY_DIRS` environment variable (supports multiple directories separated `,`)
3. `./.tflint.d/policies`
4. `~/.tflint.d/policies`

A relative path is resolved from the current directory.

### Examples

Single directory:
```hcl
plugin "opa" {
  policy_dirs = ["./policies"]
}
```

Multiple directories:
```hcl
plugin "opa" {
  policy_dirs = ["./policies", "./team-policies", "~/shared-policies"]
}
```

Using environment variable with a single directory:
```bash
export TFLINT_OPA_POLICY_DIRS="./policies"
```

Using environment variable with multiple directories:
```bash
export TFLINT_OPA_POLICY_DIRS="./policies,./team-policies,~/shared-policies"
```
