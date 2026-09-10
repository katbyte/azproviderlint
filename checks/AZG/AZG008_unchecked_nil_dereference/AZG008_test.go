package AZG008

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG008(t *testing.T) {
	t.Parallel()

	analysistest.RunWithSuggestedFixes(t, analysistest.TestData(), Analyzer, "azg008")
}

//nolint:paralleltest // mutates the package-level tests flag; must finish before parallel tests resume
func TestAZG008SkipTests(t *testing.T) {
	if err := Analyzer.Flags.Set("tests", "false"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("tests", "true") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg008notests")
}

//nolint:paralleltest // mutates the package-level fix-with flag; must finish before parallel tests resume
func TestAZG008FixWithNone(t *testing.T) {
	if err := Analyzer.Flags.Set("fix-with", "none"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("fix-with", "pointer.From") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg008fixnone")
}

//nolint:paralleltest // mutates the package-level requestbody flag; must finish before parallel tests resume
func TestAZG008RequestBody(t *testing.T) {
	if err := Analyzer.Flags.Set("requestbody", "true"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("requestbody", "false") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg008requestbody")
}
