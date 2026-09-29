# AZG002 - use new() instead of a single-use temporary's address

AZG002 reports a variable that exists only so something can take its address: `v := <expr>` followed by `&v`, with no other use of `v` in the function.

Go needs something addressable to take a pointer to, and the temporary is the old way to get one. Since Go 1.26, `new(<expr>)` does the same thing in one step at the point of use. With `use: pointer.To` the rule suggests go-azure-helpers' `pointer.To(<expr>)` instead, which is what to set for packages below Go 1.26. In new mode the rule refuses to guess on those packages and reports an error.

In new mode existing `pointer.To(x)` calls are reported too, since they should now be `new(x)`. Set `allow: pointer.To` to keep them. When a rewrite leaves nothing else using the pointer package, the fix removes the import.

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
payload.Name = &name // the initializer calls a function
```

```go
name := "hello"
name = rename(name)
payload.Name = &name // the variable has other uses
```

## What counts

The `&v` can be anywhere in a later statement of the same block: a call argument, a struct field, an assignment, a return. It must be within `max-gap` lines of the declaration.

Not reported, because moving the initializer could change what the code does:

- initializers that span lines, call a function, or receive from a channel (type conversions are fine)
- something between the declaration and the `&v` writes to, takes the address of, or shadows anything the initializer reads
- the `&v` is inside a closure, which would delay when the initializer runs

## The fix

`-fix` inlines the temporary. In pointer.To mode it adds the import if the file lacks it. A conversion like `v := sdk.Enum(x)` becomes `pointer.To(sdk.Enum(x))`, which is the exact shape [AZG003](../AZG003_pointer_to_enum_conversion) then turns into `pointer.ToEnum`, so running both fixes settles.

When the temporary copies a dereferenced pointer (`out := *p; return &out`), there are two reasonable rewrites and they mean different things. `pointer.To(*p)` keeps the copy. Using `p` directly shares the pointer. The report names both. The `fix-pointer-copy` option says which one `-fix` applies, and the default `none` leaves the choice to a person.

## Options

| Option | Default | Effect |
|---|---|---|
| `use` | `new` | what to suggest: `new` or `pointer.To`. new mode errors on packages below go1.26 |
| `allow` | | forms to leave alone where they already appear: `pointer.To` |
| `max-gap` | 100 | most lines allowed between the declaration and the `&v` |
| `fix-pointer-copy` | `none` | for `out := *p; &out`: `none` reports without a fix, `copy` uses `pointer.To(*p)` or `new(*p)`, `share` uses `p` |

Set with `-AZG002.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Sometimes the name is the point, because it documents an otherwise opaque value. Put `//azignore:AZG002 - <reason>` at the end of the declaration line, or on the line above it. The reason is required.

```go
name := "hello" //azignore:AZG002 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
