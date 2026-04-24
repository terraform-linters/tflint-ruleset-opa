# Environment Variables

Below is a list of environment variables that have meaning in the OPA ruleset:

- `TFLINT_OPA_POLICY_DIR`
  - Directory where policy files are placed. See [Configuration](./configuration.md).
- `TFLINT_OPA_BUNDLE_URL`
  - URL of a remote OPA bundle server. See [Configuration](./configuration.md).
- `TFLINT_OPA_BUNDLE_TOKEN`
  - Bearer token for authenticating with a remote bundle server. When set, the token is sent in the `Authorization: Bearer <token>` header when fetching a bundle via `bundle_url`. See [Configuration](./configuration.md).
- `TFLINT_OPA_TRACE`
  - Enable tracing. See [Debugging](./debug.md).
- `TFLINT_OPA_TEST`
  - Enable test mode. See [Testing](./testing.md)
