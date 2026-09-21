package astx

import (
	"go/ast"
	"go/token"
	"strings"

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
