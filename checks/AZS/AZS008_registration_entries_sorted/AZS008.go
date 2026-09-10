// Package AZS008 defines an analyzer that reports registration.go map and slice entries
// that are not sorted alphabetically.
package AZS008

import (
	"bytes"
	"flag"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that map and slice literals matching the result types of `Registration` methods
// in registration.go files are sorted alphabetically. Entries grouped into sections separated by
// blank lines or heading comments are validated within each section independently.
var Analyzer = &analysis.Analyzer{
	Name: "AZS008",
	Doc:  "check that registration.go map and slice entries are sorted alphabetically",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS008_registration_entries_sorted/README.md",
	Run:  run,
}

// checkGenerated also checks generated registration_gen.go files (autoRegistration receiver);
// an unsorted one means the generator's input or template needs fixing.
var checkGenerated bool

func init() {
	Analyzer.Flags.Init("AZS008", flag.ContinueOnError)
	Analyzer.Flags.BoolVar(&checkGenerated, "generated", true,
		"check generated registration_gen.go files (false skips them)")
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		base := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if base != "registration.go" && (!checkGenerated || base != "registration_gen.go") {
			continue
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if hasRegistrationReceiver(funcDecl) {
				analyzeRegistrationMethod(pass, file, funcDecl)
			}
		}
	}

	return nil, nil
}

// hasRegistrationReceiver reports whether the function has a Registration receiver, including
// the generated autoRegistration variant.
func hasRegistrationReceiver(funcDecl *ast.FuncDecl) bool {
	if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
		return false
	}

	recv := funcDecl.Recv.List[0]
	var typeName string

	switch t := recv.Type.(type) {
	case *ast.Ident:
		typeName = t.Name
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			typeName = ident.Name
		}
	}

	return strings.HasSuffix(typeName, "Registration")
}

// analyzeRegistrationMethod validates map or slice literals whose type matches one of the
// method's declared result types.
func analyzeRegistrationMethod(pass *analysis.Pass, file *ast.File, funcDecl *ast.FuncDecl) {
	if funcDecl.Body == nil {
		return
	}

	fn, ok := pass.TypesInfo.Defs[funcDecl.Name].(*types.Func)
	if !ok {
		return
	}
	signature, ok := fn.Type().(*types.Signature)
	if !ok {
		return
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}

		compositeLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		literalType := pass.TypesInfo.TypeOf(compositeLit)
		if literalType == nil {
			return true
		}
		for result := range signature.Results().Variables() {
			if types.Identical(literalType, result.Type()) {
				validateSorting(pass, file, compositeLit)
				break
			}
		}
		return true
	})
}

// splitIntoSections groups composite-literal elements into sections separated by blank lines or
// heading comments. A heading comment has a blank line before it; a comment with no blank line
// before it remains attached to the following entry.
func splitIntoSections(fset *token.FileSet, comments []*ast.CommentGroup, elts []ast.Expr) [][]ast.Expr {
	if len(elts) == 0 {
		return nil
	}

	var sections [][]ast.Expr
	current := []ast.Expr{elts[0]}

	for i := 1; i < len(elts); i++ {
		if startsSection(fset, comments, elts[i-1], elts[i]) {
			sections = append(sections, current)
			current = []ast.Expr{elts[i]}
		} else {
			current = append(current, elts[i])
		}
	}

	return append(sections, current)
}

// startsSection reports whether current begins a new section: it follows a blank line, or a
// standalone comment that has a blank line before it (a heading). A comment with no blank line
// before it is attached to the following entry and does not start a section.
func startsSection(fset *token.FileSet, comments []*ast.CommentGroup, previous, current ast.Expr) bool {
	previousEnd := fset.Position(previous.End()).Line
	currentStart := fset.Position(current.Pos()).Line

	for _, comment := range comments {
		if comment.Pos() <= previous.End() || comment.End() >= current.Pos() {
			continue
		}

		commentStart := fset.Position(comment.Pos()).Line
		if commentStart <= previousEnd {
			continue // trailing comment on the previous entry's line
		}

		// The first standalone comment starts a section only when a blank line precedes it.
		return commentStart > previousEnd+1
	}

	return currentStart > previousEnd+1
}

