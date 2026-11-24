package AZS014

import (
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS014(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs014")

	all := facts{Values: true, IsBool: true, HasDefault: true, ForceNew: true, Conflicts: true, OneOf: true, Requires: true, Items: true, Length: true}

	for _, text := range []string{
		"The name of the Thing.",
		"The owner, a user (optional) or a group.",
		"Limit the name to 64 characters to avoid collisions.",
		"The mode, which by default follows the `speed`.",
		"By default only routing information is allowed.",
		"Only one `rule` block can be specified unless `mode` is `Fast`.",
		"One of the `rule` blocks is used first.",
		"Unsetting this after configuring it will force a new resource to be created.",
		"The URL. Required if `mode` is `Fast`.",
		"If specified, `secret` must also be set to `true`.",
		"Only one is supported at a time.",
		"The range, where valid values are checked by the API.",
		"The value set by default is `Fast`.",
		"Changing this is fine.",
		"The values possible are many.",
		"The mode. Values are case sensitive.",
		"The window. If `frequency` is `PT1M`, possible values are `PT1M` and `PT5M`.",
		"The time zone, the possible values are defined [here](https://example.com/zones).",
		"The feature to use, accepted values are `a` or `b`, it follows the SKU when left out.",
		"If the `http_url` property is specified, the `file_path` is required and must be a relative path.",
		"Whether to preserve the unmatched part of the path.",
		"The owners, `a` or `b` among them.",
		"The first of the names, `a` or `b`.",
		"",
	} {
		if got := check(text, all); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// with nothing in the schema to repeat, only what is never right in a description is left
	for _, text := range []string{
		"The mode. Possible values are `a` and `b`.",
		"The mode. Changing this forces a new resource to be created.",
		"The mode. Conflicts with `speed`.",
		"The mode. Can be `a` or `b`.",
		"The key. Exactly one of `key` or `secret` must be specified.",
		"The rules. A maximum of `100` `rule` blocks may be defined.",
		"The name. Must be between 1 and 80 characters.",
		"The URL. `branch` and `token` must also be set.",
		"The type of the runbook - can be either `a`, `b` or `c`.",
		"List of permissions, must be one or more from the following: `a`, `b` and `c`.",
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
		clashes  = "description says what it conflicts with - the schema's ConflictsWith or ExactlyOneOf already does"
		needs    = "description says what else must be set - the schema's RequiredWith already does"
		either   = "description says one of several must be set - the schema's ExactlyOneOf or AtLeastOneOf already does"
		counted  = "description says how many there can be - the schema's MaxItems or MinItems already does"
		long     = "description says how long it can be - the schema's validation already does"
		page     = "description points at the docs page - a description has no page around it"
	)

	for text, want := range map[string][]string{
		"(Optional) The name.":                                                                      {`description has the tag "(Optional)" - the schema already says so`},
		"The name. (Computed)":                                                                      {`description has the tag "(Computed)" - the schema already says so`},
		"(Optional/Computed) The name.":                                                             {`description has the tag "(Optional/Computed)" - the schema already says so`},
		"(List) The names (String).":                                                                {`description has the tag "(List)" - the schema already says so`, `description has the tag "(String)" - the schema already says so`},
		"The mode. Possible values are `a` and `b`.":                                                {values},
		"The mode. Allowed values include `a`.":                                                     {values},
		"The mode. Values are `a`, `b` and `c`.":                                                    {values},
		"The mode. Values include 'a' and 'b'.":                                                     {values},
		"The mode. Acceptable values include `a`, `b` and `c`.":                                     {values},
		"The mode. Must be one of `a` or `b`.":                                                      {values},
		"The size. Should be between 1 and 5.":                                                      {values},
		"The size. The range is `100` - `65000`.":                                                   {values},
		"The size. Valid range is 1-5.":                                                             {values},
		"The mode. Default is `a`.":                                                                 {byValue},
		"The mode. The default value is `a`.":                                                       {byValue},
		"The mode. Defaults to `a`. Possible values are `a`.":                                       {byValue, values},
		"The mode. Changing this forces a new Thing to be created.":                                 {replaced},
		"The mode. Changing this value forces a new Thing to be created.":                           {replaced},
		"The mode. Cannot be used together with `speed`.":                                           {clashes},
		"The key. Only one of `key` or `secret` can be specified.":                                  {either},
		"The mode. Changing this forces a new resource.":                                            {replaced},
		"The name. Changing the name forces a new resource to be created.":                          {replaced},
		"The mode. Valid options are `a` or `b`.":                                                   {values},
		"The mode. Permitted values: `a`, `b`, or `c`.":                                             {values},
		"The mode. Possibles values are `a` and `b`.":                                               {values},
		"The mode. Currently only `a` is supported.":                                                {values},
		"The mode. Only `a` is supported currently.":                                                {values},
		"The size. This value can be between `1` and `7` days.":                                     {values},
		"The size. This number should be between 1 and 365.":                                        {values},
		"The mode of the site Possible values include: `a` and `b`.":                                {values},
		"The type of the runbook - can be either `a`, `b` or `c`.":                                  {values},
		"The type of validation, one of `a` or `b`.":                                                {values},
		"The type of validation - either `a` or `b`.":                                               {values},
		"The URL. `branch` and `token` must also be set.":                                           {needs},
		"The URL. Required with `branch`.":                                                          {needs},
		"The URL. This attribute is required when `branch` is specified.":                           {needs},
		"The URL. The `uid` and `gid` properties must be specified together or not at all.":         {needs},
		"The URL. If the `http_url` property is specified, the `file_path` is required.":            {needs},
		"The URL. When `a` is set, `b` must also be set.":                                           {needs},
		"The mode of the Thing, possible values are `a` and `b`.":                                   {values},
		"The mode of the Thing - valid values include: `a`, `b`, `c`.":                              {values},
		"The tier of the SKU, possible values are `Standard`.":                                      {values},
		"The priority of the Job, possible values can range from -1000 (lowest) to 1000 (highest).": {values},
		"The mode of the Thing, can be set to `a`, `b` or `c`.":                                     {values},
		"The mode. Currently the only possible values are `a` and `b`.":                             {values},
		"The scale unit of the Gateway, possible values are `1` through `40`.":                      {values},
		"The status of the request, possible values will be `a`, `b`, or `c`.":                      {values},
		"The mode. can be set to `a`, `b` or `c`.":                                                  {values},
		"The mode. At this time the supported values are `a` and `b`.":                              {values},
		"The URL. This is required with `branch`.":                                                  {needs},
		"The URL. Requires `branch` to be set.":                                                     {needs},
		"The URL. Required if `branch` is set.":                                                     {needs},
		"The path. Only valid in conjunction with `health_check_path`.":                             {needs},
		"The verb. Options are 'GET', 'POST', and 'PUT'.":                                           {values},
		"The mode. Allowed values: `a`, `b`.":                                                       {values},
		"The mode. Can be `a` or `b`.":                                                              {values},
		"The mode. Either `a` or `b`.":                                                              {values},
		"The mode. One of `a`, `b` or `c`.":                                                         {values},
		"The mode, which defaults to `Fast`.":                                                       {byValue},
		"The mode (defaults to `Fast`) of the Thing.":                                               {byValue},
		"The site. If blank, defaults to 'Default-First-Site-Name'.":                                {byValue},
		"The mode. This defaults to `a` if not specified.":                                          {byValue},
		"Is it on? By default, this is set to false.":                                               {byValue},
		"The key. Exactly one of `key` or `secret` must be specified.":                              {either},
		"The key. Either `id` or both `key` and `location` must be specified.":                      {either},
		"The key. One of either `key` or `secret` is required.":                                     {either},
		"The key. At least one of `key`, `secret` or `token` must be set.":                          {either},
		"The rules. A maximum of `100` `rule` blocks may be defined.":                               {counted},
		"The rules. Can support up to `5` `rule` blocks.":                                           {counted},
		"A `rule` block. Only one `rule` block can be specified.":                                   {counted},
		"The name. Must be between 1 and 80 characters.":                                            {long},
		"The name. Cannot exceed 80 characters in length.":                                          {long},
		"The name, letters and dashes up to 80 characters in length.":                               {long},
		"The rules. Each `rule` block supports fields defined below.":                               {page},
		"The `sku` block supports the fields documented below. The SKU.":                            {page},
		"A `sku` block as defined below.":                                                           {`description points at the docs page with "as defined below" - a description has no page around it`},
		"One or more `sku` blocks, described above.":                                                {`description points at the docs page with "described above" - a description has no page around it`},
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
		"(Defaults to `a`) The mode.":                                                                                   "The mode.",
		"The mode (defaults to `a`) of the Thing.":                                                                      "The mode of the Thing.",
		"The mode, which defaults to `a`.":                                                                              "The mode.",
		"The mode; defaults to `a`. Used by the API.":                                                                   "The mode. Used by the API.",
		"The mode of the site Possible values include: `a` and `b`. Used by the API.":                                   "The mode of the site. Used by the API.",
		"The type of the runbook - can be either `a`, `b` or `c`.":                                                      "The type of the runbook.",
		"The minimum TLS version for the SCM site Possible values include: `1.0`, `1.1` and  `1.2`. Defaults to `1.2`.": "The minimum TLS version for the SCM site.",
		"The rules. Each `rule` block supports fields defined below. Applied in order.":                                 "The rules. Applied in order.",
	} {
		got := text
		for _, finding := range slices.Backward(check(text, all)) {
			got = got[:finding.Start] + finding.New + got[finding.End:]
		}

		if got != want {
			t.Errorf("%q: fixed to %q, want %q", text, got, want)
		}
	}

	if got := check("Is it on? Possible values are `true` and `false`.", facts{IsBool: true}); len(got) != 1 || got[0].Message != byType {
		t.Errorf("a bool's values: got %v", got)
	}

	// a default the schema does not set is still called out where it is written, and left there
	for text, want := range map[string]string{
		"Is it on? Defaults to `false`.":           "Defaults to `false`.",
		"The mode, which defaults to `a`.":         ", which defaults to `a`",
		"The mode (defaults to `a`) of the Thing.": " (defaults to `a`)",
		"If blank, defaults to 'x'. The site.":     "If blank, defaults to 'x'.",
	} {
		got := check(text, facts{})
		if len(got) != 1 || got[0].Message != unsetDefault || got[0].Fix != "" || text[got[0].Start:got[0].End] != want {
			t.Errorf("%q with no Default in the schema: got %v", text, got)
		}
	}

	// on a read-only property each of them is reported and taken out with nothing in the schema to repeat; a tag and a pointer at the docs page
	// are reported as they are anywhere
	for text, want := range map[string][2]string{
		"The state. Possible values include `a` and `b`.":               {"description lists the values it can take" + neverSet, "The state."},
		"The type. Acceptable values include `a`, `b` and `c`.":         {"description lists the values it can take" + neverSet, "The type."},
		"The state. Defaults to `a`.":                                   {"description gives the default" + neverSet, "The state."},
		"The state, which defaults to `a`.":                             {"description gives the default" + neverSet, "The state."},
		"The state (defaults to `a`) of the Thing.":                     {"description gives the default" + neverSet, "The state of the Thing."},
		"The state. Changing this forces a new resource to be created.": {"description says a change forces a new resource" + neverSet, "The state."},
		"The state. Conflicts with `mode`.":                             {"description says what it conflicts with" + neverSet, "The state."},
		"The state. Required with `mode`.":                              {"description says what else must be set" + neverSet, "The state."},
		"The key. Exactly one of `key` or `secret` must be specified.":  {"description says one of several must be set" + neverSet, "The key."},
		"The rules. A maximum of `100` `rule` blocks may be defined.":   {"description says how many there can be" + neverSet, "The rules."},
		"The name. Must be between 1 and 80 characters.":                {"description says how long it can be" + neverSet, "The name."},
		"(Computed) The state.":                                         {`description has the tag "(Computed)" - the schema already says so`, "The state."},
		"The rules. Each `rule` block supports fields defined below.":   {page, "The rules."},
	} {
		got := check(text, facts{ReadOnly: true})
		if len(got) != 1 || got[0].Message != want[0] || text[:got[0].Start]+got[0].New+text[got[0].End:] != want[1] {
			t.Errorf("%q on a read-only property: got %v", text, got)
		}
	}

	// a read-only property described as something to set is called out by the words that say so, and left to be reworded
	for text, want := range map[string]string{
		"Whether to preserve the unmatched part of the path.": "Whether to",
		"Whether the Thing should be enabled.":                "should",
		"Set to `true` to keep the path.":                     "Set",
		"Enable to allow secured volumes.":                    "Enable",
	} {
		got := check(text, facts{ReadOnly: true})
		if len(got) != 1 || got[0].Fix != "" || text[got[0].Start:got[0].End] != want || !strings.HasSuffix(got[0].Message, "say what it holds") {
			t.Errorf("%q on a read-only property: got %v", text, got)
		}
	}

	if got := check("Whether the Thing is enabled. Settings are kept.", facts{ReadOnly: true}); len(got) > 0 {
		t.Errorf("a read-only property saying what it holds: got %v", got)
	}

	if got := check("The state of the Thing, as last read.", facts{ReadOnly: true}); len(got) > 0 {
		t.Errorf("prose on a read-only property: got %v", got)
	}

	if got := check("The value set by default is `Fast`.", facts{}); len(got) > 0 {
		t.Errorf("prose about a default: got %v", got)
	}

	// a range nothing in the schema checks is called out and left there, naming the validation when the schema tells which
	for text, want := range map[string]string{
		"The time. This value must be between 5 and 300 seconds.": "ISO8601DurationBetween",
		"The size. Should be at most `5`.":                        "ISO8601DurationBetween",
		"The size. Possible values range from 1 to 5.":            "ISO8601DurationBetween",
		"The size. The range is `100` - `65000`.":                 "ISO8601DurationBetween",
		"The name. Must be between 1 and 80 characters.":          "StringLenBetween",
	} {
		got := check(text, facts{Unchecked: true, Bounds: "ISO8601DurationBetween"})
		if len(got) != 1 || got[0].Fix != "" || !strings.HasSuffix(got[0].Message, "check it in the schema with "+want) {
			t.Errorf("%q with nothing checking it: got %v", text, got)
		}
	}

	if got := check("The size. Must be between 1 and 5.", facts{Unchecked: true}); len(got) != 1 || !strings.HasSuffix(got[0].Message, "with a validation") {
		t.Errorf("a range with no validation to name: got %v", got)
	}

	if got := check("The size, which must be between 1 and 5 unless `mode` is set.", facts{Unchecked: true}); len(got) > 0 {
		t.Errorf("a range inside a sentence that says more: got %v", got)
	}

	// each kind of repeat can be switched off on its own; the kinds are package state, so this cannot run beside anything else
	for kind, text := range map[string]string{
		"tag":               "(Optional) The name.",
		"possible-values":   "The mode. Possible values are `a` and `b`.",
		"default":           "The mode. Defaults to `a`.",
		"force-new":         "The mode. Changing this forces a new resource to be created.",
		"conflicts-with":    "The mode. Conflicts with `speed`.",
		"defined-below":     "A `sku` block as defined below.",
		"one-of":            "The key. Exactly one of `key` or `secret` must be specified.",
		"required-with":     "The URL. `branch` and `token` must also be set.",
		"item-count":        "The rules. A maximum of `100` `rule` blocks may be defined.",
		"length":            "The name, letters and dashes up to 80 characters in length.",
		"unchecked-range":   "The size. Must be between 1 and 5.",
		"read-only-setting": "Whether to keep the path.",
	} {
		known := all
		switch kind {
		case unchecked:
			known = facts{Unchecked: true}
		case setting:
			known = facts{ReadOnly: true}
		}

		if len(check(text, known)) == 0 {
			t.Errorf("%s: %q should be reported", kind, text)
		}

		*enabled[kind] = false
		if got := check(text, known); len(got) > 0 {
			t.Errorf("%s switched off: %q still reports %q", kind, text, got[0].Message)
		}
		*enabled[kind] = true
	}

	if len(enabled) != 12 {
		t.Errorf("%d kinds can be switched off, and the test covers 12", len(enabled))
	}
}
