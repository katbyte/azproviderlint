// Package AZG002 defines an analyzer that reports single-use temporaries whose only use is
// taking their address — `v := <expr>` followed by `&v` in a later statement — which should be
// `new(<expr>)` (or `pointer.To(<expr>)`) at the address-of site.
package AZG002

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"go/version"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/pointerpkg"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for a short variable declaration whose variable's only use in the entire
// function is a single `&v`, in a later statement of the same block within max-gap source
// lines. The temporary exists only because `&` needs an addressable operand, and Go 1.26's
// `new(<expr>)` — or the `pointer.To(<expr>)` helper — says the same thing in one step at the
// use site.
//
// The `use` flag picks the suggested form: `new` (the default) or `pointer.To`. `new(<expr>)`
// does not compile on packages below Go 1.26, so new mode errors there — the config must set
// `use: pointer.To` explicitly. In new mode, existing `pointer.To(x)` calls are also
// reported (they should be `new(x)`) unless `allow` lists `pointer.To`; the fix drops the
// pointer import when the rewrite leaves no other reference to the package.
//
// The initializer must be single-line and free of calls and channel receives — type
// conversions are fine — so moving its evaluation to the address-of site cannot reorder
// observable effects (in pointer.To mode a converted enum initializer inlines to
// `pointer.To(sdk.Enum(v))`, which AZG003 then rewrites to `pointer.ToEnum`). Nothing is
// reported when an intervening statement writes to, takes the address of, or shadows anything
// the initializer reads, or when the `&v` sits inside a nested function literal (the closure
// would defer the initializer's evaluation).
var Analyzer = &analysis.Analyzer{
	Name:     "AZG002",
	Doc:      "check for single-use temporaries only used as an address that should use new() or pointer.To",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG002_address_of_single_use_temporary/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const (
	useNew       = "new"
	usePointerTo = "pointer.To"
)

// maxGap bounds how many source lines may separate the declaration from the address-of site,
// matching AZG005's bound and default.
var maxGap int

// use picks the suggested pointer-creation form: the new() builtin or the pointer.To helper.
var use string

// allow lists additionally tolerated forms (comma-separated); `pointer.To` keeps existing
// pointer.To calls unreported when use is new.
var allow string

const (
	fixPointerCopyNone  = "none"
	fixPointerCopyCopy  = "copy"
	fixPointerCopyShare = "share"
)

// fixPointerCopy picks how `out := *p; &out` is fixed: none reports without a fix so a person
// chooses, copy keeps the copy-preserving `pointer.To(*p)`/`new(*p)`, share replaces the
// address with `p` itself — an aliasing change the config opts into globally.
var fixPointerCopy string

func init() {
	Analyzer.Flags.Init("AZG002", flag.ContinueOnError)
	Analyzer.Flags.IntVar(&maxGap, "max-gap", 100,
		"maximum number of source lines between the temporary's declaration and the statement taking its address")
	Analyzer.Flags.StringVar(&use, "use", useNew,
		"pointer-creation form to suggest: new or pointer.To (packages below go1.26 must set pointer.To)")
	Analyzer.Flags.StringVar(&allow, "allow", "",
		"comma-separated forms to leave unreported where they already appear: pointer.To")
	Analyzer.Flags.StringVar(&fixPointerCopy, "fix-pointer-copy", fixPointerCopyNone,
		"fix for a dereference-copy temporary (out := *p; &out): none, copy, or share")
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	if use != useNew && use != usePointerTo {
		return nil, fmt.Errorf("AZG002: invalid use flag %q: expected %q or %q", use, useNew, usePointerTo)
	}
	if fixPointerCopy != fixPointerCopyNone && fixPointerCopy != fixPointerCopyCopy && fixPointerCopy != fixPointerCopyShare {
		return nil, fmt.Errorf("AZG002: invalid fix-pointer-copy flag %q: expected %q, %q or %q", fixPointerCopy, fixPointerCopyNone, fixPointerCopyCopy, fixPointerCopyShare)
	}
	allowed := map[string]bool{}
	for form := range strings.SplitSeq(allow, ",") {
		form = strings.TrimSpace(form)
		if form == "" {
			continue
		}
		if form != usePointerTo {
			return nil, fmt.Errorf("AZG002: invalid allow flag entry %q: expected %q", form, usePointerTo)
		}
		allowed[form] = true
	}

	// new(<expr>) needs Go 1.26; below that the config must opt into the helper explicitly
	if use == useNew {
		if v := pass.Pkg.GoVersion(); version.IsValid(v) && version.Compare(version.Lang(v), "go1.26") < 0 {
			return nil, fmt.Errorf("AZG002: use: new requires go1.26 or newer but %s builds with %s — upgrade the module's go directive or set use: pointer.To", pass.Pkg.Path(), v)
		}
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		var body *ast.BlockStmt
		switch fn := n.(type) {
		case *ast.FuncDecl:
			body = fn.Body
		case *ast.FuncLit:
			body = fn.Body
		}
		if body == nil {
			return
		}

		// examine every statement list in this function, but not those of nested function
		// literals — they are visited as their own FuncLit node
		ast.Inspect(body, func(x ast.Node) bool {
			if _, isLit := x.(*ast.FuncLit); isLit && x != n {
				return false
			}

			var stmts []ast.Stmt
			switch node := x.(type) {
			case *ast.BlockStmt:
				stmts = node.List
			case *ast.CaseClause:
				stmts = node.Body
			case *ast.CommClause:
				stmts = node.Body
			default:
				return true
			}

			for i := range len(stmts) - 1 {
				declEnd := pass.Fset.Position(stmts[i].End()).Line
				for j := i + 1; j < len(stmts); j++ {
					if pass.Fset.Position(stmts[j].Pos()).Line-declEnd > maxGap {
						break
					}
					checkPair(pass, body, stmts[i], stmts[j], stmts[i+1:j], use)
				}
			}
			return true
		})
	})

	if use == useNew && !allowed[usePointerTo] {
		checkPointerToCalls(pass)
	}

	return nil, nil
}

