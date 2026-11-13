package AZR011

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZR011(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	services := "github.com/hashicorp/terraform-provider-azurerm/internal/services/"

	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, services+"azr011", services+"azr011alias")
}
