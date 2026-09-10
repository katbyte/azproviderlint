// Package AZG009 defines an analyzer that reports possibly-nil pointer dereferences that only
// a hand-written nil guard can fix: implicit dereferences (`m.Properties.Name` selects through
// a possibly-nil `m.Properties`) and explicit dereferences whose context needs the pointer
// itself (`*x = v`, `&*x`, `(*x)++`).
package AZG009

import (
	"flag"
	"go/ast"
	"go/types"
	"regexp"
	"strings"

	"github.com/katbyte/azproviderlint/lib/nilguard"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer is AZG008's sister: the same guard engine (lib/nilguard) applied to the
// dereferences pointer.From cannot rewrite, so every report needs a person to add a guard.
// Two forms are checked:
//
//   - implicit dereferences: selecting a field (`m.Properties.Name`) or calling a
//     value-receiver method through a pointer-typed chain implicitly dereferences it —
//     pointer-receiver methods don't and are ignored;
//   - explicit dereferences whose context needs the pointer itself: assignment targets
//     (`*x = v`), the operand of & (`&*x` panics too), and inc/dec statements.
//
// Dereferences of bare pointer parameters (`props.Count` where props is the parameter) are
// trusted by default — the nil check belongs at the call sites — unless include-parameters is
// set; deeper links through a parameter (`props.Sku.Name`) are always in scope. Chains
// containing calls or index expressions are out of scope, as are implicit dereferences of
// embedded pointer fields (the dereferenced chain is not spelled in the source). Pointers to
// types matching exclude-types are skipped (`*Client` silences SDK client plumbing, whose
// value-receiver methods dereference the client on every call). _test.go files are checked
// unless tests=false.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG009",
	Doc:      "check for nil pointer dereferences that need a hand-written nil guard",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG009_nil_dereference_requires_guard/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer, nilguard.ReturnsAnalyzer},
	Run:      run,
}

// includeParameters also flags dereferences of bare pointer parameters; by default the
// callers' nil-check contract is trusted, mirroring AZG008.
var includeParameters bool

// checkTests checks _test.go files too, mirroring AZG008.
var checkTests bool

// excludeTypes lists glob patterns (`*` matches anything) for the qualified name of the
// pointee type (`github.com/x/y/clients.Client`); dereferences of pointers to a matching type
// are not reported.
var excludeTypes string

func init() {
	Analyzer.Flags.Init("AZG009", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&includeParameters, "include-parameters", false,
		"also report dereferences of bare pointer parameters (callers' nil-check contract is otherwise trusted)")
	Analyzer.Flags.BoolVar(&checkTests, "tests", true,
		"check _test.go files (false skips them)")
	Analyzer.Flags.StringVar(&excludeTypes, "exclude-types", "",
		"comma-separated glob patterns for pointee types whose dereferences are not reported, matched against the qualified name (`*Client`, `*/clients.Client`)")
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	// compiled per run: packages are analysed concurrently, so no package-level state
	var exclude []*regexp.Regexp
	for pat := range strings.SplitSeq(excludeTypes, ",") {
		if pat = strings.TrimSpace(pat); pat != "" {
			exclude = append(exclude, regexp.MustCompile("^"+strings.ReplaceAll(regexp.QuoteMeta(pat), `\*`, ".*")+"$"))
		}
	}

	nilguard.ForEachFunc(pass, insp, checkTests, func(body *ast.BlockStmt, params map[types.Object]bool, parents map[ast.Node]ast.Node) {
		ast.Inspect(body, func(x ast.Node) bool {
			switch n := x.(type) {
			case *ast.StarExpr:
				checkExplicit(pass, parents, params, exclude, n)
			case *ast.SelectorExpr:
				checkImplicit(pass, parents, params, exclude, n)
			}
			return true
		})
	})

	return nil, nil
}

// report runs the shared eligibility checks on the dereferenced operand — matchable path,
// not a trusted parameter, pointee not excluded — and reports at pos when no guard covers it.
func report(pass *analysis.Pass, parents map[ast.Node]ast.Node, params map[types.Object]bool, exclude []*regexp.Regexp, at ast.Node, operand ast.Expr, message string) {
	key, ok := nilguard.PathKey(pass, operand)
	if !ok {
		return // calls, indexes, or non-variable roots: guards are not matchable
	}

	// a bare parameter's nil contract belongs to its callers
	if !includeParameters {
		if id, ok := ast.Unparen(operand).(*ast.Ident); ok && params[pass.TypesInfo.Uses[id]] {
			return
		}
	}

	if ptr, ok := pass.TypesInfo.TypeOf(operand).Underlying().(*types.Pointer); ok && len(exclude) > 0 {
		name := types.TypeString(ptr.Elem(), nil)
		if named, ok := ptr.Elem().(*types.Named); ok && named.Obj().Pkg() != nil {
			name = named.Obj().Pkg().Path() + "." + named.Obj().Name()
		}
		for _, re := range exclude {
			if re.MatchString(name) {
				return
			}
		}
	}

	if nilguard.Guarded(pass, parents, at, key) {
		return
	}

	pass.Report(analysis.Diagnostic{Pos: at.Pos(), Message: message})
}

// checkExplicit reports `*x` in a context that needs the pointer itself; value-context
// dereferences are AZG008's, which fixes them with pointer.From.
func checkExplicit(pass *analysis.Pass, parents map[ast.Node]ast.Node, params map[types.Object]bool, exclude []*regexp.Regexp, star *ast.StarExpr) {
	tv, ok := pass.TypesInfo.Types[star.X]
	if !ok || tv.Type == nil {
		return
	}
	if _, isPtr := tv.Type.Underlying().(*types.Pointer); !isPtr || !tv.IsValue() {
		return // a type expression like *T, or not a pointer
	}
	if !nilguard.DerefNeedsPointer(pass, parents, star) {
		return
	}
	report(pass, parents, params, exclude, star, star.X,
		"dereference of possibly-nil `"+types.ExprString(star.X)+"` may panic - add a nil check")
}

// checkImplicit reports a selection that implicitly dereferences a pointer-typed operand:
// a field access through a pointer, or a value-receiver method accessed through one.
func checkImplicit(pass *analysis.Pass, parents map[ast.Node]ast.Node, params map[types.Object]bool, exclude []*regexp.Regexp, sel *ast.SelectorExpr) {
	selection, ok := pass.TypesInfo.Selections[sel]
	if !ok {
		return // a package or type selector, not a field/method access
	}
	tv, ok := pass.TypesInfo.Types[sel.X]
	if !ok || tv.Type == nil || !tv.IsValue() {
		return
	}
	if _, isPtr := tv.Type.Underlying().(*types.Pointer); !isPtr {
		return // selecting on a value cannot dereference it
	}

	switch selection.Kind() {
	case types.FieldVal:
		// selecting any field through a pointer dereferences it
	case types.MethodVal:
		// only a value receiver forces a dereference; a pointer-receiver method is called
		// on the pointer itself
		if sig, ok := selection.Obj().Type().(*types.Signature); ok {
			if _, ptrRecv := sig.Recv().Type().(*types.Pointer); ptrRecv {
				return
			}
		}
	case types.MethodExpr:
		return // T.Method — the operand is a type, nothing is dereferenced
	}

	report(pass, parents, params, exclude, sel, sel.X,
		"selecting `"+sel.Sel.Name+"` implicitly dereferences possibly-nil `"+types.ExprString(sel.X)+"` and may panic - add a nil check")
}