// checkPair reports first when it declares a single-use temporary whose address second takes.
func checkPair(pass *analysis.Pass, body *ast.BlockStmt, first, second ast.Stmt, between []ast.Stmt, form string) {
	assign, ok := first.(*ast.AssignStmt)
	if !ok || assign.Tok != token.DEFINE || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	// a multi-line initializer spliced into the creation call at the use site hurts
	// readability
	if pass.Fset.Position(assign.Pos()).Line != pass.Fset.Position(assign.End()).Line {
		return
	}

	ident, ok := assign.Lhs[0].(*ast.Ident)
	if !ok || ident.Name == "_" {
		return
	}

	obj := pass.TypesInfo.Defs[ident]
	if obj == nil {
		return
	}

	// the initializer must be free of calls and receives so moving its evaluation to the
	// address-of site cannot reorder observable effects; type conversions carry none
	effectful := false
	ast.Inspect(assign.Rhs[0], func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			effectful = effectful || !pass.TypesInfo.Types[x.Fun].IsType()
		case *ast.UnaryExpr:
			effectful = effectful || x.Op == token.ARROW
		}
		return !effectful
	})
	if effectful {
		return
	}

	addr := addressUse(pass, second, obj)
	if addr == nil {
		return
	}

	// an intervening write to (or shadowing of) an operand of the initializer would change
	// its value at the address-of site — the temporary is load-bearing, not redundant
	if astx.UnsafeToMovePast(pass, assign.Rhs[0], between) {
		return
	}

	if astx.UseCount(pass, body, obj) != 1 {
		return
	}

	display := "new()"
	if form == usePointerTo {
		display = usePointerTo
	}
	message := fmt.Sprintf("%q is only used as an address by the following statement and should be inlined with %s", ident.Name, display)
	if len(between) > 0 {
		message = fmt.Sprintf("%q is only used as an address by the statement on line %d and should be inlined with %s",
			ident.Name, pass.Fset.Position(second.Pos()).Line, display)
	}
	// `v := *p` then `&v` could inline to the copy-preserving pointer.To(*p)/new(*p), but the
	// better form is often p itself — dropping the copy changes aliasing, so fix-pointer-copy picks
	// the fix: none (default) leaves the choice to a person, copy preserves the copy, share
	// substitutes p
	fixes := suggestedFixes(pass, assign, addr, form)
	if star, ok := ast.Unparen(assign.Rhs[0]).(*ast.StarExpr); ok {
		if tv, ok := pass.TypesInfo.Types[star.X]; ok && tv.Type != nil {
			if _, isPtr := tv.Type.Underlying().(*types.Pointer); isPtr {
				message += fmt.Sprintf(" - or, if sharing the original pointer is acceptable, use %s directly", types.ExprString(star.X))
				switch fixPointerCopy {
				case fixPointerCopyNone:
					fixes = nil
				case fixPointerCopyShare:
					fixes = nil
					if operandSrc, ok := astx.SourceText(pass, star.X); ok {
						tf := pass.Fset.File(assign.Pos())
						delEnd := assign.End()
						if endLine := tf.Line(assign.End()); endLine+1 <= tf.LineCount() {
							delEnd = tf.LineStart(endLine + 1)
						}
						fixes = []analysis.SuggestedFix{{
							Message: "Replace the temporary's address with the original pointer",
							TextEdits: []analysis.TextEdit{
								{Pos: tf.LineStart(tf.Line(assign.Pos())), End: delEnd},
								{Pos: addr.Pos(), End: addr.End(), NewText: operandSrc},
							},
						}}
					}
				}
			}
		}
	}
	pass.Report(analysis.Diagnostic{
		Pos:            assign.Pos(),
		Message:        message,
		SuggestedFixes: fixes,
	})
}

