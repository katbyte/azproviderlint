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
		if got := check(text, false); len(got) > 0 {
			t.Errorf("%q: unexpected %q", text, got[0].Message)
		}
	}

	// what the description becomes; nothing when there is no fix
	for text, want := range map[string]string{
		"Should the Thing be enabled?":                                      "Whether the Thing should be enabled.",
		"Should the Thing be enabled.":                                      "Whether the Thing should be enabled.",
		"Should the Thing be enabled":                                       "Whether the Thing should be enabled",
		"Should the Thing be enabled?.":                                     "Whether the Thing should be enabled.",
		"Should the Thing be enabled? Defaults to `true`":                   "Whether the Thing should be enabled. Defaults to `true`",
		"(Boolean) Should the Thing be enabled?":                            "(Boolean) Whether the Thing should be enabled.",
		"Should the `AES` cipher be enabled for all?":                       "Whether the `AES` cipher should be enabled for all.",
		"Should the Thing should be enabled?":                               "",
		"Should the Thing use HTTPS?":                                       "Whether the Thing should use HTTPS.",
		"Should the Thing prevent a Widget to be deleted?":                  "",
		"Should the Thing expose what could be reached?":                    "",
		"Should the Thing that is on be enabled?":                           "",
		"Should access to the Thing be enabled?":                            "Whether access to the Thing should be enabled.",
		"Should the Thing automatically be replaced?":                       "Whether the Thing should automatically be replaced.",
		"Can the Thing only be reached over HTTPS?":                         "Whether the Thing can only be reached over HTTPS.",
		"Should public network access be enabled?":                          "Whether public network access should be enabled.",
		"Should the use of HTTPS be required?":                              "Whether the use of HTTPS should be required.",
		"Will the Thing keep one backup?":                                   "Whether the Thing will keep one backup.",
		"Whether the Thing is enabled?":                                     "Whether the Thing is enabled.",
		"Is the required field enabled for this Thing?":                     "Whether the required field is enabled for this Thing.",
		"Is the system assigned identity enabled?":                          "Whether the system assigned identity is enabled.",
		"Is Managed Identity used for authentication?":                      "Whether Managed Identity is used for authentication.",
		"Is the Key automatically used?":                                    "Whether the Key is automatically used.",
		"Is there a policy enabled on this Thing?":                          "",
		"Is the Thing is enabled for the App?":                              "",
		"Is the key used here the primary key?":                             "",
		"Is this an external Thing?":                                        "",
		"Will the Thing will be enabled?":                                   "",
		"Should all traffic to have a Gateway applied?":                     "",
		"Should the Things `Force Delete`, which lets them delete at once?": "",
		"Is the Thing enabled?":                                             "Whether the Thing is enabled.",
		"Is the Thing enabled.":                                             "Whether the Thing is enabled.",
		"Are Web Sockets enabled":                                           "Whether Web Sockets are enabled",
		"Does the Thing use HTTPS.":                                         "",
		"Must the assignment have an expiry date.":                          "Whether the assignment must have an expiry date.",
		"(Boolean) Is approval required.":                                   "(Boolean) Whether approval is required.",
		"Enable the gateway? Only in the main region.":                      "",
		"The name of the Thing. Is it unique?":                              "",
	} {
		findings := check(text, false)
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

	// nothing is asked of a read-only property, so it is said to be what it is
	for text, want := range map[string]string{
		"Should the Thing be enabled?":                         "Whether the Thing is enabled.",
		"Should Web Sockets be enabled?":                       "Whether Web Sockets are enabled.",
		"Should the Floating IPs be enabled?":                  "Whether the Floating IPs are enabled.",
		"Should the settings for the App be suppressed?":       "Whether the settings for the App are suppressed.",
		"Should the key and the secret be rotated?":            "Whether the key and the secret are rotated.",
		"Should HTTPS be required?":                            "Whether HTTPS is required.",
		"Should public network access be enabled?":             "Whether public network access is enabled.",
		"Should Kubernetes be upgraded automatically?":         "Whether Kubernetes is upgraded automatically.",
		"Should the Thing automatically be replaced?":          "Whether the Thing is automatically replaced.",
		"Should OS Upgrades automatically be applied?":         "Whether OS Upgrades are automatically applied.",
		"Should Filter by Tags permissions be enabled?":        "Whether Filter by Tags permissions are enabled.",
		"Should the pipelines (run as the App) be kept?":       "Whether the pipelines (run as the App) are kept.",
		"Should the Thing use HTTPS?":                          "Whether the Thing uses HTTPS.",
		"Should the Thing automatically use HTTPS?":            "Whether the Thing automatically uses HTTPS.",
		"Should each node have a Public IP Address?":           "Whether each node has a Public IP Address.",
		"Should nodes in this Pool have a Public IP?":          "Whether nodes in this Pool have a Public IP.",
		"Should this API require a subscription key?":          "Whether this API requires a subscription key.",
		"Should nodes deployed in a Network have a Public IP?": "Whether nodes deployed in a Network have a Public IP.",
		"Should we retain the data?":                           "Whether we retain the data.",
		"Should the managed rules be enabled?":                 "Whether the managed rules are enabled.",
		"Should policy managed rules inspect the body?":        "Whether policy managed rules inspect the body.",
		"Can the Thing be reached?":                            "Whether the Thing can be reached.",
		"Is the Thing enabled?":                                "Whether the Thing is enabled.",
	} {
		got := check(text, true)
		if len(got) != 1 || got[0].New != want {
			t.Errorf("%q on a read-only property: got %v, want %q", text, got, want)
		}
	}

	if got := check("Is it on? Or is it off?", false); len(got) != 2 {
		t.Errorf("two questions: got %d findings", len(got))
	}
}
