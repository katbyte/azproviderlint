// Package astx defines AST helper functions that are reused in multiple checks.
package astx

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"golang.org/x/tools/go/analysis"
)

func IsTrueConstant(pass *analysis.Pass, e ast.Expr) bool {
	if e == nil {
		return false
	}
	value := pass.TypesInfo.Types[e].Value
	return value != nil && value.Kind() == constant.Bool && constant.BoolVal(value)
}

// ContainsCallExpr reports whether expr contains any function call.
func ContainsCallExpr(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if _, ok := n.(*ast.CallExpr); ok {
			found = true
			return false
		}
		return true
	})
	return found
}

// IsNilValue reports whether expr denotes a nil value: the predeclared nil identifier,
// possibly wrapped in parentheses or type conversions (`error(nil)`, `[]T(nil)`).
func IsNilValue(pass *analysis.Pass, expr ast.Expr) bool {
	switch e := ast.Unparen(expr).(type) {
	case *ast.Ident:
		_, isNil := pass.TypesInfo.Uses[e].(*types.Nil)
		return isNil
	case *ast.CallExpr:
		if len(e.Args) == 1 && pass.TypesInfo.Types[e.Fun].IsType() {
			return IsNilValue(pass, e.Args[0])
		}
	}
	return false
}

// IsUseOf reports whether expr is (modulo parentheses) an identifier resolving to obj.
func IsUseOf(pass *analysis.Pass, expr ast.Expr, obj types.Object) bool {
	id, ok := ast.Unparen(expr).(*ast.Ident)
	return ok && pass.TypesInfo.Uses[id] == obj
}

// UseCount counts references to obj within node.
func UseCount(pass *analysis.Pass, node ast.Node, obj types.Object) int {
	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj {
			count++
		}
		return true
	})
	return count
}

// UnsafeToMovePast reports whether evaluating expr after stmts could yield a different value:
// a statement writes to or takes the address of a variable expr reads — directly (`y = v`) or
// as the root of an index, field, slice, or dereference target (`column[y] = v` writes column)
// — or declares a name expr references, so moving expr past it would rebind the name to the
// shadowing variable.
func UnsafeToMovePast(pass *analysis.Pass, expr ast.Expr, stmts []ast.Stmt) bool {
	operands := map[types.Object]bool{}
	names := map[string]bool{}
	ast.Inspect(expr, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if v, ok := pass.TypesInfo.Uses[id].(*types.Var); ok {
				operands[v] = true
				names[id.Name] = true
			}
		}
		return true
	})
	if len(operands) == 0 {
		return false
	}

	writes := func(target ast.Expr) bool {
		id := rootIdent(target)
		return id != nil && operands[pass.TypesInfo.Uses[id]]
	}

	// variables the expression's own calls may write through — pointer-like arguments and
	// receivers (`astutil.Apply(f, ...)` mutates *f); an intervening statement reading one
	// would observe side effects that moving the expression re-orders
	mutables := callMutables(pass, expr)

	unsafe := false
	for _, stmt := range stmts {
		// only a declaration in this same statement list can shadow a name at the moved
		// expression's new position; declarations in nested blocks end with their block
		if shadows(stmt, names) {
			return true
		}
		ast.Inspect(stmt, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.AssignStmt:
				for _, lhs := range x.Lhs {
					unsafe = unsafe || writes(lhs)
				}
			case *ast.IncDecStmt:
				unsafe = unsafe || writes(x.X)
			case *ast.UnaryExpr:
				// a taken address may be written through by whatever receives it
				if x.Op == token.AND {
					unsafe = unsafe || writes(x.X)
				}
			case *ast.RangeStmt:
				if x.Tok == token.ASSIGN {
					unsafe = unsafe || writes(x.Key) || writes(x.Value)
				}
			case *ast.Ident:
				// reads a variable the moved expression's calls may have written through
				if v, ok := pass.TypesInfo.Uses[x].(*types.Var); ok {
					unsafe = unsafe || mutables[v]
				}
			}
			return !unsafe
		})
		// an intervening call may write through an operand it receives pointer-like, just
		// as the moved expression's calls may — either direction re-orders the effect
		if !unsafe {
			for v := range callMutables(pass, stmt) {
				unsafe = unsafe || operands[v]
			}
		}
		if unsafe {
			return true
		}
	}
	return false
}

