# AZS016 - Azure product names in resource and property descriptions must be capitalised

AZS016 reports the name of an Azure product in a resource or property description that is not capitalised the way the provider writes it everywhere else: `function app` for `Function App`, `Storage account` for `Storage Account`.

## Flagged Code

```go
"storage_account_name": {
	Type:        pluginsdk.TypeString,
	Description: "The backend storage account name which will be used by this Function app, in the same resource group.",
},
```

```text
"storage account" should be "Storage Account"
"Function app" should be "Function App"
"resource group" should be "Resource Group"
```

## Passing Code

```go
"storage_account_name": {
	Type:        pluginsdk.TypeString,
	Description: "The backend Storage Account name which will be used by this Function App, in the same Resource Group.",
},
```

## What counts

The names checked are listed in [products.go](products.go), each in the case it should have:

- the Azure services and ideas that descriptions name most, picked by hand: `Function App`, `Storage Account`, `Resource Group`
- every azurerm resource whose name is two plain words or more: `Linux Web App Slot`, `Key Vault Access Policy`

A plural counts too: `function apps` should be `Function Apps`.

These are left out, and can be added with the `name` option:

- a name of one word, such as `Subscription`, which is too often an everyday word
- a name with an acronym in it, such as `Public IP`, which [AZS013](../AZS013_description_acronym_case) already reports
- a name that is also a general term, such as `Load Balancer` or `Container Registry`
- a resource whose documentation names it differently from its type, such as `azurerm_lb_rule`, a `Load Balancer Rule`

Only prose is checked. Code in backticks, web addresses and a name joined to something else by a hyphen, such as `my-function app`, are left alone.

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Fix

The fix rewrites the name with its capitals.

## Options

| Option | Default | Effect |
|---|---|---|
| `name` | | more names to check, each written the way it should be |
| `ignore` | | names from the list above to leave alone |

Both take several values: a list in the golangci settings, or the flag repeated on the command line.

```yaml
          AZS016:
            name: [Load Balancer, Elastic Pool]
            ignore: [Virtual Machine]
```

```bash
azproviderlint -AZS016 -AZS016.name="Load Balancer" -AZS016.name="Elastic Pool" -AZS016.ignore="Virtual Machine" ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZS016 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "The name of the storage account.", //azignore:AZS016 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
