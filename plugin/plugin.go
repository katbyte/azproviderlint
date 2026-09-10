// Package plugin registers azproviderlint's analyzers as a golangci-lint module plugin.
package plugin

import (
	"fmt"
	"slices"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/katbyte/azproviderlint/checks"
)

func init() {
	register.Plugin("azproviderlint", New)
}

// the two reserved settings keys; every other top-level key is a rule name
const (
	keyEnable  = "enable"
	keyDisable = "disable"
)

// Settings allows rules to be enabled/disabled per-rule from .golangci.yml via
// linters.settings.custom.azproviderlint.settings. An empty enable list means all rules. A
// list entry names either a rule (AZS006) or a whole category (AZG — a rule name with the
// digits stripped, matching every rule in it). Any other top-level key must be a rule name
// and sets that rule's analyzer flags:
//
//	settings:
//	  enable: [AZG, AZS006]
//	  AZS006:
//	    ignore-sensitive: true
type Settings struct {
	Enable  []string `json:"enable"`
	Disable []string `json:"disable"`
	// Flags holds rule-specific analyzer flags, keyed by rule then flag name, collected from
	// the settings' rule-name keys.
	Flags map[string]map[string]string `json:"-"`
}

func New(settings any) (register.LinterPlugin, error) {
	// the typed decode rejects unknown fields, so split the raw map first: the list fields
	// decode into Settings, and every other key is a rule name carrying that rule's flags
	raw, err := register.DecodeSettings[map[string]any](settings)
	if err != nil {
		return nil, err
	}

	lists := map[string]any{}
	for _, key := range []string{keyEnable, keyDisable} {
		if v, ok := raw[key]; ok {
			lists[key] = v
		}
	}
	s, err := register.DecodeSettings[Settings](lists)
	if err != nil {
		return nil, err
	}

	// flag values are stringified so YAML booleans and numbers work unquoted
	s.Flags = map[string]map[string]string{}
	for rule, v := range raw {
		if rule == keyEnable || rule == keyDisable {
			continue
		}
		flags, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("azproviderlint setting %q: expected a map of flag values", rule)
		}
		s.Flags[rule] = map[string]string{}
		for flag, value := range flags {
			s.Flags[rule][flag] = fmt.Sprintf("%v", value)
		}
	}

	return &Plugin{settings: s}, nil
}

type Plugin struct {
	settings Settings
}

func (p *Plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	// Rule names are matched case-insensitively throughout: golangci's settings decoding
	// (viper) lowercases YAML map keys, so the rule-name keys carrying flag values arrive
	// as "azs004" no matter how the config spells them.
	known := make(map[string]bool, len(checks.All))
	categories := map[string]bool{}
	for _, a := range checks.All {
		known[strings.ToLower(a.Name)] = true
		categories[strings.ToLower(category(a.Name))] = true
	}
	// enable/disable entries may name a rule or a category; flag keys must name a rule
	for _, name := range slices.Concat(p.settings.Enable, p.settings.Disable) {
		if l := strings.ToLower(name); !known[l] && !categories[l] {
			return nil, fmt.Errorf("unknown azproviderlint rule or category %q in settings", name)
		}
	}
	flags := make(map[string]map[string]string, len(p.settings.Flags))
	for name, values := range p.settings.Flags {
		if !known[strings.ToLower(name)] {
			return nil, fmt.Errorf("unknown azproviderlint rule %q in settings", name)
		}
		flags[strings.ToLower(name)] = values
	}

	analyzers := make([]*analysis.Analyzer, 0, len(checks.All))
	for _, a := range checks.All {
		listed := func(list []string) bool {
			return slices.ContainsFunc(list, func(name string) bool {
				return strings.EqualFold(name, a.Name) || strings.EqualFold(name, category(a.Name))
			})
		}
		if len(p.settings.Enable) > 0 && !listed(p.settings.Enable) {
			continue
		}
		if listed(p.settings.Disable) {
			continue
		}
		for flag, value := range flags[strings.ToLower(a.Name)] {
			if err := a.Flags.Set(flag, value); err != nil {
				return nil, fmt.Errorf("setting %s flag %q: %w", a.Name, flag, err)
			}
		}
		analyzers = append(analyzers, a)
	}

	return analyzers, nil
}

// category returns the rule family a rule name belongs to: its name with the trailing digits
// stripped (AZS006 -> AZS).
func category(name string) string {
	return strings.TrimRight(name, "0123456789")
}

func (*Plugin) GetLoadMode() string {
	// AZS001 resolves named types and aliases via the type checker
	return register.LoadModeTypesInfo
}
