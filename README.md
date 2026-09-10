# azproviderlint

[![GitHub release](https://img.shields.io/github/v/release/katbyte/azproviderlint?color=blueviolet)](https://github.com/katbyte/azproviderlint/releases/latest)
[![Go Version](https://img.shields.io/github/go-mod/go-version/katbyte/azproviderlint?label=go&color=00ADD8)](https://github.com/katbyte/azproviderlint/blob/main/go.mod)
[![License](https://img.shields.io/github/license/katbyte/azproviderlint?color=blue)](https://github.com/katbyte/azproviderlint/blob/main/LICENSE)
![build](https://github.com/katbyte/azproviderlint/actions/workflows/build.yaml/badge.svg)
![lint](https://github.com/katbyte/azproviderlint/actions/workflows/pr-golangci-lint.yaml/badge.svg)
![CodeQL](https://github.com/katbyte/azproviderlint/actions/workflows/codeql-analysis.yml/badge.svg)

Lint rules for the [Terraform AzureRM provider](https://github.com/hashicorp/terraform-provider-azurerm). Each rule catches one mistake that comes up in provider code, from schema fields that break at plan time to pointer dereferences that panic when Azure leaves a field out. Most rules can fix what they find.

It runs on its own or as a [golangci-lint](https://golangci-lint.run/) plugin. The rules are ordinary Go `analysis` passes.

## Quick start

```bash
go install github.com/katbyte/azproviderlint@latest
azproviderlint ./...
```

That runs every rule. To run some of them, name them. A category on its own (`-AZG`) means every rule in it.

```bash
azproviderlint -AZG001 ./...
azproviderlint -AZR001 -AZR003 ./...
azproviderlint -AZG ./...
azproviderlint -AZG -AZR001 ./...
```

Add `-fix` to apply the suggested fixes. Read the diff afterwards. Each rule's README says what its fix does and when to be careful.

## Rules

Rules are named `AZ`, a category letter, and a number. The letters follow [tfproviderlint](https://github.com/bflad/tfproviderlint) where the categories overlap.

### AZG - General Go Style / Readability

| Rule | Description |
|------|-------------|
| [AZG000](checks/AZG/AZG000_azignore_missing_reason) | azignore directives must give a reason |
| [AZG001](checks/AZG/AZG001_combine_err_assignment_and_check) | combine err assignment and check into one if |
| [AZG002](checks/AZG/AZG002_address_of_single_use_temporary) | use new() instead of a single-use temporary's address |
| [AZG003](checks/AZG/AZG003_pointer_to_enum_conversion) | use pointer.ToEnum for enum conversions |
| [AZG004](checks/AZG/AZG004_zero_value_init_pointer_from) | use pointer.From instead of nil-check dereference |
| [AZG005](checks/AZG/AZG005_single_use_temporary) | inline single-use variable only used in a later assignment or return |
| [AZG006](checks/AZG/AZG006_single_use_call_argument) | inline single-use variable only used in a later function call |
| [AZG007](checks/AZG/AZG007_redundant_zero_value_field) | omit struct literal fields explicitly set to their zero value |
| [AZG008](checks/AZG/AZG008_unchecked_nil_dereference) | pointer dereferences must have a nil guard or use pointer.From |

### AZR - Resource Implementation

| Rule | Description |
|------|-------------|
| [AZR001](checks/AZR/AZR001_set_id_dereferenced_pointer) | SetId must use a resource id formatter/parser |
| [AZR002](checks/AZR/AZR002_combined_create_update_method) | separate Create and Update methods |
| [AZR003](checks/AZR/AZR003_resource_data_get_in_delete) | no d.Get in Delete functions |
| [AZR004](checks/AZR/AZR004_resource_id_equality_comparison) | compare resource id types with resourceids.Match |
| [AZR005](checks/AZR/AZR005_case_insensitive_segments_feature_flag) | do not set the case-insensitive segments feature flag |
| [AZR006](checks/AZR/AZR006_stop_context_without_timeouts) | use timeouts wrappers, not StopContext |
| [AZR007](checks/AZR/AZR007_state_change_conf_custom_poller) | use custom pollers instead of StateChangeConf |
| [AZR008](checks/AZR/AZR008_flatten_returns_nil_slice) | flatten functions must return empty slices/maps, not nil |
| [AZR009](checks/AZR/AZR009_lifecycle_logging) | no lifecycle narration logging in Create/Read/Update/Delete |
| [AZR010](checks/AZR/AZR010_flatten_handles_nil_input) | flatten functions handle nil input themselves, not their callers |

### AZD - Data Sources

| Rule | Description |
|------|-------------|
| [AZD001](checks/AZD/AZD001_data_source_empty_set_id) | data sources must error when not found, not SetId("") |
| [AZD002](checks/AZD/AZD002_data_source_mark_as_gone) | data sources must error when not found, not MarkAsGone |

### AZS - Schema & Typed SDK Models

| Rule | Description |
|------|-------------|
| [AZS001](checks/AZS/AZS001_typed_sdk_model_64bit_types) | typed SDK model numeric fields must be 64-bit (int64/float64) |
| [AZS002](checks/AZS/AZS002_schema_default_type_mismatch) | schema Default values must match the declared Type |
| [AZS003](checks/AZS/AZS003_schema_allows_empty_block) | TypeList blocks must not allow empty blocks |
| [AZS004](checks/AZS/AZS004_enum_validation_possible_values) | enum validation must use the SDK's possible-values helper |
| [AZS005](checks/AZS/AZS005_resource_missing_data_source) | registered resources must have a same-named data source |
| [AZS006](checks/AZS/AZS006_data_source_missing_properties) | data sources must expose their same-named resource's properties |
| [AZS007](checks/AZS/AZS007_optional_computed_missing_comment) | optional+computed fields must have a Note: O+C comment |
| [AZS008](checks/AZS/AZS008_registration_entries_sorted) | registration entries must be sorted alphabetically |
| [AZS009](checks/AZS/AZS009_computed_only_field_input_attributes) | computed-only fields must not set input-only schema attributes |

### AZC - Clients & SDK Usage

| Rule | Description |
|------|-------------|
| [AZC001](checks/AZC/AZC001_client_missing_resource_manager_endpoint) | clients must set an explicit resource manager endpoint |

### AZP - Provider-Wide Conventions

| Rule | Description |
|------|-------------|
| [AZP001](checks/AZP/AZP001_microsoft_docs_url_locale) | Microsoft docs URLs must not carry a locale segment like /en-us/ |

### AZT - Acceptance Testing

| Rule | Description |
|------|-------------|
| [AZT001](checks/AZT/AZT001_acceptance_test_external_package) | acceptance tests must use a _test package |
| [AZT002](checks/AZT/AZT002_credentials_from_environment) | acceptance tests must not read credentials from the environment |

### AZN - Naming Conventions

No rules yet. Reserved for property naming rules, such as percentages using a `_percentage` suffix rather than `_in_percent`.

### AZV - Validation

| Rule | Description |
|------|-------------|
| [AZV001](checks/AZV/AZV001_error_should_describe_expected_format) | 'invalid format' error messages must describe the expected format |

## Terms the rule docs use

If you are new to the provider, these come up a lot:

- **Untyped resource**: the original plugin SDK style. A `map[string]*pluginsdk.Schema`, and `d.Get` / `d.Set` to move values in and out of state.
- **Typed resource**: the newer `internal/sdk` style. A Go struct with `tfschema` tags, `Arguments()` / `Attributes()` for the schema, and `metadata.Decode` / `metadata.Encode` instead of `Get` / `Set`.
- **Framework resource**: the Terraform Plugin Framework style, registered through `FrameworkResources()`.
- **Expand / flatten**: `expandFoo` turns config into an SDK request. `flattenFoo` turns an SDK response into what goes in state.
- **Resource ID**: the provider's own parsed form of an Azure resource ID, with a generated formatter and parser per type. See [AZR001](checks/AZR/AZR001_set_id_dereferenced_pointer).
- **go-azure-sdk / go-azure-helpers**: the SDK the provider calls Azure with, and the helper library that provides `pointer.To`, `pointer.From`, and friends.

## Ignoring reports

To skip one rule on one line, add a comment at the end of the line or on the line above it. Say why. A directive without a reason is reported by [AZG000](checks/AZG/AZG000_azignore_missing_reason).

```go
d.SetId(*read.ID) //azignore:AZR001 - legacy resource, ID formatter tracked in #1234

//azignore:AZG001,AZR003 combined form obscures the retry loop here
err := client.Delete(ctx, id)
```

Several rules can be listed with commas. The `-` before the reason is optional. This works under every driver, standalone or golangci-lint.

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint rule on that line, so prefer `//azignore` when you can.

## Using it with golangci-lint

The plugin has to be compiled into a custom golangci-lint binary. Add it to `.custom-gcl.yml`:

```yaml
version: v2.12.2
plugins:
  - module: "github.com/katbyte/azproviderlint"
    import: "github.com/katbyte/azproviderlint/plugin"
    version: v0.1.0
```

Build the binary and enable the linter:

```bash
golangci-lint custom
```

```yaml
linters:
  enable:
    - azproviderlint
  settings:
    custom:
      azproviderlint:
        type: module
```

To run only azproviderlint through the custom binary:

```bash
custom-gcl run --enable-only azproviderlint ./...
```

### Choosing rules

Every rule runs by default. Under `settings`, `enable` narrows that to a list and `disable` removes from whatever is enabled. Both accept rule names or whole categories.

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        type: module
        settings:
          enable: [AZG]         # only the AZG rules...
          disable: [AZG005]     # ...except AZG005
```

There is no per-rule flag on the golangci-lint command line. Use the settings above, or the standalone binary.

### Options

Some rules take options. Each rule's README lists them. In golangci-lint they go under the rule's name in the same settings block. On the standalone binary they are `-<RULE>.<option>` flags.

```yaml
        settings:
          AZS004: {allow-extra-values: true}
          AZG005: {max-gap: 50}
```

```bash
azproviderlint -AZG005 -AZG005.max-gap=50 ./...
```

### Why bother with the plugin?

Building a custom binary is a one-time cost, and on a codebase the size of azurerm it pays off:

- **Loading the provider once, not twice.** Type-checking azurerm and its vendor tree takes minutes. A separate binary does it all over again. Inside golangci-lint it is shared, and the result cache makes warm re-runs fast.
- **One config, one output.** The path exclusions already in the provider's `.golangci.yml` (generated files, `/sdk/`, `third_party`) apply for free. So do SARIF, annotations, and `--new-from-rev`, which lets a rule be enforced on new code while a decade of existing findings is left alone.
- **`//nolint` works** alongside `//azignore`.

The standalone binary is still the right tool for a quick one-rule run, for editors that expect a plain `analysis` vet tool, and for developing new rules.
