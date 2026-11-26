package AZS027

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS027(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs027", "azs027/migration")

	// the options are package state, so this cannot run beside anything else
	for _, option := range [][2]string{{"first", "display_name"}, {"first", "name"}, {"last", "timeouts"}} {
		if err := Analyzer.Flags.Set(option[0], option[1]); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs027/options")

	first, last, givenFirst, givenLast = []string{"name", "resource_group_name", "location"}, []string{"tags"}, false, false
}
