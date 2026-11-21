// Package AZS016 defines an analyzer that reports Azure product names written in lowercase in resource and property descriptions.
package AZS016

import (
	"cmp"
	"flag"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/katbyte/azproviderlint/lib/description"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer checks that the name of an Azure product in a description is capitalised the way the provider writes it everywhere else: "Function
// App", not "function app". Code in backticks and web addresses are left alone, since there the case is part of the value.
var Analyzer = &analysis.Analyzer{
	Name:     "AZS016",
	Doc:      "check that Azure product names in resource and property descriptions are capitalised",
	URL:      "https://github.com/katbyte/azproviderlint/blob/main/checks/AZS/AZS016_description_product_name_case/README.md",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// names holds how each product name is written, by its lowercase form, with its plural beside it. Each is two plain words or more: a single word
// is too often an everyday one, and an acronym is AZS013's to check. Names that are also general terms, such as "load balancer", are left out.
var names = map[string]string{}

// find matches any of the names, in any case.
var find *regexp.Regexp

func init() {
	list(true, strings.Split("Function App, Web App, Static Web App, Logic App, Container App, Container App Environment, App Service, App Service Plan, App Service Environment, Service Plan, App Configuration, Application Insights, Application Gateway, Key Vault, Key Vault Key, Key Vault Secret, Key Vault Certificate, Storage Account, Storage Container, Storage Share, Storage Queue, Storage Table, Resource Group, Management Group, Resource Manager, Resource Provider, Virtual Network, Virtual Network Gateway, Local Network Gateway, Virtual Machine, Virtual Machine Scale Set, Virtual Hub, Network Security Group, Network Watcher, Private Endpoint, Private Link, Front Door, Traffic Manager, Managed Identity, Managed Service Identity, User Assigned Identity, User Assigned Managed Identity, System Assigned Identity, System Assigned Managed Identity, Service Principal, Role Assignment, Role Definition, Active Directory, Log Analytics, Log Analytics Workspace, Service Bus, Event Hub, Event Grid, Notification Hub, Data Factory, Data Lake, Container Group, Automation Account, Batch Account, Cognitive Account, Recovery Services Vault, Backup Vault, Site Recovery, Managed Disk, Disk Encryption Set, Availability Zone, Availability Set, Dedicated Host, Proximity Placement Group, Shared Image Gallery, Redis Cache, Search Service, Machine Learning Workspace, Stream Analytics Job, Digital Twins, Spring Cloud, Service Fabric, Synapse Workspace, Databricks Workspace, Deployment Slot", ", ")...)

	Analyzer.Flags.Init("AZS016", flag.ContinueOnError)
	Analyzer.Flags.Func("name", "another product name to check, written the way it should be; repeat for several", func(name string) error {
		list(true, name)
		return nil
	})
	Analyzer.Flags.Func("ignore", "a product name to leave alone; repeat for several", func(name string) error {
		list(false, name)
		return nil
	})
}

// list changes which names are checked: each one given is added, or taken away, along with its plural.
func list(checked bool, given ...string) {
	for _, name := range given {
		plural := name + "s"
		switch {
		case strings.HasSuffix(name, "s"):
			plural = name

		case strings.HasSuffix(name, "y") && !strings.ContainsAny(name[len(name)-2:len(name)-1], "aeiou"):
			plural = strings.TrimSuffix(name, "y") + "ies"
		}

		for _, form := range []string{name, plural} {
			if checked {
				names[strings.ToLower(form)] = form
			} else {
				delete(names, strings.ToLower(form))
			}
		}
	}

	// the longest comes first, so an "App Service Plan" is not read as an "App Service"
	forms := slices.SortedFunc(maps.Keys(names), func(a, b string) int { return cmp.Or(cmp.Compare(len(b), len(a)), cmp.Compare(a, b)) })
	for i, form := range forms {
		forms[i] = regexp.QuoteMeta(form)
	}

	find = regexp.MustCompile(`(?i)\b(?:` + strings.Join(forms, "|") + `)\b`)
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil
	}

	for _, found := range description.All(pass, insp) {
		for _, finding := range check(found.Text) {
			found.Report(pass, finding)
		}
	}

	return nil, nil
}

// check returns each product name in the description that is not capitalised the way it should be.
func check(text string) []description.Finding {
	masked := description.Mask(text)
	joined := "-" + string(description.Blank)

	var findings []description.Finding
	for _, at := range find.FindAllStringIndex(masked, -1) {
		// joined to something else by a hyphen, or running into code, it is part of a longer name: "my-function app", "function app-level"
		if (at[0] > 0 && strings.ContainsRune(joined, rune(masked[at[0]-1]))) || (at[1] < len(masked) && strings.ContainsRune(joined, rune(masked[at[1]]))) {
			continue
		}

		written := masked[at[0]:at[1]]
		if want, ok := names[strings.ToLower(written)]; ok && written != want {
			findings = append(findings, description.Finding{Start: at[0], End: at[1], Message: fmt.Sprintf("%q should be %q", written, want), Fix: "Write " + want, New: want})
		}
	}

	return findings
}
