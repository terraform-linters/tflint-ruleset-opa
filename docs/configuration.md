# Configuration

This plugin can take advantage of additional features by configuring the plugin block.

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

A relative path is resolved from the current directory.

## `bundle_url`

Default: (none)

Fetch policies from a remote [OPA bundle](https://www.openpolicyagent.org/docs/latest/management-bundles/) server over HTTP(S) at startup. The URL should point to a valid OPA bundle (a tar.gz archive containing `.rego` files and optional data files).

The priority is as follows:

1. `bundle_url` in the config
2. `TFLINT_OPA_BUNDLE_URL` environment variable

```hcl
plugin "opa" {
  enabled    = true
  bundle_url = "https://policy-server.example.com/bundles/tflint.tar.gz"
}
```

`bundle_url` and a local policy directory are mutually exclusive. When `bundle_url` is set, the bundle is the sole source of policies and data; if the local policy directory (`policy_dir`, or one of its [default locations](#policy_dir)) contains any policies or data, an error is raised. Only an empty (or absent) local directory is allowed alongside a bundle.
