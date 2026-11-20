# AZS011 - resource and property descriptions must not be blank

AZS011 reports a resource or property description that is empty, or only spaces.

A blank description looks done and documents nothing. Leaving it out is more honest, and writing one is better.

## Flagged Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "",
},
```

```text
description is blank - write one, or remove it
```

## Passing Code

```go
"name": {
	Type:        pluginsdk.TypeString,
	Description: "The name of the Thing.",
},
```

## What counts

A description is text written out in the code, outside test files, in one of these places:

- `Description` on a plugin SDK schema property or resource (`pluginsdk.Schema`, `pluginsdk.Resource`)
- `Description` or `MarkdownDescription` on a framework schema, attribute or block
- `Description` or `MarkdownDescription` on a provider function or one of its parameters
- what a typed resource or data source returns from `Description() string`

Provider and `features` block descriptions count too. Text built when the provider runs, such as with `fmt.Sprintf`, is not checked.

A property with no `Description` at all is not reported.

## No fix

Removing the line is easy to do by hand, and writing the description is the better answer.

## Ignoring Reports

Put `//azignore:AZS011 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "", //azignore:AZS011 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
