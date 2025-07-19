package funcs

import (
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
			got, err := IssueFunc().Func(ctx, ast.NewTerm(msg), ast.NewTerm(rng))
			if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(want.String(), got.Value.String()); diff != "" {
				t.Error(diff)
			}
		})
	}
}
