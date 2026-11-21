package AZS016

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS016(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs016")

	for _, text := range []string{
		"The Storage Account used by the Function App, its App Service Plan and their Managed Identities.",
		"The `storage account` of the function_app.",
		"See https://example.com/key vault or [the docs](https://example.com/resource group) for more.",
		"A name such as my-function app, or a function app-level setting.",
		"Removed from the load balancer, the network interface and the container registry.",
		"The storage accounting period and the webapp service.",
		"The Api Management service, which the acronym check reports.",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	for text, want := range map[string][]string{
		"The storage account.":                                               {`"storage account" should be "Storage Account"`},
		"Function app or FUNCTION APPS in a Web app":                         {`"Function app" should be "Function App"`, `"FUNCTION APPS" should be "Function Apps"`, `"Web app" should be "Web App"`},
		"An app service plan, not an app service.":                           {`"app service plan" should be "App Service Plan"`, `"app service" should be "App Service"`},
		"The key vault keys of the user assigned Managed Identity.":          {`"key vault keys" should be "Key Vault Keys"`, `"user assigned Managed Identity" should be "User Assigned Managed Identity"`},
		"The *resource group* (or management group)'s application insights.": {`"resource group" should be "Resource Group"`, `"management group" should be "Management Group"`, `"application insights" should be "Application Insights"`},
	} {
		findings := check(text)
		got := make([]string, 0, len(findings))
		for _, finding := range findings {
			got = append(got, finding.Message)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	// the options add a name and drop one; they are package state, so this cannot run beside another test in this package
	for option, name := range map[string]string{"name": "Load Balancer", "ignore": "Storage Account"} {
		if err := Analyzer.Flags.Set(option, name); err != nil {
			t.Fatal(err)
		}
	}

	findings := check("The storage accounts behind the load balancers.")
	got := make([]string, 0, len(findings))
	for _, finding := range findings {
		got = append(got, finding.Message)
	}

	if want := []string{`"load balancers" should be "Load Balancers"`}; !slices.Equal(got, want) {
		t.Errorf("with Load Balancer added and Storage Account dropped: got %q, want %q", got, want)
	}

	list(false, "Load Balancer")
	list(true, "Storage Account")
}
