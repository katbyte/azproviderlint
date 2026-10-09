package AZT006

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZT006(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "github.com/hashicorp/terraform-provider-azurerm/internal/services/azt006")
}
