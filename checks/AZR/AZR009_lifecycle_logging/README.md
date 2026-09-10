# AZR009 - no lifecycle narration logging

AZR009 reports log statements inside a resource's Create, Read, Update, or Delete function that only announce the step: `log.Printf("[INFO] preparing arguments for ...")`, `log.Printf("[DEBUG] Retrieving %s", id)`, `metadata.Logger.Infof("creating %s", id)`, `metadata.Logger.Info("Decoding state..")`. The framework already logs each step, so these lines add noise and drift out of date. [hashicorp/terraform-provider-azurerm#32423](https://github.com/hashicorp/terraform-provider-azurerm/pull/32423) removed the pattern across the provider; this check keeps it out.

Lifecycle functions are found the same way as in AZR003: untyped resources' functions registered under `Create:`, `Read:`, `Update:`, or `Delete:`, and typed resources' `Create() sdk.ResourceFunc` methods and siblings, including the function literal inside. A log call counts when its first argument is a string literal that, after any `[LEVEL]` tag, starts with a lifecycle verb: preparing arguments, creating, updating, deleting, retrieving, reading, importing, import check, decoding state, checking for presence/existence, or their past tenses. Helper functions the lifecycle calls are not searched.

Left alone, because they carry information:

- messages about the resource ID (`Updating ID from %s to %s`), which record a state migration
- `not found`, `removing from state`, `does not exist` messages
- `Waiting for ...` polling messages
- messages explaining a skipped step (`... was nil, skipping`)
- any call whose format is not a string literal

The suggested fix deletes the statement and its line. When nothing else in the file still uses the `log` package, the fix on the file's last report also removes the import, so `-fix` leaves the file compiling.

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

## Ignoring Reports

When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZR009 - <reason>` comment instead, in the same positions:

```go
log.Printf("[DEBUG] Creating %s", id) //azignore:AZR009 - <reason>
```
