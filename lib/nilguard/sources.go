// Non-nil sources: expressions that can never be nil, and the field descent through literals
// and composite IDs that carries their guarantee.

package nilguard

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
)

// isNonNilSource reports whether expr can never be nil: an address-of, new(...), an
// immediately-invoked func literal whose returns all qualify, or a call to a function
// ReturnsAnalyzer proved never returns nil in that position.
func isNonNilSource(pass *analysis.Pass, expr ast.Expr) bool {
	switch x := ast.Unparen(expr).(type) {
	case *ast.UnaryExpr:
		return x.Op == token.AND
	case *ast.CallExpr:
		if id, ok := ast.Unparen(x.Fun).(*ast.Ident); ok {
			if _, isBuiltin := pass.TypesInfo.Uses[id].(*types.Builtin); isBuiltin {
				return id.Name == "new"
			}
		}
		// an immediately-invoked func literal whose every return is itself non-nil
		if lit, ok := ast.Unparen(x.Fun).(*ast.FuncLit); ok {
			return funcLitReturnsNonNil(pass, lit)
		}
		// any other call is trusted only on its own proof: pointer.To (`return &v`) and the
		// flag constructors (`p := new(T); ...; return p`) earn their verdict like any
		// other function
		return nonNilResult(pass, astx.CalledFunc(pass, x), 0)
	}
	return false
}

// funcLitReturnsNonNil reports whether lit has a single result and every return statement in
// its body (nested literals excluded) returns a provably non-nil value.
func funcLitReturnsNonNil(pass *analysis.Pass, lit *ast.FuncLit) bool {
	if lit.Type.Results == nil || len(lit.Type.Results.List) != 1 || len(lit.Type.Results.List[0].Names) > 1 {
		return false
	}
	returns := 0
	nonNil := true
	ast.Inspect(lit.Body, func(n ast.Node) bool {
		switch r := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ReturnStmt:
			returns++
			if len(r.Results) != 1 || !isNonNilSource(pass, r.Results[0]) {
				nonNil = false
			}
		}
		return nonNil
	})
	return returns > 0 && nonNil
}

// literalField returns the value a composite literal (or &literal) gives the field path in
// suffix (".F.G"), descending nested literals; nil when expr is not a literal, a segment is
// absent (the zero value) or unkeyed, or an intermediate value is not itself a literal.
func literalField(expr ast.Expr, suffix string) ast.Expr {
	for seg := range strings.SplitSeq(strings.TrimPrefix(suffix, "."), ".") {
		e := ast.Unparen(expr)
		if u, ok := e.(*ast.UnaryExpr); ok && u.Op == token.AND {
			e = ast.Unparen(u.X)
		}
		lit, ok := e.(*ast.CompositeLit)
		if !ok {
			return nil
		}
		expr = nil
		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			if id, ok := kv.Key.(*ast.Ident); ok && id.Name == seg {
				expr = kv.Value
				break
			}
		}
		if expr == nil {
			return nil
		}
	}
	return expr
}

// compositeIDArg returns the argument a commonids composite-ID constructor or parser stores
// under suffix (".First" or ".Second"), nil when expr is neither call or suffix is another
// field. Both functions take the two IDs as their trailing arguments.
func compositeIDArg(pass *analysis.Pass, expr ast.Expr, suffix string) ast.Expr {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok || len(call.Args) < 2 {
		return nil
	}
	fn := astx.CalledFunc(pass, call)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != commonIDsPkgPath ||
		(fn.Name() != "NewCompositeResourceID" && fn.Name() != "ParseCompositeResourceID") {
		return nil
	}
	switch suffix {
	case ".First":
		return call.Args[len(call.Args)-2]
	case ".Second":
		return call.Args[len(call.Args)-1]
	}
	return nil
}

const commonIDsPkgPath = "github.com/hashicorp/go-azure-helpers/resourcemanager/commonids"
