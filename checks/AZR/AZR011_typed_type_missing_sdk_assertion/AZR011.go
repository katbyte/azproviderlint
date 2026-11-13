// Package AZR011 defines an analyzer that reports a typed resource, data source or other sdk type with no line asserting what it is.
package AZR011

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer checks that a type the sdk can register has a `var _ sdk.Resource = FooResource{}` line beside it. The line has the compiler check
// the type where it is declared, and tells a reader what the type is.
var Analyzer = &analysis.Analyzer{
	Name: "AZR011",
	Doc:  "check that typed resources, data sources and other sdk types assert the sdk interface they implement",
	URL:  "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR011_typed_type_missing_sdk_assertion/README.md",
	Run:  run,
}

// what the sdk registers, in the order they are tried: a resource is also everything a data source is, so it is asked first
var kinds = []string{"Resource", "DataSource", "FrameworkWrappedResource", "FrameworkWrappedDataSource", "FrameworkListWrappedResource", "EphemeralResource", "Action"}

func run(pass *analysis.Pass) (any, error) {
	var sdk *types.Package
	for _, imported := range pass.Pkg.Imports() {
		if strings.HasSuffix(imported.Path(), "/internal/sdk") {
			sdk = imported
		}
	}

	if sdk == nil {
		return nil, nil
	}

	// the types that some line already asserts against one of the sdk's interfaces
	asserted := map[*types.TypeName]bool{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}

			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok || value.Type == nil || !checksAgainst(sdk, pass.TypesInfo.TypeOf(value.Type)) {
					continue
				}

				for i, name := range value.Names {
					if name.Name != "_" || i >= len(value.Values) {
						continue
					}

					if obj := named(pass.TypesInfo.TypeOf(value.Values[i])); obj != nil {
						asserted[obj] = true
					}
				}
			}
		}
	}

	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.TypeParams != nil {
					continue
				}

				obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
				if !ok || asserted[obj] || types.IsInterface(obj.Type()) {
					continue
				}

				kind, value := kindOf(sdk, obj)
				if kind == "" {
					continue
				}

				// the line can only be written into a file that already imports the sdk; elsewhere it is shown as it is usually spelt
				qualifier := sdkName(file, sdk)
				line := fmt.Sprintf("var _ %s.%s = %s", cmp.Or(qualifier, "sdk"), kind, value)
				diagnostic := analysis.Diagnostic{
					Pos:     ts.Name.Pos(),
					End:     ts.Name.End(),
					Message: fmt.Sprintf("%s has no %q line to have the compiler check it where it is declared", obj.Name(), line),
				}

				if qualifier != "" {
					diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
						Message:   "add " + line,
						TextEdits: []analysis.TextEdit{{Pos: lineEnd(pass, gen.End()), End: lineEnd(pass, gen.End()), NewText: []byte("\n\n" + line)}},
					}}
				}

				pass.Report(diagnostic)
			}
		}
	}

	return nil, nil
}

// lineEnd returns where the line holding pos ends, so that what is added goes after a comment at the end of that line and not before it.
func lineEnd(pass *analysis.Pass, pos token.Pos) token.Pos {
	file := pass.Fset.File(pos)
	if line := file.Line(pos); line < file.LineCount() {
		return file.LineStart(line+1) - 1
	}

	return token.Pos(file.Base() + file.Size())
}

// kindOf returns the sdk interface a type should assert and the value to assert it with, or two empty strings when the sdk has no use for the
// type. An update is optional and found when the provider runs, so a type that has one asserts the interface that includes it.
func kindOf(sdk *types.Package, obj *types.TypeName) (kind, value string) {
	for _, kind := range kinds {
		for _, candidate := range []struct {
			t     types.Type
			value string
		}{{obj.Type(), obj.Name() + "{}"}, {types.NewPointer(obj.Type()), "&" + obj.Name() + "{}"}} {
			if !implements(sdk, candidate.t, kind) {
				continue
			}

			if implements(sdk, candidate.t, kind+"WithUpdate") {
				return kind + "WithUpdate", candidate.value
			}

			return kind, candidate.value
		}
	}

	return "", ""
}

func implements(sdk *types.Package, t types.Type, name string) bool {
	obj := sdk.Scope().Lookup(name)
	if obj == nil {
		return false
	}

	iface, ok := obj.Type().Underlying().(*types.Interface)

	return ok && types.Implements(t, iface)
}

// checksAgainst reports whether asserting a type against t checks it against the sdk: t is an interface the sdk declares, or the same methods
// under another name, as sdk.Action is the framework's ActionWithConfigure.
func checksAgainst(sdk *types.Package, t types.Type) bool {
	if t == nil || !types.IsInterface(t) {
		return false
	}

	if n, ok := types.Unalias(t).(*types.Named); ok && n.Obj().Pkg() == sdk {
		return true
	}

	for _, kind := range kinds {
		if obj := sdk.Scope().Lookup(kind); obj != nil && types.Identical(t.Underlying(), obj.Type().Underlying()) {
			return true
		}
	}

	return false
}

// named returns the type declared under t's name, through a pointer: FooResource for both FooResource{} and &FooResource{}.
func named(t types.Type) *types.TypeName {
	if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
		t = ptr.Elem()
	}

	if n, ok := types.Unalias(t).(*types.Named); ok {
		return n.Obj()
	}

	return nil
}

// sdkName returns the name a file imports the sdk package under, or an empty string when it does not import it.
func sdkName(file *ast.File, sdk *types.Package) string {
	for _, spec := range file.Imports {
		if strings.Trim(spec.Path.Value, `"`) != sdk.Path() {
			continue
		}

		switch {
		case spec.Name == nil:
			return sdk.Name()
		case spec.Name.Name != "_" && spec.Name.Name != ".":
			return spec.Name.Name
		}
	}

	return ""
}
