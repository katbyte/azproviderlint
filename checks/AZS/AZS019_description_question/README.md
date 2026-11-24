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

`Should the Thing be enabled?` becomes `Whether the Thing should be enabled.` The other questions have no fix, since turning them round takes more than moving a word.

## Options

There are none.

## Ignoring Reports

Put `//azignore:AZS019 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Should the Thing be enabled?", //azignore:AZS019 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
