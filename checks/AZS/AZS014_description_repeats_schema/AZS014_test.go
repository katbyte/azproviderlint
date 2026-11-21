package AZS014

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS014(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs014")

	all := facts{values: true, isBool: true, hasDefault: true, forceNew: true, conflicts: true}

	for _, text := range []string{
		"The name of the Thing.",
		"The owner, a user (optional) or a group.",
		"A String of up to 64 characters, or a List of them.",
		"The mode, which defaults to `Fast`.",
		"The value set by default is `Fast`.",
		"Changing this is fine.",
		"The values possible are many.",
		"",
	} {
		if got := check(text, all); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// with nothing in the schema to repeat, only what is never right in a description is left
	for _, text := range []string{
		"The mode. Possible values are `a` and `b`.",
		"The mode. Defaults to `a`.",
		"The mode. Changing this forces a new resource to be created.",
		"The mode. Conflicts with `speed`.",
	} {
		if got := check(text, facts{}); len(got) > 0 {
			t.Errorf("%q with no facts: unexpected %q", text, got[0].Message)
		}
	}

	const (
		values   = "description lists the values it can take - the schema's validation already does"
		byType   = "description lists the values it can take - a bool is always true or false"
		byValue  = "description gives the default - the schema's Default already does"
		replaced = "description says a change forces a new resource - the schema's ForceNew already does"
		clashes  = "description says what it conflicts with - the schema's ConflictsWith already does"
	)

	for text, want := range map[string][]string{
		"(Optional) The name.":                                            {`description has the tag "(Optional)" - the schema already says so`},
		"The name. (Computed)":                                            {`description has the tag "(Computed)" - the schema already says so`},
		"(Optional/Computed) The name.":                                   {`description has the tag "(Optional/Computed)" - the schema already says so`},
		"(List) The names (String).":                                      {`description has the tag "(List)" - the schema already says so`, `description has the tag "(String)" - the schema already says so`},
		"The mode. Possible values are `a` and `b`.":                      {values},
		"The mode. Allowed values include `a`.":                           {values},
		"The mode. Must be one of `a` or `b`.":                            {values},
		"The size. Should be between 1 and 5.":                            {values},
		"The mode. Default is `a`.":                                       {byValue},
		"The mode. The default value is `a`.":                             {byValue},
		"The mode. Defaults to `a`. Possible values are `a`.":             {byValue, values},
		"The mode. Changing this forces a new Thing to be created.":       {replaced},
		"The mode. Changing this value forces a new Thing to be created.": {replaced},
		"The mode. Cannot be used together with `speed`.":                 {clashes},
		"A `sku` block as defined below.":                                 {`description points at the docs page with "as defined below" - a description has no page around it`},
		"One or more `sku` blocks, described above.":                      {`description points at the docs page with "described above" - a description has no page around it`},
	} {
		var got []string
		for _, finding := range check(text, all) {
			got = append(got, finding.Message)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	if got := check("Is it on? Possible values are `true` and `false`.", facts{isBool: true}); len(got) != 1 || got[0].Message != byType {
		t.Errorf("a bool's values: got %v", got)
	}

	// each kind of repeat can be switched off on its own; the kinds are package state, so this cannot run beside anything else
	for kind, text := range map[string]string{
		"tag":             "(Optional) The name.",
		"possible-values": "The mode. Possible values are `a` and `b`.",
		"default":         "The mode. Defaults to `a`.",
		"force-new":       "The mode. Changing this forces a new resource to be created.",
		"conflicts-with":  "The mode. Conflicts with `speed`.",
		"defined-below":   "A `sku` block as defined below.",
	} {
		if len(check(text, all)) == 0 {
			t.Errorf("%s: %q should be reported", kind, text)
		}

		*enabled[kind] = false
		if got := check(text, all); len(got) > 0 {
			t.Errorf("%s switched off: %q still reports %q", kind, text, got[0].Message)
		}
		*enabled[kind] = true
	}

	if len(enabled) != 6 {
		t.Errorf("%d kinds can be switched off, and the test covers 6", len(enabled))
	}
}
