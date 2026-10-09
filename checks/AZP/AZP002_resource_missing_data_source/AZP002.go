// Package AZP002 defines an analyzer that reports registered resources that have no
// corresponding data source of the same name, across every registration flavour: untyped
// plugin SDK maps, typed sdk.Resource slices and framework wrapped resource slices.
package AZP002

import (
	"go/ast"
	"strings"

	"github.com/katbyte/azproviderlint/lib/tf"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer correlates a service package's registered resources and data sources by their
// terraform type name and reports every resource without a matching data source. All
// registration flavours contribute to both sides:
//
//   - untyped plugin SDK: `SupportedResources()` / `SupportedDataSources()` map keys
//   - typed SDK: `Resources()` / `DataSources()` elements' `ResourceType()` methods
//   - framework: `FrameworkResources()` / `FrameworkDataSources()` elements' `ResourceType()` methods
//
// Conditionally registered entries (`m["azurerm_x"] = ...`, `append(out, Foo{})` behind
// feature flags) are collected too. If any data source entry's name cannot be resolved the
// package is skipped entirely, so an unresolvable data source can never produce a false
// "missing data source" report.
var Analyzer = &analysis.Analyzer{
	Name:     "AZP002",
	Doc:      "check for registered resources that have no corresponding data source",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZP/AZP002_resource_missing_data_source/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// actionStyleSuffixes marks invoke-style resources — ones that perform an operation or mint a
// credential rather than manage a durable object (the kind that will eventually become
// framework Actions). These have no meaningful data source form, so they are never reported.
// There is no structural marker for them in any registration flavour, so this is a name
// convention list.
var actionStyleSuffixes = []string{
	"_run_command",
	"_sas_token",
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	var resources []tf.Entry
	dataSources := map[string]bool{}
	resolvable := true

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Body == nil {
			return
		}

		isResource := tf.ResourceMethods[fn.Name.Name]
		isDataSource := tf.DataSourceMethods[fn.Name.Name]
		if !isResource && !isDataSource {
			return
		}

		if !tf.RegistrationReturnShape(pass, fn) {
			return
		}

		entries, allResolved := tf.RegistrationEntries(pass, fn.Body)
		if isDataSource {
			if !allResolved {
				// an unresolvable data source name could hide a match — skip the package
				resolvable = false
				return
			}
			for _, e := range entries {
				dataSources[e.Name] = true
			}
			return
		}

		// unresolvable resource entries are simply skipped: they can only under-report
		resources = append(resources, entries...)
	})

	if !resolvable {
		return nil, nil
	}

	reported := map[string]bool{}
	for _, r := range resources {
		if dataSources[r.Name] || reported[r.Name] || isActionStyle(r.Name) {
			continue
		}
		reported[r.Name] = true
		pass.Reportf(r.Pos, "resource %q has no corresponding data source", r.Name)
	}

	return nil, nil
}

// isActionStyle reports whether a resource name marks an invoke-style resource that has no
// meaningful data source form.
func isActionStyle(name string) bool {
	for _, suffix := range actionStyleSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
