# AZG011 - nil dereferences that need a hand-written guard

AZG011 reports a dereference of a pointer that may be nil, where `pointer.From` cannot help and a person has to add a nil check.

It is the other half of [AZG008](../AZG008_unchecked_nil_dereference). AZG008 reports `*x` where the value is only read, and can fix it. AZG011 reports the rest:

- **No `*` is written.** `m.Properties.Name` reads through `m.Properties` and panics when it is nil. So does calling a method with a value receiver through a pointer, `props.Method()`. This is the most common nil panic in provider code, and AZG008 never sees it.
- **The pointer itself is needed.** `*x = v`, `(*x).F = v`, `(*x)[k] = v`, `&*x`, `(*x)++`, slicing an array with `(*x)[:]`, or a pointer-receiver method on `*x`.

## Flagged Code

```go
d.Set("name", model.Properties.Name) // panics when Properties is nil
```

```go
if model.Properties.Sku != nil { // the condition itself reads through Properties
	...
}
```

```go
*props.Count = 1
```

## Passing Code

```go
if model.Properties != nil {
	d.Set("name", model.Properties.Name)
}
```

```go
if props.Count == nil {
	return
}
*props.Count = 1
```

## What counts

The same checks as AZG008; see [what counts as a check](../AZG008_unchecked_nil_dereference/README.md#what-counts-as-a-check) there.

A check covers exactly what it names. `if m.Properties != nil` covers reading `m.Properties.Sku`, but not `m.Properties.Sku.Name`, which is reported on its own.

Not reported:

- bare pointer parameters (`func f(p *T) { p.Name }`), closures included, since the nil check belongs at the call site. Fields reached through a parameter, like `p.Sku.Name`, are always checked
- a pointer-receiver method called on a pointer, which dereferences nothing
- chains that contain a call or an index, like `f().Name` or `items[0].Name`
- fields promoted from an embedded pointer, since the pointer being read is not written in the code
- pointers to a type that matches `exclude-types`

## The fix

There is none. Every report needs a nil check written by hand.

## Options

| Option | Default | Effect |
|---|---|---|
| `include-parameters` | false | also report dereferences of bare pointer parameters |
| `tests` | true | check `_test.go` files |
| `exclude-types` | | comma-separated patterns; pointers to a matching type are not reported |

An `exclude-types` pattern is matched against the type's full name, such as `github.com/x/y/clients.Client`. `*` matches anything, `/` included, so `*Client` and `*/clients.Client` both match it.

SDK clients are the main use. Their methods have value receivers, so `client.Get(ctx, id)` dereferences `client` on every call, and a client is set up at startup rather than read from a response. On azurerm, `*Client,*Authorizers,*ResourceManagerAccount` removes about three quarters of the reports.

Set with `-AZG011.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

A dereference can be deliberate, when something else guarantees the field or a panic is the right way to fail. Put `//azignore:AZG011 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
d.Set("name", model.Properties.Name) //azignore:AZG011 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
