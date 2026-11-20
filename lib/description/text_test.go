package description

import (
	"strings"
	"testing"
)

func TestMask(t *testing.T) {
	t.Parallel()

	blank := string(Blank)
	for text, want := range map[string]string{
		"The name.":                           "The name.",
		"One of `a` or `b,c`.":                "One of ### or #####.",
		"An odd ` backtick.":                  "An odd ` backtick.",
		"See [the docs](https://x.io/a(b)c).": "See [the docs]####################.",
		"See https://x.io/a,b.":               "See ################.",
		"See <https://x.io/a>.":               "See ################.",
		`In ["10.0.0.0/24"] form.`:            `In ["###########"] form.`,
		"Of type Microsoft.Compute/disks.":    "Of type #######################.",
	} {
		if got := strings.ReplaceAll(Mask(text), blank, "#"); got != want {
			t.Errorf("%q: got %q, want %q", text, got, want)
		}
	}
}

func TestWords(t *testing.T) {
	t.Parallel()

	words := Words(" The (name),  then\tmore. ")
	if len(words) != 4 {
		t.Fatalf("got %d words, want 4", len(words))
	}

	if w := words[1].Trim(",)"); w.Text != "name" || w.Start != 6 || w.End != 10 {
		t.Errorf("got %q at %d to %d, want %q at 6 to 10", w.Text, w.Start, w.End, "name")
	}
}
