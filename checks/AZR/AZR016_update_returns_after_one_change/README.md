# AZR016 - an update must not return after handling one changed property while others are still to be looked at

AZR016 reports an update that returns once it has handled one changed property, with others still to be looked at below.

When two properties change in the same apply, the second is never sent. Terraform saves it as done, and the next plan shows the same change again.

## Flagged Code

```go
if d.HasChange("sku_name") {
	if err := client.UpdateSku(ctx, id, sku); err != nil {
		return fmt.Errorf("updating sku of %s: %+v", id, err)
	}

	return resourceThingRead(d, meta)
}

if d.HasChange("tags") {
	...
}
```

```text
this returns once "sku_name" is handled, so a change to "tags" in the same apply is never sent
```

## Passing Code

```go
if d.HasChange("sku_name") {
	if err := client.UpdateSku(ctx, id, sku); err != nil {
		return fmt.Errorf("updating sku of %s: %+v", id, err)
	}
}

if d.HasChange("tags") {
	...
}

return resourceThingRead(d, meta)
```

## What is checked

An `if` that asks a `ResourceData` whether something has changed and ends by returning `nil` or the result of a call, when a later statement beside it asks about a change too.

These are left alone:

- a return of `fmt.Errorf` or `errors.New`, which refuses the change
- a return of an error held in a variable
- an `if` with an `else`
- an `if` that asks whether something has not changed, as in `if !d.HasChange("sku") { return nil }`

## Fix

None is offered. Whether the properties can be sent together is the author's call.

## Ignoring Reports

Put `//azignore:AZR016 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return resourceThingRead(d, meta) //azignore:AZR016 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
