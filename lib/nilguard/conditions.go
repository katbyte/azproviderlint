// Conditions and companions: what an if/for/case condition proves about a path, and the
// err/ok contract of multi-result calls.

package nilguard

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"golang.org/x/tools/go/analysis"
)

// companionKeys returns the keys of the error and ok-bool results assign binds alongside its
// other results.
func companionKeys(pass *analysis.Pass, assign *ast.AssignStmt) (errKey, okKey string) {
	errType := types.Universe.Lookup("error").Type()
	for _, lhs := range assign.Lhs {
		id, ok := ast.Unparen(lhs).(*ast.Ident)
		if !ok {
			continue
		}
		obj := pass.TypesInfo.Uses[id]
		if obj == nil {
			obj = pass.TypesInfo.Defs[id]
		}
		if obj == nil {
			continue
		}
		if types.Identical(obj.Type(), errType) {
			errKey, _ = PathKey(pass, lhs)
		} else if basic, isBasic := obj.Type().Underlying().(*types.Basic); isBasic && basic.Kind() == types.Bool {
			okKey, _ = PathKey(pass, lhs)
		}
	}
	return errKey, okKey
}

// companionCheckedBetween reports whether assign returns an error or ok-bool alongside the
// pointer and one of the following statements exits on it (`if err != nil { return ... }`,
// `if !ok { return ... }`) — the conventional Go contracts under which the other results are
// valid.
func companionCheckedBetween(pass *analysis.Pass, assign *ast.AssignStmt, between []ast.Stmt) bool {
	errKey, okKey := companionKeys(pass, assign)
	for _, s := range between {
		if errKey == "" && okKey == "" {
			return false
		}
		if ifs, ok := s.(*ast.IfStmt); ok && ifs.Else == nil && terminates(ifs.Body) && condProvesInvalid(pass, ifs.Cond, errKey, okKey) {
			return true
		}
		// a companion overwritten before the check, anywhere in the statement, no longer
		// speaks for the call
		if errKey != "" && writesPath(pass, s, errKey) {
			errKey = ""
		}
		if okKey != "" && writesPath(pass, s, okKey) {
			okKey = ""
		}
	}
	return false
}

// condProvesInvalid reports whether cond being true means the companion signalled failure —
// `err != nil` or `!ok`, possibly inside a pure-|| chain — so cond being false (an exit not
// taken, an else branch) proves the other results valid.
func condProvesInvalid(pass *analysis.Pass, cond ast.Expr, errKey, okKey string) bool {
	return (errKey != "" && orContainsNeqNil(pass, cond, errKey)) || (okKey != "" && orContainsNot(pass, cond, okKey))
}

// condProvesValid reports whether cond being true means the companion signalled success —
// `err == nil` or `ok`, possibly inside a pure-&& chain.
func condProvesValid(pass *analysis.Pass, cond ast.Expr, errKey, okKey string) bool {
	if errKey == "" && okKey == "" {
		return false
	}
	switch x := ast.Unparen(cond).(type) {
	case *ast.BinaryExpr:
		if x.Op == token.LAND {
			return condProvesValid(pass, x.X, errKey, okKey) || condProvesValid(pass, x.Y, errKey, okKey)
		}
		return errKey != "" && x.Op == token.EQL && nilComparison(pass, x, errKey)
	case *ast.Ident:
		k, ok := PathKey(pass, x)
		return ok && okKey != "" && k == okKey
	}
	return false
}

// orContainsNot reports whether cond is `!ok` or a pure-|| chain containing it.
func orContainsNot(pass *analysis.Pass, cond ast.Expr, key string) bool {
	switch x := ast.Unparen(cond).(type) {
	case *ast.BinaryExpr:
		if x.Op == token.LOR {
			return orContainsNot(pass, x.X, key) || orContainsNot(pass, x.Y, key)
		}
	case *ast.UnaryExpr:
		if x.Op == token.NOT {
			k, ok := PathKey(pass, x.X)
			return ok && k == key
		}
	}
	return false
}

// orContainsNeqNil reports whether cond is `key != nil` or a pure-|| chain containing it, so
// cond being false proves key == nil.
func orContainsNeqNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LOR {
		return orContainsNeqNil(pass, x.X, key) || orContainsNeqNil(pass, x.Y, key)
	}
	return x.Op == token.NEQ && nilComparison(pass, x, key)
}

