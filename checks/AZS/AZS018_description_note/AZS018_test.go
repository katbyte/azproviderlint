package AZS018

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS018(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs018")

	for _, text := range []string{
		"The name of the Thing.",
		"A note about this user.",
		"The key. Please note these values are only shown once.",
		"The key, note: shown once.",
		"The footnote: a line of text.",
		"The message to show, such as `Note: closed`.",
		"The name. **Important:** it cannot be changed.",
		"The name. Notes: none.",
		"",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// what is left once the note is taken out; nothing when there is no fix
	for text, want := range map[string]string{
		"The name. **Note:** it cannot be changed.":      "The name.",
		"The name. **NOTE:** it cannot be changed.":      "The name.",
		"The name. **Note**: it cannot be changed.":      "The name.",
		"The name. **Note** it cannot be changed.":       "The name.",
		"The name. __Note:__ it cannot be changed.":      "The name.",
		"The name. Note: it cannot be changed.":          "The name.",
		"The name. NOTE: it cannot be changed. Ever.":    "The name.",
		"Is it on? Note: it cannot be changed.":          "Is it on?",
		"The name (if any). Note: it cannot be changed.": "The name (if any).",
		"The name of the Thing **Note:** it is fixed.":   "The name of the Thing",
		"The name. ~> **Note:** it cannot be changed.":   "The name.",
		"The name. -> **Note:** it cannot be changed.":   "The name.",
		"The name.  **Note:** it cannot be changed.  ":   "The name.",
		"The name.\n\n**Note:** it cannot be changed.":   "The name.",
		"The name. **Note:** one. **Note:** two.":        "The name.",
		"**Note:** This value is ignored.":               "",
		"Note: this value is ignored.":                   "",
		" ~> **Note:** this value is ignored.":           "",
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

	// the options add a label and drop one; they are package state, so this cannot run beside another test in this package
	for option, label := range map[string]string{"label": "Important", "ignore": "Note"} {
		if err := Analyzer.Flags.Set(option, label); err != nil {
			t.Fatal(err)
		}
	}

	if got := check("The name. **Important:** it cannot be changed."); len(got) != 1 || got[0].Fix == "" {
		t.Errorf("with Important added: got %v", got)
	}

	if got := check("The name. **Note:** it cannot be changed."); len(got) != 0 {
		t.Errorf("with Note dropped: got %v", got)
	}

	if err := Analyzer.Flags.Set("ignore", "Important"); err != nil {
		t.Fatal(err)
	}

	if got := check("The name. **Important:** it cannot be changed."); len(got) != 0 {
		t.Errorf("with no labels left: got %v", got)
	}

	if err := Analyzer.Flags.Set("label", "Note"); err != nil {
		t.Fatal(err)
	}
}
