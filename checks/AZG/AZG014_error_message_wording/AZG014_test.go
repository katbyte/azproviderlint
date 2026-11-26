package AZG014

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG014(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azg014")

	// each kind can be switched off; the kinds are package state, so this cannot run beside anything else
	for _, kind := range []string{"error-verb", "contraction", "quoted-value"} {
		if err := Analyzer.Flags.Set(kind, "false"); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azg014/off")

	for _, kind := range []string{"error-verb", "contraction", "quoted-value"} {
		if err := Analyzer.Flags.Set(kind, "true"); err != nil {
			t.Fatal(err)
		}
	}
}