// validateSorting checks each section independently and reports every unsorted one at its
// first out-of-order entry, naming the keys, with a fix when the section can be safely
// rewritten. Per-section reports keep `//azignore:AZS008` scoped to one section instead of
// silencing the whole method.
func validateSorting(pass *analysis.Pass, file *ast.File, compositeLit *ast.CompositeLit) {
	if compositeLit.Type == nil {
		return
	}

	typ := pass.TypesInfo.TypeOf(compositeLit)
	if typ == nil {
		return
	}

	var isMap bool
	switch typ.Underlying().(type) {
	case *types.Map:
		isMap = true
	case *types.Slice:
	default:
		return
	}

	for _, section := range splitIntoSections(pass.Fset, file.Comments, compositeLit.Elts) {
		names := make([]string, 0, len(section))
		resolvable := true
		for _, elt := range section {
			name, ok := sortKey(elt, isMap)
			if !ok {
				resolvable = false
				break
			}
			names = append(names, name)
		}

		// Skip sections with an unresolvable key rather than judge a partial subset.
		if !resolvable {
			continue
		}

		for i := 1; i < len(names); i++ {
			if lessFold(names[i], names[i-1]) {
				pass.Report(analysis.Diagnostic{
					Pos: section[i].Pos(),
					Message: "registration entries should be sorted alphabetically: `" +
						names[i] + "` should come before `" + names[i-1] + "`",
					SuggestedFixes: sortFix(pass, file, compositeLit, section, isMap),
				})
				break
			}
		}
	}
}

// sortFix builds a suggested fix for one unsorted section when it can be safely rewritten.
func sortFix(pass *analysis.Pass, file *ast.File, compositeLit *ast.CompositeLit, section []ast.Expr, isMap bool) []analysis.SuggestedFix {
	tf := pass.Fset.File(compositeLit.Pos())
	if tf == nil {
		return nil
	}
	if pass.ReadFile == nil {
		return nil
	}

	// Bail out when a brace shares an entry's line: entries are copied as whole lines, so moving
	// one would drag the brace with it (and a trailing brace makes LineStart(endLine+1) panic).
	if elts := compositeLit.Elts; len(elts) > 0 {
		if section[0] == elts[0] && tf.Line(compositeLit.Lbrace) == tf.Line(elts[0].Pos()) {
			return nil
		}
		if section[len(section)-1] == elts[len(elts)-1] &&
			tf.Line(compositeLit.Rbrace) == tf.Line(elts[len(elts)-1].End()) {
			return nil
		}
	}

	content, err := pass.ReadFile(tf.Name())
	if err != nil {
		return nil
	}

	if hasSpanningComment(file.Comments, tf, section) {
		return nil
	}
	edit, ok := sortSectionEdit(tf, content, file.Comments, compositeLit, section, isMap)
	if !ok {
		return nil
	}

	return []analysis.SuggestedFix{{
		Message:   "Sort registration entries alphabetically",
		TextEdits: []analysis.TextEdit{edit},
	}}
}

// hasSpanningComment reports whether a multiline comment crosses source lines occupied by a
// section. Reordering whole lines would split such a comment into separate entry blocks.
func hasSpanningComment(comments []*ast.CommentGroup, tf *token.File, section []ast.Expr) bool {
	startLine := tf.Line(section[0].Pos())
	endLine := tf.Line(section[len(section)-1].End())
	for _, comment := range comments {
		commentStart := tf.Line(comment.Pos())
		commentEnd := tf.Line(comment.End())
		if commentStart != commentEnd && commentStart <= endLine && commentEnd >= startLine {
			return true
		}
	}

	return false
}

