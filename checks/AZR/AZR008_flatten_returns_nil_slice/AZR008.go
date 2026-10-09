// Package AZR008 defines an analyzer that reports flatten* functions that return nil slices
// or maps.
package AZR008

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that flatten* helpers return empty slices and maps instead of nil on
// non-error paths. It catches literal nils (including conversions like `[]T(nil)`), naked
// returns with still-unassigned named results, and variables that are provably still nil —
// which also stops a nil error value (`var noErr error`) from masking a return as an error
// path. Pointer results are exempt (nil is a deliberate absent signal) and `interface{}`
// results are out of scope (the container shape is not declared).
var Analyzer = &analysis.Analyzer{
	Name:     "AZR008",
	Doc:      "check that flatten functions returning slices or maps return an empty container instead of nil",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR008_flatten_returns_nil_slice/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// resultInfo describes one logical return position.
type resultInfo struct {
	typeStr string
	kind    string // "slice", "map" or "" for other types
	isError bool
	name    types.Object // named result object, nil when results are unnamed
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	errorInterface, ok := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)
	if !ok {
		return nil, errors.New("AZR008: could not resolve the built-in error interface type")
	}

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}

		name := funcDecl.Name.Name
		if len(name) < 7 || !strings.EqualFold(name[:7], "flatten") {
			return
		}

		if funcDecl.Type.Results == nil || funcDecl.Body == nil {
			return
		}

		// Keep one entry per return position, including multi-name fields.
		var results []resultInfo
		haveContainer := false
		for _, field := range funcDecl.Type.Results.List {
			info := resultInfo{typeStr: types.ExprString(field.Type)}
			if t := pass.TypesInfo.TypeOf(field.Type); t != nil {
				switch t.Underlying().(type) {
				case *types.Slice:
					info.kind = "slice"
				case *types.Map:
					info.kind = "map"
				}
				info.isError = types.Implements(t, errorInterface)
			}
			haveContainer = haveContainer || info.kind != ""
			if len(field.Names) == 0 {
				results = append(results, info)
				continue
			}
			for _, fieldName := range field.Names {
				withName := info
				withName.name = pass.TypesInfo.Defs[fieldName]
				results = append(results, withName)
			}
		}

		if !haveContainer {
			return
		}

		namedResults := map[types.Object]bool{}
		for _, res := range results {
			if res.name != nil {
				namedResults[res.name] = true
			}
		}

		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}

			retStmt, ok := node.(*ast.ReturnStmt)
			if !ok {
				return true
			}

			if len(retStmt.Results) == 0 {
				checkNakedReturn(pass, funcDecl.Body, retStmt, results, name)
			} else {
				checkExplicitReturn(pass, funcDecl.Body, retStmt, results, namedResults, name)
			}
			return true
		})
	})

	return nil, nil
}

// checkExplicitReturn reports a return whose container positions hold provably nil values on a
// non-error path, rewriting each nil into an empty container literal.
func checkExplicitReturn(pass *analysis.Pass, body *ast.BlockStmt, retStmt *ast.ReturnStmt, results []resultInfo, namedResults map[types.Object]bool, funcName string) {
	// An error path (`return nil, err`) legitimately returns nil for the container; only the
	// nil-input branch, whose error positions are all provably nil, is a real finding.
	for i, res := range retStmt.Results {
		if i < len(results) && results[i].isError && !isProvablyNil(pass, body, res, retStmt.Pos(), namedResults) {
			return
		}
	}

	var edits []analysis.TextEdit
	var kinds []string
	for i, res := range retStmt.Results {
		if i >= len(results) || results[i].kind == "" {
			continue
		}
		if !isProvablyNil(pass, body, res, retStmt.Pos(), namedResults) {
			continue
		}

		kinds = append(kinds, results[i].kind)
		edit := analysis.TextEdit{
			Pos:     res.Pos(),
			End:     res.End(),
			NewText: []byte(results[i].typeStr + "{}"),
		}
		edits = append(edits, edit)
		// replacing the sole use of a declaration-only variable would leave it unused, which
		// does not compile — delete the declaration's line too when it is cleanly deletable,
		// otherwise report without a fix
		if ident, ok := ast.Unparen(res).(*ast.Ident); ok {
			if obj := pass.TypesInfo.Uses[ident]; obj != nil && astx.UseCount(pass, body, obj) == 1 && !namedResults[obj] {
				declEdit, ok := declarationLineEdit(pass, body, obj)
				if !ok {
					edits = nil
					break
				}
				edits = append(edits, *declEdit)
			}
		}
	}

	if len(kinds) == 0 {
		return
	}

	report(pass, retStmt, kinds, edits, funcName)
}

