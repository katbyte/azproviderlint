// Package AZT004 defines an analyzer that reports test functions that run an acceptance test
// but are not named TestAcc.
package AZT004

import (
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/acctest"
)

// Analyzer checks that every test function reaching the acceptance test harness is named
// `TestAcc...`. Acceptance runs select tests by that prefix, so one named otherwise is never
// run against Azure, however complete it is.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT004",
	Doc:      "check that test functions running an acceptance test are named TestAcc",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT004_acceptance_test_not_named_test_acc/README.md",
	Requires: []*analysis.Analyzer{acctest.Analyzer},
	Run:      run,
}

// strayAcc matches Acc as a whole word inside a camel-case name: followed by an upper-case
// letter, an underscore or the end, so Account and Access are not matched.
var strayAcc = regexp.MustCompile(`Acc([A-Z_]|$)`)

func run(pass *analysis.Pass) (any, error) {
	runs, ok := pass.ResultOf[acctest.Analyzer].(acctest.Funcs)
	if !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			// what go test runs: a top-level Test function; its signature is enforced by the compiler
			if !ok || fd.Body == nil || fd.Recv != nil || !strings.HasPrefix(fd.Name.Name, "Test") || strings.HasPrefix(fd.Name.Name, "TestAcc") {
				continue
			}

			fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok || runs[fn] == nil {
				continue
			}

			// Acc already sits somewhere in the name as its own word (TestWebAppAccActiveSlot_basic),
			// so prefixing it would double up; Account and Access do not count
			rest := strings.TrimPrefix(fn.Name(), "Test")
			if strayAcc.MatchString(rest) {
				pass.Reportf(fd.Name.Pos(), "%s runs an acceptance test, rename it to start with TestAcc so acceptance runs pick it up", fn.Name())
				continue
			}

			pass.Reportf(fd.Name.Pos(), "%s runs an acceptance test, name it TestAcc%s so acceptance runs pick it up", fn.Name(), rest)
		}
	}

	return nil, nil
}
