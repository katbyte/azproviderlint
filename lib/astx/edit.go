package astx

import (
	"go/ast"
	"go/token"
	"slices"
	"strings"
	"sync"

	"golang.org/x/tools/go/analysis"
)

// DeleteLine returns the edit removing node together with the rest of its line (a trailing
// comma and comment included), so no blank line is left behind; when other code shares the
// line only the node itself goes.
func DeleteLine(pass *analysis.Pass, node ast.Node) analysis.TextEdit {
	tf := pass.Fset.File(node.Pos())
	start, end := node.Pos(), node.End()
	if tf != nil {
		line := tf.Line(start)
		lineStart := tf.LineStart(line)
		var lineEnd token.Pos
		if line < tf.LineCount() {
			lineEnd = tf.LineStart(line + 1)
		} else {
			lineEnd = token.Pos(tf.Base() + tf.Size())
		}
		if content, err := pass.ReadFile(tf.Name()); err == nil {
			before := strings.TrimSpace(string(content[tf.Offset(lineStart):tf.Offset(start)]))
			after := strings.TrimSpace(string(content[tf.Offset(end):tf.Offset(lineEnd)]))
			after = strings.TrimSpace(strings.TrimPrefix(after, ","))
			if before == "" && (after == "" || strings.HasPrefix(after, "//")) {
				start, end = lineStart, lineEnd
			}
		}
	}
	return analysis.TextEdit{Pos: start, End: end}
}

// claims holds, for each file, the bytes each rule's fixes rewrite.
var (
	claims     = map[string][]claim{}
	claimsLock sync.Mutex
)

type claim struct {
	file       string
	start, end int
	by         string
}

// OneFixPerPlace makes every rule give up a fix that rewrites bytes another rule's fix already does. When two rules rewrite the same bytes
// golangci-lint drops every fix in the file, so the first rule to fix a place keeps its fix and the others report without one until it has
// been applied. A fix is kept or given up whole, however many edits it is made of. One that rewrites an import takes the whole file, since
// another rule's fix may be about to use the name that import gives.
func OneFixPerPlace(analyzers []*analysis.Analyzer) []*analysis.Analyzer {
	for _, a := range analyzers {
		run, name := a.Run, a.Name
		a.Run = func(pass *analysis.Pass) (any, error) {
			report := pass.Report
			pass.Report = func(d analysis.Diagnostic) {
				var kept []analysis.SuggestedFix
				for _, fix := range d.SuggestedFixes {
					var mine []claim
					for _, edit := range fix.TextEdits {
						at, size := pass.Fset.Position(edit.Pos), 0
						if edit.End.IsValid() {
							size = int(edit.End - edit.Pos)
						}
						mine = append(mine, claim{file: at.Filename, start: at.Offset, end: at.Offset + size, by: name})

						if file := EnclosingFile(pass, edit.Pos); file != nil && slices.ContainsFunc(file.Imports, func(spec *ast.ImportSpec) bool { return edit.Pos < spec.End() && spec.Pos() < edit.End }) {
							mine = append(mine, claim{file: at.Filename, end: pass.Fset.File(edit.Pos).Size(), by: name})
						}
					}

					claimsLock.Lock()
					taken := slices.ContainsFunc(mine, func(m claim) bool {
						return slices.ContainsFunc(claims[m.file], func(c claim) bool { return c.by != m.by && m.start <= c.end && c.start <= m.end })
					})
					if !taken {
						for _, m := range mine {
							claims[m.file] = append(claims[m.file], m)
						}
						kept = append(kept, fix)
					}
					claimsLock.Unlock()
				}

				d.SuggestedFixes = kept
				report(d)
			}

			return run(pass)
		}
	}

	return analyzers
}
