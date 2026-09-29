# AZR002 - separate Create and Update methods

AZR002 reports resources that register one `CreateUpdate` function as both `Create` and `Update`.

A combined function hides which properties can actually change after creation, makes `ignore_changes` behaviour hard to reason about, and gets in the way of moving the resource to the typed SDK later. New resources should have a separate `Create` and `Update`. Older resources with a combined function are being split over time.

## Flagged Code

```go
return &pluginsdk.Resource{
	Create: resourceExampleCreateUpdate,
	Read:   resourceExampleRead,
	Update: resourceExampleCreateUpdate,
	Delete: resourceExampleDelete,
}
```

## Passing Code

```go
return &pluginsdk.Resource{
	Create: resourceExampleCreate,
	Read:   resourceExampleRead,
	Update: resourceExampleUpdate,
	Delete: resourceExampleDelete,
}
```

## Ignoring Reports

Put `//azignore:AZR002 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
Create: resourceExampleCreateUpdate, //azignore:AZR002 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
