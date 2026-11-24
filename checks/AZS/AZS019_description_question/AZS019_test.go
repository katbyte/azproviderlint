package AZS019

import (
	"path/filepath"
	"runtime"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAZS019(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	analysistest.RunWithSuggestedFixes(t, filepath.Join(filepath.Dir(filename), "testdata"), Analyzer, "azs019")

	for _, text := range []string{
		"Whether the Thing is enabled.",
		"The name of the Thing.",
		"Must be between `1` and `5`.",
		"Is required when `mode` is set.",
		"Can only be set when `mode` is `Fast`.",
		"Has no effect on existing Things.",
		"Do not set this with `mode`.",
		"The name. Should be unique.",
		"The query to run, such as `enabled?`.",
		"See https://example.com/things?v=1 for more.",
		"",
	} {
		if got := check(text); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// what the description becomes; nothing when there is no fix
	for text, want := range map[string]string{
		"Should the Thing be enabled?":                    "Whether the Thing should be enabled.",
		"Should the Thing be enabled.":                    "Whether the Thing should be enabled.",
		"Should the Thing be enabled":                     "Whether the Thing should be enabled",
		"Should the Thing be enabled?.":                   "Whether the Thing should be enabled.",
		"Should the Thing be enabled? Defaults to `true`": "Whether the Thing should be enabled. Defaults to `true`",
		"(Boolean) Should the Thing be enabled?":          "(Boolean) Whether the Thing should be enabled.",
		"Should the `AES` cipher be enabled for all?":     "Whether the `AES` cipher should be enabled for all.",
		"Should the Thing should be enabled?":             "",
		"Should the Thing use HTTPS?":                     "",
		"Is the Thing enabled?":                           "",
		"Is the Thing enabled.":                           "",
		"Are Web Sockets enabled":                         "",
		"Does the Thing use HTTPS.":                       "",
		"Must the assignment have an expiry date.":        "",
		"(Boolean) Is approval required.":                 "",
		"Enable the gateway? Only in the main region.":    "",
		"The name of the Thing. Is it unique?":            "",
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

	if got := check("Is it on? Or is it off?"); len(got) != 2 {
		t.Errorf("two questions: got %d findings", len(got))
	}
}
