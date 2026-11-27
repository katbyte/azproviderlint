# AZR015 - the create callback of a resource with an identity must set the identity too

AZR015 reports a create whose callback sets the ID of a resource that has an identity, but not the identity.

The callback saves the resource before a long-running create has finished, so a create that fails part-way is not lost. A resource with an identity has to save that as well, or Terraform fails with "Missing Resource Identity After Create".

## Flagged Code

```go
func resourceThing() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceThingCreate,
		Identity: &schema.ResourceIdentity{
			SchemaFunc: pluginsdk.GenerateIdentitySchema(&things.ThingId{}),
		},
	}
}

func resourceThingCreate(d *pluginsdk.ResourceData, meta any) error {
	if err := client.CreateCallbackThenPoll(ctx, id, payload, sdk.SetIDCallback(meta, &id, d)); err != nil {
```

## Passing Code

```go
func resourceThingCreate(d *pluginsdk.ResourceData, meta any) error {
	if err := client.CreateCallbackThenPoll(ctx, id, payload, sdk.SetIDAndIdentityCallback(meta, &id, d)); err != nil {
```

## What is checked

A call to `SetIDCallback` in the create of a resource with an identity:

- the function a plugin SDK resource names as its `Create`, when the resource sets `Identity`
- the `Create` method of a typed resource that has an `Identity` method, where it is `metadata.SetIDCallback`

An update is left alone: by then both are saved.

## Fix

`SetIDCallback` becomes `SetIDAndIdentityCallback`, which takes the same values.

## Ignoring Reports

Put `//azignore:AZR015 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
if err := client.CreateCallbackThenPoll(ctx, id, payload, sdk.SetIDCallback(meta, &id, d)); err != nil { //azignore:AZR015 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
