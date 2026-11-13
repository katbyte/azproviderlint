package AZP007

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZP007(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")
	services := "github.com/hashicorp/terraform-provider-azurerm/internal/services/"

	analysistest.Run(t, dir, Analyzer, services+"azp007/...")

	// the flag is package state read during run, so its fixture runs after the others rather than beside them
	if err := Analyzer.Flags.Set("allow", "helper.go, customizediff.go"); err != nil {
		t.Fatal(err)
	}

	analysistest.Run(t, dir, Analyzer, services+"azp007allow/...")
	allow = nil
}
