package AZP001

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP001(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azp001")
}