// attachedLeadingComment returns a comment directly above entry with no blank lines around it.
// Such a comment is treated as part of the entry so a suggested fix moves them together. For
// the first entry the span starts at the opening brace instead of a previous entry.
func attachedLeadingComment(comments []*ast.CommentGroup, tf *token.File, compositeLit *ast.CompositeLit, entry ast.Expr) *ast.CommentGroup {
	previousEnd := compositeLit.Lbrace
	for i, elt := range compositeLit.Elts {
		if elt == entry {
			if i > 0 {
				previousEnd = compositeLit.Elts[i-1].End()
			}
			break
		}
	}

	for _, comment := range comments {
		if comment.Pos() > previousEnd && comment.End() < entry.Pos() &&
			tf.Line(comment.Pos()) == tf.Line(previousEnd)+1 &&
			tf.Line(entry.Pos()) == tf.Line(comment.End())+1 {
			return comment
		}
	}

	return nil
}

// sortSectionEdit produces a text edit reordering one section's entries alphabetically. Each
// entry is copied as the whole source lines it spans, so its attached comments move with it.
func sortSectionEdit(tf *token.File, content []byte, comments []*ast.CommentGroup, compositeLit *ast.CompositeLit, section []ast.Expr, isMap bool) (analysis.TextEdit, bool) {
	type block struct {
		key  string
		text []byte
	}

	blocks := make([]block, len(section))
	var editStart, editEnd token.Pos
	prevEndLine := 0
	for i, elt := range section {
		key, ok := sortKey(elt, isMap)
		if !ok {
			return analysis.TextEdit{}, false
		}

		startLine := tf.Line(elt.Pos())
		if leadingComment := attachedLeadingComment(comments, tf, compositeLit, elt); leadingComment != nil {
			startLine = tf.Line(leadingComment.Pos())
		}
		endLine := tf.Line(elt.End())
		if i > 0 && startLine <= prevEndLine {
			return analysis.TextEdit{}, false // entries share a line; not safely reorderable
		}
		prevEndLine = endLine

		lineStart := tf.LineStart(startLine)
		lineEnd := tf.LineStart(endLine+1) - 1
		if i == 0 {
			editStart = lineStart
		}
		editEnd = lineEnd
		blocks[i] = block{key: key, text: content[tf.Offset(lineStart):tf.Offset(lineEnd)]}
	}

	slices.SortStableFunc(blocks, func(a, b block) int {
		return strings.Compare(strings.ToLower(a.key), strings.ToLower(b.key))
	})

	var buf bytes.Buffer
	for i, b := range blocks {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.Write(b.text)
	}

	return analysis.TextEdit{Pos: editStart, End: editEnd, NewText: buf.Bytes()}, true
}

// lessFold orders registration keys case-insensitively.
func lessFold(a, b string) bool {
	return strings.ToLower(a) < strings.ToLower(b)
}

// sortKey returns the alphabetical sort key for a registration entry: the map key for map
// literals, the resource struct or constructor name for slice literals.
func sortKey(elt ast.Expr, isMap bool) (string, bool) {
	if !isMap {
		return sliceEntryKey(elt)
	}

	kv, ok := elt.(*ast.KeyValueExpr)
	if !ok {
		return "", false
	}
	if lit, ok := kv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		return sliceEntryKey(lit)
	}
	return "", false
}

// sliceEntryKey extracts the sort key from a typed or framework slice entry: the struct type
// name of a composite literal (`FooResource{}`, `&FooResource{}`, `pkg.FooResource{}`), the
// constructor identifier of a framework entry (`newFooResource`, `pkg.NewFooResource`), or the
// unquoted value of a string literal (`[]string` lists). Any package
// qualifier is dropped so `pkg.FooResource` and `FooResource` sort identically.
func sliceEntryKey(elt ast.Expr) (string, bool) {
	switch e := elt.(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			return sliceEntryKey(e.X)
		}
	case *ast.CompositeLit:
		return sliceEntryKey(e.Type)
	case *ast.Ident:
		return e.Name, true
	case *ast.SelectorExpr:
		return e.Sel.Name, true
	case *ast.BasicLit:
		if e.Kind == token.STRING {
			if s, err := strconv.Unquote(e.Value); err == nil {
				return s, true
			}
		}
	}
	return "", false
}
