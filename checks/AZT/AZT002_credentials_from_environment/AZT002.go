// Package AZT002 defines an analyzer that reports tests reading provider
// credentials from the environment instead of provisioning their own identity.
package AZT002

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks test files for `os.Getenv("ARM_CLIENT_ID")` / `os.Getenv("ARM_CLIENT_SECRET")` /
// `os.Getenv("ARM_CLIENT_SECRET_ALT")`. Test configurations should not reuse the testing
// credentials - instead create an azurerm_user_assigned_identity as part of the test
// configuration with as minimal permissions as possible, which is cleaned up with the test.
var Analyzer = &analysis.Analyzer{
	Name: "AZT002",
	Doc:  "check for tests reading provider credentials from the environment instead of creating a user assigned identity",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT002_credentials_from_environment/README.md",
	Run:  run,
}

var bannedVariables = map[string]bool{
	"ARM_CLIENT_ID":         true,
	"ARM_CLIENT_SECRET":     true,
	"ARM_CLIENT_SECRET_ALT": true,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		// Only test files are checked: the provider runtime and the acceptance test
		// framework legitimately read the credentials they authenticate with.
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Getenv" {
				return true
			}

			pkg, ok := sel.X.(*ast.Ident)
			if !ok || pkg.Name != "os" {
				return true
			}

			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}

			name, err := strconv.Unquote(lit.Value)
			if err != nil || !bannedVariables[name] {
				return true
			}

			pass.Reportf(call.Pos(), "tests should not obtain credentials via os.Getenv(%q), create an azurerm_user_assigned_identity with minimal permissions as part of the test configuration instead", name)
			return true
		})
	}

	return nil, nil
}
