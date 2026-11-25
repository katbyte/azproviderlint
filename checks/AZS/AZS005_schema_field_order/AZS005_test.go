package AZS005

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS005(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs005")

	// each case is a file's declarations, the messages reported and the file once every fix is made
	for name, tc := range map[string]struct {
		src, fixed string
		want       []string
	}{
		"in order": {
			src:   "var _ = Schema{\n\tDescription: \"The name.\",\n\tType:        1,\n\tOptional:    true,\n\tComputed:    true,\n\tElem:        nil,\n}\n",
			fixed: "var _ = Schema{\n\tDescription: \"The name.\",\n\tType:        1,\n\tOptional:    true,\n\tComputed:    true,\n\tElem:        nil,\n}\n",
		},
		"on one line": {
			src:   "var _ = Schema{Computed: true, Type: 1, Optional: true}\n",
			fixed: "var _ = Schema{Type: 1, Optional: true, Computed: true}\n",
			want:  []string{"schema field Type should come before Computed"},
		},
		"a comment moves with its field": {
			src:   "var _ = Schema{\n\tType: 1,\n\t// checked by the API too\n\tValidateFunc: nil, // for now\n\tRequired:     true,\n}\n",
			fixed: "var _ = Schema{\n\tType:     1,\n\tRequired: true,\n\t// checked by the API too\n\tValidateFunc: nil, // for now\n}\n",
			want:  []string{"schema field Required should come before ValidateFunc"},
		},
		"the note on a property both set and read back stays between the two": {
			src:   "var _ = Schema{\n\tType:     1,\n\tComputed: true,\n\t// Note: O+C the API sets one\n\tOptional: true,\n}\n",
			fixed: "var _ = Schema{\n\tType:     1,\n\tOptional: true,\n\t// Note: O+C the API sets one\n\tComputed: true,\n}\n",
			want:  []string{"schema field Optional should come before Computed"},
		},
		"wherever it was written": {
			src:   "var _ = Schema{\n\tType:     1,\n\tOptional: true,\n\t// NOTE: O+C the API sets one\n\tForceNew: true,\n\tComputed: true,\n}\n",
			fixed: "var _ = Schema{\n\tType:     1,\n\tOptional: true,\n\t// NOTE: O+C the API sets one\n\tComputed: true,\n\tForceNew: true,\n}\n",
			want:  []string{"schema field Computed should come before ForceNew"},
		},
		"the values are lined up afresh": {
			src:   "var _ = Schema{\n\tType: 1,\n\tElem: &Schema{\n\t\tType: 2,\n\t},\n\tValidateFunc: nil,\n\tOptional:     true,\n}\n",
			fixed: "var _ = Schema{\n\tType:         1,\n\tOptional:     true,\n\tValidateFunc: nil,\n\tElem: &Schema{\n\t\tType: 2,\n\t},\n}\n",
			want:  []string{"schema field Optional should come before Elem"},
		},
		"one fix for a property and the one inside it": {
			src:   "var _ = Schema{\n\tElem: &Schema{\n\t\tOptional: true,\n\t\tType:     2,\n\t},\n\tType: 1,\n}\n",
			fixed: "var _ = Schema{\n\tType: 1,\n\tElem: &Schema{\n\t\tType:     2,\n\t\tOptional: true,\n\t},\n}\n",
			want:  []string{"schema field Type should come before Elem", "schema field Type should come before Optional"},
		},
		"a field the order does not name stays put": {
			src:   "var _ = Schema{\n\tOther:    1,\n\tOptional: true,\n\tExtra:    2,\n\tType:     1,\n}\n",
			fixed: "var _ = Schema{\n\tOther:    1,\n\tType:     1,\n\tExtra:    2,\n\tOptional: true,\n}\n",
			want:  []string{"schema field Type should come before Optional"},
		},
		"two fields sharing a line cannot be moved": {
			src:   "var _ = Schema{\n\tOptional: true, Type: 1,\n}\n",
			fixed: "var _ = Schema{\n\tOptional: true, Type: 1,\n}\n",
			want:  []string{"schema field Type should come before Optional"},
		},
		"nor can a field on the closing line": {
			src:   "var _ = Schema{\n\tOptional: true,\n\tType:     1}\n",
			fixed: "var _ = Schema{\n\tOptional: true,\n\tType:     1}\n",
			want:  []string{"schema field Type should come before Optional"},
		},
	} {
		fset := token.NewFileSet()
		src := "package p\n\n" + tc.src
		file, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}

		var properties []property
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.CompositeLit); ok {
				properties = append(properties, property{lit: lit, want: pluginOrder})
			}

			return true
		})

		var got []string
		fixed := src
		found := check(fset, func(string) ([]byte, error) { return []byte(src), nil }, properties)
		for _, f := range slices.Backward(found) {
			got = append([]string{f.message}, got...)
			if f.fix != "" {
				fixed = fixed[:fset.Position(f.lit.Lbrace).Offset] + f.fix + fixed[fset.Position(f.lit.Rbrace).Offset+1:]
			}
		}

		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %q, want %q", name, got, tc.want)
		}
		if fixed != "package p\n\n"+tc.fixed {
			t.Errorf("%s: fixed to\n%s\nwant\n%s", name, fixed, "package p\n\n"+tc.fixed)
		}
	}

	// an order given in the settings replaces the default; the orders are package state, so this cannot run beside anything else
	plugin, framework := pluginOrder, frameworkOrder
	for _, field := range []string{"Type", "Description"} {
		if err := Analyzer.Flags.Set("order-plugin", field); err != nil {
			t.Fatal(err)
		}
	}

	if !slices.Equal(pluginOrder, []string{"Type", "Description"}) {
		t.Errorf("order-plugin given in the settings: got %q", pluginOrder)
	}

	for _, field := range []string{"Type", "not a field"} {
		if err := Analyzer.Flags.Set("order-plugin", field); err == nil {
			t.Errorf("order-plugin %q should be refused", field)
		}
	}

	// order and order-framework are two names for the one setting
	for name, field := range map[string]string{"order": "Required", "order-framework": "Description"} {
		if err := Analyzer.Flags.Set(name, field); err != nil {
			t.Fatal(err)
		}
	}

	if len(frameworkOrder) != 2 || !slices.Contains(frameworkOrder, "Required") || !slices.Contains(frameworkOrder, "Description") {
		t.Errorf("order and order-framework given in the settings: got %q", frameworkOrder)
	}
	pluginOrder, frameworkOrder = plugin, framework
}
