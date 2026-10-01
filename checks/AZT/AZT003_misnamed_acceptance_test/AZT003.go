// Package AZT003 defines an analyzer that reports test functions named TestAcc that never
// run an acceptance test.
package AZT003

import (
	"go/ast"
	"go/constant"
	"go/types"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/lib/astx"
)

// Analyzer checks that every `TestAcc...` function reaches the acceptance test harness.
// Acceptance runs select tests by that prefix, so a unit test carrying it is scheduled and
// reported as an acceptance test although it never touches Azure.
var Analyzer = &analysis.Analyzer{
	Name: "AZT003",
	Doc:  "check that test functions named TestAcc run an acceptance test",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZT/AZT003_misnamed_acceptance_test/README.md",
	Run:  run,
}

// The Terraform test harnesses; a provider's own internal/acceptance package counts too.
var harnessPackages = map[string]bool{
	"github.com/hashicorp/terraform-plugin-testing/helper/resource": true,
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource":  true,
}

func run(pass *analysis.Pass) (any, error) {
	decls := map[*types.Func]*ast.FuncDecl{}
	var candidates []*types.Func
	for _, file := range pass.Files {
		isTest := strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok {
				continue
			}
			decls[fn] = fd
			// the bare prefix, as test runners match it: TestAccountName is picked up too
			if isTest && fd.Recv == nil && strings.HasPrefix(fd.Name.Name, "TestAcc") {
				candidates = append(candidates, fn)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// runs marks the functions that are acceptance tests on their own: they hand t to a
	// harness or look at TF_ACC. mentions lists the same-package functions each one calls or
	// passes around as a value (`t.Run("basic", testAccThing_basic)`).
	runs := map[*types.Func]bool{}
	mentions := map[*types.Func][]*types.Func{}
	for fn, fd := range decls {
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CallExpr:
				callee := astx.CalledFunc(pass, n)
				if callee == nil || callee.Pkg() == nil {
					break
				}
				if path := callee.Pkg().Path(); !harnessPackages[path] && !strings.HasSuffix(path, "/internal/acceptance") {
					break
				}
				for _, arg := range n.Args {
					ptr, ok := pass.TypesInfo.TypeOf(arg).(*types.Pointer)
					if !ok {
						continue
					}
					if named, ok := ptr.Elem().(*types.Named); ok && named.Obj().Name() == "T" && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == "testing" {
						runs[fn] = true
					}
				}
			case *ast.Ident:
				if used, ok := pass.TypesInfo.Uses[n].(*types.Func); ok && used.Pkg() == pass.Pkg {
					mentions[fn] = append(mentions[fn], used.Origin())
				}
			}
			// by value, so resource.EnvTfAcc and local constants count as well as the literal
			if e, ok := n.(ast.Expr); ok {
				if v := pass.TypesInfo.Types[e].Value; v != nil && v.Kind() == constant.String && constant.StringVal(v) == "TF_ACC" {
					runs[fn] = true
				}
			}
			return true
		})
	}

	// a function that mentions an acceptance test is one too
	for changed := true; changed; {
		changed = false
		for fn, used := range mentions {
			if !runs[fn] && slices.ContainsFunc(used, func(u *types.Func) bool { return runs[u] }) {
				runs[fn] = true
				changed = true
			}
		}
	}

	for _, fn := range candidates {
		if runs[fn] {
			continue
		}
		// Acc that starts a word (TestAccountName) cannot simply be dropped
		rest := strings.TrimPrefix(fn.Name(), "TestAcc")
		if rest != "" && unicode.IsLower(rune(rest[0])) {
			pass.Reportf(decls[fn].Name.Pos(),
				"%s does not run an acceptance test, rename it so it does not start with TestAcc and is not picked up as one", fn.Name())
			continue
		}
		pass.Reportf(decls[fn].Name.Pos(),
			"%s does not run an acceptance test, name it Test%s so it is not picked up as one", fn.Name(), rest)
	}

	return nil, nil
}
