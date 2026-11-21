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

	all := facts{values: true, isBool: true, hasDefault: true, forceNew: true, conflicts: true, oneOf: true, items: true, length: true}

	for _, text := range []string{
		"The name of the Thing.",
		"The owner, a user (optional) or a group.",
		"Limit the name to 64 characters to avoid collisions.",
		"The mode, which by default follows the `speed`.",
		"By default only routing information is allowed.",
		"Only one `rule` block can be specified unless `mode` is `Fast`.",
		"One of the `rule` blocks is used first.",
		"Unsetting this after configuring it will force a new resource to be created.",
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
		"The mode, which defaults to `a`.",
		"The mode. Can be `a` or `b`.",
		"The key. Exactly one of `key` or `secret` must be specified.",
		"The rules. A maximum of `100` `rule` blocks may be defined.",
		"The name. Must be between 1 and 80 characters.",
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
		either   = "description says one of several must be set - the schema's ExactlyOneOf or AtLeastOneOf already does"
		counted  = "description says how many there can be - the schema's MaxItems or MinItems already does"
		long     = "description says how long it can be - the schema's validation already does"
		page     = "description points at the docs page - a description has no page around it"
	)

	for text, want := range map[string][]string{
		"(Optional) The name.":                                             {`description has the tag "(Optional)" - the schema already says so`},
		"The name. (Computed)":                                             {`description has the tag "(Computed)" - the schema already says so`},
		"(Optional/Computed) The name.":                                    {`description has the tag "(Optional/Computed)" - the schema already says so`},
		"(List) The names (String).":                                       {`description has the tag "(List)" - the schema already says so`, `description has the tag "(String)" - the schema already says so`},
		"The mode. Possible values are `a` and `b`.":                       {values},
		"The mode. Allowed values include `a`.":                            {values},
		"The mode. Must be one of `a` or `b`.":                             {values},
		"The size. Should be between 1 and 5.":                             {values},
		"The mode. Default is `a`.":                                        {byValue},
		"The mode. The default value is `a`.":                              {byValue},
		"The mode. Defaults to `a`. Possible values are `a`.":              {byValue, values},
		"The mode. Changing this forces a new Thing to be created.":        {replaced},
		"The mode. Changing this value forces a new Thing to be created.":  {replaced},
		"The mode. Cannot be used together with `speed`.":                  {clashes},
		"The key. Only one of `key` or `secret` can be specified.":         {either},
		"The mode. Changing this forces a new resource.":                   {replaced},
		"The name. Changing the name forces a new resource to be created.": {replaced},
		"The mode. Valid options are `a` or `b`.":                          {values},
		"The verb. Options are 'GET', 'POST', and 'PUT'.":                  {values},
		"The mode. Allowed values: `a`, `b`.":                              {values},
		"The mode. Can be `a` or `b`.":                                     {values},
		"The mode. Either `a` or `b`.":                                     {values},
		"The mode. One of `a`, `b` or `c`.":                                {values},
		"The mode, which defaults to `Fast`.":                              {byValue},
		"The mode (defaults to `Fast`) of the Thing.":                      {byValue},
		"The site. If blank, defaults to 'Default-First-Site-Name'.":       {byValue},
		"The mode. This defaults to `a` if not specified.":                 {byValue},
		"Is it on? By default, this is set to false.":                      {byValue},
		"The key. Exactly one of `key` or `secret` must be specified.":     {either},
		"The key. One of either `key` or `secret` is required.":            {either},
		"The key. At least one of `key`, `secret` or `token` must be set.": {either},
		"The rules. A maximum of `100` `rule` blocks may be defined.":      {counted},
		"The rules. Can support up to `5` `rule` blocks.":                  {counted},
		"A `rule` block. Only one `rule` block can be specified.":          {counted},
		"The name. Must be between 1 and 80 characters.":                   {long},
		"The name. Cannot exceed 80 characters in length.":                 {long},
		"The name, letters and dashes up to 80 characters in length.":      {long},
		"The rules. Each `rule` block supports fields defined below.":      {page},
		"The `sku` block supports the fields documented below. The SKU.":   {page},
		"A `sku` block as defined below.":                                  {`description points at the docs page with "as defined below" - a description has no page around it`},
		"One or more `sku` blocks, described above.":                       {`description points at the docs page with "described above" - a description has no page around it`},
	} {
		var got []string
		for _, finding := range check(text, all) {
			got = append(got, finding.Message)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	// what is taken out goes with one space, so none is left doubled or leading
	for text, want := range map[string]string{
		"(Defaults to `a`) The mode.":                                                   "The mode.",
		"The mode (defaults to `a`) of the Thing.":                                      "The mode of the Thing.",
		"The mode, which defaults to `a`.":                                              "The mode.",
		"The mode; defaults to `a`. Used by the API.":                                   "The mode. Used by the API.",
		"The rules. Each `rule` block supports fields defined below. Applied in order.": "The rules. Applied in order.",
	} {
		got := text
		for _, finding := range slices.Backward(check(text, all)) {
			got = got[:finding.Start] + finding.New + got[finding.End:]
		}

		if got != want {
			t.Errorf("%q: fixed to %q, want %q", text, got, want)
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
		"one-of":          "The key. Exactly one of `key` or `secret` must be specified.",
		"item-count":      "The rules. A maximum of `100` `rule` blocks may be defined.",
		"length":          "The name, letters and dashes up to 80 characters in length.",
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

	if len(enabled) != 9 {
		t.Errorf("%d kinds can be switched off, and the test covers 9", len(enabled))
	}
}
