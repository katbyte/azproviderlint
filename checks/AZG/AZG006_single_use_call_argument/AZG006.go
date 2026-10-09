// Package AZG006 defines an analyzer that reports single-use variables whose only use is an
// argument of a later call whose other arguments are all literals — `x := flatten(...)`
// followed by `d.Set("key", x)` — which should be inlined.
package AZG006

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for a short variable declaration with a single-line initializer whose
// variable is used exactly once in the entire function, as a bare argument of a later call
// statement — or the call in an if statement's init (`if err := d.Set("key", x); err != nil`)
// — whose every other argument is a basic literal. When the sibling arguments are literals the
// name documents nothing the literal does not already say, and inlining cannot reorder side
// effects, so the temporary should be inlined.
//
// Sibling arguments must be basic literals or plain identifiers (`ctx`, `id`, `client`);
// anything more complex — selector chains, calls, type assertions — marks the name as
// documentation among expressions (`client.CreateOrUpdate(ctx, id, payload)`) and the call is
// left alone. The only-when-literals flag restricts siblings to literals only, and
// maximum-arguments skips calls carrying more than that many arguments (0 = unlimited).
// Multi-line initializers are out of scope since splicing one into an argument list hurts
// readability — SDK payload literals passed to client calls are the common case. Nothing is
// reported when an intervening statement writes to, takes the address of, or shadows anything
// the initializer reads — there the temporary preserves a value the intervening code changes.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG006",
	Doc:      "check for single-use variables only used in a later function call that should be inlined",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG006_single_use_call_argument/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// maxGap bounds how many source lines may separate the declaration from the consuming call,
// matching AZG005's bound and default.
var maxGap int

func init() {
	Analyzer.Flags.Init("AZG006", flag.ContinueOnError)
	Analyzer.Flags.IntVar(&maxGap, "max-gap", 100, "maximum number of source lines between the variable's declaration and the consuming call")
	Analyzer.Flags.BoolVar(&onlyWhenLiterals, "only-when-literals", false, "only inline when every sibling argument is a basic literal (plain identifiers are otherwise also accepted)")
	Analyzer.Flags.IntVar(&maximumArguments, "maximum-arguments", 0, "skip calls with more than this many arguments (0 = unlimited)")
}

// onlyWhenLiterals restricts sibling arguments to basic literals; by default plain
// identifiers are also accepted, which reads fine (`client.ImportThenPoll(ctx, id, expandImport(d))`)
// but lets several temps feeding one call co-inline into a long line in rare cases.
var onlyWhenLiterals bool

// maximumArguments skips consuming calls with more than this many arguments, bounding how
// crowded an argument list the inline may land in; 0 means unlimited.
var maximumArguments int

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
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
					checkPair(pass, body, stmts[i], stmts[j], stmts[i+1:j])
				}
			}
			return true
		})
	})

	return nil, nil
}

// checkPair reports first when it declares a single-use, single-line temporary that second
// consumes as a call argument among literal siblings.
func checkPair(pass *analysis.Pass, body *ast.BlockStmt, first, second ast.Stmt, between []ast.Stmt) {
	adjacent := len(between) == 0
	assign, ok := first.(*ast.AssignStmt)
	if !ok || assign.Tok != token.DEFINE || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
		return
	}

	// a multi-line initializer spliced into an argument list hurts readability
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

	call := callConsumer(second)
	if call == nil {
		return
	}

	if maximumArguments > 0 && len(call.Args) > maximumArguments {
		return
	}

	useExpr, ok := literalSiblingUse(pass, call, obj)
	if !ok {
		return
	}

	// an intervening write to (or shadowing of) an operand of the initializer would change
	// its value at the consumer — the temporary is load-bearing, not redundant
	if astx.UnsafeToMovePast(pass, assign.Rhs[0], between) {
		return
	}

	if astx.UseCount(pass, body, obj) != 1 {
		return
	}

	message := fmt.Sprintf("%q is only used by the following call and should be inlined", ident.Name)
	if !adjacent {
		message = fmt.Sprintf("%q is only used by the call on line %d and should be inlined", ident.Name, pass.Fset.Position(second.Pos()).Line)
	}

	pass.Report(analysis.Diagnostic{
		Pos:            assign.Pos(),
		Message:        message,
		SuggestedFixes: suggestedFixes(pass, assign, second, useExpr, adjacent),
	})
}

// callConsumer returns the call a statement executes: a bare call statement, or the call an if
// statement's init assigns (`if err := d.Set("key", x); err != nil`).
func callConsumer(stmt ast.Stmt) *ast.CallExpr {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			return call
		}
	case *ast.IfStmt:
		if as, ok := s.Init.(*ast.AssignStmt); ok && len(as.Rhs) == 1 {
			if call, ok := as.Rhs[0].(*ast.CallExpr); ok {
				return call
			}
		}
	}
	return nil
}

// literalSiblingUse returns the argument that is a bare reference to obj when it appears
// exactly once and every other argument is a basic literal — the condition under which the
// name documents nothing and inlining cannot reorder evaluation.
func literalSiblingUse(pass *analysis.Pass, call *ast.CallExpr, obj types.Object) (ast.Expr, bool) {
	var use ast.Expr
	for _, arg := range call.Args {
		if astx.IsUseOf(pass, arg, obj) {
			if use != nil {
				return nil, false
			}
			use = arg
			continue
		}
		switch ast.Unparen(arg).(type) {
		case *ast.BasicLit:
		case *ast.Ident:
			if onlyWhenLiterals {
				return nil, false
			}
		default:
			return nil, false
		}
	}
	return use, use != nil
}

// suggestedFixes deletes the temporary's declaration and replaces the consuming argument with
// the declaration's initializer text.
func suggestedFixes(pass *analysis.Pass, assign *ast.AssignStmt, consumer ast.Stmt, useExpr ast.Expr, adjacent bool) []analysis.SuggestedFix {
	exprSrc, ok := astx.SourceText(pass, assign.Rhs[0])
	if !ok {
		return nil
	}

	// adjacent pairs delete straight through to the consumer; a distant consumer must leave
	// the intervening statements alone, so only the declaration's own line is removed
	del := analysis.TextEdit{Pos: assign.Pos(), End: consumer.Pos()}
	if !adjacent {
		tf := pass.Fset.File(assign.Pos())
		delEnd := assign.End()
		if endLine := tf.Line(assign.End()); endLine+1 <= tf.LineCount() {
			delEnd = tf.LineStart(endLine + 1)
		}
		del = analysis.TextEdit{Pos: tf.LineStart(tf.Line(assign.Pos())), End: delEnd}
	}

	return []analysis.SuggestedFix{{
		Message: "Inline the temporary into the consuming call",
		TextEdits: []analysis.TextEdit{
			del,
			{Pos: useExpr.Pos(), End: useExpr.End(), NewText: exprSrc},
		},
	}}
}
