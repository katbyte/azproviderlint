# AZR007

The AZR007 analyzer reports use of `StateChangeConf{...}`. Going forward the provider prefers custom pollers that implement the [go-azure-sdk](https://github.com/hashicorp/go-azure-sdk) `pollers.PollerType` interface and are driven via `pollers.NewPoller(...).PollUntilDone(ctx)`.

The check matches by resolving the composite literal's type to the underlying `StateChangeConf` declared in `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry`. This catches every spelling — `pluginsdk.StateChangeConf{...}`, `retry.StateChangeConf{...}`, `resource.StateChangeConf{...}` — regardless of import alias. It matches whether the literal is taken by value or by pointer.

Caveat: this check intentionally covers explicit `StateChangeConf{...}` composite literals only. Helper APIs such as `pluginsdk.Retry()` hide the `StateChangeConf` construction internally, so they are not reported by AZR007 even though they use the same deprecated retry machinery.

Reference: [terraform-provider-azurerm#guide-new-resource](https://github.com/hashicorp/terraform-provider-azurerm/blob/main/contributing/topics/guide-new-resource.md)

## Flagged Code

```go
stateConf := &pluginsdk.StateChangeConf{
	Pending: []string{"Creating"},
	Target:  []string{"Created"},
	Refresh: refreshFunc,
	Timeout: 10 * time.Minute,
}
result, err := stateConf.WaitForStateContext(ctx)
```

## Passing Code

```go
pollerType := custompollers.NewMyCustomPoller(...)
poller := pollers.NewPoller(pollerType, 10*time.Second, pollers.DefaultNumberOfDroppedConnectionsToAllow)
if err := poller.PollUntilDone(ctx); err != nil {
	return err
}
```

## Ignoring Reports

When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it:

```go
result, err := stateConf.WaitForStateContext(ctx) //nolint:azproviderlint
```
To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZR007` comment instead, in the same positions:

```go
stateConf := &pluginsdk.StateChangeConf{ //azignore:AZR007
```
