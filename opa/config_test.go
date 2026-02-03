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
		want       []string
		err        error
	}{
		{
			name:   "default (not exists)",
			config: &Config{},
			root:   filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies"),
			want:   []string{filepath.Join(cwd, "test-fixtures", "config", "root-not-exists", ".tflint.d", "policies")},
			err:    os.ErrNotExist,
		},
		{
			name:   "default (exists)",
			config: &Config{},
			root:   filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies"),
			want:   []string{filepath.Join(cwd, "test-fixtures", "config", "root-exists", ".tflint.d", "policies")},
		},
		{
			name:       "local",
			config:     &Config{},
			currentDir: filepath.Join(cwd, "test-fixtures", "config", "local"),
			want:       []string{"./.tflint.d/policies"},
		},
		{
			name:   "env",
			config: &Config{},
			env: map[string]string{
				"TFLINT_OPA_POLICY_DIRS": "policies",
			},
			want: []string{"policies"},
		},
		{
			name:   "env multiple directories",
			config: &Config{},
			env: map[string]string{
				"TFLINT_OPA_POLICY_DIRS": "policies,other/policies",
			},
			want: []string{"policies", "other/policies"},
		},
		{
			name:   "env multiple directories with spaces",
			config: &Config{},
			env: map[string]string{
				"TFLINT_OPA_POLICY_DIRS": " policies , other/policies ",
			},
			want: []string{"policies", "other/policies"},
		},
		{
			name:   "env with tilde expansion",
			config: &Config{},
			env: map[string]string{
				"TFLINT_OPA_POLICY_DIRS": "~/policies",
			},
			want: []string{filepath.Join(os.Getenv("HOME"), "policies")},
		},
		{
			name:   "config single directory",
			config: &Config{PolicyDirs: []string{"config/policies"}},
			want:   []string{"config/policies"},
		},
		{
			name:   "config multiple directories",
			config: &Config{PolicyDirs: []string{"config/policies", "other/policies"}},
			want:   []string{"config/policies", "other/policies"},
		},
		{
			name:   "config with tilde expansion",
			config: &Config{PolicyDirs: []string{"~/policies"}},
			want:   []string{filepath.Join(os.Getenv("HOME"), "policies")},
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

			got, err := test.config.policyDirs()
			if err != nil {
				if errors.Is(err, test.err) {
					return
				}
				t.Fatal(err)
			}
			if err == nil && test.err != nil {
				t.Fatal("should return an error, but it does not")
			}

			if len(got) != len(test.want) {
				t.Fatalf("want: %v, got: %v", test.want, got)
			}

			for i := range got {
				if got[i] != test.want[i] {
					t.Fatalf("want: %v, got: %v", test.want, got)
				}
			}
		})
	}
}
