package azr013

import (
	"azr013/cty"
	"azr013/schema"
)

type ResourceDiff = schema.ResourceDiff

func customizeDiff(diff *schema.ResourceDiff) error {
	rules := diff.GetRawConfig().GetAttr("rule")
	if rules.IsNull() || rules.LengthInt() == 0 { // want `rules.LengthInt\(\) panics while planning when the value is not known yet - check rules.IsKnown\(\) first`
		return nil
	}

	// reported once, where it is first read
	rule := rules.AsValueSlice()[0]
	if rule.GetAttr("name").AsString() == "" { // want `rule.GetAttr\("name"\).AsString\(\) panics while planning when the value is not known yet - check rule.GetAttr\("name"\).IsKnown\(\) first`
		return nil
	}

	return nil
}

func customizeDiffChecked(diff *ResourceDiff) error {
	rules := diff.GetRawConfig().GetAttr("rule")
	if !rules.IsKnown() || rules.IsNull() || rules.LengthInt() == 0 {
		return nil
	}

	// each value is asked for itself: a list that is known can hold one that is not
	rule := rules.AsValueSlice()[0]
	name := rule.GetAttr("name")
	if name.AsString() == "" { // want `name.AsString\(\) panics while planning`
		return nil
	}

	enabled := rule.GetAttr("enabled")
	if enabled.IsKnown() && enabled.True() {
		return nil
	}

	return nil
}

// asking after the read is too late
func customizeDiffLate(diff *schema.ResourceDiff) error {
	zones := diff.GetRawConfig().GetAttr("zones")
	if zones.LengthInt() > 0 && zones.IsKnown() { // want `zones.LengthInt\(\) panics while planning`
		return nil
	}

	return nil
}

// a config that is wholly known has nothing in it that is not
func customizeDiffWholly(diff *schema.ResourceDiff) error {
	config := diff.GetRawConfig()
	if !config.IsWhollyKnown() {
		return nil
	}

	_ = config.GetAttr("rule").LengthInt()

	return nil
}

// when the change is made every value is known
func create(d *schema.ResourceData) error {
	rules := d.GetRawConfig().GetAttr("rule")
	if rules.IsNull() || rules.LengthInt() == 0 {
		return nil
	}

	return nil
}

// a value that is not the config of a plan is not this rule's
func other(v cty.Value) int { return v.LengthInt() }

// the config as a whole is always known, and so is a block in a list that is
func customizeDiffWhole(diff *schema.ResourceDiff) error {
	rules := diff.GetRawConfig().AsValueMap()["rule"]
	if !rules.IsKnown() || rules.IsNull() || rules.LengthInt() == 0 {
		return nil
	}

	raw := diff.GetRawConfig()
	_ = raw.AsValueMap()["zones"]

	first := rules.AsValueSlice()[0]
	_ = first.AsValueMap()["name"]
	_ = rules.AsValueSlice()[0].AsValueMap()["name"]

	for _, rule := range rules.AsValueSlice() {
		_ = rule.AsValueMap()["name"]
	}

	for _, rule := range rules.AsValueSet().Values() {
		_ = rule.AsValueMap()["name"]
	}

	// what a block holds can still be not known yet
	name := first.AsValueMap()["name"]
	_ = name.AsString() // want `name.AsString\(\) panics while planning`

	// and a plain value out of a list can be too
	for _, zone := range raw.AsValueMap()["zones"].AsValueSlice() { // want `raw.AsValueMap\(\)\["zones"\].AsValueSlice\(\) panics while planning`
		_ = zone.AsString() // want `zone.AsString\(\) panics while planning`
	}

	return nil
}

// what decides whether a change is shown runs while planning too
var suppressed = map[string]*schema.Schema{
	"address_prefixes": {
		DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
			pool := d.GetRawConfig().AsValueMap()["ip_address_pool"]
			if !pool.IsNull() && len(pool.AsValueSlice()) > 0 { // want `pool.AsValueSlice\(\) panics while planning`
				return true
			}

			return false
		},
	},

	"zones": {
		DiffSuppressFunc: suppressZones,
	},

	"name": {
		DiffSuppressFunc: func(_, old, new string, d *schema.ResourceData) bool {
			pool := d.GetRawConfig().AsValueMap()["ip_address_pool"]

			return pool.IsKnown() && !pool.IsNull() && pool.LengthInt() > 0
		},
	},
}

func suppressZones(_, old, new string, d *schema.ResourceData) bool {
	return d.GetRawConfig().AsValueMap()["zones"].LengthInt() > 0 // want `LengthInt\(\) panics while planning`
}