// terminates reports whether a block's final statement leaves the enclosing flow: a return,
// branch, panic, or fatal call.
func terminates(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}
	switch last := block.List[len(block.List)-1].(type) {
	case *ast.ReturnStmt, *ast.BranchStmt:
		return true
	case *ast.ExprStmt:
		call, ok := last.X.(*ast.CallExpr)
		if !ok {
			return false
		}
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			return fun.Name == "panic"
		case *ast.SelectorExpr:
			name := fun.Sel.Name
			return name == "Exit" || strings.HasPrefix(name, "Fatal")
		}
	}
	return false
}

// impliesNonNil reports whether cond being true proves key != nil.
func impliesNonNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LAND {
		return impliesNonNil(pass, x.X, key) || impliesNonNil(pass, x.Y, key)
	}
	return (x.Op == token.NEQ && nilComparison(pass, x, key)) || fromNonZero(pass, x, key)
}

// fromNonZero reports whether cmp is `pointer.From(x) != <zero>` or `len(pointer.From(x)) > 0`
// (either operand order) for key's path — both false when x is nil, so true proves x non-nil.
func fromNonZero(pass *analysis.Pass, cmp *ast.BinaryExpr, key string) bool {
	// fromArg returns x when e is pointer.From(x), nil otherwise
	fromArg := func(e ast.Expr) ast.Expr {
		call, ok := ast.Unparen(e).(*ast.CallExpr)
		if !ok || len(call.Args) != 1 {
			return nil
		}
		if fn := astx.CalledFunc(pass, call); fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != pointerpkg.PkgPath || fn.Name() != "From" {
			return nil
		}
		return call.Args[0]
	}
	// proves reports whether `e <op> zero` (op already normalised to have e on the left)
	// proves key non-nil: `From(x) != 0` or `len(From(x)) > 0` / `len(From(x)) != 0`
	proves := func(e, zero ast.Expr, op token.Token) bool {
		if !isZeroConst(pass, zero) {
			return false
		}
		arg := fromArg(e)
		if arg == nil {
			call, ok := ast.Unparen(e).(*ast.CallExpr)
			if !ok || len(call.Args) != 1 || (op != token.NEQ && op != token.GTR) {
				return false
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "len" {
				return false
			} else if _, isBuiltin := pass.TypesInfo.Uses[id].(*types.Builtin); !isBuiltin {
				return false
			}
			arg = fromArg(call.Args[0])
		} else if op != token.NEQ {
			return false
		}
		if arg == nil {
			return false
		}
		k, ok := PathKey(pass, arg)
		return ok && k == key
	}
	switch cmp.Op { //nolint:exhaustive // only the three comparisons that can prove non-nil matter
	case token.NEQ:
		return proves(cmp.X, cmp.Y, token.NEQ) || proves(cmp.Y, cmp.X, token.NEQ)
	case token.GTR:
		return proves(cmp.X, cmp.Y, token.GTR)
	case token.LSS:
		return proves(cmp.Y, cmp.X, token.GTR)
	default:
		return false
	}
}

// isZeroConst reports whether e is a constant "" or 0.
func isZeroConst(pass *analysis.Pass, e ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[e]
	if !ok || tv.Value == nil {
		return false
	}
	switch tv.Value.Kind() {
	case constant.String:
		return constant.StringVal(tv.Value) == ""
	case constant.Int, constant.Float:
		return constant.Sign(tv.Value) == 0
	case constant.Unknown, constant.Bool, constant.Complex:
	}
	return false
}

// impliedByNil reports whether key == nil forces cond to be true — so cond being false (an
// else branch, a short-circuited ||, a not-taken early exit) proves key != nil.
func impliedByNil(pass *analysis.Pass, cond ast.Expr, key string) bool {
	x, ok := ast.Unparen(cond).(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if x.Op == token.LOR {
		return impliedByNil(pass, x.X, key) || impliedByNil(pass, x.Y, key)
	}
	return x.Op == token.EQL && nilComparison(pass, x, key)
}

// nilComparison reports whether cmp compares key's path against nil (either operand order).
func nilComparison(pass *analysis.Pass, cmp *ast.BinaryExpr, key string) bool {
	matches := func(e, other ast.Expr) bool {
		k, ok := PathKey(pass, e)
		return ok && k == key && astx.IsNilValue(pass, other)
	}
	return matches(cmp.X, cmp.Y) || matches(cmp.Y, cmp.X)
}
