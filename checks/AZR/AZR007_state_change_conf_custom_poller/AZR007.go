// Package AZR007 defines an analyzer that reports StateChangeConf and Retry usage, which
// should be replaced with a custom poller implementing the pollers.PollerType interface.
package AZR007

import (
	"go/ast"
	"go/types"

	"github.com/katbyte/azproviderlint/lib/astx"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// stateChangeConfPkgPath is the import path of the Plugin SDK helper package that actually
// declares StateChangeConf. Provider wrappers such as azurerm's
// `internal/tf/pluginsdk.StateChangeConf` are just type aliases for this type.
const stateChangeConfPkgPath = "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

// Analyzer checks for `StateChangeConf{...}` composite literals, and for calls to the `Retry` helpers that build one. Going forward the
// provider prefers custom pollers that implement the go-azure-sdk `pollers.PollerType`
// interface and are driven via `pollers.NewPoller(...).PollUntilDone(ctx)`.
//
// Reference: https://github.com/hashicorp/terraform-provider-azurerm/pull/30066
var Analyzer = &analysis.Analyzer{
	Name:     "AZR007",
	Doc:      "check for StateChangeConf and Retry usage that should use a Custom Poller instead",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZR/AZR007_state_change_conf_custom_poller/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	insp.WithStack([]ast.Node{(*ast.CompositeLit)(nil), (*ast.CallExpr)(nil)}, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}

		// Retry and RetryContext build a StateChangeConf: the SDK's own, and a provider's wrapper of them in its pluginsdk package.
		if call, isCall := n.(*ast.CallExpr); isCall {
			fn := astx.CalledFunc(pass, call)
			if fn == nil || fn.Pkg() == nil || fn.Signature().Recv() != nil || (fn.Name() != "Retry" && fn.Name() != "RetryContext") || (fn.Pkg().Path() != stateChangeConfPkgPath && fn.Pkg().Name() != "pluginsdk") {
				return true
			}

			// the wrapper itself is not a wait
			for _, outer := range stack {
				if decl, isDecl := outer.(*ast.FuncDecl); isDecl && (decl.Name.Name == "Retry" || decl.Name.Name == "RetryContext") {
					return true
				}
			}

			pass.Reportf(call.Pos(), "prefer a Custom Poller over %s", fn.Name())

			return true
		}

		lit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Unalias so provider aliases like `pluginsdk.StateChangeConf` resolve to `retry.StateChangeConf`.
		named, ok := types.Unalias(pass.TypesInfo.TypeOf(lit)).(*types.Named)
		if !ok {
			return true
		}

		obj := named.Obj()
		if obj.Name() != "StateChangeConf" {
			return true
		}

		if pkg := obj.Pkg(); pkg == nil || pkg.Path() != stateChangeConfPkgPath {
			return true
		}

		pass.Reportf(lit.Pos(), "prefer a Custom Poller over StateChangeConf")

		return true
	})

	return nil, nil
}