// addressUse returns the `&v` expression within stmt whose operand resolves to obj, looking
// through parentheses but not into nested function literals — an address captured by a
// closure is evaluated when the closure runs, not where it is written.
func addressUse(pass *analysis.Pass, stmt ast.Stmt, obj types.Object) *ast.UnaryExpr {
	var addr *ast.UnaryExpr
	ast.Inspect(stmt, func(n ast.Node) bool {
		if _, isLit := n.(*ast.FuncLit); isLit {
			return false
		}
		if u, ok := n.(*ast.UnaryExpr); ok && u.Op == token.AND && astx.IsUseOf(pass, u.X, obj) {
			addr = u
			return false
		}
		return addr == nil
	})
	return addr
}

// suggestedFixes deletes the temporary's declaration line and replaces the `&v` with the
// chosen creation form around the initializer, adding the pointer import when pointer.To is
// suggested and the file lacks it.
func suggestedFixes(pass *analysis.Pass, assign *ast.AssignStmt, addr *ast.UnaryExpr, form string) []analysis.SuggestedFix {
	exprSrc, ok := astx.SourceText(pass, assign.Rhs[0])
	if !ok {
		return nil
	}

	creation := "new(" + string(exprSrc) + ")"
	var importEdit *analysis.TextEdit
	if form == usePointerTo {
		file := astx.EnclosingFile(pass, assign.Pos())
		if file == nil {
			return nil
		}
		pkgName, edit, ok := pointerpkg.Ref(file)
		if !ok {
			return nil
		}
		creation = pkgName + ".To(" + string(exprSrc) + ")"
		importEdit = edit
	}

	tf := pass.Fset.File(assign.Pos())
	delEnd := assign.End()
	if endLine := tf.Line(assign.End()); endLine+1 <= tf.LineCount() {
		delEnd = tf.LineStart(endLine + 1)
	}

	edits := []analysis.TextEdit{
		{Pos: tf.LineStart(tf.Line(assign.Pos())), End: delEnd},
		{Pos: addr.Pos(), End: addr.End(), NewText: []byte(creation)},
	}
	if importEdit != nil {
		edits = append(edits, *importEdit)
	}

	return []analysis.SuggestedFix{{
		Message:   "Replace the temporary's address with " + form,
		TextEdits: edits,
	}}
}

// checkPointerToCalls reports pointer.To calls that should use the new() builtin. When a
// file's flagged calls are its only references to the pointer package, the first call's fix
// also deletes the then-unused import; files that keep other pointer helpers keep the import.
func checkPointerToCalls(pass *analysis.Pass) {
	for _, file := range pass.Files {
		var toCalls []*ast.CallExpr
		pkgRefs := 0
		ast.Inspect(file, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.Ident:
				if pkg, ok := pass.TypesInfo.Uses[x].(*types.PkgName); ok && pkg.Imported().Path() == pointerpkg.PkgPath {
					pkgRefs++
				}
			case *ast.CallExpr:
				sel, ok := x.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
				if ok && fn.Name() == "To" && fn.Pkg() != nil && fn.Pkg().Path() == pointerpkg.PkgPath && len(x.Args) == 1 {
					toCalls = append(toCalls, x)
				}
			}
			return true
		})

		for i, call := range toCalls {
			edits := []analysis.TextEdit{
				{Pos: call.Fun.Pos(), End: call.Fun.End(), NewText: []byte("new")},
			}
			// rewriting every reference orphans the import; deleting it from the first fix
			// keeps the joint application compiling
			if i == 0 && pkgRefs == len(toCalls) {
				if edit := deleteImportEdit(pass, file); edit != nil {
					edits = append(edits, *edit)
				}
			}
			pass.Report(analysis.Diagnostic{
				Pos:     call.Pos(),
				Message: "pointer.To should be replaced with the new(...) builtin",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message:   "Replace pointer.To with new",
					TextEdits: edits,
				}},
			})
		}
	}
}

// deleteImportEdit removes the pointer package's import spec line from file.
func deleteImportEdit(pass *analysis.Pass, file *ast.File) *analysis.TextEdit {
	for _, imp := range file.Imports {
		if strings.Trim(imp.Path.Value, `"`) != pointerpkg.PkgPath {
			continue
		}
		tf := pass.Fset.File(imp.Pos())
		end := imp.End()
		if endLine := tf.Line(imp.End()); endLine+1 <= tf.LineCount() {
			end = tf.LineStart(endLine + 1)
		}
		return &analysis.TextEdit{Pos: tf.LineStart(tf.Line(imp.Pos())), End: end}
	}
	return nil
}
