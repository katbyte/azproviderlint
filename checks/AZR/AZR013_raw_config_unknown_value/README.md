# AZR013 - raw config read while planning must be checked with IsKnown first

AZR013 reports a raw config value that is read while planning with no check that it is known yet.

A value that comes from a resource not created yet is not known during a plan. Asking it for its length, its elements or its text panics, and the user sees a crash where a plan should be. `IsNull()` does not help: a value that is not known is not null either.

## Flagged Code

```go
func(ctx context.Context, diff *pluginsdk.ResourceDiff, meta any) error {
	rules := diff.GetRawConfig().GetAttr("rule")
	if rules.IsNull() || rules.LengthInt() == 0 {
		return nil
	}
```

```text
rules.LengthInt() panics while planning when the value is not known yet - check rules.IsKnown() first
```

## Passing Code

```go
func(ctx context.Context, diff *pluginsdk.ResourceDiff, meta any) error {
	rules := diff.GetRawConfig().GetAttr("rule")
	if !rules.IsKnown() || rules.IsNull() || rules.LengthInt() == 0 {
		return nil
	}
```

## What is checked

A function that reads `GetRawConfig()` or `GetRawPlan()` from a `ResourceDiff`. In it, each `cty.Value` asked for `LengthInt`, `AsValueSlice`, `AsValueMap`, `AsValueSet`, `ElementIterator`, `AsString`, `AsBigFloat`, `True` or `False` has to be asked `IsKnown()` earlier in the function.

Each value is asked for itself: a list that is known can hold a value that is not. A value read twice with no check is reported once, where it is first read. A function that asks any value `IsWhollyKnown()` is left alone.

These are not checked:

- the config as a whole, `GetRawConfig().AsValueMap()`, which is always known
- a block out of a list or set read as a map, `rules.AsValueSlice()[0].AsValueMap()`: the list is what has to be known
- the raw config of a `ResourceData`, where every value is known by the time the change is made
- a helper that is handed a value and does not read the plan itself

## Fix

None is offered. What to do when the value is not known, most often return and leave the check for the apply, is the author's call.

## Ignoring Reports

Put `//azignore:AZR013 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
if rules.IsNull() || rules.LengthInt() == 0 { //azignore:AZR013 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
