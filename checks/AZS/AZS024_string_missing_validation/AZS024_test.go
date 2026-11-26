package AZS024

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS024(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs024", "azs024/migration")

	// when strict, a string only checked for not being empty is reported too; the setting is package state, so this cannot run beside anything else
	if err := Analyzer.Flags.Set("strict", "true"); err != nil {
		t.Fatal(err)
	}
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs024/strict")
	strict = false
}
