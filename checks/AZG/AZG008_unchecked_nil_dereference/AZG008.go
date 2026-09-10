// Package AZG008 defines an analyzer that reports pointer dereferences with no reachable nil
// guard — `string(*props.Status)` where nothing established `props.Status != nil` — which
// panic when an optional SDK field is absent.
package AZG008

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/nilguard"
	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"github.com/katbyte/azproviderlint/lib/requestbody"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks every explicit dereference `*x` in value context — one a fix can rewrite to
// `pointer.From(x)` — where x is a pointer-typed variable or selector chain (`props.Status`,
// `model.Properties.Sku`) with no guard proving x non-nil at that point (the guard vocabulary
// lives in lib/nilguard: enclosing nil conditions, early exits, provably non-nil assignments,
// err/ok companion checks, aliases of guarded chains, all invalidated by reassignment from an
// unknown source).
//
// Dereferences of bare pointer parameters are trusted by default — the nil check belongs at
// the call sites — unless include-parameters is set; field dereferences through a parameter
// (`*props.Status`) are always in scope.
//
// The suggested fix rewrites `string(*x)` of a string-kinded enum pointer to
// `pointer.FromEnum(x)` and other dereferences to `pointer.From(x)` — note both change
// behaviour from panic to zero value, which is the desired semantics in flatten/read paths
// but should be reviewed elsewhere; fix-with: none reports without fixes. Dereferences whose
// context needs the pointer itself (`*x = v`, `&*x`, `(*x)++`) are AZG009's, as are implicit
// dereferences (`m.Properties.Name` with a nil `Properties`); chains containing calls or
// index expressions (`*resp.Items[i].Name`) are out of scope. _test.go files are checked
// unless tests=false.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG008",
	Doc:      "check for pointer dereferences with no reachable nil guard that should use pointer.From",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG008_unchecked_nil_dereference/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer, nilguard.ReturnsAnalyzer, requestbody.Analyzer},
	Run:      run,
}

const (
	fixPointerFrom = "pointer.From"
	fixNone        = "none"
)

// includeParameters also flags dereferences of bare pointer parameters (`*id` where id is a
// parameter). By default those are trusted — the nil check belongs at the call sites, and a
// helper dereferencing its own parameter is usually relying on a caller-side contract; field
// dereferences through a parameter (`*props.Status`) are always in scope.
var includeParameters bool

// checkTests checks _test.go files too; a nil deref there panics the test rather than the
// provider, but still fails the run.
var checkTests bool

// reportRequestBody also reports dereferences whose value is sent as a PUT/PATCH/POST body.
// Those get no fix (pointer.From would send an empty request) and need a hand-written guard,
// so they are opt-in: off by default, on for a dedicated pass over the write paths.
var reportRequestBody bool

// fixWith picks the suggested-fix form: pointer.From (string conversions of enum pointers
// upgrade to pointer.FromEnum) or none to report without fixes.
var fixWith string

func init() {
	Analyzer.Flags.Init("AZG008", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&includeParameters, "include-parameters", false,
		"also report dereferences of bare pointer parameters (callers' nil-check contract is otherwise trusted)")
	Analyzer.Flags.BoolVar(&checkTests, "tests", true,
		"check _test.go files (false skips them)")
	Analyzer.Flags.BoolVar(&reportRequestBody, "requestbody", false,
		"also report dereferences sent as a PUT/PATCH/POST body (reported without a fix)")
	Analyzer.Flags.StringVar(&fixWith, "fix-with", fixPointerFrom,
		"suggested-fix form: pointer.From or none")
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	if fixWith != fixPointerFrom && fixWith != fixNone {
		return nil, fmt.Errorf("AZG008: invalid fix-with flag %q: expected %q or %q", fixWith, fixPointerFrom, fixNone)
	}

	nilguard.ForEachFunc(pass, insp, checkTests, func(body *ast.BlockStmt, params map[types.Object]bool, parents map[ast.Node]ast.Node) {
		ast.Inspect(body, func(x ast.Node) bool {
			if star, ok := x.(*ast.StarExpr); ok {
				checkDeref(pass, parents, params, star)
			}
			return true
		})
	})

	return nil, nil
}

