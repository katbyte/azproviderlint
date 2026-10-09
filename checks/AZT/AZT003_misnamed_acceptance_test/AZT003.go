// Package AZT003 defines an analyzer that reports test functions named TestAcc that never
// run an acceptance test.
package AZT003

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/acctest"
)

// Analyzer checks that every `TestAcc...` function reaches the acceptance test harness.
// Acceptance runs select tests by that prefix, so a unit test carrying it is scheduled and
// reported as an acceptance test although it never touches Azure.
var Analyzer = &analysis.Analyzer{
	Name:     "AZT003",
	Doc:      "check that test functions named TestAcc run an acceptance test",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT003_misnamed_acceptance_test/README.md",
	Requires: []*analysis.Analyzer{acctest.Analyzer},
	Run:      run,
}

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
			// the bare prefix, as test runners match it: TestAccountName is picked up too
			if !ok || fd.Body == nil || fd.Recv != nil || !strings.HasPrefix(fd.Name.Name, "TestAcc") {
				continue
			}

			fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok || runs[fn] != nil {
				continue
			}

			// Acc that starts a word (TestAccountName) cannot simply be dropped
			rest := strings.TrimPrefix(fn.Name(), "TestAcc")
			if rest != "" && unicode.IsLower(rune(rest[0])) {
				pass.Reportf(fd.Name.Pos(), "%s does not run an acceptance test, rename it so it does not start with TestAcc and is not picked up as one", fn.Name())
				continue
			}

			pass.Reportf(fd.Name.Pos(), "%s does not run an acceptance test, name it Test%s so it is not picked up as one", fn.Name(), rest)
		}
	}

	return nil, nil
}
