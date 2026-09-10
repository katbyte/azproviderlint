package plugin

import (
	"slices"
	"strings"
	"testing"

	"github.com/katbyte/azproviderlint/checks"
)

func buildAnalyzers(t *testing.T, settings any) ([]string, error) {
	t.Helper()

	p, err := New(settings)
	if err != nil {
		return nil, err
	}

	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(analyzers))
	for _, a := range analyzers {
		names = append(names, a.Name)
	}
	return names, nil
}

func TestBuildAnalyzersDefault(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != len(checks.All) {
		t.Fatalf("expected all %d analyzers with no settings, got %d", len(checks.All), len(names))
	}
}

func TestBuildAnalyzersDisable(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, map[string]any{"disable": []string{"AZR002"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != len(checks.All)-1 {
		t.Fatalf("expected %d analyzers, got %d", len(checks.All)-1, len(names))
	}
	for _, name := range names {
		if name == "AZR002" {
			t.Fatal("AZR002 should have been disabled")
		}
	}
}

func TestBuildAnalyzersEnable(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, map[string]any{"enable": []string{"AZS001", "AZT001"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "AZS001" || names[1] != "AZT001" {
		t.Fatalf("expected exactly [AZS001 AZT001], got %v", names)
	}
}

func TestBuildAnalyzersEnableCategory(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, map[string]any{"enable": []string{"AZG"}, "disable": []string{"AZG005"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("expected the AZG rules")
	}
	for _, name := range names {
		if !strings.HasPrefix(name, "AZG") {
			t.Fatalf("expected only AZG rules, got %v", names)
		}
		if name == "AZG005" {
			t.Fatal("AZG005 should have been disabled")
		}
	}
	if !slices.Contains(names, "AZG001") {
		t.Fatalf("expected AZG001 in %v", names)
	}
}

func TestBuildAnalyzersDisableCategory(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, map[string]any{"disable": []string{"azt"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if strings.HasPrefix(name, "AZT") {
			t.Fatalf("expected no AZT rules, got %v", names)
		}
	}
	if len(names) == len(checks.All) {
		t.Fatal("expected the AZT rules to be removed")
	}
}

func TestBuildAnalyzersUnknownRule(t *testing.T) {
	t.Parallel()

	if _, err := buildAnalyzers(t, map[string]any{"disable": []string{"AZX999"}}); err == nil {
		t.Fatal("expected an error for an unknown rule name")
	}
}

func TestBuildAnalyzersRuleFlags(t *testing.T) {
	t.Parallel()

	names, err := buildAnalyzers(t, map[string]any{
		"enable": []string{"AZP003"},
		"AZP003": map[string]any{"ignore-sensitive": true}, // unquoted YAML bool
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "AZP003" {
		t.Fatalf("expected exactly [AZP003], got %v", names)
	}

	p, err := New(map[string]any{"AZP003": map[string]any{"ignore-sensitive": "true"}})
	if err != nil {
		t.Fatal(err)
	}
	pl, ok := p.(*Plugin)
	if !ok {
		t.Fatalf("expected *Plugin, got %T", p)
	}
	analyzers, err := pl.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range analyzers {
		if a.Name != "AZP003" {
			continue
		}
		if f := a.Flags.Lookup("ignore-sensitive"); f == nil || f.Value.String() != "true" {
			t.Fatalf("expected AZP003 ignore-sensitive flag to be true, got %v", f)
		}
	}
}

func TestBuildAnalyzersRuleFlagErrors(t *testing.T) {
	t.Parallel()

	if _, err := buildAnalyzers(t, map[string]any{"AZX999": map[string]any{"some-flag": "true"}}); err == nil {
		t.Fatal("expected an error for flags on an unknown rule name")
	}
	if _, err := New(map[string]any{"AZP003": "not-a-map"}); err == nil {
		t.Fatal("expected an error for a non-map rule settings value")
	}
	if _, err := buildAnalyzers(t, map[string]any{"AZP003": map[string]any{"no-such-flag": "true"}}); err == nil {
		t.Fatal("expected an error for an unknown flag name")
	}
}

func TestBuildAnalyzersLowercasedSettings(t *testing.T) {
	t.Parallel()

	// golangci's settings decoding (viper) lowercases YAML map keys, so rule-name keys
	// carrying flags arrive as "azs004" — they must still resolve to AZS004.
	p, err := New(map[string]any{"azs004": map[string]any{"allow-missing-values": true}})
	if err != nil {
		t.Fatal(err)
	}
	pl, ok := p.(*Plugin)
	if !ok {
		t.Fatalf("expected *Plugin, got %T", p)
	}
	analyzers, err := pl.BuildAnalyzers()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, a := range analyzers {
		if a.Name != "AZS004" {
			continue
		}
		found = true
		if f := a.Flags.Lookup("allow-missing-values"); f == nil || f.Value.String() != "true" {
			t.Fatalf("expected AZS004 allow-missing-values flag to be true, got %v", f)
		}
	}
	if !found {
		t.Fatal("AZS004 missing from built analyzers")
	}

	// enable/disable list values keep their case, but tolerate lowercase for consistency
	names, err := buildAnalyzers(t, map[string]any{"disable": []string{"azr002"}})
	if err != nil {
		t.Fatal(err)
	}
	if slices.Contains(names, "AZR002") {
		t.Fatal("expected lowercase disable entry to disable AZR002")
	}
}
