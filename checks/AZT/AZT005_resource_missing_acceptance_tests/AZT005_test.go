package AZT005

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZT005(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.Run(t, dir, Analyzer, "azt005")

	disabled = map[string]bool{"requiresImport": true, "complete": true, "update": true}
	analysistest.Run(t, dir, Analyzer, "azt005disable")
	disabled = map[string]bool{}
}
