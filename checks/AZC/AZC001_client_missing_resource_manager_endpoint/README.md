# AZC001 - clients must set an explicit resource manager endpoint

AZC001 reports Azure SDK clients (track1 and kermit) created with `NewFoosClient(o.SubscriptionId)`.

Without an endpoint the client quietly talks to Azure Public. That breaks the provider in Azure China, Azure US Government, and any other non-public cloud. Use the `WithBaseURI` constructor and pass in the resource manager endpoint so the right cloud is always used.

## Flagged Code

```go
client := servers.NewServersClient(o.SubscriptionId)
o.ConfigureClient(&client.Client, o.ResourceManagerAuthorizer)
```

## Passing Code

```go
client := servers.NewServersClientWithBaseURI(o.ResourceManagerEndpoint, o.SubscriptionId)
o.ConfigureClient(&client.Client, o.ResourceManagerAuthorizer)
```

## Ignoring Reports

Put `//azignore:AZC001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
client := servers.NewServersClient(o.SubscriptionId) //azignore:AZC001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
