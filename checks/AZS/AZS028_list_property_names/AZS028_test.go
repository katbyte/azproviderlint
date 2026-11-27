package AZS028

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS028(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs028", "azs028/migration")

	// the options are package state, so this cannot run beside anything else
	for _, option := range [][2]string{{"block-singular", "false"}, {"singular-word", "settings"}, {"plural-word", "sheep"}, {"trailing-word", "dark"}} {
		if err := Analyzer.Flags.Set(option[0], option[1]); err != nil {
			t.Fatal(err)
		}
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs028/options")

	blockSingular, singular, plural, trailing = true, singular[:len(singular)-1], plural[:len(plural)-1], trailing[:len(trailing)-1]
}
