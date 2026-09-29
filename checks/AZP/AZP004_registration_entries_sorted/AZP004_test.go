package AZP004

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP004(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azp004")
}

//nolint:paralleltest // mutates the package-level generated flag; must finish before parallel tests resume
func TestAZP004SkipGenerated(t *testing.T) {
	if err := Analyzer.Flags.Set("generated", "false"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("generated", "true") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azp004nogen")
}
