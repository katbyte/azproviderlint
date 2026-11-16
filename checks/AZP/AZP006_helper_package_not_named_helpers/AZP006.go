// Package AZP006 defines an analyzer that reports a service's shared code kept in a package named helper where the provider uses helpers.
package AZP006

import (
	"go/ast"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a service keeps its shared code in `helpers`, not `helper`. Both names are in use, and a service that ends up with both
// leaves every reader guessing which one holds what.
var Analyzer = &analysis.Analyzer{
	Name: "AZP006",
	Doc:  "check that a service's shared code is in a package named helpers, not helper",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP006_helper_package_not_named_helpers/README.md",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// internal/services/<service>/helper, and nothing deeper
	_, within, ok := strings.Cut(pass.Pkg.Path(), "/internal/services/")
	if !ok || strings.Count(within, "/") != 1 || path.Base(within) != "helper" {
		return nil, nil
	}

	// one report for the package, on its first file by name; the tests are a package of their own and say nothing new
	var first *ast.File
	for _, file := range pass.Files {
		name := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		if first == nil || name < pass.Fset.Position(first.Pos()).Filename {
			first = file
		}
	}

	if first == nil {
		return nil, nil
	}

	dir := filepath.Dir(pass.Fset.Position(first.Pos()).Filename)
	if st, err := os.Stat(filepath.Join(dir, "..", "helpers")); err == nil && st.IsDir() {
		pass.Reportf(first.Name.Pos(), "this service has both a helper and a helpers package: move this one into helpers")
		return nil, nil
	}

	pass.Reportf(first.Name.Pos(), "a service's shared code belongs in a package named helpers, not helper")

	return nil, nil
}
