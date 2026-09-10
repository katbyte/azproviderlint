// Package AZR009 defines an analyzer that reports lifecycle narration logging — "preparing
// arguments for", "Creating %s", "Decoding state.." — inside a resource's Create, Read, Update,
// or Delete function, which the provider no longer does.
package AZR009

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/lifecycle"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks Create/Read/Update/Delete functions (untyped resources' registered functions
// and typed resources' `Create() sdk.ResourceFunc` methods) for log statements that narrate
// the step — `log.Printf("[INFO] preparing arguments for ...")`, `log.Printf("[DEBUG]
// Retrieving %s", id)`, `metadata.Logger.Infof("creating %s", id)`, `metadata.Logger.Info(
// "Decoding state..")`. hashicorp/terraform-provider-azurerm#32423 removed the pattern
// provider-wide: the framework already logs each step, so these lines are noise. Messages about
// the resource ID (`Updating ID from %s to %s`) record a state migration and are left alone,
// as are "not found - removing from state" and "Waiting for" messages. The suggested fix
// deletes the statement, and the `log` import once nothing else in the file uses it.
var Analyzer = &analysis.Analyzer{
	Name:     "AZR009",
	Doc:      "check for lifecycle narration logging inside Create/Read/Update/Delete functions",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR009_lifecycle_logging/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	// levelPrefix is the log-level tag the message may start with.
	levelPrefix = regexp.MustCompile(`^\[(TRACE|DEBUG|INFO|WARN|ERROR)\]\s*`)
	// narration matches a message that merely announces the lifecycle step or a sub-step.
	narration = regexp.MustCompile(`(?i)^(preparing (the )?arguments|creating|updating|deleting|retrieving|reading|importing|import check|decoding( the)? state|checking for( the| an| a)? (presence|existence|existing)|created|updated|deleted|retrieved)\b`)
	// keep matches messages that carry information beyond narration: ID rewrites (a state
	// migration), state removal, polling, and a step being skipped.
	keep = regexp.MustCompile(`(?i)\bid\b|not found|removing from state|does not exist|waiting|skip`)
)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	type finding struct {
		stmt *ast.ExprStmt
		call *ast.CallExpr
	}
	byFile := map[*ast.File][]finding{}
	for fn := range lifecycle.Funcs(insp) {
		file := enclosingFile(pass, fn.Pos())
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			stmt, ok := n.(*ast.ExprStmt)
			if !ok {
				return true
			}
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok || !isLogCall(pass, call) || len(call.Args) == 0 {
				return true
			}
			lit, ok := call.Args[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			msg, err := strconv.Unquote(lit.Value)
			if err != nil {
				return true
			}
			msg = levelPrefix.ReplaceAllString(strings.TrimSpace(msg), "")
			if narration.MatchString(msg) && !keep.MatchString(msg) {
				byFile[file] = append(byFile[file], finding{stmt: stmt, call: call})
			}
			return true
		})
	}

	for file, findings := range byFile {
		// the `log` import can go with the last fix in the file when every use of the
		// package is one of the statements being removed
		var importEdit *analysis.TextEdit
		if uses := logPackageUses(pass, file); len(uses) > 0 {
			removed := 0
			for _, f := range findings {
				for _, u := range uses {
					if f.stmt.Pos() <= u.Pos() && u.Pos() < f.stmt.End() {
						removed++
					}
				}
			}
			if removed == len(uses) {
				importEdit = deleteLogImport(pass, file)
			}
		}
		for i, f := range findings {
			edits := []analysis.TextEdit{astx.DeleteLine(pass, f.stmt)}
			if i == len(findings)-1 && importEdit != nil {
				edits = append(edits, *importEdit)
			}
			pass.Report(analysis.Diagnostic{
				Pos:     f.call.Pos(),
				Message: "lifecycle logging should be removed - the framework already logs each Create/Read/Update/Delete step (hashicorp/terraform-provider-azurerm#32423)",
				SuggestedFixes: []analysis.SuggestedFix{{
					Message:   "Remove the log statement",
					TextEdits: edits,
				}},
			})
		}
	}
	return nil, nil
}

// isLogCall reports whether call is log.Print/Printf/Println from the standard library or a
// metadata.Logger Info/Infof/Warn/Warnf/Debug/Debugf method.
func isLogCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Print", "Printf", "Println":
		id, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}
		pkg, ok := pass.TypesInfo.Uses[id].(*types.PkgName)
		return ok && pkg.Imported().Path() == "log"
	case "Info", "Infof", "Warn", "Warnf", "Debug", "Debugf":
		inner, ok := sel.X.(*ast.SelectorExpr)
		return ok && inner.Sel.Name == "Logger"
	}
	return false
}

// logPackageUses returns every identifier in file that refers to the imported `log` package.
func logPackageUses(pass *analysis.Pass, file *ast.File) []*ast.Ident {
	var uses []*ast.Ident
	ast.Inspect(file, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if pkg, ok := pass.TypesInfo.Uses[id].(*types.PkgName); ok && pkg.Imported().Path() == "log" {
				uses = append(uses, id)
			}
		}
		return true
	})
	return uses
}

// deleteLogImport returns the edit removing the `log` import: its whole declaration when it
// stands alone, otherwise just its line inside the import block.
func deleteLogImport(pass *analysis.Pass, file *ast.File) *analysis.TextEdit {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		for _, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if !ok || imp.Path.Value != `"log"` {
				continue
			}
			if len(gen.Specs) == 1 {
				e := astx.DeleteLine(pass, gen)
				return &e
			}
			e := astx.DeleteLine(pass, imp)
			return &e
		}
	}
	return nil
}

// enclosingFile returns the file in the pass containing pos.
func enclosingFile(pass *analysis.Pass, pos token.Pos) *ast.File {
	for _, f := range pass.Files {
		if f.FileStart <= pos && pos < f.FileEnd {
			return f
		}
	}
	return nil
}
