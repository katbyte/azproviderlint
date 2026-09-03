# AZG009 - nil dereferences that need a hand-written guard

The AZG009 analyzer is [AZG008](../AZG008_unchecked_nil_dereference)'s sister: the same guard engine applied to the possibly-nil dereferences `pointer.From` cannot rewrite, so every report needs a person to add a nil check. Two forms are reported:

- **implicit dereferences** — `m.Properties.Name` selects through `m.Properties`, dereferencing it, and panics when it is nil; likewise calling a value-receiver method through a pointer (`props.Method()`). Pointer-receiver methods don't dereference and are ignored. This is the most common nil panic in provider code — no `*` appears, so AZG008 never sees it.
- **explicit dereferences whose context needs the pointer itself** — assignment targets (`*x = v`), the operand of `&` (`&*x` panics on nil too), and inc/dec statements (`(*x)++`).

A dereference counts as guarded under the same rules as AZG008: an enclosing `x != nil` condition (including `&&`/`||` short-circuits and else branches), an `if x == nil` early exit, provably non-nil assignments, an `err`/`ok` companion check, or an alias of a guarded chain — all matched by path equality on the exact dereferenced chain and invalidated by reassignment from an unknown source. Each link needs its own guard: `m.Properties != nil` covers the `Sku` selection but not the deeper `m.Properties.Sku.Name`, which reports separately.

Dereferences of bare pointer parameters are trusted by default (the nil check belongs at the call sites); deeper links through a parameter are always in scope. Chains containing calls or index expressions are out of scope, as are dereferences of embedded pointer fields reached by promotion. `_test.go` files are checked by default; set `tests` to false to skip them.

## Flagged Code

```go
d.Set("name", model.Properties.Name) // panics when Properties is nil
```

```go
if model.Properties.Sku != nil { // the condition itself dereferences Properties
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

## Options

| Option | Default | Effect |
|---|---|---|
| `include-parameters` | false | also report dereferences of bare pointer parameters |
| `tests` | true | check `_test.go` files |

Set via `-AZG009.<option>` on the CLI or a rule-name key in the plugin's golangci settings.

## Ignoring Reports

A dereference can be deliberate — an invariant guarantees the field, or a panic is the correct failure mode. When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZG009 - <reason>` comment instead, in the same positions:

```go
d.Set("name", model.Properties.Name) //azignore:AZG009 - <reason>
```
