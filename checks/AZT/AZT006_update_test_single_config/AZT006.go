// Package AZT006 defines an analyzer that reports update tests that do not apply two
// different configs.
package AZT006

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a test named for an update step applies at least two different
// configs. An update test with one config, or the same config twice, only proves the
// resource can be created, and the update path it is named for is never run.
var Analyzer = &analysis.Analyzer{
	Name: "AZT006",
	Doc:  "check that update tests apply two different configs",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT006_update_test_single_config/README.md",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil || fd.Recv != nil || !isUpdateTest(fd.Name.Name) {
				continue
			}

			// the configs the test applies, by their source text: Config keys of TestStep literals
			var configs []string
			distinct := map[string]bool{}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				cl, ok := n.(*ast.CompositeLit)
				if !ok {
					return true
				}
				named, ok := types.Unalias(pass.TypesInfo.TypeOf(cl)).(*types.Named)
				if !ok || named.Obj().Name() != "TestStep" {
					return true
				}
				for _, elt := range cl.Elts {
					kv, ok := elt.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "Config" {
						text := types.ExprString(kv.Value)
						configs = append(configs, text)
						distinct[text] = true
					}
				}
				return true
			})

			// no steps here: the test delegates to a helper, which is checked on its own
			switch {
			case len(configs) == 0 || len(distinct) >= 2:
			case len(configs) == 1:
				pass.Reportf(fd.Name.Pos(), "%s applies only one config, an update test needs a second, different config to apply", fd.Name.Name)
			default:
				pass.Reportf(fd.Name.Pos(), "%s applies the same config every time, an update test needs a second, different config to apply", fd.Name.Name)
			}
		}
	}

	return nil, nil
}

// isUpdateTest reports whether a test name is for an update step: a segment after the first
// underscore that starts with update, as in TestAccThing_update, TestAccThing_updateSku and
// testAccThing_list_update, but not TestAccThing_autoUpdate.
func isUpdateTest(name string) bool {
	_, rest, ok := strings.Cut(name, "_")
	if !ok || !(strings.HasPrefix(name, "TestAcc") || strings.HasPrefix(name, "testAcc")) {
		return false
	}
	for _, seg := range strings.Split(rest, "_") {
		if strings.HasPrefix(strings.ToLower(seg), "update") {
			return true
		}
	}
	return false
}
