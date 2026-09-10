package AZP003

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP003(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.Run(t, dir, Analyzer, "azp003")

	// the ignore-sensitive flag is package state read during run, so the flag-on fixtures
	// must run sequentially within the same test rather than as a parallel sibling
	ignoreSensitive = true
	defer func() { ignoreSensitive = false }()

	analysistest.Run(t, dir, Analyzer, "azp003sensitive")
}
