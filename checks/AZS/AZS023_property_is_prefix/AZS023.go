// Package AZS023 defines an analyzer that reports a schema property whose name starts with is_.
package AZS023

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks the names properties are given in a schema for a leading is_, which says nothing the rest of the name does not: is_storage_enabled
// is storage_enabled.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS023",
	Doc:      "check for schema properties whose name starts with is_",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS023_property_is_prefix/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const prefix = "is_"

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// a state migration's schema is what old state looked like, names and all
	if !ok || tf.InStateMigration(pass) {
		return nil, nil
	}

	insp.Preorder([]ast.Node{(*ast.CompositeLit)(nil), (*ast.AssignStmt)(nil)}, func(n ast.Node) {
		// a property is named where a map of them is written out, and where one is added to a map later
		var properties ast.Expr
		var names []ast.Expr
		switch n := n.(type) {
		case *ast.CompositeLit:
			properties = n
			for _, elt := range n.Elts {
				if kv, isPair := elt.(*ast.KeyValueExpr); isPair {
					names = append(names, kv.Key)
				}
			}

		case *ast.AssignStmt:
			if at, isIndex := n.Lhs[0].(*ast.IndexExpr); len(n.Lhs) == 1 && isIndex {
				properties, names = at.X, []ast.Expr{at.Index}
			}
		}
		if properties == nil || pass.TypesInfo.TypeOf(properties) == nil {
			return
		}

		// the plugin sdk and the framework both keep their schema types in a package named schema
		byName, isMap := pass.TypesInfo.TypeOf(properties).Underlying().(*types.Map)
		if !isMap {
			return
		}

		of := byName.Elem()
		if ptr, isPointer := types.Unalias(of).(*types.Pointer); isPointer {
			of = ptr.Elem()
		}

		named, isNamed := types.Unalias(of).(*types.Named)
		if !isNamed || named.Obj().Pkg() == nil || named.Obj().Pkg().Name() != "schema" || (named.Obj().Name() != "Schema" && named.Obj().Name() != "Attribute" && named.Obj().Name() != "Block") {
			return
		}

		for _, name := range names {
			tv, isName := pass.TypesInfo.Types[name]
			if !isName || tv.Value == nil || tv.Value.Kind() != constant.String {
				continue
			}

			if rest, found := strings.CutPrefix(constant.StringVal(tv.Value), prefix); found && rest != "" {
				pass.Reportf(name.Pos(), "property %q starts with %s - name it %q", prefix+rest, prefix, rest)
			}
		}
	})

	return nil, nil
}
