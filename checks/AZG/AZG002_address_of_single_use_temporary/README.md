# AZG002 - use new() instead of a single-use temporary's address

The AZG002 analyzer reports single-use temporaries whose only use is taking their address: `v := <expr>` followed by `&v` in a later statement, where `v` has no other use in the function. The temporary exists only because `&` needs an addressable operand — Go 1.26's `new(<expr>)` says the same thing in one step at the use site. With `use: pointer.To` the rule suggests the go-azure-helpers `pointer.To(<expr>)` helper instead. `new(<expr>)` does not compile below Go 1.26, so on older packages new mode errors rather than guessing — set `use: pointer.To` explicitly there.

In new mode, existing `pointer.To(x)` calls are also reported — they should be `new(x)` — unless `allow: pointer.To` is set; when a rewrite leaves a file with no other reference to the pointer package, the fix deletes the then-unused import.

The `&v` may sit anywhere in a later statement of the same block — a call argument, a composite literal field, an assignment, a return — at most `max-gap` source lines below the declaration (default 100). The initializer must be single-line and free of calls and channel receives (type conversions are fine), so moving its evaluation to the address-of site cannot reorder observable effects. Nothing is reported when a statement between the declaration and the `&v` writes to, takes the address of, or shadows anything the initializer reads, or when the `&v` is captured by a nested function literal (the closure would defer the initializer's evaluation).

The report carries a suggested fix, so `azproviderlint -AZG002 -fix` inlines the temporary automatically; in pointer.To mode the fix adds the pointer import when the file lacks it, and a conversion initializer fixes to `pointer.To(sdk.Enum(v))` — exactly the shape [AZG003](../AZG003_pointer_to_enum_conversion) then rewrites to `pointer.ToEnum`, so running both fixes converges.

When the temporary copies a dereferenced pointer (`out := *p; return &out`), the message names both options and `fix-pointer-copy` picks the fix: `none` (default) offers no fix so a person chooses, `copy` keeps the copy-preserving `pointer.To(*p)`/`new(*p)`, and `share` substitutes `p` itself — an aliasing change the config opts into globally.

## Flagged Code

```go
name := "hello"
payload.Name = &name
```

```go
enabled := true
props := Properties{
	Enabled: &enabled,
}
```

```go
payload.Count = pointer.To(3) // new mode, without allow: pointer.To
```

## Passing Code

```go
payload.Name = new("hello")
```

```go
name := flattenName(input)
payload.Name = &name // initializer calls a function
```

```go
name := "hello"
name = rename(name)
payload.Name = &name // the variable has other uses
```

## Options

| Option | Default | Effect |
|---|---|---|
| `use` | `new` | suggested creation form: `new` or `pointer.To`; new mode errors on packages below go1.26 |
| `allow` | | comma-separated forms to leave unreported where they already appear: `pointer.To` |
| `max-gap` | 100 | maximum source lines between the declaration and the statement taking its address |
| `fix-pointer-copy` | `none` | fix for a dereference-copy temporary (`out := *p; &out`): `none` reports without a fix, `copy` keeps the copy (`pointer.To(*p)`/`new(*p)`), `share` substitutes `p` — an aliasing change |

Set via `-AZG002.<option>` on the CLI or a rule-name key in the plugin's golangci settings.

## Ignoring Reports

A named temporary can be deliberate documentation, so suppressions are expected where the name genuinely helps. When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZG002 - <reason>` comment instead, in the same positions:

```go
name := "hello" //azignore:AZG002 - <reason>
```
