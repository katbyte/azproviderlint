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
}
