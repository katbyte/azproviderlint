package AZG009

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG009(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg009")
}

//nolint:paralleltest // mutates the package-level tests flag; must finish before parallel tests resume
func TestAZG009SkipTests(t *testing.T) {
	if err := Analyzer.Flags.Set("tests", "false"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("tests", "true") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg009notests")
}
