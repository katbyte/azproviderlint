# AZG007 - omit struct literal fields explicitly set to their zero value

AZG007 reports struct literal fields set to the value they would have anyway: a pointer set to `nil`, a string to `""`, a number to `0`, a bool to `false`. Leave the field out.

An omitted field already gets its zero value, so the line changes nothing and gives the reader one more thing to check.

## Flagged Code

```go
return &profiles.ProfileLogScrubbing{
	State:    &policyDisabled,
	Selector: nil, // Selector is *string
}

return KubeConfigModel{
	Host:              cluster.Server,
	Username:          name,
	Password:          "",
	ClientCertificate: "",
	ClientKey:         "",
}
```

## Passing Code

```go
return &profiles.ProfileLogScrubbing{
	State: &policyDisabled,
}

// slices, maps, and interfaces are left alone
return &Config{
	Items: nil,
	Data:  nil,
}

// a named constant says something, even when it is zero
return &Settings{
	Mode: ModeNone,
}
```

## What counts

Any compile-time zero however it is spelled: `0`, `-0.0`, `'\x00'`, `int64(0)`, `""`, `false`.

Not reported:

- slices, maps, and interfaces, where an explicit `nil` can be a deliberate signal
- named constants (`Mode: ModeNone`, `N: int64(ZeroCount)`), where the name carries meaning
- `_test.go` files, unless `tests` is on, since a zero entry in a test table is often a meaningful row

## The fix

`-fix` removes the field, its comma, and any trailing comment, and leaves gofmt to tidy the whitespace. A field with its own comment on the line above is reported without a fix, because deleting the field would leave that comment attached to the next field. A person decides whether to drop it, keep it, or turn the comment into an `//azignore:AZG007 - <reason>`.

## Options

| Option | Default | Effect |
|---|---|---|
| `tests` | false | also check `_test.go` files |

Set with `-AZG007.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Put `//azignore:AZG007 - <reason>` at the end of the field's line, or on the line above it. The reason is required.

```go
Selector: nil, //azignore:AZG007 - <reason>
```

On the opening line of a struct literal it covers every field in that literal, nested literals included:

```go
return UserFeatures{ //azignore:AZG007 - all nested objects must be fully populated
	// ...
}
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
