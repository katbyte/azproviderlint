# AZR009 - no lifecycle narration logging in Create/Read/Update/Delete

AZR009 reports log lines in a resource's Create, Read, Update, or Delete that only say which step is running: `log.Printf("[INFO] preparing arguments for ...")`, `log.Printf("[DEBUG] Retrieving %s", id)`, `metadata.Logger.Infof("creating %s", id)`, `metadata.Logger.Info("Decoding state..")`.

The framework already logs every step, so these lines add nothing and go stale as the code changes. [hashicorp/terraform-provider-azurerm#32423](https://github.com/hashicorp/terraform-provider-azurerm/pull/32423) removed them across the provider. This check keeps them from coming back.

## Flagged Code

```go
func resourceExampleCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	log.Printf("[INFO] preparing arguments for Azure Example creation.")
	// ...
	log.Printf("[DEBUG] Creating %s", id)
}
```

```go
func (r ExampleResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			metadata.Logger.Info("Decoding state..")
			metadata.Logger.Infof("retrieving %s", id)
			// ...
		},
	}
}
```

## Passing Code

```go
func resourceExampleRead(d *pluginsdk.ResourceData, meta interface{}) error {
	// ...
	if response.WasNotFound(resp.HttpResponse) {
		log.Printf("[DEBUG] %s was not found - removing from state!", id)
		d.SetId("")
		return nil
	}
	log.Printf("[DEBUG] Updating ID from %q to %q", d.Id(), id.ID())
}
```

## What counts

The lifecycle functions are found the same way as in AZR003: whatever is registered under `Create:`, `Read:`, `Update:`, or `Delete:` in an untyped resource, and the `Create() sdk.ResourceFunc` methods and their siblings in a typed one, including the function literal inside. Helpers called from those functions are not searched.

A log call is reported when its format is a string literal that, after any `[LEVEL]` tag, starts with a step verb: preparing arguments, creating, updating, deleting, retrieving, reading, importing, import check, decoding state, checking for presence or existence, or the past tense of those.

Kept, because they say something the framework does not:

- messages about the resource ID (`Updating ID from %s to %s`), which record a state migration
- `not found`, `removing from state`, and `does not exist` messages
- `Waiting for ...` polling messages
- messages explaining a skipped step (`... was nil, skipping`)
- any call whose format is not a string literal

## The fix

`-fix` deletes the line. When nothing else in the file uses the `log` package, the import goes too, so the file still compiles.

## Ignoring Reports

Put `//azignore:AZR009 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
log.Printf("[DEBUG] Creating %s", id) //azignore:AZR009 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
