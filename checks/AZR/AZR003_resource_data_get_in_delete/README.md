# AZR003 - no d.Get in Delete functions

AZR003 reports schema reads inside a resource's Delete function: `d.Get(...)` in untyped resources, and `metadata.ResourceData.Get(...)` in typed resources.

When Terraform deletes a resource, the config may already be gone and the state may be partial, so `Get` can return empty or stale values. Everything Delete needs should come from parsing the resource ID.

The Delete function is whatever is registered under `Delete:` for an untyped resource, or the `Delete() sdk.ResourceFunc` method for a typed one.

## Flagged Code

```go
func resourceExampleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	// ...
}
```

```go
func (r ExampleResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			name := metadata.ResourceData.Get("name").(string)
			// ...
		},
	}
}
```

## Passing Code

```go
func resourceExampleDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	id, err := parse.ExampleID(d.Id())
	if err != nil {
		return err
	}
	// use id.Name, id.ResourceGroup, ...
}
```

## Ignoring Reports

Put `//azignore:AZR003 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
name := d.Get("name").(string) //azignore:AZR003 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
