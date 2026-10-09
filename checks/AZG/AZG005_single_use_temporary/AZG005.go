// Package AZG005 defines an analyzer that reports single-use temporaries consumed by a later
// statement in the same block — `x := <expr>` followed by `y = x` or `return x` where x has no
// other use — which should be inlined. The consumer may be at most max-gap source lines below
// the declaration (default 100), keeping the inline local enough to read.
package AZG005

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

// Analyzer checks for a short variable declaration whose variable is used exactly once in the
// entire function, by a later statement in the same block within max-gap source lines, as the
// bare right-hand side of a plain assignment or as the sole return value. Such a temporary adds a name without adding
// information — `output.Format = pointer.From(input.Format)` reads as well as the two-line
// form — so it should be inlined.
//
// The consuming statement must be `y = x` (single pair, plain `=`) or `return x`; call
// arguments are deliberately out of scope since naming an argument is usually intentional
// documentation. Assignments whose left-hand side contains a function call are skipped, since
// inlining would reorder the call relative to the temporary's initializer. Nothing is reported
// when an intervening statement writes to, takes the address of, or shadows anything the
// initializer reads — there the temporary preserves a value the intervening code changes.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG005",
	Doc:      "check for single-use temporaries immediately consumed by the next statement that should be inlined",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG005_single_use_temporary/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// maxGap bounds how many source lines may separate the declaration from its consumer: far
// apart pairs are legal to inline but moving an initializer hundreds of lines is a
// readability loss, not a win (a provider config literal teleporting 450 lines down its
// function is what motivated the default).
var maxGap int

func init() {
	Analyzer.Flags.Init("AZG005", flag.ContinueOnError)
	Analyzer.Flags.IntVar(&maxGap, "max-gap", 100, "maximum number of source lines between the temporary's declaration and its consumer")
}

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

// checkPair reports first when it declares a single-use temporary that second consumes.
func checkPair(pass *analysis.Pass, body *ast.BlockStmt, first, second ast.Stmt, between []ast.Stmt) {
	adjacent := len(between) == 0
	assign, ok := first.(*ast.AssignStmt)
	if !ok || assign.Tok != token.DEFINE || len(assign.Lhs) != 1 || len(assign.Rhs) != 1 {
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

	useExpr, ok := consumesAsBareValue(pass, second, obj)
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

	message := fmt.Sprintf("%q is only used by the following statement and should be inlined", ident.Name)
	if !adjacent {
		message = fmt.Sprintf("%q is only used by the statement on line %d and should be inlined", ident.Name, pass.Fset.Position(second.Pos()).Line)
	}

	pass.Report(analysis.Diagnostic{
		Pos:            assign.Pos(),
		Message:        message,
		SuggestedFixes: suggestedFixes(pass, assign, second, useExpr, adjacent),
	})
}

// consumesAsBareValue reports whether stmt consumes obj as the bare right-hand side of a plain
// single assignment or as the sole return value, returning the consuming expression.
func consumesAsBareValue(pass *analysis.Pass, stmt ast.Stmt, obj types.Object) (ast.Expr, bool) {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok != token.ASSIGN || len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return nil, false
		}
		if !astx.IsUseOf(pass, s.Rhs[0], obj) {
			return nil, false
		}
		// a blank assignment is a discard, not a consumption worth inlining into
		if lhs, ok := s.Lhs[0].(*ast.Ident); ok && lhs.Name == "_" {
			return nil, false
		}
		// inlining moves the initializer after the left-hand side's operands in evaluation
		// order, so a call on the left-hand side could observe the swap
		if astx.ContainsCallExpr(s.Lhs[0]) {
			return nil, false
		}
		return s.Rhs[0], true
	case *ast.ReturnStmt:
		if len(s.Results) == 1 && astx.IsUseOf(pass, s.Results[0], obj) {
			return s.Results[0], true
		}
	}
	return nil, false
}

// suggestedFixes deletes the temporary's declaration and replaces the consuming reference with
// the declaration's initializer, spliced as raw source text so multi-line initializers keep
// their exact original formatting.
func suggestedFixes(pass *analysis.Pass, assign *ast.AssignStmt, consumer ast.Stmt, useExpr ast.Expr, adjacent bool) []analysis.SuggestedFix {
	exprSrc, ok := astx.SourceText(pass, assign.Rhs[0])
	if !ok {
		return nil
	}

	// adjacent pairs delete straight through to the consumer; a distant consumer must leave
	// the intervening statements alone, so only the declaration's own lines are removed
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
		Message: "Inline the temporary into the consuming statement",
		TextEdits: []analysis.TextEdit{
			del,
			{Pos: useExpr.Pos(), End: useExpr.End(), NewText: exprSrc},
		},
	}}
}
