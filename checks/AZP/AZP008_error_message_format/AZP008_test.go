package AZP008

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP008(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azp008")

	// the options are package state, so none of this can run beside anything else
	for _, option := range [][2]string{{"error-verb", "%w"}, {"quoted-error", "%q"}} {
		if err := Analyzer.Flags.Set(option[0], option[1]); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azp008/chosen")

	// each kind can be switched off
	for _, option := range [][2]string{{"capital", "false"}, {"error-verb", "none"}, {"contraction", "false"}, {"quoted-error", "none"}} {
		if err := Analyzer.Flags.Set(option[0], option[1]); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azp008/off")

	capital, errorVerb, contraction, quote = true, "%+v", true, "backticks"
}
