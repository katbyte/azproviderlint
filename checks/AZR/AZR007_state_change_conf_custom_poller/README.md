# AZR007 - use custom pollers instead of StateChangeConf and Retry

AZR007 reports `StateChangeConf{...}` literals, and calls to `Retry` and `RetryContext`.

`StateChangeConf` is the plugin SDK's old way of waiting for something to reach a state, and `Retry` builds one to run a function until it stops failing. The provider now prefers custom pollers: a small type that implements the [go-azure-sdk](https://github.com/hashicorp/go-azure-sdk) `pollers.PollerType` interface, run with `pollers.NewPoller(...).PollUntilDone(ctx)`. See the [new resource guide](https://github.com/hashicorp/terraform-provider-azurerm/blob/main/contributing/topics/guide-new-resource.md).

The check resolves the literal's type, so it catches `pluginsdk.StateChangeConf`, `retry.StateChangeConf`, and `resource.StateChangeConf` under any import alias, by value or by pointer.

`Retry` and `RetryContext` are caught as the plugin SDK's own, and as a provider's wrapper of them in a package named `pluginsdk`.

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

```go
return pluginsdk.Retry(d.Timeout(pluginsdk.TimeoutCreate), func() *pluginsdk.RetryError {
	if _, err := client.Create(ctx, id, payload); err != nil {
		return pluginsdk.RetryableError(err)
	}

	return nil
})
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

Put `//azignore:AZR007 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
stateConf := &pluginsdk.StateChangeConf{ //azignore:AZR007 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
