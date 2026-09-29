# AZG006 - inline single-use variable only used in a later function call

AZG006 reports a one-line `x := <expr>` where `x` is used exactly once, as an argument to a later call whose other arguments are all literals or plain names. The usual shape is `x := flattenThing(...)` followed by `d.Set("key", x)`. Inline it.

Next to `"key"`, `ctx`, or `id`, the temporary's name is not telling the reader anything those do not already say. `client.ImportThenPoll(ctx, id, expandMsSqlServerImport(d))` reads no worse than the two-line version.

## Flagged Code

```go
apns := flattenNotificationHubsAPNSCredentials(props.ApnsCredential)
if err := d.Set("apns_credential", apns); err != nil {
	return fmt.Errorf("setting `apns_credential`: %+v", err)
}
```

## Passing Code

```go
if err := d.Set("apns_credential", flattenNotificationHubsAPNSCredentials(props.ApnsCredential)); err != nil {
	return fmt.Errorf("setting `apns_credential`: %+v", err)
}
```

```go
// not reported: next to other expressions, the name is documentation
payload := expandThing(d)
client.CreateOrUpdate(ctx, id, payload)
```

## What counts

The call can be a statement on its own or the init of an `if` (`if err := d.Set("key", x); err != nil`), in the same block, within `max-gap` lines.

Not reported:

- calls with any more complex sibling argument, such as a selector chain, another call, or a type assertion. There the name earns its place, and inlining could reorder evaluation.
- multi-line initializers, which would make the call hard to read
- something between the declaration and the call writes to, takes the address of, or shadows anything the initializer reads

## The fix

`-fix` moves the initializer into the call.

## Options

| Option | Default | Effect |
|---|---|---|
| `max-gap` | 100 | most lines allowed between the declaration and the call |
| `only-when-literals` | false | require every sibling argument to be a literal, not just a plain name |
| `maximum-arguments` | 0 | skip calls with more arguments than this (0 means no limit) |

The last two are for tightening the rule further when several temporaries feeding one call would inline into a long line. Set with `-AZG006.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Put `//azignore:AZG006 - <reason>` at the end of the declaration line, or on the line above it. The reason is required.

```go
apns := flattenNotificationHubsAPNSCredentials(props.ApnsCredential) //azignore:AZG006 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
