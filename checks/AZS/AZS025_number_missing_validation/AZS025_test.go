package AZS025

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS025(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs025", "azs025/migration")

	// when strict, a number only checked for not being zero or negative is reported too; the setting is package state, so this cannot run beside
	// anything else
	if err := Analyzer.Flags.Set("strict", "true"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs025/strict")
	strict = false
}
