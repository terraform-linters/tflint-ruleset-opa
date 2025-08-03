package funcs

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/open-policy-agent/opa/v1/ast"
	"github.com/open-policy-agent/opa/v1/rego"
)

func TestIssueFunc(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		rng  map[string]any
		want map[string]any
	}{
		{
			name: "issue",
			msg:  "test",
			rng: map[string]any{
				"filename": "main.tf",
				"start": map[string]int{
					"line":   1,
					"column": 1,
					"byte":   0,
				},
				"end": map[string]int{
					"line":   1,
					"column": 1,
					"byte":   0,
				},
			},
			want: map[string]any{
				"msg": "test",
				"range": map[string]any{
					"filename": "main.tf",
					"start": map[string]int{
						"line":   1,
						"column": 1,
						"byte":   0,
					},
					"end": map[string]int{
						"line":   1,
						"column": 1,
						"byte":   0,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg, err := ast.InterfaceToValue(test.msg)
			if err != nil {
				t.Fatal(err)
			}
			rng, err := ast.InterfaceToValue(test.rng)
			if err != nil {
				t.Fatal(err)
			}
			want, err := ast.InterfaceToValue(test.want)
			if err != nil {
				t.Fatal(err)
			}

			ctx := rego.BuiltinContext{}
			got, err := IssueFunc().Impl(ctx, ast.NewTerm(msg), ast.NewTerm(rng))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestAsIssue(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  *Issue
		err   string
	}{
		{
			name: "valid issue",
			input: map[string]any{
				"msg": "message",
				"range": map[string]any{
					"filename": "",
					"start":    map[string]any{"line": json.Number("0"), "column": json.Number("0"), "byte": json.Number("0")},
					"end":      map[string]any{"line": json.Number("0"), "column": json.Number("0"), "byte": json.Number("0")},
				},
			},
			want: &Issue{Message: "message"},
		},
		{
			name:  "invalid type",
			input: "",
			err:   "issue is not object, got string",
		},
		{
			name: "invalid message type",
			input: map[string]any{
				"msg": 1,
				"range": map[string]any{
					"filename": "",
					"start":    map[string]any{"line": json.Number("0"), "column": json.Number("0"), "byte": json.Number("0")},
					"end":      map[string]any{"line": json.Number("0"), "column": json.Number("0"), "byte": json.Number("0")},
				},
			},
			err: "issue.msg is not string, got int",
		},
		{
			name: "invalid range type",
			input: map[string]any{
				"msg":   "test",
				"range": "main.tf:1,1-1",
			},
			err: "issue.range is not object, got string",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := AsIssue(test.input)
			if err != nil {
				if err.Error() != test.err {
					t.Fatalf(`expect "%s", but got "%s"`, test.err, err.Error())
				}
				return
			}
			if test.err != "" {
				t.Fatal("should return an error, but it does not")
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
