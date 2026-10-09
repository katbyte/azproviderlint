// Package AZS001 defines an analyzer that reports tfschema-tagged typed SDK model fields
// using non-64-bit numeric types where the SDK's Encode/Decode requires int64/float64.
package AZS001

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
	"strconv"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that typed SDK model structs (identified by `tfschema` struct tags)
// use 64-bit numeric types: int64 rather than int/int16/int32, and float64 rather than float32.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS001",
	Doc:      "check that tfschema-tagged model fields use 64-bit numeric types (int64/float64) as required by the typed SDK's Encode/Decode",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS001_typed_sdk_model_64bit_types/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return
		}
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return
		}

		checkModelStruct(pass, typeSpec.Name.Name, structType)
	})

	return nil, nil
}

func checkModelStruct(pass *analysis.Pass, modelName string, structType *ast.StructType) {
	qualifier := types.RelativeTo(pass.Pkg)

	for _, field := range structType.Fields.List {
		// Only fields with a `tfschema` tag are encoded/decoded by the typed SDK
		if !hasTFSchemaTag(field) {
			continue
		}

		fieldType := pass.TypesInfo.TypeOf(field.Type)
		if fieldType == nil {
			continue
		}

		want, underlying, bad := badType(fieldType, qualifier)
		if !bad {
			continue
		}

		got := types.ExprString(field.Type)
		if got != underlying {
			// named type or alias, show what it resolves to
			got = fmt.Sprintf("%s (underlying %s)", got, underlying)
		}

		for _, name := range fieldNames(field) {
			pass.Reportf(field.Pos(), "property %s in model %s should be type %s, got %s", name, modelName, want, got)
		}
	}
}

// hasTFSchemaTag reports whether the field carries a `tfschema:"..."` struct tag.
func hasTFSchemaTag(field *ast.Field) bool {
	if field.Tag == nil {
		return false
	}

	tag, err := strconv.Unquote(field.Tag.Value)
	if err != nil {
		return false
	}

	_, ok := reflect.StructTag(tag).Lookup("tfschema")
	return ok
}

// badType walks the type's underlying structure (so named types and aliases resolve,
// matching what reflect.Kind() sees at runtime) and, if it bottoms out in a disallowed
// numeric type, returns the equivalent type built on int64/float64 along with the
// underlying structural type for the diagnostic.
func badType(t types.Type, qualifier types.Qualifier) (want, underlying string, bad bool) {
	switch u := t.Underlying().(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.Int, types.Int16, types.Int32:
			return "int64", u.Name(), true
		case types.Float32:
			return "float64", u.Name(), true
		default:
		}
	case *types.Pointer:
		if w, un, ok := badType(u.Elem(), qualifier); ok {
			return "*" + w, "*" + un, true
		}
	case *types.Slice:
		if w, un, ok := badType(u.Elem(), qualifier); ok {
			return "[]" + w, "[]" + un, true
		}
	case *types.Array:
		if w, un, ok := badType(u.Elem(), qualifier); ok {
			return "[]" + w, "[]" + un, true
		}
	case *types.Map:
		if w, un, ok := badType(u.Elem(), qualifier); ok {
			key := types.TypeString(u.Key(), qualifier)
			return "map[" + key + "]" + w, "map[" + key + "]" + un, true
		}
	}

	return "", "", false
}

func fieldNames(field *ast.Field) []string {
	if len(field.Names) == 0 {
		// embedded field, use the rendered type as the name
		return []string{types.ExprString(field.Type)}
	}

	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	return names
}
