# AZD002 - data sources must error when not found, not MarkAsGone

AZD002 reports data sources that call `metadata.MarkAsGone(...)`.

`MarkAsGone` is for resources: it removes something from state after it was deleted outside of Terraform. A data source is different: the user asked for something by name, and if it is not there they need to be told why their config cannot be applied. Return an error.

Only files with `data_source` in their name are checked.

## Flagged Code

```go
if response.WasNotFound(resp.HttpResponse) {
	return metadata.MarkAsGone(id)
}
```

## Passing Code

```go
if response.WasNotFound(resp.HttpResponse) {
	return fmt.Errorf("%s was not found", id)
}
```

## Ignoring Reports

Put `//azignore:AZD002 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
return metadata.MarkAsGone(id) //azignore:AZD002 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
