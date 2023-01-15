package opa

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		config     *Config
		root       string
		currentDir string
		env        map[string]string
		want       string
		err        error
	}{
		{
			name:   "default (not exists)",
			config: &Config{},
			root:   filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies"),
			want:   filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies"),
			err:    os.ErrNotExist,
		},
		{
			name:   "default (exists)",
			config: &Config{},
			root:   filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies"),
			want:   filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies"),
		},
		{
			name:       "local",
			config:     &Config{},
			currentDir: filepath.Join(cwd, "test-fixtures", "config", "local"),
			want:       "./.tflint.d/policies",
		},
		{
			name:   "env",
			config: &Config{},
			env: map[string]string{
				"TFLINT_OPA_POLICY_DIR": "policies",
			},
			want: "policies",
		},
		{
			name:   "config",
			config: &Config{PolicyDir: "config/policies"},
			want:   "config/policies",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.root != "" {
				original := policyRoot
				policyRoot = test.root
				defer func() { policyRoot = original }()
			}
			if test.currentDir != "" {
				os.Chdir(test.currentDir)
				defer os.Chdir(cwd)
			}
			for k, v := range test.env {
				t.Setenv(k, v)
			}

			got, err := test.config.policyDir()
			if err != nil {
				if errors.Is(err, test.err) {
					return
				}
				t.Fatal(err)
			}
			if err == nil && test.err != nil {
				t.Fatal("should return an error, but it does not")
			}

			if got != test.want {
				t.Fatalf("want: %s, got: %s", test.want, got)
			}
		})
	}
}
