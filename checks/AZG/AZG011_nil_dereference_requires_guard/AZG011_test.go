package AZG011

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG011(t *testing.T) {
	t.Parallel()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg011")
}

//nolint:paralleltest // mutates the package-level tests flag; must finish before parallel tests resume
func TestAZG011SkipTests(t *testing.T) {
	if err := Analyzer.Flags.Set("tests", "false"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("tests", "true") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg011notests")
}

//nolint:paralleltest // mutates the package-level exclude-types flag; must finish before parallel tests resume
func TestAZG011ExcludeTypes(t *testing.T) {
	if err := Analyzer.Flags.Set("exclude-types", "*Client, azg011excludetypes.account"); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = Analyzer.Flags.Set("exclude-types", "") }()

	analysistest.Run(t, analysistest.TestData(), Analyzer, "azg011excludetypes")
}
