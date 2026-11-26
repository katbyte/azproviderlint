// Package AZR014 defines an analyzer that reports an ID parser's error wrapped in another message.
package AZR014

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks what is done with the error of a resource ID parser. That error already names the ID, what is wrong with it and what a right
// one looks like, so wrapping it in "parsing ..." only makes the message the user reads longer.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR014",
	Doc:      "check that an ID parser's error is returned as it is",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR014_id_parser_error_wrapped/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// parser is what the SDK names the function that reads one kind of resource ID.
var parser = regexp.MustCompile(`^Parse\w+ID(Insensitively)?$`)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	wrapped := map[*ast.File][]*ast.CallExpr{}
	insp.WithStack([]ast.Node{(*ast.IfStmt)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		check, isIf := n.(*ast.IfStmt)
		if !push || !isIf || len(check.Body.List) != 1 || len(stack) < 2 {
			return true
		}

		// the parse is the statement the check starts with, or the one before it
		parse := check.Init
		if parse == nil {
			var stmts []ast.Stmt
			switch in := stack[len(stack)-2].(type) {
			case *ast.BlockStmt:
				stmts = in.List

			case *ast.CaseClause:
				stmts = in.Body
			}

			for i, stmt := range stmts {
				if stmt == check && i > 0 {
					parse = stmts[i-1]
				}
			}
		}

		assign, isAssign := parse.(*ast.AssignStmt)
		if !isAssign || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
			return true
		}

		call, isCall := assign.Rhs[0].(*ast.CallExpr)
		failed, isIdent := assign.Lhs[1].(*ast.Ident)
		if !isCall || !isIdent || pass.TypesInfo.ObjectOf(failed) == nil {
			return true
		}

		// a parser gives back an ID, which can be written out again
		fn := astx.CalledFunc(pass, call)
		if fn == nil || !parser.MatchString(fn.Name()) || fn.Signature().Results().Len() != 2 {
			return true
		}
		if id, _, _ := types.LookupFieldOrMethod(fn.Signature().Results().At(0).Type(), true, fn.Pkg(), "ID"); id == nil {
			return true
		}

		cond, isCompare := check.Cond.(*ast.BinaryExpr)
		ret, isReturn := check.Body.List[0].(*ast.ReturnStmt)
		if !isCompare || cond.Op != token.NEQ || !astx.IsUseOf(pass, cond.X, pass.TypesInfo.ObjectOf(failed)) || !isReturn || len(ret.Results) == 0 {
			return true
		}

		wrap, isWrap := ret.Results[len(ret.Results)-1].(*ast.CallExpr)
		if !isWrap || len(wrap.Args) == 0 || !astx.IsUseOf(pass, wrap.Args[len(wrap.Args)-1], pass.TypesInfo.ObjectOf(failed)) {
			return true
		}

		if errorf := astx.CalledFunc(pass, wrap); errorf != nil && errorf.Pkg() != nil && errorf.Pkg().Path() == "fmt" && errorf.Name() == "Errorf" {
			wrapped[astx.EnclosingFile(pass, wrap.Pos())] = append(wrapped[astx.EnclosingFile(pass, wrap.Pos())], wrap)
		}

		return true
	})

	for file, wraps := range wrapped {
		// taking the last use of fmt out of a file leaves an import that does not compile, so then there is no fix
		uses := 0
		ast.Inspect(file, func(n ast.Node) bool {
			if id, isIdent := n.(*ast.Ident); isIdent {
				if pkg, isPkg := pass.TypesInfo.Uses[id].(*types.PkgName); isPkg && pkg.Imported().Path() == "fmt" {
					uses++
				}
			}

			return true
		})

		for _, wrap := range wraps {
			failed := wrap.Args[len(wrap.Args)-1]
			d := analysis.Diagnostic{Pos: wrap.Pos(), End: wrap.End(), Message: "the ID parser's error already says what is wrong with the ID and what a right one looks like - return it as it is"}
			if uses > len(wraps) {
				d.SuggestedFixes = []analysis.SuggestedFix{{Message: "return the parser's error", TextEdits: []analysis.TextEdit{{Pos: wrap.Pos(), End: failed.Pos()}, {Pos: failed.End(), End: wrap.End()}}}}
			}
			pass.Report(d)
		}
	}

	return nil, nil
}
