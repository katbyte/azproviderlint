// Package AZG009 reports string(pointer.From(enum)) calls that should use pointer.FromEnum.
package AZG009

import (
	"go/ast"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/azuresdk"
	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "AZG009",
	Doc:      "check for string(pointer.From(...)) with a go-azure-sdk enum that should use pointer.FromEnum instead",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG009_pointer_from_enum_conversion/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return
		}
		conversion, ok := ast.Unparen(call.Fun).(*ast.Ident)
		if !ok || pass.TypesInfo.Uses[conversion] != types.Universe.Lookup("string") {
			return
		}
		from, ok := ast.Unparen(call.Args[0]).(*ast.CallExpr)
		if !ok || len(from.Args) != 1 {
			return
		}
		fun := ast.Unparen(from.Fun)
		if index, isIndex := fun.(*ast.IndexExpr); isIndex {
			fun = ast.Unparen(index.X)
		}
		sel, ok := fun.(*ast.SelectorExpr)
		if !ok {
			return
		}
		fn := astx.CalledFunc(pass, from)
		if fn == nil || fn.Name() != "From" || fn.Pkg() == nil || fn.Pkg().Path() != pointerpkg.PkgPath {
			return
		}
		named, ok := types.Unalias(pass.TypesInfo.TypeOf(from)).(*types.Named)
		if !ok || !azuresdk.IsEnumType(named) {
			return
		}

		pass.Report(analysis.Diagnostic{
			Pos:            call.Pos(),
			End:            call.End(),
			Message:        "string(pointer.From(...)) with a go-azure-sdk enum should use pointer.FromEnum(...) instead",
			SuggestedFixes: suggestedFixes(pass, call, from, sel),
		})
	})
	return nil, nil
}

// suggestedFixes rewrites `string(pointer.From(v))` into `pointer.FromEnum(v)`, or returns no
// fix when a comment inside the conversion but outside pointer.From would be lost.
func suggestedFixes(pass *analysis.Pass, call, from *ast.CallExpr, sel *ast.SelectorExpr) []analysis.SuggestedFix {
	for _, file := range pass.Files {
		if call.Pos() < file.Pos() || call.End() > file.End() {
			continue
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if comment.Pos() > call.Pos() && comment.End() < call.End() &&
					(comment.Pos() < from.Pos() || comment.End() > from.End()) {
					return nil
				}
			}
		}
	}
	source, ok := astx.SourceText(pass, from)
	if !ok {
		return nil
	}
	replacement := string(source[:sel.Sel.Pos()-from.Pos()]) + "FromEnum" + string(source[sel.Sel.End()-from.Pos():])
	return []analysis.SuggestedFix{{
		Message:   "Replace string(pointer.From(...)) with pointer.FromEnum(...)",
		TextEdits: []analysis.TextEdit{{Pos: call.Pos(), End: call.End(), NewText: []byte(replacement)}},
	}}
}
