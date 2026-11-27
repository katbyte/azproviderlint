# AZT011 - a test callback must give its context a deadline before polling with it

AZT011 reports an acceptance test callback that polls with the context it was handed.

`CheckWithClient` and its siblings call back with the provider's own context, which never ends. A poller refuses a context with no deadline, so the test fails with "the context used must have a deadline attached for polling purposes" before it has waited at all.

## Flagged Code

```go
data.CheckWithClientForResource(func(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	return client.Things.DeleteThenPoll(ctx, id)
}, data.ResourceName),
```

## Passing Code

```go
data.CheckWithClientForResource(func(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	return client.Things.DeleteThenPoll(ctx, id)
}, data.ResourceName),
```

## What is checked

The callback handed to `CheckWithClient`, `CheckWithClientForResource` or `CheckWithClientWithoutResource`: a function written in place, or a function or method of the same package.

In it, the context it is handed must not be passed to a call whose name ends in `ThenPoll`, or to `PollUntilDone`, before it has been swapped for one from `context.WithTimeout` or `context.WithDeadline`.

These are not checked:

- a helper the callback passes its context on to
- a test resource's `Exists` and `Destroy`, which are handed a context that has a deadline

## Fix

None is offered. How long to wait is the author's call.

## Ignoring Reports

Put `//azignore:AZT011 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
return client.Things.DeleteThenPoll(ctx, id) //azignore:AZT011 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