// callMutables collects the variables that calls inside n may write through: the root of any
// argument or method receiver whose type is pointer-like (pointer, slice, map, chan, func,
// or interface), since the callee shares that memory with the caller. The go-azure-helpers
// pointer package is pure by construction, so its calls mark nothing — `pointer.From(x)`
// never justifies keeping a temporary.
func callMutables(pass *analysis.Pass, n ast.Node) map[*types.Var]bool {
	mutables := map[*types.Var]bool{}
	mark := func(e ast.Expr) {
		root := rootIdent(e)
		if root == nil {
			return
		}
		v, ok := pass.TypesInfo.Uses[root].(*types.Var)
		if !ok {
			return
		}
		if tv, ok := pass.TypesInfo.Types[e]; ok && tv.Type != nil && pointerLike(tv.Type) {
			mutables[v] = true
		}
	}
	ast.Inspect(n, func(x ast.Node) bool {
		call, ok := x.(*ast.CallExpr)
		if !ok {
			return true
		}
		if tv, ok := pass.TypesInfo.Types[call.Fun]; ok && tv.IsType() {
			return true // a conversion copies its operand
		}
		if sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr); ok {
			if fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func); ok && fn.Pkg() != nil &&
				fn.Pkg().Path() == pointerpkg.PkgPath {
				return true // pure: reads through its arguments, never writes
			}
			mark(sel.X) // method receiver
		}
		for _, arg := range call.Args {
			mark(arg)
		}
		return true
	})
	return mutables
}

// pointerLike reports whether a value of type t shares memory with the caller when passed to
// a call.
func pointerLike(t types.Type) bool {
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return true
	}
	return false
}

// shadows reports whether stmt itself declares any of names — a `x := ...` or `var x ...`
// whose new variable would capture references in an expression moved below it.
func shadows(stmt ast.Stmt, names map[string]bool) bool {
	switch d := stmt.(type) {
	case *ast.AssignStmt:
		if d.Tok == token.DEFINE {
			for _, lhs := range d.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && names[id.Name] {
					return true
				}
			}
		}
	case *ast.DeclStmt:
		if gd, ok := d.Decl.(*ast.GenDecl); ok && gd.Tok == token.VAR {
			for _, spec := range gd.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, id := range vs.Names {
						if names[id.Name] {
							return true
						}
					}
				}
			}
		}
	}
	return false
}

// rootIdent returns the base identifier an lvalue writes through: column for `column[y]`,
// out for `out.Format`, p for `*p`; nil when the base is not an identifier.
func rootIdent(e ast.Expr) *ast.Ident {
	for {
		switch x := e.(type) {
		case *ast.Ident:
			return x
		case *ast.IndexExpr:
			e = x.X
		case *ast.IndexListExpr:
			e = x.X
		case *ast.SelectorExpr:
			e = x.X
		case *ast.SliceExpr:
			e = x.X
		case *ast.StarExpr:
			e = x.X
		case *ast.ParenExpr:
			e = x.X
		default:
			return nil
		}
	}
}

// EnclosingFile returns the *ast.File containing pos.
func EnclosingFile(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, file := range pass.Files {
		if file.FileStart <= pos && pos < file.FileEnd {
			return file
		}
	}
	return nil
}

// SourceText returns the raw source bytes of node.
func SourceText(pass *analysis.Pass, node ast.Node) ([]byte, bool) {
	tf := pass.Fset.File(node.Pos())
	if tf == nil {
		return nil, false
	}
	content, err := pass.ReadFile(tf.Name())
	if err != nil {
		return nil, false
	}
	start, end := tf.Offset(node.Pos()), tf.Offset(node.End())
	if start < 0 || end > len(content) || start > end {
		return nil, false
	}
	return content[start:end], true
}

// CalledFunc resolves the function a call invokes, through parens and explicit generic
// instantiation; nil for builtins, function values, and method values on unresolved types.
func CalledFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	fun := ast.Unparen(call.Fun)
	switch ix := fun.(type) {
	case *ast.IndexExpr:
		fun = ix.X
	case *ast.IndexListExpr:
		fun = ix.X
	}
	var id *ast.Ident
	switch f := fun.(type) {
	case *ast.Ident:
		id = f
	case *ast.SelectorExpr:
		id = f.Sel
	default:
		return nil
	}
	fn, _ := pass.TypesInfo.Uses[id].(*types.Func)
	return fn
}
