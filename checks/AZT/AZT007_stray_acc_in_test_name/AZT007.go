// Package AZT007 defines an analyzer that reports acceptance test names carrying a second
// Acc after the TestAcc prefix.
package AZT007

import (
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a TestAcc name does not contain Acc again as a word of its own:
// TestAccThing_storageAccBehindFirewall. Acc is the marker that makes a function an
// acceptance test, so a second one reads as a mistake, and as an abbreviation of account it
// is unclear. Account and Access are whole words and are left alone.
var Analyzer = &analysis.Analyzer{
	Name: "AZT007",
	Doc:  "check that acceptance test names do not carry a second Acc after the TestAcc prefix",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT007_stray_acc_in_test_name/README.md",
	Run:  run,
}

// strayAcc matches Acc as a whole word inside a camel-case name: followed by an upper-case
// letter, an underscore or the end, so Account and Access are not matched.
var strayAcc = regexp.MustCompile(`Acc([A-Z_]|$)`)

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Recv != nil {
				continue
			}

			rest, ok := strings.CutPrefix(fd.Name.Name, "TestAcc")
			if !ok {
				rest, ok = strings.CutPrefix(fd.Name.Name, "testAcc")
			}

			if ok && strayAcc.MatchString(rest) {
				pass.Reportf(fd.Name.Pos(), "%s has a second Acc in its name, spell the word out so it is not read as another acceptance marker", fd.Name.Name)
			}
		}
	}

	return nil, nil
}
