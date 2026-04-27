package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/open-policy-agent/opa/v1/bundle"
	"github.com/open-policy-agent/opa/v1/loader"
)

type Result struct {
	Issue  []Issue `json:"issues"`
	Errors []any   `json:"errors"`
}

type Issue struct {
	Rule    Rule    `json:"rule"`
	Message string  `json:"message"`
	Range   Range   `json:"range"`
	Callers []Range `json:"callers"`
}

type Rule struct {
	Name     string `json:"name"`
	Severity string `json:"severity"`
	Link     string `json:"link"`
}

type Range struct {
	Filename string `json:"filename"`
	Start    Pos    `json:"start"`
	End      Pos    `json:"end"`
}

type Pos struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type bundleSetup struct {
	policyDir string
	token     string
}

func TestIntegration(t *testing.T) {
	tests := []struct {
		name    string
		command *exec.Cmd
		dir     string
		test    bool
		bundle  *bundleSetup
	}{
		{
			name:    "instance type",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "instance_type",
		},
		{
			name:    "instance type (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "instance_type",
			test:    true,
		},
		{
			name:    "naming convention",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "naming_convention",
		},
		{
			name:    "naming convention (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "naming_convention",
			test:    true,
		},
		{
			name:    "volume type",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "volume_type",
		},
		{
			name:    "volume type (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "volume_type",
			test:    true,
		},
		{
			name:    "volume size",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "volume_size",
		},
		{
			name:    "volume size (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "volume_size",
			test:    true,
		},
		{
			name:    "tagged",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "tagged",
		},
		{
			name:    "tagged (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "tagged",
			test:    true,
		},
		{
			name:    "resources",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "resources",
		},
		{
			name:    "resources (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "resources",
			test:    true,
		},
		{
			name:    "providers",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "providers",
		},
		{
			name:    "providers (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "providers",
			test:    true,
		},
		{
			name:    "data sources",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "data_sources",
		},
		{
			name:    "data sources (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "data_sources",
			test:    true,
		},
		{
			name:    "module calls",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "module_calls",
		},
		{
			name:    "module calls (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "module_calls",
			test:    true,
		},
		{
			name:    "settings",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "settings",
		},
		{
			name:    "settings (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "settings",
			test:    true,
		},
		{
			name:    "variables",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "variables",
		},
		{
			name:    "variables (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "variables",
			test:    true,
		},
		{
			name:    "outputs",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "outputs",
		},
		{
			name:    "outputs (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "outputs",
			test:    true,
		},
		{
			name:    "locals",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "locals",
		},
		{
			name:    "locals (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "locals",
			test:    true,
		},
		{
			name:    "moved blocks",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "moved",
		},
		{
			name:    "moved blocks (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "moved",
			test:    true,
		},
		{
			name:    "imports",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "imports",
		},
		{
			name:    "imports (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "imports",
			test:    true,
		},
		{
			name:    "checks",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "checks",
		},
		{
			name:    "checks (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "checks",
			test:    true,
		},
		{
			name:    "removed blocks",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "removed",
		},
		{
			name:    "removed blocks (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "removed",
			test:    true,
		},
		{
			name:    "ephemerals",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "ephemerals",
		},
		{
			name:    "ephemerals (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "ephemerals",
			test:    true,
		},
		{
			name:    "expr without eval",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "expr_without_eval",
		},
		{
			name:    "expr without eval (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "expr_without_eval",
			test:    true,
		},
		{
			name:    "actions",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "actions",
		},
		{
			name:    "actions (test)",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "actions",
			test:    true,
		},
		{
			name:    "remote bundle",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "remote_bundle",
			bundle:  &bundleSetup{policyDir: "policies"},
		},
		{
			name:    "remote bundle with auth",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "remote_bundle",
			bundle:  &bundleSetup{policyDir: "policies", token: "test-secret-token"},
		},
		{
			name:    "remote bundle with local override",
			command: exec.Command("tflint", "--format", "json", "--force"),
			dir:     "remote_bundle_override",
			bundle:  &bundleSetup{policyDir: "bundle_policies"},
		},
	}

	dir, _ := os.Getwd()
	defer os.Chdir(dir)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir := filepath.Join(dir, test.dir)
			t.Chdir(testDir)

			if test.test {
				test.command.Env = append(os.Environ(), "TFLINT_OPA_TEST=1")
			}

			if test.bundle != nil {
				bundleDir := filepath.Join(testDir, test.bundle.policyDir)
				server := serveBundleFromDir(t, bundleDir, test.bundle.token)
				defer server.Close()

				t.Setenv("TFLINT_OPA_BUNDLE_URL", server.URL)
				if test.bundle.token != "" {
					t.Setenv("TFLINT_OPA_BUNDLE_TOKEN", test.bundle.token)
				}
			}

			var stdout, stderr bytes.Buffer
			test.command.Stdout = &stdout
			test.command.Stderr = &stderr
			if err := test.command.Run(); err != nil {
				t.Fatalf("%s, stdout=%s stderr=%s", err, stdout.String(), stderr.String())
			}

			b, err := readResultFile(testDir, test.test)
			if err != nil {
				t.Fatal(err)
			}

			var want Result
			if err := json.Unmarshal(b, &want); err != nil {
				t.Fatal(err)
			}

			var got Result
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatal(err)
			}

			opts := cmp.Options{
				cmpopts.SortSlices(func(a, b Issue) bool {
					if a.Range.Start.Line != b.Range.Start.Line {
						return a.Range.Start.Line > b.Range.Start.Line
					}
					return a.Rule.Name > b.Rule.Name
				}),
				cmpopts.AcyclicTransformer("Link", func(path string) string {
					return filepath.ToSlash(path)
				}),
			}
			if diff := cmp.Diff(want, got, opts...); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func serveBundleFromDir(t *testing.T, policyDir string, token string) *httptest.Server {
	t.Helper()

	ret, err := loader.NewFileLoader().Filtered([]string{policyDir}, nil)
	if err != nil {
		t.Fatal(err)
	}

	absDir, err := filepath.Abs(policyDir)
	if err != nil {
		t.Fatal(err)
	}
	parentDir := filepath.Dir(absDir)

	var mods []bundle.ModuleFile
	for _, regoFile := range ret.Modules {
		relPath, err := filepath.Rel(parentDir, regoFile.Name)
		if err != nil {
			t.Fatal(err)
		}
		mods = append(mods, bundle.ModuleFile{
			URL:  relPath,
			Path: relPath,
			Raw:  regoFile.Raw,
		})
	}

	b := bundle.Bundle{
		Modules: mods,
		Data:    map[string]interface{}{},
	}
	var buf bytes.Buffer
	if err := bundle.NewWriter(&buf).Write(b); err != nil {
		t.Fatal(err)
	}
	bundleBytes := buf.Bytes()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if token != "" && r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Write(bundleBytes)
	}))
}

func readResultFile(dir string, test bool) ([]byte, error) {
	var resultFile string
	if test {
		resultFile = "result_test.json"
	} else {
		resultFile = "result.json"
	}

	tflintVersion := os.Getenv("TFLINT_VERSION")
	if tflintVersion != "latest" {
		var versionResultFile string
		if test {
			versionResultFile = fmt.Sprintf("result_test-%s.json", tflintVersion)
		} else {
			versionResultFile = fmt.Sprintf("result-%s.json", tflintVersion)
		}

		if _, err := os.Stat(filepath.Join(dir, versionResultFile)); !os.IsNotExist(err) {
			resultFile = versionResultFile
		}
	}
	return os.ReadFile(filepath.Join(dir, resultFile))
}
