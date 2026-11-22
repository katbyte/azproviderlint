# AZP007 - URLs must use the address the page has moved to

AZP007 reports a URL, in a comment or a string, whose page has moved, and fixes it to the new address.

An old address works only for as long as its redirect is kept up.

## Flagged Code

```go
// See https://docs.microsoft.com/rest/api/automation/webhook/generate-uri
Description: "The name of an [Authorization Server](https://www.terraform.io/docs/providers/azurerm/r/api_management_authorization_server.html).",
```

```text
URL starts with `docs.microsoft.com/`, which has moved - use `learn.microsoft.com/`
URL starts with `www.terraform.io/docs/providers/azurerm/r/api_management_authorization_server.html`, which has moved - use `registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/api_management_authorization_server`
```

## Passing Code

```go
// See https://learn.microsoft.com/rest/api/automation/webhook/generate-uri
Description: "The name of an [Authorization Server](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/api_management_authorization_server).",
```

## What has moved

| Old | New |
|---|---|
| `docs.microsoft.com/...` | `learn.microsoft.com/...` |
| `www.terraform.io/docs/providers/<provider>/r/<name>.html` | `registry.terraform.io/providers/hashicorp/<provider>/latest/docs/resources/<name>` |
| `www.terraform.io/docs/providers/<provider>/d/<name>.html` | `registry.terraform.io/providers/hashicorp/<provider>/latest/docs/data-sources/<name>` |
| `www.terraform.io/docs/configuration/functions/<name>.html` | `developer.hashicorp.com/terraform/language/functions/<name>` |
| `github.com/terraform-providers/terraform-provider-<provider>` | `github.com/hashicorp/terraform-provider-<provider>` |
| `www.terraform.io/docs/...`, anything else, and `www.terraform.io/language/...` | reported with no fix |

`<provider>` is `azurerm`, `azuread` or `azurestack`. `terraform.io` without the `www.` counts too. Whatever follows the part that moved, such as the rest of the path or a `#section`, is kept.

The last row is a page that has moved to somewhere the old address does not give away, so it has to be looked up.

A locale in a Microsoft docs URL is [AZP001](../AZP001_microsoft_docs_url_locale)'s to report; both can fire on one URL, and both fixes apply.

## Options

| Option | Default | Effect |
|---|---|---|
| `moved` | none | more moves to report and fix, as comma-separated `old=new` pairs of what the address starts with, after the `https://` |

```yaml
          AZP007:
            moved: "old.example.com/docs/=new.example.com/,wiki.example.com/=docs.example.com/wiki/"
```

```bash
azproviderlint -AZP007 -AZP007.moved="old.example.com/docs/=new.example.com/" ./...
```

A pair given here is tried before the ones built in. See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZP007 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
const doc = "https://docs.microsoft.com/azure/aks/" //azignore:AZP007 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
