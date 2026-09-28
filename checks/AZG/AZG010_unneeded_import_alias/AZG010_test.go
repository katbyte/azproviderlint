package AZG010

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG010(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg010")
	analysistest.Run(t, dir, Analyzer, "azg010testfile")

	// the flags are package state read during run, so flag fixtures must run sequentially
	// within the same test rather than as parallel siblings
	allowRenames = true
	analysistest.Run(t, dir, Analyzer, "azg010renames")
	allowRenames = false

	if err := Analyzer.Flags.Set("ignore", `^v\d{4}_`); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, dir, Analyzer, "azg010ignore")
	ignore = nil
}
