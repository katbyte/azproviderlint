package tf

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// IsSchemaHelperType reports whether the composite literal's type is the named type from a
// package named "schema", so direct helper/schema use, azurerm's pluginsdk aliases and
// literals elided inside map[string]*Schema all resolve the same.
func IsSchemaHelperType(pass *analysis.Pass, cl *ast.CompositeLit, name string) bool {
	t := pass.TypesInfo.TypeOf(cl)
	if t == nil {
		return false
	}
	if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
		t = ptr.Elem()
	}

	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}

	obj := named.Obj()
	return obj.Name() == name && obj.Pkg() != nil && obj.Pkg().Name() == "schema"
}
