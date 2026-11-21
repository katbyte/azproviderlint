package AZP007

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP007(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	testdata := filepath.Join(filepath.Dir(filename), "testdata")
	analysistest.RunWithSuggestedFixes(t, testdata, Analyzer, "azp007")

	// the consumer's own moves are package state, so this cannot run beside another test in this package
	extra = "old.example.com/docs/=new.example.com/, =skipped"
	analysistest.Run(t, testdata, Analyzer, "azp007extra")
	extra = ""
}
