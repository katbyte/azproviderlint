# AZS020 - resource and property descriptions must not name a resource or property that does not exist

AZS020 reports a name in backticks in a description when the provider has no resource or property called that.

A name goes stale when a property is renamed and the text is not, or is mistyped to begin with. A reader who follows it finds nothing.

## Flagged Code

```go
"client_certificate_mode": {
	Type:        pluginsdk.TypeString,
	Description: "The client certificate mode. This has no effect when `client_cert_enabled` is `false`.",
},
```

```text
description names "client_cert_enabled" - nothing in the provider is called that
```

```go
Description: "The ID of the Subnet. Use `azurerm_subnet_nsg_association` to link a Network Security Group.",
```

```text
description names the resource "azurerm_subnet_nsg_association" - the provider has no resource called that
```

## Passing Code

```go
"client_certificate_mode": {
	Type:        pluginsdk.TypeString,
	Description: "The client certificate mode. This has no effect when `client_certificate_enabled` is `false`.",
},
```

## What counts

Only what is in backticks is read.

- **A resource** starts with the provider's name, which comes from the module in `go.mod`: `terraform-provider-azurerm` has resources called `azurerm_...`. Only the type is checked, so in `azurerm_subnet.example.id` it is `azurerm_subnet`. `azurerm_key_vault_*` stands for several and is not checked.
- **A property** is a lowercase name with an underscore in it, on its own or on the way to one: each part of `site_config.0.linux_fx_version` is checked.
- **A one-word name**, such as `identity`, is checked only where the text calls it a block, property, field, argument or attribute. Otherwise it may as well be a value.

A name exists when it is written out anywhere in the provider's code outside its tests: as a schema key, a resource's type, a value, or a model's `tfschema` tag. A value may also be one of the constants of a package the file's package imports, such as an SDK's.

Terraform's own `for_each`, `depends_on`, `ignore_changes`, `create_before_destroy`, `prevent_destroy` and `replace_triggered_by` pass.

The same descriptions are read as by [AZS011](../AZS011_description_blank#what-counts).

## Limits

- A name is looked for across the whole provider, not the one resource, so a real property of some other resource passes.
- Another provider's resource, such as `azuread_application`, is reported. Add it with `allow`.
- golangci-lint keeps a package's reports until the package changes. One cleared by adding the missing resource somewhere else stays until `golangci-lint cache clean`.

## No fix

The right name cannot be guessed.

## Options

`allow` adds a name to accept though the provider does not define it. Repeat it for several.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          AZS020:
            allow: [azuread_application, azapi_resource]
```

```sh
azproviderlint -AZS020 -AZS020.allow=azuread_application -AZS020.allow=azapi_resource ./...
```

## Ignoring Reports

Put `//azignore:AZS020 - <reason>` at the end of the description's line, or on the line above it. The reason is required.

```go
Description: "Use `azuread_application` to create one.", //azignore:AZS020 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
