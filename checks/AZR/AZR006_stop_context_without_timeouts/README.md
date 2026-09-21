# AZR006 - use timeouts wrappers, not StopContext

AZR006 reports `ctx := meta.(*clients.Client).StopContext`.

Users can set custom timeouts on a resource, but they only take effect when the context is wrapped with the matching helper: `timeouts.ForCreate`, `ForCreateUpdate`, `ForRead`, `ForUpdate`, or `ForDelete`. The resource also needs a `Timeouts` block in its schema. Using `StopContext` directly means the user's timeout is ignored.

## Flagged Code

```go
func resourceExampleCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	ctx := meta.(*clients.Client).StopContext
	// ...
}
```

## Passing Code

```go
func resourceExampleCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	// ...
}
```

with `Timeouts` set on the resource:

```go
Timeouts: &pluginsdk.ResourceTimeout{
	Create: pluginsdk.DefaultTimeout(30 * time.Minute),
	Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
	Update: pluginsdk.DefaultTimeout(30 * time.Minute),
	Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
},
```

## Ignoring Reports

Put `//azignore:AZR006 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
ctx := meta.(*clients.Client).StopContext //azignore:AZR006 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
