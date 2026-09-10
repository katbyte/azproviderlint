// Package AZP005 defines an analyzer that reports assignments to the
// TreatUserSpecifiedSegmentsAsCaseInsensitive feature flag, which must not be configured.
package AZP005

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks for assignments to `features.TreatUserSpecifiedSegmentsAsCaseInsensitive`.
// The case-aware comparisons feature has a substantial number of unresolved dependencies and
// is not ready for use, so it must not be configured or exposed at this time.
var Analyzer = &analysis.Analyzer{
	Name:     "AZP005",
	Doc:      "check for assignments to the TreatUserSpecifiedSegmentsAsCaseInsensitive feature flag which is not ready for use",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP005_case_insensitive_segments_feature_flag/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	nodeFilter := []ast.Node{
		(*ast.AssignStmt)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return
		}

		for _, lhs := range assign.Lhs {
			sel, ok := lhs.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "TreatUserSpecifiedSegmentsAsCaseInsensitive" {
				continue
			}

			pass.Reportf(assign.Pos(),
				"TreatUserSpecifiedSegmentsAsCaseInsensitive must not be set, the case-aware comparisons feature is not ready for use")
		}
	})

	return nil, nil
}
