package description_test

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

// Two rules that would rewrite the same words cannot both fix them, once every rule is held to one fix per place: the first keeps its fix, the
// second reports without one.
func TestReportOneFixPerPlace(t *testing.T) {
	t.Parallel()

	rule := func(name, word string) *analysis.Analyzer {
		return &analysis.Analyzer{
			Name:     name,
			Doc:      "report " + word,
			Requires: []*analysis.Analyzer{description.Analyzer},
			Run: func(pass *analysis.Pass) (any, error) {
				for _, found := range description.All(pass) {
					if at := strings.Index(found.Text, word); at >= 0 {
						found.Report(pass, description.Finding{Start: at, End: at + len(word), Message: name, Fix: "Remove it"})
					}
				}

				return nil, nil
			},
		}
	}

	_, filename, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(filename), "testdata")

	fixes := func(a *analysis.Analyzer) int {
		n := 0
		for _, result := range analysistest.Run(discard{t}, testdata, astx.OneFixPerPlace([]*analysis.Analyzer{a})[0], "claims") {
			for _, d := range result.Diagnostics {
				n += len(d.SuggestedFixes)
			}
		}

		return n
	}

	// "Thing, please" and "please" share bytes; "name" shares none with either
	for _, tc := range []struct {
		name, word string
		want       int
	}{
		{"first", "Thing, please", 1},
		{"second", "please", 0},
		{"third", "name", 1},
		{"first", "Thing, please", 1},
	} {
		if got := fixes(rule(tc.name, tc.word)); got != tc.want {
			t.Errorf("%s on %q: %d fixes, want %d", tc.name, tc.word, got, tc.want)
		}
	}
}

// discard lets analysistest run without the fixture's "want" comments, since the test counts fixes itself.
type discard struct{ *testing.T }

func (discard) Errorf(string, ...any) {}
