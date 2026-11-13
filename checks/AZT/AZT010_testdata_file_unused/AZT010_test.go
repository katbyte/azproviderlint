package AZT010

import (
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZT010(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")
	services := "github.com/hashicorp/terraform-provider-azurerm/internal/services/"

	analysistest.Run(t, dir, Analyzer, services+"azt010/...", services+"azt010used", services+"azt010fixtures", services+"azt010tests")

	ignore = []*regexp.Regexp{regexp.MustCompile(`\.md$`)}
	analysistest.Run(t, dir, Analyzer, services+"azt010ignore")
	ignore = nil
}
