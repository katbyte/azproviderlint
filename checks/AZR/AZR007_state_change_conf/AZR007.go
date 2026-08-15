// Package AZR007 defines an analyzer that reports StateChangeConf usage, which
// should be replaced with a custom poller implementing the pollers.PollerType interface.
package AZR007

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// stateChangeConfPkgPath is the import path of the Plugin SDK helper package that actually
// declares StateChangeConf. Provider wrappers such as azurerm's
// `internal/tf/pluginsdk.StateChangeConf` are just type aliases for this type.
const stateChangeConfPkgPath = "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

// Analyzer checks for `StateChangeConf{...}` composite literals. Going forward the
// provider prefers custom pollers that implement the go-azure-sdk `pollers.PollerType`
// interface and are driven via `pollers.NewPoller(...).PollUntilDone(ctx)`.
//
// Reference: https://github.com/hashicorp/terraform-provider-azurerm/pull/30066
var Analyzer = &analysis.Analyzer{
	Name:     "AZR007",
	Doc:      "check for StateChangeConf usage that should use a Custom Poller instead",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR007_state_change_conf/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
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
		lit, ok := n.(*ast.CompositeLit)
		if !ok || lit.Type == nil {
			return
		}

		// Resolve the literal's type, unwrapping any alias so that provider wrappers such as
		// azurerm's `pluginsdk.StateChangeConf` (an alias for `retry.StateChangeConf`) resolve
		// to the same underlying named type as a direct `retry.StateChangeConf` usage.
		named, ok := types.Unalias(pass.TypesInfo.TypeOf(lit.Type)).(*types.Named)
		if !ok {
			return
		}

		obj := named.Obj()
		if obj.Name() != "StateChangeConf" {
			return
		}

		if pkg := obj.Pkg(); pkg == nil || pkg.Path() != stateChangeConfPkgPath {
			return
		}

		pass.Reportf(lit.Pos(),
			"prefer a Custom Poller over StateChangeConf")
	})

	return nil, nil
}
