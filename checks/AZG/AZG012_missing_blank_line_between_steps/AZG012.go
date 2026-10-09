// Package AZG012 defines an analyzer that reports a block statement run straight into the
// next step of a longer function body without a blank line between them.
package AZG012

import (
	"flag"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/katbyte/azproviderlint/lib/astx"
)

// Analyzer checks that in a body with more than a few statements, an if, for, switch or select
// is separated from the statement that follows it by a blank line, so a reader sees where one
// step ends and the next begins. A short body is one group and is left alone, as is a run of
// one-line ifs, which read as a single set of guards.
var Analyzer = &analysis.Analyzer{
	Name:     "AZG012",
	Doc:      "check that the steps of a longer function body are separated by blank lines",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZG/AZG012_missing_blank_line_between_steps/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// A body with at most this many statements, or spanning at most this many lines, is one
// group: no blank lines are asked for inside it.
var (
	maxStatements = 5
	maxLines      = 10
)

func init() {
	Analyzer.Flags.Init("AZG012", flag.ContinueOnError)
	Analyzer.Flags.IntVar(&maxStatements, "statements", maxStatements, "bodies with this many statements or fewer are one group")
	Analyzer.Flags.IntVar(&maxLines, "lines", maxLines, "bodies spanning this many lines or fewer are one group")
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	line := func(p token.Pos) int { return pass.Fset.Position(p).Line }
	check := func(stmts []ast.Stmt, open, closing token.Pos) {
		if len(stmts) <= maxStatements || line(closing)-line(open)+1 <= maxLines {
			return
		}

		for i := 0; i+1 < len(stmts); i++ {
			a, s := stmts[i], stmts[i+1]
			kind := blockKind(a)
			if kind == "" {
				continue
			}

			switch s.(type) {
			case *ast.CaseClause, *ast.CommClause:
				continue
			}

			if _, isIf := s.(*ast.IfStmt); isIf && oneLiner(a) && oneLiner(s) {
				continue
			}

			// a comment between them belongs to the next step, so the blank goes before it
			at := s.Pos()
			for _, cg := range astx.EnclosingFile(pass, s.Pos()).Comments {
				if cg.Pos() > a.End() && cg.End() < s.Pos() && line(cg.Pos()) > line(a.End()) && line(cg.Pos()) < line(at) {
					at = cg.Pos()
				}
			}

			if line(at) != line(a.End())+1 {
				continue
			}

			tf := pass.Fset.File(at)
			start := tf.LineStart(tf.Line(at))
			pass.Report(analysis.Diagnostic{
				Pos:     at,
				Message: "a blank line should separate this from the " + kind + " above, which ends a step",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message:   "Insert a blank line",
					TextEdits: []analysis.TextEdit{{Pos: start, End: start, NewText: []byte("\n")}},
				}},
			})
		}
	}

	insp.Preorder([]ast.Node{(*ast.BlockStmt)(nil), (*ast.CaseClause)(nil), (*ast.CommClause)(nil)}, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.BlockStmt:
			check(n.List, n.Lbrace, n.Rbrace)
		case *ast.CaseClause:
			check(n.Body, n.Pos(), n.End())
		case *ast.CommClause:
			check(n.Body, n.Pos(), n.End())
		}
	})

	return nil, nil
}

// blockKind names a statement with a body, or is empty for anything else: only these end a
// step. A literal assignment that happens to close with a brace does not.
func blockKind(s ast.Stmt) string {
	switch s.(type) {
	case *ast.IfStmt:
		return "if"
	case *ast.ForStmt, *ast.RangeStmt:
		return "for"
	case *ast.SwitchStmt, *ast.TypeSwitchStmt:
		return "switch"
	case *ast.SelectStmt:
		return "select"
	case *ast.BlockStmt:
		return "block"
	}
	return ""
}

// oneLiner reports whether s is an if without else whose body is a single statement: a guard.
func oneLiner(s ast.Stmt) bool {
	i, ok := s.(*ast.IfStmt)
	return ok && i.Else == nil && len(i.Body.List) == 1
}
