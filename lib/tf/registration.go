// Package tf defines helpers shared by checks that reason about terraform provider code:
// plugin SDK schema literals, service registration methods and data source files.
package tf

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// RegistrationReturnShape reports whether fn returns either a map[string]*Resource (untyped
// registration) or a slice of a named Resource/DataSource/FrameworkWrapped* type
// (typed/framework registration), which guards against unrelated methods sharing the names.
func RegistrationReturnShape(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}

	ret := pass.TypesInfo.TypeOf(fn.Type.Results.List[0].Type)
	switch t := types.Unalias(ret).(type) {
	case *types.Map:
		basic, ok := types.Unalias(t.Key()).(*types.Basic)
		if !ok || basic.Kind() != types.String {
			return false
		}

		ptr, ok := types.Unalias(t.Elem()).(*types.Pointer)
		if !ok {
			return false
		}

		named, ok := types.Unalias(ptr.Elem()).(*types.Named)
		return ok && named.Obj().Name() == "Resource"
	case *types.Slice:
		named, ok := types.Unalias(t.Elem()).(*types.Named)
		if !ok {
			return false
		}
		switch named.Obj().Name() {
		case "Resource", "DataSource", "FrameworkWrappedResource", "FrameworkWrappedDataSource":
			return true
		}
	}

	return false
}
