# AZP001 - Microsoft docs URLs must not carry a locale

The AZP001 analyzer reports a `learn.microsoft.com`, `docs.microsoft.com`, or `msdn.microsoft.com` URL whose first path segment is a locale, such as `/en-us/` or `/en-gb/`, anywhere in a comment or string literal. Microsoft serves a reader their own language when the segment is absent, and pins them to English when it is present. hashicorp/terraform-provider-azurerm#33403 stripped the segment across the provider and added a website check so it does not come back; this rule covers the Go side.

The fix deletes the segment and nothing else, so `https://learn.microsoft.com/en-us/azure/aks/` becomes `https://learn.microsoft.com/azure/aks/`. Apply with `-fix`.

## Flagged Code

```go
// See https://learn.microsoft.com/en-us/azure/aks/egress-outboundtype
Description: "https://docs.microsoft.com/en-us/azure/batch/batch-api-basics#pool",
```

## Passing Code

```go
// See https://learn.microsoft.com/azure/aks/egress-outboundtype
Description: "https://docs.microsoft.com/azure/batch/batch-api-basics#pool",
```

## Ignoring Reports

When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZP001 - <reason>` comment instead, in the same positions:

```go
const doc = "https://learn.microsoft.com/en-us/azure/aks/" //azignore:AZP001 - <reason>
```
