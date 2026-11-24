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
	start, end int
	by         string
}

// Claim reports whether the rule's fix may rewrite the bytes from start to end. When two rules rewrite the same bytes golangci-lint drops every
// fix in the file, so the first rule to claim a place keeps its fix and the others report without one until it has been applied.
func Claim(pass *analysis.Pass, start, end token.Pos) bool {
	at := pass.Fset.Position(start)
	mine := claim{start: at.Offset, end: at.Offset + int(end-start), by: pass.Analyzer.Name}

	claimsLock.Lock()
	defer claimsLock.Unlock()

	if slices.ContainsFunc(claims[at.Filename], func(c claim) bool { return c.by != mine.by && mine.start <= c.end && c.start <= mine.end }) {
		return false
	}
	if !slices.Contains(claims[at.Filename], mine) {
		claims[at.Filename] = append(claims[at.Filename], mine)
	}

	return true
}
