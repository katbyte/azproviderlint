package tf

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// InDataSourceFile reports whether the node sits in a file whose name marks it as a data
// source file.
func InDataSourceFile(pass *analysis.Pass, n ast.Node) bool {
	filename := filepath.Base(pass.Fset.Position(n.Pos()).Filename)
	return strings.Contains(filename, "data_source")
}

// InStateMigration reports whether the package holds state migrations, whose schemas are what old state looked like: nobody sets them and no
// docs are made from them. Providers name the package migration or migrations.
func InStateMigration(pass *analysis.Pass) bool {
	return pass.Pkg.Name() == "migration" || pass.Pkg.Name() == "migrations"
}
