package AZG012

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG012(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg012")

	maxStatements, maxLines = 2, 4
	analysistest.Run(t, dir, Analyzer, "azg012short")
	maxStatements, maxLines = 5, 10
}
