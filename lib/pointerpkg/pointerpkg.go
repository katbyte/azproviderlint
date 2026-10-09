// Package pointerpkg locates the go-azure-helpers pointer package within a file so fixes can
// reference it, adding an import edit when the file does not import it yet.
package pointerpkg

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	// PkgPath is the import path of the go-azure-helpers pointer package.
	PkgPath = "github.com/hashicorp/go-azure-helpers/lang/pointer"
	// PkgName is the package's default reference name when imported without an alias.
	PkgName = "pointer"
)

// Ref returns the name the pointer package is (or would be) referenced by in file, and, when
// the file does not yet import it, a TextEdit inserting the import in sorted position within
// the file's import block. Files with no parenthesized import block (or a dot/blank import of
// the package) get no reference.
func Ref(file *ast.File) (string, *analysis.TextEdit, bool) {
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path != PkgPath {
			continue
		}
		if imp.Name == nil {
			return PkgName, nil, true
		}
		if imp.Name.Name == "." || imp.Name.Name == "_" {
			return "", nil, false
		}
		return imp.Name.Name, nil, true
	}

	// not imported: insert into the first parenthesized import declaration. Import blocks are
	// conventionally organised into gci-style sections (standard library, then side-effect
	// imports, then everything else), so the new import goes in sorted position among the
	// existing non-stdlib imports only — never between standard-library ones — and opens a new
	// section after the block's final import when the file has none. Layouts beyond that
	// remain gci's job.
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT || !gen.Lparen.IsValid() || len(gen.Specs) == 0 {
			continue
		}

		newImport := `"` + PkgPath + `"`
		insertAfter := token.NoPos
		var insertBefore *ast.ImportSpec
		for _, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}

			path := strings.Trim(imp.Path.Value, `"`)
			if !strings.Contains(strings.SplitN(path, "/", 2)[0], ".") {
				continue // standard library: a different section
			}
			if imp.Name != nil && imp.Name.Name == "_" {
				continue // side-effect imports form their own section
			}

			if path < PkgPath {
				// a trailing comment is not part of the spec's End; inserting between the two
				// would re-attach the comment (e.g. a nolint directive) to the new import
				insertAfter = imp.End()
				if imp.Comment != nil {
					insertAfter = imp.Comment.End()
				}
			} else if insertBefore == nil {
				insertBefore = imp
			}
		}

		if insertAfter.IsValid() {
			return PkgName, &analysis.TextEdit{Pos: insertAfter, End: insertAfter, NewText: []byte("\n\t" + newImport)}, true
		}

		if insertBefore != nil {
			first := insertBefore.Pos()
			// keep a doc comment attached to the spec it documents rather than the new import
			if insertBefore.Doc != nil {
				first = insertBefore.Doc.Pos()
			}
			return PkgName, &analysis.TextEdit{Pos: first, End: first, NewText: []byte(newImport + "\n\t")}, true
		}

		// only standard-library or side-effect imports: open a new section after the last one
		end := gen.Specs[len(gen.Specs)-1].End()
		if imp, ok := gen.Specs[len(gen.Specs)-1].(*ast.ImportSpec); ok && imp.Comment != nil {
			end = imp.Comment.End()
		}

		return PkgName, &analysis.TextEdit{Pos: end, End: end, NewText: []byte("\n\n\t" + newImport)}, true
	}

	return "", nil, false
}
