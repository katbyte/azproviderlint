// Package AZT001 defines an analyzer that reports acceptance test files (for resources,
// data sources, actions and ephemeral resources) that do not use an external _test package.
package AZT001

import (
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that acceptance test files use an external test package
// (`package foo_test` rather than `package foo`) to prevent a circular dependency between
// the service package and the acceptance test framework.
var Analyzer = &analysis.Analyzer{
	Name: "AZT001",
	Doc:  "check that acceptance test files (resources, data sources, actions, ephemeral resources, including generated and list tests) use an external _test package to prevent circular dependencies",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT001_acceptance_test_external_package/README.md",
	Run:  run,
}

// Anchored to the file name end: "resource"/"data_source" elsewhere in the name does not
// make a file an acceptance test — parse/validate unit tests like
// resource_group_assignment_test.go or storage_queue_resource_manager_id_test.go must not
// match. Covers plain (_resource_test.go), list (_resource_list_test.go), and generated
// (_resource_gen_test.go, _resource_identity_gen_test.go) variants of each kind.
var acceptanceTestFile = regexp.MustCompile(`(_resource|_data_source|_action|_ephemeral)(_list)?(_identity)?(_gen)?_test\.go$`)

// The uppercase requirement after TestAcc keeps unit test functions like TestAccountName
// from counting as acceptance tests.
var acceptanceTestFunc = regexp.MustCompile(`^TestAcc[A-Z]`)

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		filename := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if !acceptanceTestFile.MatchString(filename) {
			continue
		}

		if strings.HasSuffix(file.Name.Name, "_test") {
			continue
		}

		// A matching name alone is not proof (e.g. a tool's check_resource_test.go unit
		// tests) — the file must also contain an acceptance test function.
		hasAccFunc := false
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && acceptanceTestFunc.MatchString(fn.Name.Name) {
				hasAccFunc = true
				break
			}
		}

		if !hasAccFunc {
			continue
		}

		pass.Reportf(file.Name.Pos(), "acceptance test files must use a _test package to prevent circular dependencies")
	}

	return nil, nil
}
