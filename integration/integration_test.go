package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func TestIntegration(t *testing.T) {
	tests := []struct {
		name    string
		command *exec.Cmd
		dir     string
		test    bool
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
	}

	dir, _ := os.Getwd()
	defer os.Chdir(dir)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testDir := filepath.Join(dir, test.dir)

			t.Cleanup(func() {
				if err := os.Chdir(dir); err != nil {
					t.Fatal(err)
				}
			})

			if err := os.Chdir(testDir); err != nil {
				t.Fatal(err)
			}

			retFile := "result"
			if test.test {
				test.command.Env = append(os.Environ(), "TFLINT_OPA_TEST=1")
				retFile = "result_test"
			}

			var stdout, stderr bytes.Buffer
			test.command.Stdout = &stdout
			test.command.Stderr = &stderr
			if err := test.command.Run(); err != nil {
				t.Fatalf("%s, stdout=%s stderr=%s", err, stdout.String(), stderr.String())
			}

			b, err := os.ReadFile(filepath.Join(testDir, fmt.Sprintf("%s.json", retFile)))
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
