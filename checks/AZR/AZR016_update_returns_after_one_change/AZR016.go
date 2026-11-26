// Package AZR016 defines an analyzer that reports an update that returns after handling one changed property, before it has looked at the rest.
package AZR016

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks an update that handles its changed properties one after another. One that returns when the first is done never looks at the
// ones below it, so two changes made in the same apply leave the second unsent, and the next plan shows it again.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR016",
	Doc:      "check that an update does not return after handling one changed property while others are still to be looked at",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR016_update_returns_after_one_change/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// changed gives what a piece of code asks about with HasChange: the property when it is written out, and whether it asks at all
	changed := func(n ast.Node) (string, bool) {
		property, asks := "", false
		ast.Inspect(n, func(n ast.Node) bool {
			call, isCall := n.(*ast.CallExpr)
			if !isCall || asks {
				return !asks
			}

			fn := astx.CalledFunc(pass, call)
			if fn == nil || fn.Signature().Recv() == nil || !strings.HasPrefix(fn.Name(), "HasChange") {
				return true
			}

			on := fn.Signature().Recv().Type()
			if ptr, isPointer := types.Unalias(on).(*types.Pointer); isPointer {
				on = ptr.Elem()
			}
			if named, isNamed := types.Unalias(on).(*types.Named); !isNamed || named.Obj().Name() != "ResourceData" {
				return true
			}

			asks = true
			if len(call.Args) > 0 {
				if name, isText := pass.TypesInfo.Types[call.Args[0]]; isText && name.Value != nil && name.Value.Kind() == constant.String {
					property = strconv.Quote(constant.StringVal(name.Value))
				}
			}

			return false
		})

		return property, asks
	}

	insp.Preorder([]ast.Node{(*ast.BlockStmt)(nil)}, func(n ast.Node) {
		block, isBlock := n.(*ast.BlockStmt)
		if !isBlock {
			return
		}

		for i, stmt := range block.List {
			check, isIf := stmt.(*ast.IfStmt)
			if !isIf || check.Else != nil || len(check.Body.List) == 0 {
				continue
			}

			ret, isReturn := check.Body.List[len(check.Body.List)-1].(*ast.ReturnStmt)
			handled, asks := changed(check.Cond)
			if !asks || !isReturn || len(ret.Results) == 0 {
				continue
			}

			// returning an error made on the spot is refusing the change, not finishing it
			result := ret.Results[len(ret.Results)-1]
			if call, isCall := result.(*ast.CallExpr); isCall {
				if fn := astx.CalledFunc(pass, call); fn != nil && (fn.FullName() == "fmt.Errorf" || fn.FullName() == "errors.New") {
					continue
				}
			} else if !astx.IsNilValue(pass, result) {
				continue
			}

			for _, later := range block.List[i+1:] {
				missed, isMissed := changed(later)
				if !isMissed {
					continue
				}

				if handled == "" {
					handled = "this change"
				}
				if missed == "" {
					missed = "the properties below"
				}
				pass.Reportf(ret.Pos(), "this returns once %s is handled, so a change to %s in the same apply is never sent", handled, missed)

				break
			}
		}
	})

	return nil, nil
}
