# AZS019 - resource and property descriptions must not be a question

AZS019 reports a resource or property description written as a question, such as `Should the Thing be enabled?`.

A description says what the property is. The docs describe a setting; they do not ask the reader about it.

## Flagged Code

```go
"audit_log_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Should the log auditing be enabled?",
},
```

```text
description asks a question - say what it is, such as "Whether the Thing is enabled."
```

## Passing Code

```go
"audit_log_enabled": {
	Type:        pluginsdk.TypeBool,
	Description: "Whether log auditing is enabled.",
},
```

## What counts

A description is a question when:

- its first sentence opens with `Should`, `Is`, `Are`, `Does`, `Do`, `Can`, `Must`, `Has`, `Have`, `Will`, `Was` or `Were`, whatever it ends with, and also behind a tag such as `(Boolean)`
- any of its sentences ends in a question mark

These pass, since the next word leaves the subject out and states a rule: `Must be between 1 and 5.`, `Can only be set when...`, `Is required when...`, `Has no effect...`.

A question mark inside code in backticks or a web address is not counted.

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Fix

The question is turned round where that only takes moving a word:

| Question | Becomes |
|---|---|
| `Should the Thing be enabled?` | `Whether the Thing should be enabled.` |
| `Should the Thing use HTTPS?` | `Whether the Thing should use HTTPS.` |
| `Can the Thing only be reached over HTTPS?` | `Whether the Thing can only be reached over HTTPS.` |
| `Is the Thing enabled?`, `Are Web Sockets enabled?` | `Whether the Thing is enabled.`, `Whether Web Sockets are enabled.` |
| `Whether the Thing is enabled?` | `Whether the Thing is enabled.` |

On a read-only property nothing is asked of the user, so `Should the Thing be enabled?` becomes `Whether the Thing is enabled.` and `Should the Thing use HTTPS?` becomes `Whether the Thing uses HTTPS.`

The rest have no fix: `Does the Thing use HTTPS?`, a question where the verb cannot be told from the subject, and a garbled one such as `Should the Thing prevent it to be deleted?`

## Options

There are none.

## Ignoring Reports

Put `//azignore:AZS019 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Should the Thing be enabled?", //azignore:AZS019 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
