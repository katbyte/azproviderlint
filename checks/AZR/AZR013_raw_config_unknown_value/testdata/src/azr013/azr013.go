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

	rule := rules.AsValueSlice()[0]            // want `rules.AsValueSlice\(\) panics while planning`
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
