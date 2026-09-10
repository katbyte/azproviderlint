# AZG005 - inline single-use variable only used in a later assignment or return

AZG005 reports `x := <expr>` where the only thing that ever happens to `x` is `y = x` or `return x` a little further down. Inline it.

A name that is used once and adds no information is just a line to read. `output.Format = pointer.From(input.Format)` says everything `format := pointer.From(input.Format); output.Format = format` does.

## Flagged Code

```go
format := pointer.From(input.Format)
output.Format = format
```

```go
name := buildName(input)
return name
```

## Passing Code

```go
output.Format = pointer.From(input.Format)
```

```go
return buildName(input)
```

```go
oldKey := column[y]
column[y] = minimumOf3(column[y]+1, column[y-1]+1, lastKey+incr)
lastKey = oldKey // column[y] changed in between, so the temporary is doing real work
```

## What counts

The consumer must be a plain single assignment or a single-value `return`, in the same block, within `max-gap` lines. Call arguments are left to [AZG006](../AZG006_single_use_call_argument), since naming an argument is often deliberate.

Not reported, because inlining could change what the code does:

- something between the declaration and the consumer writes to, takes the address of, or shadows anything the initializer reads, as in the `oldKey` example above
- the assignment's left side contains a call, which would run before the initializer once inlined
- multi-value declarations, `_` discards, and variables captured by closures

## The fix

`-fix` splices the initializer's source text into the consumer, so multi-line initializers keep their formatting.

## Options

| Option | Default | Effect |
|---|---|---|
| `max-gap` | 100 | most lines allowed between the declaration and its consumer |

Set with `-AZG005.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Sometimes the name is the point, because it explains an otherwise opaque expression. Put `//azignore:AZG005 - <reason>` at the end of the declaration line, or on the line above it. The reason is required.

```go
format := pointer.From(input.Format) //azignore:AZG005 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