// checkDeref reports star when it dereferences a pointer-typed path in value context with no
// reachable guard.
func checkDeref(pass *analysis.Pass, parents map[ast.Node]ast.Node, params map[types.Object]bool, star *ast.StarExpr) {
	tv, ok := pass.TypesInfo.Types[star.X]
	if !ok || tv.Type == nil {
		return
	}
	if _, isPtr := tv.Type.Underlying().(*types.Pointer); !isPtr || !tv.IsValue() {
		return // a type expression like *T, or not a pointer
	}

	key, ok := nilguard.PathKey(pass, star.X)
	if !ok {
		return // calls, indexes, or non-variable roots: guards are not matchable
	}

	// a bare parameter's nil contract belongs to its callers
	if !includeParameters {
		if id, ok := ast.Unparen(star.X).(*ast.Ident); ok && params[pass.TypesInfo.Uses[id]] {
			return
		}
	}

	// contexts that need an addressable pointee (*x = v, (*x).F = v, &*x, (*x)++, a pointer
	// method on *x) cannot take pointer.From and are AZG009's to report
	if nilguard.DerefNeedsPointer(pass, parents, star) {
		return
	}

	if nilguard.Guarded(pass, parents, star, key) {
		return
	}

	// a value sent as the body of a PUT/PATCH/POST must not be rewritten: pointer.From would
	// turn the panic into an empty (or partly empty) write request. The report stands,
	// without a fix, so the site gets a real guard. The dereferenced value counts when,
	// climbing out through parens, conversions, composite literals, and address-of, it lands
	// on a write-body argument (requestbody facts) or on an assignment into a local — whole,
	// or one of its fields — that reaches such an argument anywhere in the function.
	rootObj := func(e ast.Expr) types.Object {
		for {
			switch x := ast.Unparen(e).(type) {
			case *ast.Ident:
				if o := pass.TypesInfo.Uses[x]; o != nil {
					return o
				}
				return pass.TypesInfo.Defs[x]
			case *ast.SelectorExpr:
				e = x.X
			case *ast.StarExpr:
				e = x.X
			case *ast.IndexExpr:
				e = x.X
			case *ast.UnaryExpr:
				if x.Op != token.AND {
					return nil
				}
				e = x.X
			default:
				return nil
			}
		}
	}
	var body ast.Node = star
	for parents[body] != nil {
		body = parents[body]
	}
	sinks := map[types.Object]bool{}
	ast.Inspect(body, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			for j, a := range call.Args {
				if requestbody.CallBodyArg(pass, call, j) {
					if o := rootObj(a); o != nil {
						sinks[o] = true
					}
				}
			}
		}
		return true
	})
	payload := false
	node := ast.Node(star)
climb:
	for {
		switch p := parents[node].(type) {
		case *ast.ParenExpr, *ast.CompositeLit, *ast.KeyValueExpr:
			node = p
		case *ast.UnaryExpr:
			if p.Op != token.AND {
				break climb
			}
			node = p
		case *ast.CallExpr:
			for i, a := range p.Args {
				if a == node && requestbody.CallBodyArg(pass, p, i) {
					payload = true
				}
			}
			// a conversion (T(*x)) is still building the value
			if tv, ok := pass.TypesInfo.Types[p.Fun]; !payload && ok && tv.IsType() && len(p.Args) == 1 && p.Args[0] == node {
				node = p
				continue
			}
			break climb
		case *ast.AssignStmt:
			if len(p.Lhs) == len(p.Rhs) {
				for i, r := range p.Rhs {
					if r == node && sinks[rootObj(p.Lhs[i])] {
						payload = true
					}
				}
			}
			break climb
		default:
			break climb
		}
	}

	if payload && !reportRequestBody {
		return
	}

	msg := "dereference of possibly-nil `" + types.ExprString(star.X) + "` may panic - add a nil check or use pointer.From"
	var fixes []analysis.SuggestedFix
	switch {
	case payload:
		msg = "dereference of possibly-nil `" + types.ExprString(star.X) + "` may panic - add a nil check (no fix: the value is sent as a write request body, pointer.From would send it empty)"
	case fixWith == fixPointerFrom:
		fixes = suggestedFixes(pass, parents, star)
	}
	pass.Report(analysis.Diagnostic{
		Pos:            star.Pos(),
		Message:        msg,
		SuggestedFixes: fixes,
	})
}

// suggestedFixes rewrites the dereference to pointer.From, or the enclosing string conversion
// of an enum pointer to pointer.FromEnum.
func suggestedFixes(pass *analysis.Pass, parents map[ast.Node]ast.Node, star *ast.StarExpr) []analysis.SuggestedFix {
	// walk out of parens to find the true consuming context
	var outer ast.Node = star
	for {
		p, ok := parents[outer].(*ast.ParenExpr)
		if !ok {
			break
		}
		outer = p
	}

	operandSrc, ok := astx.SourceText(pass, star.X)
	if !ok {
		return nil
	}
	file := astx.EnclosingFile(pass, star.Pos())
	if file == nil {
		return nil
	}
	pkgName, importEdit, ok := pointerpkg.Ref(file)
	if !ok {
		return nil
	}

	// string(*enumPtr) becomes pointer.FromEnum(enumPtr) when the pointee is a string-kinded
	// named type; otherwise the dereference alone becomes pointer.From(x)
	target, replacement := ast.Node(star), pkgName+".From("+string(operandSrc)+")"
	if conv, ok := parents[outer].(*ast.CallExpr); ok && len(conv.Args) == 1 && conv.Args[0] == outer {
		if tv, ok := pass.TypesInfo.Types[conv.Fun]; ok && tv.IsType() {
			if basic, ok := tv.Type.Underlying().(*types.Basic); ok && basic.Kind() == types.String {
				if ptr, ok := pass.TypesInfo.Types[star.X].Type.Underlying().(*types.Pointer); ok {
					if named, ok := ptr.Elem().(*types.Named); ok {
						if elem, ok := named.Underlying().(*types.Basic); ok && elem.Info()&types.IsString != 0 {
							target, replacement = conv, pkgName+".FromEnum("+string(operandSrc)+")"
						}
					}
				}
			}
		}
	}

	edits := []analysis.TextEdit{{Pos: target.Pos(), End: target.End(), NewText: []byte(replacement)}}
	if importEdit != nil {
		edits = append(edits, *importEdit)
	}
	return []analysis.SuggestedFix{{
		Message:   "Replace the dereference with " + pkgName + ".From",
		TextEdits: edits,
	}}
}
