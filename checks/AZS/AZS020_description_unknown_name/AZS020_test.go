package AZS020

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS020(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.Run(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs020")

	known := func(name string) bool {
		return slices.Contains([]string{"example_thing", "site_config", "linux_fx_version", "read_only"}, name)
	}

	for _, text := range []string{
		"The name of the Thing.",
		"See `example_thing`, `example_thing.main.id` and `example_thing.main.no_such_thing`.",
		"Any of `example_thing_*` or `example_*`.",
		"Set in `site_config.0.linux_fx_version`, `site_config[0].linux_fx_version` or `site_config.*.linux_fx_version`.",
		"One of `read_only`, `enabled`, `Standard_LRS`, `PT1M`, `application/json`, `a-b`, `my_key = value` or `MY_VAR`.",
		"Another provider's `other_thing`, when it is known.",
		"One word is a value unless it is called a name: `skew`, `standard` (the default).",
		"A name with no backticks, such as no_such_thing.",
		"",
	} {
		if got := check(text, "example", func(name string) bool { return known(name) || name == "other_thing" }); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	for text, want := range map[string][]string{
		"See `example_thingy`.":                              {"example_thingy"},
		"See `example_thingy.main.id`.":                      {"example_thingy"},
		"See `example_thingy.main.id;rev=1`.":                {"example_thingy"},
		"Set in `site_config.0.linux_fx`.":                   {"linux_fx"},
		"Set in the `skew` block.":                           {"skew"},
		"The `skew` property and `tier` fields.":             {"skew", "tier"},
		"Set in `site_conf[0].linux_fx_version`.":            {"site_conf"},
		"Conflicts with `no_such_thing` and `nor_this_one`.": {"no_such_thing", "nor_this_one"},
	} {
		findings := check(text, "example", known)
		got := make([]string, 0, len(findings))
		for _, finding := range findings {
			if finding.Fix != "" {
				t.Errorf("%q: a name cannot be guessed, so there is no fix", text)
			}
			got = append(got, text[finding.Start:finding.End])
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	// with no provider name to go by, a resource is checked as any other name
	if got := check("See `example_thingy`.", "", known); len(got) != 1 || got[0].Message != `description names "example_thingy" - nothing in the provider is called that` {
		t.Errorf("no provider name: got %v", got)
	}

	// a name the consumer allows is accepted; the allowed names are package state, so this cannot run beside anything else
	if err := Analyzer.Flags.Set("allow", "azuread_application"); err != nil {
		t.Fatal(err)
	}
	if !allowed["azuread_application"] || !allowed["ignore_changes"] {
		t.Errorf("allowed names: got %v", allowed)
	}
	delete(allowed, "azuread_application")
}
