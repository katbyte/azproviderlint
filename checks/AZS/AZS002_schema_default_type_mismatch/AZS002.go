// Package AZS002 defines an analyzer that reports schema declarations whose Default value
// does not match the declared schema Type.
package AZS002

import (
	"go/ast"
	"go/constant"
	"slices"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that a Schema's Default value is compatible with its declared Type, e.g.
// `Type: schema.TypeInt` must not declare `Default: true`. The plugin SDK's InternalValidate
// does not type-check Default, so a mismatch only surfaces as an error at plan time.
// Constant values are resolved through the type checker, so named constants
// (`Default: SkuStandard`) are checked too; non-constant defaults are skipped.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS002",
	Doc:      "check that a schema's Default value matches its declared Type",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS002_schema_default_type_mismatch/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// compatibleKinds maps a schema value type to the constant kinds its Default may use.
// TypeFloat accepts int constants (`Default: 1` is a valid float default); list/set/map
// schema types cannot have literal defaults and are out of scope.
var compatibleKinds = map[string][]constant.Kind{
	"TypeBool":   {constant.Bool},
	"TypeInt":    {constant.Int},
	"TypeFloat":  {constant.Int, constant.Float},
	"TypeString": {constant.String},
}

var kindNames = map[constant.Kind]string{
	constant.Bool:    "bool",
	constant.Int:     "int",
	constant.Float:   "float",
	constant.String:  "string",
	constant.Complex: "complex",
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return
		}

		if !tf.IsSchemaHelperType(pass, cl, "Schema") {
			return
		}

		var typeExpr, defaultExpr ast.Expr
		for _, elt := range cl.Elts {
			kv, isKV := elt.(*ast.KeyValueExpr)
			if !isKV {
				continue
			}
			key, isIdent := kv.Key.(*ast.Ident)
			if !isIdent {
				continue
			}
			switch key.Name {
			case "Type":
				typeExpr = kv.Value
			case "Default":
				defaultExpr = kv.Value
			}
		}

		if typeExpr == nil || defaultExpr == nil {
			return
		}

		var schemaType string
		switch v := typeExpr.(type) {
		case *ast.SelectorExpr:
			schemaType = v.Sel.Name
		case *ast.Ident:
			schemaType = v.Name
		}

		compatible, ok := compatibleKinds[schemaType]
		if !ok {
			return
		}

		// nil or a non-constant expression - nothing to compare statically
		value := pass.TypesInfo.Types[defaultExpr].Value
		if value == nil {
			return
		}

		kindName, ok := kindNames[value.Kind()]
		if !ok || slices.Contains(compatible, value.Kind()) {
			return
		}

		pass.Reportf(defaultExpr.Pos(), "schema Default value type %s does not match the declared Type %s", kindName, schemaType)
	})

	return nil, nil
}
