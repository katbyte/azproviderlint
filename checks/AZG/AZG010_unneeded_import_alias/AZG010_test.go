package AZG010

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZG010(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg010")
	analysistest.Run(t, dir, Analyzer, "azg010testfile")
	analysistest.Run(t, dir, Analyzer, "azg010nogenerated")

	// the flags are package state read during run, so flag fixtures must run sequentially
	// within the same test rather than as parallel siblings
	for _, pattern := range []string{`^v\d{4}_`, `^devices$`} {
		if err := Analyzer.Flags.Set("ignore", pattern); err != nil {
			t.Fatal(err)
		}
	}

	analysistest.Run(t, dir, Analyzer, "azg010ignore")
	ignore = nil

	for _, aliases := range []string{"s", "networkValidate, iothub"} {
		if err := Analyzer.Flags.Set("allow", aliases); err != nil {
			t.Fatal(err)
		}
	}

	analysistest.Run(t, dir, Analyzer, "azg010allow")
	allow = nil

	// the main package go test generates is skipped even when generated files are checked
	checkGenerated = true
	analysistest.Run(t, dir, Analyzer, "azg010generated")
	analysistest.Run(t, dir, Analyzer, "azg010.test")
	checkGenerated = false

	ownPackage = true
	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg010ownpackage")
	ownPackage = false

	mismatched = false
	analysistest.RunWithSuggestedFixes(t, dir, Analyzer, "azg010nomismatched")
	mismatched = true

	if err := Analyzer.Flags.Set("ignore", "("); err == nil {
		t.Fatal("expected an error for an invalid ignore pattern")
	}
}
