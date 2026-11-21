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
		"Specifies whether or not it is on.":        "Whether or not it is on.",
		"Indicates if the Thing is on.":             "Whether the Thing is on.",
		"Determines what the Thing does.":           "What the Thing does.",
		"Specifies where the version is read from.": "Where the version is read from.",
		"Specify the, well, name.":                  "The, well, name.",
		"Specifies starting IP address.":            "",
		"Denotes this action uses Linux.":           "",
		"Specifies `name` of the Thing.":            "",
		"Specifies (the) name.":                     "",
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
}
