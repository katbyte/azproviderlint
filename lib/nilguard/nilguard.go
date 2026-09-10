// Package nilguard is the shared guard engine behind AZG008 and AZG009: it decides whether a
// pointer-typed variable or selector chain is provably non-nil at a given point in a function
// body.
package nilguard

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

// ForEachFunc visits every function body in the pass exactly once — package-level function
// literals (var f = func() {...}) are reached through their own visit, nested literals through
// their enclosing FuncDecl, keeping guards outside a closure visible to dereferences inside
// it. Bodies in _test.go files are skipped when tests is false. visit receives the body, the
// function's (and receiver's) parameter objects, and a child-to-parent map covering the body.
func ForEachFunc(pass *analysis.Pass, insp *inspector.Inspector, tests bool, visit func(body *ast.BlockStmt, params map[types.Object]bool, parents map[ast.Node]ast.Node)) {
	var declRanges [][2]token.Pos
	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(n ast.Node) {
		declRanges = append(declRanges, [2]token.Pos{n.Pos(), n.End()})
	})
	inDecl := func(pos token.Pos) bool {
		for _, r := range declRanges {
			if r[0] <= pos && pos < r[1] {
				return true
			}
		}
		return false
	}

	insp.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(n ast.Node) {
		var body *ast.BlockStmt
		params := map[types.Object]bool{}
		collectParams := func(fl *ast.FieldList) {
			if fl == nil {
				return
			}
			for _, field := range fl.List {
				for _, name := range field.Names {
					if obj := pass.TypesInfo.Defs[name]; obj != nil {
						params[obj] = true
					}
				}
			}
		}
		switch fn := n.(type) {
		case *ast.FuncDecl:
			body = fn.Body
			collectParams(fn.Recv)
			collectParams(fn.Type.Params)
		case *ast.FuncLit:
			if inDecl(fn.Pos()) {
				return // visited via its enclosing FuncDecl
			}
			body = fn.Body
			collectParams(fn.Type.Params)
		}
		if body == nil {
			return
		}

		if !tests && strings.HasSuffix(pass.Fset.Position(body.Pos()).Filename, "_test.go") {
			return
		}

		visit(body, params, parentMap(body))
	})
}

// PathKey canonicalizes a variable or field chain (`props.Status`) so guard conditions and
// dereference operands can be compared; the root must resolve to a variable.
func PathKey(pass *analysis.Pass, e ast.Expr) (string, bool) {
	switch x := ast.Unparen(e).(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[x]
		if obj == nil {
			obj = pass.TypesInfo.Defs[x] // the ident being defined by `x := ...`
		}
		if v, ok := obj.(*types.Var); ok {
			return v.Id() + "@" + pass.Fset.Position(v.Pos()).String(), true
		}
	case *ast.SelectorExpr:
		if base, ok := PathKey(pass, x.X); ok {
			return base + "." + x.Sel.Name, true
		}
	}
	return "", false
}

// DerefNeedsPointer reports whether star's context needs the pointee to be addressable — so a
// pointer.From copy cannot stand in: an assignment or inc/dec target (`*x = v`, `(*x).F = v`,
// `(*x)[i] = v` on an array, `(*x)++`), the operand of & (`&*x`, `&(*x).F`), a slice of an
// array (`(*x)[:]`), or the receiver of a pointer method (`(*x).M()`). Field selections and
// array indexes propagate the requirement; a selection that itself goes through a pointer
// field, or an index into a slice or map, is addressable on its own.
func DerefNeedsPointer(pass *analysis.Pass, parents map[ast.Node]ast.Node, star *ast.StarExpr) bool {
	var outer ast.Node = star
	for {
		switch p := parents[outer].(type) {
		case *ast.ParenExpr:
			outer = p
		case *ast.SelectorExpr:
			sel := pass.TypesInfo.Selections[p]
			if p.X != outer || sel == nil || sel.Indirect() {
				return false
			}
			if sel.Kind() == types.MethodVal {
				_, ptrRecv := sel.Obj().(*types.Func).Type().(*types.Signature).Recv().Type().(*types.Pointer)
				return ptrRecv
			}
			outer = p
		case *ast.IndexExpr:
			// an array index needs the addressable array; a slice or map index is
			// addressable on its own, but writing through it after pointer.From still
			// panics on a nil map or empty slice, so the target contexts below apply to all
			if p.X != outer {
				return false
			}
			outer = p
		case *ast.SliceExpr:
			return p.X == outer && isArray(pass, p.X)
		case *ast.AssignStmt:
			for _, lhs := range p.Lhs {
				if lhs == outer {
					return true
				}
			}
			return false
		case *ast.UnaryExpr:
			return p.Op == token.AND
		case *ast.IncDecStmt:
			return true
		default:
			return false
		}
	}
}

func isArray(pass *analysis.Pass, e ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(e)
	if t == nil {
		return false
	}
	_, ok := t.Underlying().(*types.Array)
	return ok
}
