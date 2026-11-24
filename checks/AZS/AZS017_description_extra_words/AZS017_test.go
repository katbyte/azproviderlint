package AZS017

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS017(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs017")

	for _, text := range []string{
		"The name of the Thing.",
		"The name of the Thing. Specifies the Thing to use.",
		"Whether the Thing is enabled.",
		"Configures the minimum version of TLS.",
		"Specifications of the Thing.",
		"Specifies",
		"The value to use when it is empty or not set.",
		"Whether the Thing is on or not yet ready.",
		"The order to apply the rules in.",
		"The message, such as `please wait`.",
		"",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// what the description reads as once fixed; nothing when there is no fix
	for text, want := range map[string]string{
		"Specifies the name of the Thing.":          "The name of the Thing.",
		"specifies a list of Things.":               "A list of Things.",
		"Specifies an ID.":                          "An ID.",
		"Indicates if the Thing is on.":             "Whether the Thing is on.",
		"Determines what the Thing does.":           "What the Thing does.",
		"Specifies where the version is read from.": "Where the version is read from.",
		"Specify the, well, name.":                  "The, well, name.",
		"Specifies starting IP address.":            "",
		"Denotes this action uses Linux.":           "",
		"Specifies `name` of the Thing.":            "",
		"Specifies (the) name.":                     "",
		"Whether to enable caching or not.":         "Whether to enable caching.",
		"Whether or not caching is enabled.":        "Whether caching is enabled.",
		"If it is allowed to access or not, or so.": "If it is allowed to access, or so.",
		"The key (required or not).":                "The key (required).",
		"The key to use in order to sign in.":       "The key to use to sign in.",
		"In order to sign in, set the key.":         "To sign in, set the key.",
		"The terms to agree to in order to join.":   "",
		"The Thing. Prior to v2 it is ignored.":     "The Thing. Before v2 it is ignored.",
		"Whether the Thing is able to scale.":       "Whether the Thing can scale.",
		"For more, please read the docs.":           "For more, read the docs.",
		"The key. Please note it is shown once.":    "The key. It is shown once.",
		"The key. Note that `id` is needed.":        "The key. `id` is needed.",
		"The key. Please note: it is shown once.":   "",
		"please note that it is shown once.":        "it is shown once.",
		"The key (please keep it safe).":            "The key (keep it safe).",
	} {
		findings := check(text)
		if len(findings) != 1 {
			t.Errorf("%q: got %d findings, want 1", text, len(findings))
			continue
		}

		got := ""
		if findings[0].Fix != "" {
			got = text[:findings[0].Start] + findings[0].New + text[findings[0].End:]
		}

		if got != want {
			t.Errorf("%q: fixed to %q, want %q", text, got, want)
		}
	}

	// the options add a verb and drop one; they are package state, so this cannot run beside another test in this package
	for option, verb := range map[string]string{"verb": "Configures", "ignore": "Specifies"} {
		if err := Analyzer.Flags.Set(option, verb); err != nil {
			t.Fatal(err)
		}
	}

	if got := check("Configures the minimum version of TLS."); len(got) != 1 || got[0].New != "The" {
		t.Errorf("with Configures added: got %v", got)
	}

	if got := check("Specifies the name of the Thing."); len(got) != 0 {
		t.Errorf("with Specifies dropped: got %v", got)
	}

	delete(verbs, "configures")
	verbs["specifies"] = true

	// a description can open with a verb and be wordy as well
	if got := check("Specifies whether or not it is on."); len(got) != 2 || got[0].New != "Whether" || got[1].Fix != "Remove it" {
		t.Errorf("verb and phrase: got %v", got)
	}

	for option, phrase := range map[string]string{"phrase": "At this time=now", "ignore": "or not"} {
		if err := Analyzer.Flags.Set(option, phrase); err != nil {
			t.Fatal(err)
		}
	}

	if got := check("At this time only `v1` works."); len(got) != 1 || got[0].New != "Now" {
		t.Errorf("with a phrase added: got %v", got)
	}

	if got := check("Whether to enable caching or not."); len(got) != 0 {
		t.Errorf("with a phrase dropped: got %v", got)
	}

	delete(phrases, "at this time")
	phrases["or not"] = ""
	rebuild()
}
