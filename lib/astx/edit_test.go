package astx_test

import (
	"go/ast"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// A fix that rewrites an import takes the whole file, so no other rule's fix can go on using the name that import gave; two fixes that share no
// bytes and leave the imports alone are both kept.
func TestOneFixPerPlace(t *testing.T) {
	t.Parallel()

	// each rule fixes the declaration it is named after: the import, or one of the two functions
	rule := func(name, fixes string) *analysis.Analyzer {
		return astx.OneFixPerPlace([]*analysis.Analyzer{{
			Name: name,
			Doc:  "fix " + fixes,
			Run: func(pass *analysis.Pass) (any, error) {
				for _, decl := range pass.Files[0].Decls {
					var node ast.Node
					switch decl := decl.(type) {
					case *ast.GenDecl:
						if fixes == "import" {
							node = decl.Specs[0]
						}

					case *ast.FuncDecl:
						if fixes == decl.Name.Name {
							node = decl.Name
						}
					}

					if node != nil {
						pass.Report(analysis.Diagnostic{Pos: node.Pos(), Message: name, SuggestedFixes: []analysis.SuggestedFix{{Message: "Rewrite it", TextEdits: []analysis.TextEdit{{Pos: node.Pos(), End: node.End()}}}}})
					}
				}

				return nil, nil
			},
		}})[0]
	}

	_, filename, _, _ := runtime.Caller(0)
	for _, tc := range []struct {
		name, fixes string
		want        int
	}{
		{"first", "upper", 1},
		{"second", "lower", 1},
		{"third", "import", 0},
		{"first", "upper", 1},
	} {
		got := 0
		for _, result := range analysistest.Run(discard{t}, filepath.Join(filepath.Dir(filename), "testdata"), rule(tc.name, tc.fixes), "onefix") {
			for _, d := range result.Diagnostics {
				got += len(d.SuggestedFixes)
			}
		}

		if got != tc.want {
			t.Errorf("%s fixing %s: %d fixes, want %d", tc.name, tc.fixes, got, tc.want)
		}
	}
}

// discard lets analysistest run without "want" comments, since the test counts the fixes itself.
type discard struct{ *testing.T }

func (discard) Errorf(string, ...any) {}
