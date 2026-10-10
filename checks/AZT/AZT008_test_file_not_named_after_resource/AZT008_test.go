package AZT008

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZT008(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analysistest.Run(t, dir, Analyzer, "github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008")

	suffix = false
	analysistest.Run(t, dir, Analyzer, "github.com/hashicorp/terraform-provider-azurerm/internal/services/azt008exact")
	suffix = true
}
