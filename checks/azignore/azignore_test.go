package azignore

import (
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	AZG001 "github.com/katbyte/azproviderlint/checks/AZG/AZG001_combine_err_assignment_and_check"
)

func TestLintIgnore(t *testing.T) {
	t.Parallel()

	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "testdata")

	analyzers := Wrap([]*analysis.Analyzer{AZG001.Analyzer})

	analysistest.Run(t, dir, analyzers[0], "azignore")
}

func TestParseDirective(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		rules  []string
		reason string
		ok     bool
	}{
		{"not a directive", "// plain comment", nil, "", false},
		{"prose mentioning the syntax", "// suppress with //azignore:AZG001", nil, "", false},
		{"bare single rule", "//azignore:AZG001", []string{"AZG001"}, "", true},
		{"bare rule list with spaces", "//azignore:AZG001, AZR001", []string{"AZG001", "AZR001"}, "", true},
		{"hyphen reason", "//azignore:AZG001 - deliberate subset", []string{"AZG001"}, "deliberate subset", true},
		{"reason without a dash", "//azignore:AZG001 the combined form obscures the retry loop", []string{"AZG001"}, "the combined form obscures the retry loop", true},
		{"list then reason without a dash", "//azignore:AZG001, AZR001 deliberate", []string{"AZG001", "AZR001"}, "deliberate", true},
		{"bare with trailing comma", "//azignore:AZG001,", []string{"AZG001"}, "", true},
		{"en dash reason", "//azignore:AZG001 – deliberate", []string{"AZG001"}, "deliberate", true},
		{"em dash reason", "//azignore:AZP003 — deliberately not exposed", []string{"AZP003"}, "deliberately not exposed", true},
		{"rule list with reason", "//azignore:AZG001,AZR001 - why not", []string{"AZG001", "AZR001"}, "why not", true},
		{"empty reason after dash", "//azignore:AZG001 - ", []string{"AZG001"}, "", true},
		{"reason containing dashes", "//azignore:AZG001 - see ADR-042 - long story", []string{"AZG001"}, "see ADR-042 - long story", true},
		{"space after the comment marker", "// azignore:AZG001 - ok", []string{"AZG001"}, "ok", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rules, reason, ok := ParseDirective(tc.text)
			if ok != tc.ok || reason != tc.reason || !slices.Equal(rules, tc.rules) {
				t.Fatalf("ParseDirective(%q) = %v, %q, %v; want %v, %q, %v",
					tc.text, rules, reason, ok, tc.rules, tc.reason, tc.ok)
			}
		})
	}
}
