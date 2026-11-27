// Package AZR014 defines an analyzer that reports an ID parser's error wrapped in another message.
package AZR014

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/tf"
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

	// a state migration reads IDs as they used to be, and says so, and a test's errors are not read by a user
	if tf.InStateMigration(pass) {
		return nil, nil
	}

	// each wrap by file, and the statement that read the ID whose error it wraps
	wrapped, parsedBy := map[*ast.File][]*ast.CallExpr{}, map[*ast.CallExpr]*ast.AssignStmt{}
	insp.WithStack([]ast.Node{(*ast.IfStmt)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		check, isIf := n.(*ast.IfStmt)
		if !push || !isIf || len(stack) < 2 || strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return true
		}

		// the error that is checked
		cond, isCompare := check.Cond.(*ast.BinaryExpr)
		if !isCompare || cond.Op != token.NEQ {
			return true
		}

		failed, isIdent := ast.Unparen(cond.X).(*ast.Ident)
		if !isIdent || pass.TypesInfo.ObjectOf(failed) == nil {
			return true
		}

		// where it was last set: the statement the check starts with, or the nearest one above it that sets it
		before := []ast.Stmt{check.Init}
		if check.Init == nil {
			before = nil
			switch in := stack[len(stack)-2].(type) {
			case *ast.BlockStmt:
				before = in.List[:slices.Index(in.List, ast.Stmt(check))]

			case *ast.CaseClause:
				before = in.Body[:slices.Index(in.Body, ast.Stmt(check))]
			}
		}

		var call *ast.CallExpr
		var parsed *ast.AssignStmt
		for i := len(before) - 1; i >= 0 && call == nil; i-- {
			// a statement that sets it anywhere inside, such as in one arm of an if, is where the search ends
			sets := false
			ast.Inspect(before[i], func(n ast.Node) bool {
				inner, isInner := n.(*ast.AssignStmt)
				if !isInner {
					return true
				}

				for _, to := range inner.Lhs {
					name, isName := to.(*ast.Ident)
					sets = sets || (isName && pass.TypesInfo.ObjectOf(name) == pass.TypesInfo.ObjectOf(failed))
				}

				return true
			})
			if !sets {
				continue
			}

			assign, isAssign := before[i].(*ast.AssignStmt)
			if !isAssign || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
				return true
			}

			parse, isCall := assign.Rhs[0].(*ast.CallExpr)
			if !isCall {
				return true
			}
			call, parsed = parse, assign
		}

		if call == nil {
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

		// wrapped wherever the check does it: returned, or added to a list of errors
		ast.Inspect(check.Body, func(n ast.Node) bool {
			wrap, isWrap := n.(*ast.CallExpr)
			if !isWrap || len(wrap.Args) == 0 || !astx.IsUseOf(pass, wrap.Args[len(wrap.Args)-1], pass.TypesInfo.ObjectOf(failed)) {
				return true
			}

			if errorf := astx.CalledFunc(pass, wrap); errorf != nil && errorf.FullName() == "fmt.Errorf" {
				wrapped[astx.EnclosingFile(pass, wrap.Pos())], parsedBy[wrap] = append(wrapped[astx.EnclosingFile(pass, wrap.Pos())], wrap), parsed
			}

			return true
		})

		return true
	})

	for file, wraps := range wrapped {
		// when every use of fmt in the file goes with the fixes, the last one takes the import with it
		uses, removed := 0, 0
		ast.Inspect(file, func(n ast.Node) bool {
			if id, isIdent := n.(*ast.Ident); isIdent {
				if pkg, isPkg := pass.TypesInfo.Uses[id].(*types.PkgName); isPkg && pkg.Imported().Path() == "fmt" {
					uses++
					if slices.ContainsFunc(wraps, func(wrap *ast.CallExpr) bool { return wrap.Pos() <= id.Pos() && id.End() <= wrap.End() }) {
						removed++
					}
				}
			}

			return true
		})

		var dropImport *analysis.TextEdit
		if uses == removed {
			dropImport = astx.DeleteImport(pass, file, "fmt")
		}

		renamed := map[types.Object]bool{}
		for i, wrap := range wraps {
			failed := wrap.Args[len(wrap.Args)-1]
			d := analysis.Diagnostic{Pos: wrap.Pos(), End: wrap.End(), Message: "the ID parser's error already says what is wrong with the ID and what a right one looks like - return it as it is"}
			edits := []analysis.TextEdit{{Pos: wrap.Pos(), End: failed.Pos()}, {Pos: failed.End(), End: wrap.End()}}
			fixable := uses > removed || dropImport != nil

			// an ID read only to be printed in the message is not used once the message goes, so it is not kept
			if name, isName := parsedBy[wrap].Lhs[0].(*ast.Ident); isName && name.Name != "_" && !renamed[pass.TypesInfo.ObjectOf(name)] {
				// a variable that is only ever set is unused too, so what counts is where it is read
				printed, written := 0, 0
				for _, other := range wraps {
					printed += astx.UseCount(pass, other, pass.TypesInfo.ObjectOf(name))
				}
				ast.Inspect(file, func(n ast.Node) bool {
					if assign, isAssign := n.(*ast.AssignStmt); isAssign {
						for _, to := range assign.Lhs {
							if astx.IsUseOf(pass, to, pass.TypesInfo.ObjectOf(name)) {
								written++
							}
						}
					}

					return true
				})

				if printed > 0 && astx.UseCount(pass, file, pass.TypesInfo.ObjectOf(name))-written == printed {
					// a variable declared somewhere else would be left unused there
					fixable = fixable && pass.TypesInfo.Defs[name] != nil
					renamed[pass.TypesInfo.ObjectOf(name)] = true
					edits = append(edits, analysis.TextEdit{Pos: name.Pos(), End: name.End(), NewText: []byte("_")})

					// := needs a new name on its left, which the error may not be
					if failure, isFailure := parsedBy[wrap].Lhs[1].(*ast.Ident); isFailure && parsedBy[wrap].Tok == token.DEFINE && pass.TypesInfo.Defs[failure] == nil {
						edits = append(edits, analysis.TextEdit{Pos: parsedBy[wrap].TokPos, End: parsedBy[wrap].TokPos + 2, NewText: []byte("=")})
					}
				}
			}

			if i == len(wraps)-1 && dropImport != nil {
				edits = append(edits, *dropImport)
			}
			if fixable {
				d.SuggestedFixes = []analysis.SuggestedFix{{Message: "return the parser's error", TextEdits: edits}}
			}
			pass.Report(d)
		}
	}

	return nil, nil
}
