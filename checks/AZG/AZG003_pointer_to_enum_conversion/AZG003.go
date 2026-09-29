// Package AZG003 defines an analyzer that reports pointer.To being used with an explicit
// go-azure-sdk enum type conversion (pointer.To(sdk.Enum(v))) where the generic
// pointer.ToEnum[sdk.Enum](v) helper should be used instead.
package AZG003

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/azuresdk"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// pointerPkgPath is the import path of the go-azure-helpers pointer package.
	pointerPkgPath = "github.com/hashicorp/go-azure-helpers/lang/pointer"
)

// Analyzer checks for `pointer.To(sdk.SomeEnum(v))` calls that convert a go-azure-sdk enum
// type with an explicit conversion. The generic `pointer.ToEnum[sdk.SomeEnum](v)` helper is
// clearer and type-safe, so those call sites should use it instead.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG003",
	Doc:      "check for pointer.To with an explicit go-azure-sdk enum conversion that should use pointer.ToEnum instead",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG003_pointer_to_enum_conversion/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return
		}

		argCall, ok := call.Args[0].(*ast.CallExpr)
		if !ok || len(argCall.Args) != 1 {
			return
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || sel.Sel.Name != "To" {
			return
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "pointer" {
			return
		}

		pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
		if !ok || pkgName.Imported().Path() != pointerPkgPath {
			return
		}

		// Unalias before the *types.Named assertion so SDK enums referenced through a `=` alias
		// (whose type is *types.Alias, not *types.Named) are still matched.
		named, ok := types.Unalias(pass.TypesInfo.TypeOf(argCall.Fun)).(*types.Named)
		if !ok || !azuresdk.IsEnumType(named) {
			return
		}

		valueIsString := convertedValueIsString(pass, argCall.Args[0])

		message := fmt.Sprintf("pointer.To with an explicit go-azure-sdk enum conversion should use pointer.ToEnum[%s](...) instead", named.Obj().Name())
		if !valueIsString {
			message = fmt.Sprintf("pointer.To with an explicit go-azure-sdk enum conversion should use pointer.ToEnum[%s](string(...)) instead", named.Obj().Name())
		}

		pass.Report(analysis.Diagnostic{
			Pos:            call.Pos(),
			Message:        message,
			SuggestedFixes: suggestedFixes(pass, sel, argCall, valueIsString),
		})
	})

	return nil, nil
}

// suggestedFixes rewrites `pointer.To(sdk.Enum(v))` into `pointer.ToEnum[sdk.Enum](v)`: the
// `To` selector becomes `ToEnum[<type as written>]`, and the explicit conversion is replaced
// by its bare argument — or, when the argument is not assignable to string, only the
// conversion's type expression is replaced with `string`, turning `sdk.Enum(v)` into
// `string(v)` in place.
func suggestedFixes(pass *analysis.Pass, sel *ast.SelectorExpr, argCall *ast.CallExpr, valueIsString bool) []analysis.SuggestedFix {
	var typeBuf bytes.Buffer
	if err := printer.Fprint(&typeBuf, pass.Fset, argCall.Fun); err != nil {
		return nil
	}

	edits := []analysis.TextEdit{{
		Pos:     sel.Sel.Pos(),
		End:     sel.Sel.End(),
		NewText: []byte("ToEnum[" + typeBuf.String() + "]"),
	}}

	if valueIsString {
		var argBuf bytes.Buffer
		if err := printer.Fprint(&argBuf, pass.Fset, argCall.Args[0]); err != nil {
			return nil
		}
		edits = append(edits, analysis.TextEdit{Pos: argCall.Pos(), End: argCall.End(), NewText: argBuf.Bytes()})
	} else {
		edits = append(edits, analysis.TextEdit{Pos: argCall.Fun.Pos(), End: argCall.Fun.End(), NewText: []byte("string")})
	}

	return []analysis.SuggestedFix{{
		Message:   "Replace pointer.To with the explicit enum conversion by pointer.ToEnum",
		TextEdits: edits,
	}}
}

// convertedValueIsString reports whether expr, the value being wrapped in an explicit enum
// conversion, can be passed directly to pointer.ToEnum (whose parameter is string) without an
// added string(...) conversion.
func convertedValueIsString(pass *analysis.Pass, expr ast.Expr) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return true
	}

	return types.AssignableTo(types.Default(pass.TypesInfo.TypeOf(expr)), types.Typ[types.String])
}
