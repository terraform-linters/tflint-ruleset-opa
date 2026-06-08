# Configuration

This plugin can take advantage of additional features by configuring the plugin block.

Here's an example:

```hcl
plugin "opa" {
  // Plugin common attributes

  policy_dir = "./policies"
  bundle_url = "https://policy-server.example.com/bundles/tflint.tar.gz"
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

Both `bundle_url` and `policy_dir` can be used together. When both are set, local policies replace bundle policies that share the same package path. Since all policies use the `tflint` package, local policies effectively replace the entire bundle when both are present. Data documents (JSON/YAML) from both sources are deep-merged, with local values taking precedence on conflicts.

### Caching

When the bundle server supports [ETag](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/ETag) headers, the plugin caches the downloaded bundle locally to avoid re-downloading unchanged bundles on subsequent runs. The cache is stored at `~/.tflint.d/cache/opa/bundles`. Caching is automatic and requires no configuration.
