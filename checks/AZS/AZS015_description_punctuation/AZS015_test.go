package AZS015

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS015(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs015")

	for _, text := range []string{
		"The name of the Thing.",
		"Should the Thing be enabled?",
		"Should the Thing be enabled? Defaults to `true`.",
		"Must be between `1` and `5`.",
		"Is required when `mode` is set.",
		"Can be one of `A` or `B`.",
		"`sku_name` is made of two parts.",
		"iSCSI targets to connect to.",
		"The time (in seconds.)",
		"The version of .NET to use.",
		"The tag, such as `app:latest`, or an address such as ::1.",
		"A value e.g. `a`, or i.e., the other, etc.",
		"A list of things, etc.",
		"The value... or not.",
		"The options are: a) the first, b) the second.",
		"The rate, between 1,000 and 2,000 at 10:30.",
		"The type, such as Microsoft.Compute/virtualMachines.",
		"See https://example.com/a,b for more.",
		"See <https://example.com/a>.",
		"The foo & bar&nbsp;baz.",
		"Version 2. is the default.",
		"Is set to `true` unless `mode` is given.",
		"Should only be set when `mode` is `Fast`.",
		"Has no effect on existing Things.",
		"The `subnet`s to use, as `a:b` or host:port.",
		"Possible values are ` ` (a space) and `,`.",
		"Uses ``double`` backticks.",
		"The value(s) and address(es) to use, as [a link](https://example.com) or `max(1)`.",
		"The time now() gives.",
		"The mode known as \"bring your own key\".",
		"Terraform's own value, or the user's.",
		"The value, such as \"`a`\" in quotes.",
		"-1 turns it off.",
		"",
		"   ",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	for text, want := range map[string][]string{
		"the name.":                            {"description should start with a capital letter"},
		"[the name](https://example.com).":     {"description should start with a capital letter"},
		"The name":                             {"description should end with a full stop"},
		"The name!":                            {"description should end with a full stop"},
		"The name of `foo`":                    {"description should end with a full stop"},
		"The name (optional)":                  {"description should end with a full stop"},
		"Is `foo` enabled.":                    {"description asks a question, so it should end with a question mark"},
		"Is HTTPS required.":                   {"description asks a question, so it should end with a question mark"},
		"Does this apply to all rules":         {"description asks a question, so it should end with a question mark"},
		"Should public access be enabled.":     {"description asks a question, so it should end with a question mark"},
		"Is detailed logging enabled":          {"description asks a question, so it should end with a question mark"},
		"(Boolean) Is approval required.":      {"description asks a question, so it should end with a question mark"},
		"Should only HTTPS be allowed.":        {"description asks a question, so it should end with a question mark"},
		"Should it be on. Defaults to off.":    {"description asks a question, so it should end with a question mark"},
		"- The name.":                          {`description starts with "-" - start with its first word`},
		"* the name.":                          {`description starts with "*" - start with its first word`, "description should start with a capital letter"},
		"The name\tof it.":                     {"description has a tab in it - use a space"},
		"The name.\nMore here.":                {"description has a line break in it - use a space"},
		"One of`a` or `b`and more.":            {"description has no space before the code", "description has no space after the code"},
		"One of ` a` or `b `.":                 {"description has a space just inside a backtick", "description has a space just inside a backtick"},
		"One of `a``b`.":                       {"description has two pieces of code with nothing between them"},
		"For example:`a`.":                     {"description has no space after ':'"},
		"The Thing.If unset it is off.":        {"description has no space after '.'"},
		"The name. changing this is fine.":     {"sentence should start with a capital letter"},
		"Is it on? defaults to off.":           {"sentence should start with a capital letter"},
		"The name.Changing this is fine.":      {"description has no space after '.'"},
		"The `name`.Changing this is fine.":    {"description has no space after '.'"},
		"One;two.":                             {"description has no space after ';'"},
		"The name , then more.":                {"description has a space before ','"},
		"Is it on ?":                           {"description has a space before '?'"},
		"The name ,then more.":                 {"description has a space before ','", "description has no space after ','"},
		"The [ name ] here.":                   {"description has a space just inside a bracket", "description has a space just inside a bracket"},
		"The name(of the Thing).":              {"description has no space before '('"},
		"Only \"TemporaryDisk\" is supported.": {`description has "TemporaryDisk" in quotes - a value goes in backticks`},
		"Of the form 'https://{account}.file.core.windows.net/'.": {"description has 'https://{account}.file.core.windows.net/' in quotes - a value goes in backticks"},
		"Either \"OsDisk\" or 'TemporaryDisk' (or \"None\").":     {`description has "OsDisk" in quotes - a value goes in backticks`, "description has 'TemporaryDisk' in quotes - a value goes in backticks", `description has "None" in quotes - a value goes in backticks`},
		"The `name`(of the Thing).":                               {"description has no space before '('"},
		"The size(1 to 5)(in GB).":                                {"description has no space before '('"},
		"The name,, then more.":                                   {`description has doubled punctuation ",,"`},
		"Is it on?.":                                              {`description has doubled punctuation "?."`},
		"The name, then more,.":                                   {`description has doubled punctuation ",."`},
		"The name....":                                            {`description has doubled punctuation "...."`},
		`The "name.`:                                              {"description has a quote with no partner"},
		"The name).":                                              {"description has a ')' that nothing opened"},
		"The [name.":                                              {"description has a '[' that is never closed"},
		"The name (a] here.":                                      {"description has a '(' that is never closed", "description has a ']' that nothing opened"},
		"The name, ie the label.":                                 {`"ie" should be "i.e."`},
		"E.g `foo`.":                                              {`"E.g" should be "E.g."`},
		"The name\u00a0here.":                                     {`description has a non-breaking space '\u00a0' - use " "`},
		"The name\u200b here.":                                    {"description has an invisible character (U+200B) - remove it"},
		"The name—here.":                                          {`description has an em dash '—' - use "-"`},
	} {
		var got []string
		for _, finding := range check(text) {
			got = append(got, finding.Message)
		}

		if !slices.Equal(got, want) {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}

	// each kind of mistake can be switched off on its own; the kinds are package state, so this cannot run beside anything else
	if got := check(`Only "TemporaryDisk" is supported.`); len(got) != 1 || got[0].New != "`TemporaryDisk`" {
		t.Errorf("a value in quotes: got %v", got)
	}

	for kind, text := range map[string]string{
		"capital":             "the name.",
		"full-stop":           "The name",
		"question-mark":       "Is the Thing on.",
		"sentence-capital":    "The name. changing it is fine.",
		"leading-mark":        "- The name.",
		"line-break":          "The name\tof it.",
		"code-space":          "One of`a`.",
		"edge-space":          " The name. ",
		"double-space":        "The  name.",
		"space-before":        "The name , then more.",
		"space-after":         "The name,then more.",
		"bracket-space":       "The ( name ) here.",
		"doubled-punctuation": "The name..",
		"unbalanced":          "The `name.",
		"abbreviation":        "The name, eg the label.",
		"pasted-characters":   "The \u201cname\u201d.",
		"quoted-value":        "Only \"Fast\" is supported.",
	} {
		if len(check(text)) == 0 {
			t.Errorf("%s: %q should be reported", kind, text)
		}

		*enabled[kind] = false
		if got := check(text); len(got) > 0 {
			t.Errorf("%s switched off: %q still reports %q", kind, text, got[0].Message)
		}
		*enabled[kind] = true
	}

	if len(enabled) != 17 {
		t.Errorf("%d kinds can be switched off, and the test covers 17", len(enabled))
	}
}
