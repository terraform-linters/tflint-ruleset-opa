package opa

import (
	"os"

	"github.com/mitchellh/go-homedir"
)

// Config is the configuration for the ruleset.
type Config struct {
	PolicyDir string `hclext:"policy_dir,optional"`
	BundleURL string `hclext:"bundle_url,optional"`
}

var (
	policyRoot      = "~/.tflint.d/policies"
	localPolicyRoot = "./.tflint.d/policies"
)

// policyDir returns the base policy directory.
// Adopted with the following priorities:
//
//  1. `policy_dir` in a config file
//  2. `TFLINT_OPA_POLICY_DIR` environment variable
//  3. Current directory (./.tflint.d/policies)
//  4. Home directory (~/.tflint.d/policies)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func (c *Config) policyDir() (string, error) {
	if c.PolicyDir != "" {
		return homedir.Expand(c.PolicyDir)
	}

	if dir := os.Getenv("TFLINT_OPA_POLICY_DIR"); dir != "" {
		return dir, nil
	}

	_, err := os.Stat(localPolicyRoot)
	if os.IsNotExist(err) {
		return policyRootDir()
	}

	return localPolicyRoot, err
}

func policyRootDir() (string, error) {
	dir, err := homedir.Expand(policyRoot)
	if err != nil {
		return "", err
	}

	// Returning os.ErrNotExist allows checking to continue even if it doesn't exist
	_, err = os.Stat(dir)
	return dir, err
}

// bundleURL returns the bundle URL if configured.
// Adopted with the following priorities:
//
//  1. `bundle_url` in a config file
//  2. `TFLINT_OPA_BUNDLE_URL` environment variable
func (c *Config) bundleURL() string {
	if c.BundleURL != "" {
		return c.BundleURL
	}
	return os.Getenv("TFLINT_OPA_BUNDLE_URL")
}

var bundleCacheRoot = "~/.tflint.d/cache/bundles"

func bundleCacheDir() string {
	dir, err := homedir.Expand(bundleCacheRoot)
	if err != nil {
		return ""
	}
	return dir
}
