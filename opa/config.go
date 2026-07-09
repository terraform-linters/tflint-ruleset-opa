package opa

import (
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
)

// Config is the configuration for the ruleset.
type Config struct {
	PolicyDirs []string `hclext:"policy_dirs,optional"`
}

var (
	policyRoot      = "~/.tflint.d/policies"
	localPolicyRoot = "./.tflint.d/policies"
)

// policyDirs returns the policy directories to load.
// Adopted with the following priorities:
//
//  1. `policy_dirs` in a config file
//  2. `TFLINT_OPA_POLICY_DIRS` environment variable (supports multiple directories separated by `,`)
//  3. Current directory (./.tflint.d/policies)
//  4. Home directory (~/.tflint.d/policies)
//
// If the environment variable is set, other directories will not be considered,
// but if the current directory does not exist, it will fallback to the home directory.
func (c *Config) policyDirs() ([]string, error) {
	var expandedDirs []string

	// Priority 1: policy_dirs from config
	for _, dir := range c.PolicyDirs {
		expanded, err := homedir.Expand(dir)
		if err != nil {
			return nil, err
		}
		expandedDirs = append(expandedDirs, expanded)
	}

	if len(expandedDirs) > 0 {
		return expandedDirs, nil
	}

	// Priority 2: TFLINT_OPA_POLICY_DIRS environment variable
	// Supports multiple directories separated by `,`
	for dir := range strings.SplitSeq(os.Getenv("TFLINT_OPA_POLICY_DIRS"), ",") {
		dir = strings.TrimSpace(dir)
		if dir != "" {
			expanded, err := homedir.Expand(dir)
			if err != nil {
				return nil, err
			}
			expandedDirs = append(expandedDirs, expanded)
		}
	}

	if len(expandedDirs) > 0 {
		return expandedDirs, nil
	}

	// Priority 3 & 4: Check local directory, fallback to home directory
	_, err := os.Stat(localPolicyRoot)
	if os.IsNotExist(err) {
		dir, err := policyRootDir()
		return []string{dir}, err 
	}

	return []string{localPolicyRoot}, err
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
