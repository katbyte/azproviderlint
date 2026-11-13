package AZP006

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP006(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	services := "github.com/hashicorp/terraform-provider-azurerm/internal/services/"

	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, services+"azp006/...", services+"azp006only/...", services+"azp006fine/...", "elsewhere/helper")
}
