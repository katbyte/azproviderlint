// Package azuresdk identifies generated go-azure-sdk types.
package azuresdk

import (
	"go/types"
	"strings"
)

// IsEnumType reports whether named is a go-azure-sdk enum type: a named type with a
// string underlying type, declared in a go-azure-sdk package, that exposes the generated
// `PossibleValuesFor<Name>() []T` helper.
func IsEnumType(named *types.Named) bool {
	basic, ok := named.Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsString == 0 {
		return false
	}

	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !strings.HasPrefix(pkg.Path(), "github.com/hashicorp/go-azure-sdk/") {
		return false
	}

	fn, ok := pkg.Scope().Lookup("PossibleValuesFor" + obj.Name()).(*types.Func)
	if !ok {
		return false
	}

	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Params().Len() != 0 || sig.Results().Len() != 1 {
		return false
	}

	slice, ok := sig.Results().At(0).Type().(*types.Slice)
	if !ok {
		return false
	}

	elem, ok := slice.Elem().(*types.Basic)
	return ok && elem.Kind() == basic.Kind()
}
