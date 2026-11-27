// Package AZS021 defines an analyzer that reports a block that can be given once and holds a single property.
package AZS021

import (
	"fmt"
	"go/ast"
	"go/constant"
	"strings"

	"github.com/katbyte/azproviderlint/lib/astx"
	"github.com/katbyte/azproviderlint/lib/property"
	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for a MaxItems: 1 block with one property in it, which only adds a level of nesting: the property can sit where the block does.
// A comment on the block says why it is kept, such as more properties being on the way.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS021",
	Doc:      "check for MaxItems: 1 blocks that hold a single property and have no comment saying why",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS021_single_property_block/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer, property.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, which cannot be given another shape
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	// a block is named by its key in the map it sits in, which is met before the block is
	keys := map[*ast.CompositeLit]ast.Expr{}
	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(n ast.Node) {
		// a schema made up inside a test is never shipped
		lit, isLit := n.(*ast.CompositeLit)
		if !isLit || strings.HasSuffix(pass.Fset.Position(n.Pos()).Filename, "_test.go") {
			return
		}

		fields := map[string]ast.Expr{}
		for _, elt := range lit.Elts {
			kv, isPair := elt.(*ast.KeyValueExpr)
			if !isPair {
				continue
			}

			value := kv.Value
			if ref, ok := value.(*ast.UnaryExpr); ok {
				value = ref.X
			}
			if inner, ok := value.(*ast.CompositeLit); ok {
				keys[inner] = kv.Key
			}
			if key, ok := kv.Key.(*ast.Ident); ok {
				fields[key.Name] = kv.Value
			}
		}

		if !tf.IsSchemaHelperType(pass, lit, "Schema") {
			return
		}

		most, ok := pass.TypesInfo.Types[fields["MaxItems"]]
		if !ok || most.Value == nil || most.Value.Kind() != constant.Int || most.Value.ExactString() != "1" {
			return
		}

		// the block's properties are written out in it, or handed back by a function as its Elem or Schema, and then counted but not named
		count, only := 0, ast.Expr(nil)
		if given, isGiven := property.BlockOf(pass, fields["Elem"]); isGiven {
			count = given.Properties
		}

		if ref, isRef := fields["Elem"].(*ast.UnaryExpr); isRef {
			if resource, isResource := ref.X.(*ast.CompositeLit); isResource && tf.IsSchemaHelperType(pass, resource, "Resource") {
				for _, elt := range resource.Elts {
					kv, isPair := elt.(*ast.KeyValueExpr)
					if key, isField := kv.Key.(*ast.Ident); !isPair || !isField || key.Name != "Schema" {
						continue
					}

					if given, isGiven := property.BlockOf(pass, kv.Value); isGiven {
						count = given.Properties
					}
					if written, isWritten := kv.Value.(*ast.CompositeLit); isWritten {
						count = len(written.Elts)
						if count == 1 {
							if first, isFirst := written.Elts[0].(*ast.KeyValueExpr); isFirst {
								only = first.Key
							}
						}
					}
				}
			}
		}

		if count != 1 {
			return
		}

		start := ast.Node(lit)
		if key := keys[lit]; key != nil {
			start = key
		}

		tfile, file := pass.Fset.File(lit.Pos()), astx.EnclosingFile(pass, lit.Pos())
		src, err := pass.ReadFile(tfile.Name())
		if file == nil || err != nil {
			return
		}

		// a comment on a line of its own right above the block says why it is kept; one inside it is about a field, and a note for a linter is not a reason
		for _, comment := range file.Comments {
			line := tfile.Line(comment.Pos())
			if strings.TrimSpace(string(src[tfile.Offset(tfile.LineStart(line)):tfile.Offset(comment.Pos())])) != "" || tfile.Line(comment.End()) != tfile.Line(start.Pos())-1 {
				continue
			}

			for _, written := range comment.List {
				if text := strings.TrimSpace(strings.TrimPrefix(written.Text, "//")); text != "" && !strings.Contains(text, "lintignore:") && !strings.Contains(text, "nolint") && !strings.Contains(text, "azignore:") {
					return
				}
			}
		}

		block, held := "block", "one property"
		if name, ok := pass.TypesInfo.Types[keys[lit]]; ok && name.Value != nil && name.Value.Kind() == constant.String {
			block = fmt.Sprintf("block %q", constant.StringVal(name.Value))
		}
		if name, ok := pass.TypesInfo.Types[only]; ok && name.Value != nil && name.Value.Kind() == constant.String {
			held = fmt.Sprintf("only %q", constant.StringVal(name.Value))
		}

		pass.Reportf(start.Pos(), "%s can be given once and holds %s - move it out of the block, or say why the block is kept in a comment above it", block, held)
	})

	return nil, nil
}
