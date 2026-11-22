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

400 names are checked, in any other case than the one shown. They are listed in [products.go](products.go):

- the Azure services and ideas that descriptions name most, picked by hand: `Function App`, `Storage Account`, `Resource Group`, `Virtual Network`, `Managed Identity`, `Role Assignment` and 74 more
- every azurerm resource whose name is two plain words or more, as its documentation gives it: `Linux Web App Slot`, `Key Vault Access Policy`, `Automation Runbook` and 317 more

A plural counts too: `function apps` should be `Function Apps`, `managed identities` should be `Managed Identities`.

Each name is two plain words or more. These are left out, and can be added with the `name` option:

- a name of one word, such as `Subscription` or `Subnet`, which is too often an everyday word
- a name with an acronym in it, such as `API Management` or `Public IP`, where [AZS013](../AZS013_description_acronym_case) already reports the acronym
- a name that is also a general term, such as `Load Balancer`, `Network Interface` or `Container Registry`, which a description often uses for something that is not the Azure product
- a resource whose documentation writes its name another way than its type, such as `azurerm_lb_rule`, a `Load Balancer Rule`

Only prose is checked. These are left alone, since there the case is part of a name or a value:

- code in backticks: `` `function app` ``
- web addresses, and the address a link points to
- a name joined to something else by a hyphen: `my-function app`

A description is text written out in the code, outside test files: `Description` on a plugin SDK schema property, block or resource, `Description` or `MarkdownDescription` on a framework schema, attribute or block, or on a provider function or one of its parameters, and what a typed resource or data source returns from `Description() string`. Provider and `features` block descriptions count too.

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
