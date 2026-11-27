package AZS026

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS026(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs026", "azs026/migration")

	// the endings are package state, so this cannot run beside anything else
	defaults := endings
	for _, ending := range []string{"_enabled", "_required"} {
		if err := Analyzer.Flags.Set("ending", ending); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs026/options")

	endings, givenEndings = defaults, false
}