// checkNakedReturn reports a bare return whose named container results are provably still
// nil, rewriting it into an explicit return with empty container literals.
func checkNakedReturn(pass *analysis.Pass, body *ast.BlockStmt, retStmt *ast.ReturnStmt, results []resultInfo, funcName string) {
	// a bare return returns the named results as they are; an error position that may hold a
	// real error marks an error path
	for _, res := range results {
		if res.isError && (res.name == nil || !isUnassignedBefore(pass, body, res.name, retStmt.Pos())) {
			return
		}
	}

	values := make([]string, 0, len(results))
	var kinds []string
	for _, res := range results {
		if res.name == nil {
			return // unnamed results cannot reach a bare return in compiling code
		}
		if res.kind != "" && isUnassignedBefore(pass, body, res.name, retStmt.Pos()) {
			kinds = append(kinds, res.kind)
			values = append(values, res.typeStr+"{}")
			continue
		}
		values = append(values, res.name.Name())
	}

	if len(kinds) == 0 {
		return
	}

	edits := []analysis.TextEdit{{
		Pos:     retStmt.Pos(),
		End:     retStmt.End(),
		NewText: []byte("return " + strings.Join(values, ", ")),
	}}
	report(pass, retStmt, kinds, edits, funcName)
}

// report emits the diagnostic, naming the container kind(s) involved; edits may be nil when no
// safe rewrite exists.
func report(pass *analysis.Pass, retStmt *ast.ReturnStmt, kinds []string, edits []analysis.TextEdit, funcName string) {
	word := kinds[0]
	for _, k := range kinds[1:] {
		if k != word {
			word = "slice/map"
			break
		}
	}

	diag := analysis.Diagnostic{
		Pos:     retStmt.Pos(),
		Message: fmt.Sprintf("flatten function %q should return an empty %s instead of nil", funcName, word),
	}
	if len(edits) > 0 {
		diag.SuggestedFixes = []analysis.SuggestedFix{{
			Message:   fmt.Sprintf("Return an empty %s instead of nil", word),
			TextEdits: edits,
		}}
	}
	pass.Report(diag)
}

// isProvablyNil reports whether expr denotes a nil value at retPos: the predeclared nil
// (possibly wrapped in conversions), or a variable that is a zero-value declaration or named
// result with no assignment before retPos and its address never taken.
func isProvablyNil(pass *analysis.Pass, body *ast.BlockStmt, expr ast.Expr, retPos token.Pos, namedResults map[types.Object]bool) bool {
	if astx.IsNilValue(pass, expr) {
		return true
	}

	ident, ok := ast.Unparen(expr).(*ast.Ident)
	if !ok {
		return false
	}

	obj := pass.TypesInfo.Uses[ident]
	if obj == nil {
		return false
	}
	if !namedResults[obj] && !isZeroValueDeclared(pass, body, obj) {
		return false
	}

	return isUnassignedBefore(pass, body, obj, retPos)
}

// isZeroValueDeclared reports whether obj is declared by a `var x T` declaration with no
// assignment expression (named results are handled by the caller via the namedResults set, so
// parameters — also declared outside the body — never count).
func isZeroValueDeclared(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object) bool {
	declaredNil := false
	ast.Inspect(body, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, name := range spec.Names {
			if pass.TypesInfo.Defs[name] == obj {
				declaredNil = len(spec.Values) == 0
				return false
			}
		}
		return true
	})
	return declaredNil
}

// isUnassignedBefore reports whether obj has no assignment before pos and its address is never
// taken — the conservative condition under which its declared nil value must still hold.
// Nested function literals are scanned too, since a closure may assign a captured variable.
func isUnassignedBefore(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object, pos token.Pos) bool {
	clean := true
	ast.Inspect(body, func(n ast.Node) bool {
		if !clean {
			return false
		}
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if id, ok := ast.Unparen(lhs).(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj && node.Pos() < pos {
					clean = false
				}
			}
		case *ast.UnaryExpr:
			// an address-taken variable can be assigned anywhere, so nothing is provable
			if id, ok := ast.Unparen(node.X).(*ast.Ident); ok && node.Op == token.AND && pass.TypesInfo.Uses[id] == obj {
				clean = false
			}
		case *ast.RangeStmt:
			if node.Tok == token.ASSIGN {
				for _, e := range []ast.Expr{node.Key, node.Value} {
					if id, ok := e.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj && node.Pos() < pos {
						clean = false
					}
				}
			}
		}
		return true
	})
	return clean
}

// declarationLineEdit returns an edit deleting obj's `var` declaration line when the
// declaration cleanly holds just this one name, and ok=false when the declaration cannot be
// safely removed.
func declarationLineEdit(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object) (*analysis.TextEdit, bool) {
	var edit *analysis.TextEdit
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		decl, ok := n.(*ast.DeclStmt)
		if !ok {
			return true
		}

		gen, ok := decl.Decl.(*ast.GenDecl)
		if !ok || len(gen.Specs) != 1 {
			return true
		}

		spec, ok := gen.Specs[0].(*ast.ValueSpec)
		if !ok || len(spec.Names) != 1 || pass.TypesInfo.Defs[spec.Names[0]] != obj {
			return true
		}

		found = true
		tf := pass.Fset.File(decl.Pos())
		delEnd := decl.End()
		if endLine := tf.Line(decl.End()); endLine+1 <= tf.LineCount() {
			delEnd = tf.LineStart(endLine + 1)
		}

		edit = &analysis.TextEdit{Pos: tf.LineStart(tf.Line(decl.Pos())), End: delEnd}
		return false
	})
	if !found {
		return nil, false
	}
	return edit, true
}
