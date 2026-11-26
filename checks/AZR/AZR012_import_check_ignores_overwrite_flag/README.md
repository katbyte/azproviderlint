# AZR012 - the already-exists check on create must sit inside the overwrite feature flag check

AZR012 reports a create that returns the "already exists, import it" error without asking whether the provider is set to overwrite.

A user who has switched on `skip_import_check_on_create_and_allow_overwriting_existing_resources` has asked for create to go ahead. A create that does not ask refuses them anyway.

## Flagged Code

```go
existing, err := client.Get(ctx, id)
if err != nil && !response.WasNotFound(existing.HttpResponse) {
	return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
}
if !response.WasNotFound(existing.HttpResponse) {
	return tf.ImportAsExistsError("azurerm_thing", id.ID())
}
```

## Passing Code

```go
if !meta.(*clients.Client).Features.SkipImportCheckOnCreateAndAllowOverwritingExistingResources {
	existing, err := client.Get(ctx, id)
	if err != nil && !response.WasNotFound(existing.HttpResponse) {
		return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
	}
	if !response.WasNotFound(existing.HttpResponse) {
		return tf.ImportAsExistsError("azurerm_thing", id.ID())
	}
}
```

A typed resource asks `metadata.Client.Features` the same way, around `metadata.ResourceRequiresImport`.

## What is checked

A call to `tf.ImportAsExistsError` or `metadata.ResourceRequiresImport` has to be inside an `if` whose condition names `SkipImportCheckOnCreateAndAllowOverwritingExistingResources`. Asking somewhere else in the function does not count.

## Fix

None is offered. The lookup of the existing resource usually moves inside the `if` too.

## Ignoring Reports

Put `//azignore:AZR012 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return tf.ImportAsExistsError("azurerm_thing", id.ID()) //azignore:AZR012 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
