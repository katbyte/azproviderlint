# AZR010 - flatten functions handle nil input themselves

The AZR010 analyzer reports a `flatten*` call whose pointer argument is nil-checked by the caller: `if v.Ids != nil { out = flattenIds(v.Ids) }`. The nil handling belongs inside the flatten function, where one early return covers every call site and callers read as plain assignments. A guard at the call site has to be repeated by every caller, and the one that forgets panics on a sparse API response.

A call is reported when it sits in the body of an `if` whose condition proves the argument non-nil (`x != nil`, possibly one conjunct of an `&&` chain, or the else branch of `x == nil`), the argument is a variable or field chain (an `x := v.Field` init makes the two interchangeable), and the branch uses that chain for nothing but flatten arguments. A branch that also dereferences or selects through the chain needs its check and is left alone. Any function whose name starts with `flatten` counts, in any package.

When the flatten function already returns early on a nil parameter — a leading `if input == nil { return ... }` (`||` chains included), or a straight `return other(input)` to a function that does — the report says the check is redundant, and a bare `if x != nil { <one statement> }` gets a fix that unwraps it. The verdict is exported as an analysis fact, so helpers in other packages count too.

## Flagged Code

```go
if props.SuppressionIds != nil {
	suppressionIds = flattenSuppressionIds(props.SuppressionIds)
}

func flattenSuppressionIds(input *[]string) []interface{} {
	out := make([]interface{}, 0, len(*input))
	...
}
```

## Passing Code

```go
suppressionIds = flattenSuppressionIds(props.SuppressionIds)

func flattenSuppressionIds(input *[]string) []interface{} {
	if input == nil {
		return []interface{}{}
	}
	out := make([]interface{}, 0, len(*input))
	...
}
```

## Ignoring Reports

When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZR010 - <reason>` comment instead, in the same positions:

```go
suppressionIds = flattenSuppressionIds(props.SuppressionIds) //azignore:AZR010 - <reason>
```
