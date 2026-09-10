# AZR008 - flatten functions must return empty slices/maps, not nil

AZR008 reports `flatten*` functions that return `nil` where a slice or map is expected. Return an empty one instead: `[]T{}` or `map[K]V{}`.

Flatten functions turn an API response into what goes into state. There, a nil slice and an empty slice are not the same thing. A nil can show up as a plan diff that never goes away, or as a nil that something downstream trips over.

## Flagged Code

```go
func flattenNetworkACLs(input *NetworkRuleSet) []NetworkACLs {
	if input == nil {
		return nil
	}
	// ...
}

func flattenTags(input *Resource) map[string]interface{} {
	if input == nil {
		return nil
	}
	// ...
}

// naked return: ret is still nil here
func flattenReplicaSets(input *[]ReplicaSet) (ret []interface{}) {
	if input == nil {
		return
	}
	// ...
}

// out is still nil here
func flattenRules(input *RuleSet) []Rule {
	var out []Rule
	if input == nil {
		return out
	}
	// ...
}
```

## Passing Code

```go
func flattenNetworkACLs(input *NetworkRuleSet) []NetworkACLs {
	if input == nil {
		return []NetworkACLs{} // or make([]NetworkACLs, 0)
	}
	// ...
}

// error path: a nil container alongside an error is fine
func flattenSku(input *Sku) ([]interface{}, error) {
	if input.Name == nil {
		return nil, fmt.Errorf("`name` was nil")
	}
	// ...
}

// a nil pointer means "absent", which is a different thing
func flattenOptionalTags(input *Resource) *map[string]string {
	if input == nil {
		return nil
	}
	// ...
}
```

## What counts

Any function whose name starts with `flatten` (any casing) and returns a slice or map, named types included. A result is reported when it is provably nil at the return:

- a literal `nil`, including conversions like `[]T(nil)`
- a naked `return` before the named result was assigned
- a variable that is still nil: declared with `var` or as a named result, never assigned, address never taken

Not reported:

- `return nil, err` when the error can be non-nil. `var noErr error; return nil, noErr` is still reported.
- pointer results (`*T`, `*[]T`, `*map[K]V`), where nil means absent
- `interface{}` results, where the container shape is not declared
- `expand*` functions, where nil is normal

## The fix

`-fix` replaces the nil with an empty literal, turns a naked return into an explicit one (`return []T{}, err`), and deletes the returned variable's declaration when nothing else uses it.

## Ignoring Reports

Put `//azignore:AZR008 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
return nil //azignore:AZR008 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
