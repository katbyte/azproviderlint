# AZD001 - data sources must error when not found, not SetId("")

AZD001 reports data sources that call `d.SetId("")`.

In a resource, clearing the ID is how you tell Terraform the thing was deleted outside of Terraform, so it should be recreated. A data source is different: the user asked for something by name, and if it is not there they need to be told. Setting an empty ID instead leaves them with empty attributes and confusing diffs downstream, with no explanation. Return an error.

Only files with `data_source` in their name are checked.

## Flagged Code

```go
if response.WasNotFound(resp.HttpResponse) {
	d.SetId("")
	return nil
}
```

## Passing Code

```go
if response.WasNotFound(resp.HttpResponse) {
	return fmt.Errorf("%s was not found", id)
}
```

## Ignoring Reports

Put `//azignore:AZD001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
d.SetId("") //azignore:AZD001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
