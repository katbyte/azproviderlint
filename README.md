# azproviderlint

[![GitHub release](https://img.shields.io/github/v/release/katbyte/azproviderlint?color=blueviolet)](https://github.com/katbyte/azproviderlint/releases/latest)
![build](https://github.com/katbyte/azproviderlint/actions/workflows/build.yaml/badge.svg)
![test](https://github.com/katbyte/azproviderlint/actions/workflows/test.yaml/badge.svg)
![lint](https://github.com/katbyte/azproviderlint/actions/workflows/lint.yaml/badge.svg)
![govulncheck](https://github.com/katbyte/azproviderlint/actions/workflows/govulncheck.yaml/badge.svg)
![CodeQL](https://github.com/katbyte/azproviderlint/actions/workflows/codeql-analysis.yml/badge.svg)
[![Go Version](https://img.shields.io/github/go-mod/go-version/katbyte/azproviderlint?color=00ADD8)](https://github.com/katbyte/azproviderlint/blob/main/go.mod)
[![License](https://img.shields.io/github/license/katbyte/azproviderlint?color=blue)](https://github.com/katbyte/azproviderlint/blob/main/LICENSE)

A custom [golangci-lint](https://golangci-lint.run/) module plugin providing Azure provider-specific linting rules built on Go's `analysis` framework.

## Installation

### Standalone

```bash
go install github.com/katbyte/azproviderlint@latest
```

Then run directly (all rules run by default):
```bash
azproviderlint ./...
```

Each rule is also a flag, and setting any rule flag switches to running only the named rules; a category on its own (`-AZG`) runs every rule in that category:
```bash
azproviderlint -AZG001 ./...
azproviderlint -AZR001 -AZR003 ./...
azproviderlint -AZG ./...
azproviderlint -AZG -AZR001 ./...
```

### As a golangci-lint Plugin

Add to your `.custom-gcl.yml`:
```yaml
version: v2.12.2
plugins:
  - module: "github.com/katbyte/azproviderlint"
    import: "github.com/katbyte/azproviderlint/plugin"
    version: v0.1.0
```

Build the custom binary:
```bash
golangci-lint custom
```

Then enable in `.golangci.yml`:
```yaml
linters:
  enable:
    - azproviderlint
  settings:
    custom:
      azproviderlint:
        type: module
```

Individual rules can be enabled/disabled via plugin settings (an empty `enable` list means all rules):
```yaml
linters:
  settings:
    custom:
      azproviderlint:
        type: module
        settings:
          disable: [AZR002]
```

To run just azproviderlint through the custom binary, skipping every other linter:
```bash
custom-gcl run --enable-only azproviderlint ./...
```

There is no CLI flag for a single rule — combine `--enable-only` with an `enable: [AZG001]` list in the plugin settings above, or use the standalone binary's per-rule flags.

### Why the plugin over the standalone binary?

The plugin requires every consumer to build a custom golangci-lint binary (`golangci-lint custom`), but that one-time cost buys a lot on a codebase the size of a provider:

- **One package-load instead of two.** On azurerm this is the dominant cost — loading and type-checking the provider codebase (with its enormous vendor tree) takes minutes, and every separate analysis binary pays it again from scratch. Folding checks into golangci-lint amortizes it, and golangci-lint's result cache makes warm local re-runs dramatically faster; a standalone multichecker reloads the world every single time.
- **Unified config and reporting.** `.golangci.yml` path exclusions (generated files, `/sdk/`, `third_party` — already curated in the provider repo) apply to these checks for free; one output stream, one CI job, SARIF/annotations, and `--new-from-rev` — the killer feature for a codebase with 22+ pre-existing findings per service, since checks can be enforced on new code only instead of azignoring a decade of history.
- **`//nolint` works uniformly** alongside `//azignore` (see [Ignoring Reports](#ignoring-reports)).

The standalone binary remains the right tool for one-off or single-rule runs (`azproviderlint -AZG001 ./...`), editor integrations that expect a plain `analysis`-style vet tool, and quick iteration while developing new checks.

## Rules

Rules are named `AZ<category letter><number>`, aligned with [tfproviderlint](https://github.com/bflad/tfproviderlint)'s category letters (`R`, `S`, `V`, `AT`) where they overlap.

### AZG — General Go Style / Readability

| Rule | Description |
|------|-------------|
| [AZG001](checks/AZG/AZG001_combine_err_assignment_and_check) | `err := SomeFunc()` or `_, err := SomeFunc()` followed by `if err != nil` should be combined into a single `if` init statement |
| [AZG002](checks/AZG/AZG002_error_should_describe_expected_format) | Error messages should describe the expected format instead of saying `invalid format of ...` |
| [AZG003](checks/AZG/AZG003_pointer_to_enum_conversion) | `pointer.To(sdk.SomeEnum(v))` explicit go-azure-sdk enum conversions must use the generic `pointer.ToEnum[sdk.SomeEnum](v)` helper instead |
| [AZG004](checks/AZG/AZG004_zero_value_init_pointer_from) | `y := <zero>; if x != nil { y = *x }` zero-value initialization followed by a nil check and pointer dereference must use the generic `pointer.From(x)` helper instead |
| [AZG005](checks/AZG/AZG005_single_use_temporary) | `x := <expr>` immediately followed by `y = x` or `return x`, with no other use of `x`, should be inlined into the consuming statement |

### AZR — Resource Implementation

| Rule | Description |
|------|-------------|
| [AZR001](checks/AZR/AZR001_set_id_dereferenced_pointer) | `SetId` must not be passed a dereferenced pointer (`d.SetId(*read.ID)`) — use a generated Resource ID Formatter/Parser and `d.SetId(id.ID())` |
| [AZR002](checks/AZR/AZR002_combined_create_update_method) | Resources must register separate `Create` and `Update` methods instead of a combined `CreateUpdate` method |
| [AZR003](checks/AZR/AZR003_resource_data_get_in_delete) | `d.Get` / `metadata.ResourceData.Get` must not be used inside a resource's Delete function, where it does not work as expected |
| [AZR004](checks/AZR/AZR004_resource_id_equality_comparison) | Resource IDs must not be compared with `==`/`!=` — use `resourceids.Match` |
| [AZR005](checks/AZR/AZR005_case_insensitive_segments_feature_flag) | `features.TreatUserSpecifiedSegmentsAsCaseInsensitive` must not be set — the case-aware comparisons feature is not ready for use |
| [AZR006](checks/AZR/AZR006_stop_context_without_timeouts) | `ctx` must not be assigned directly from `meta.(*clients.Client).StopContext` — use `timeouts.ForCreate`/`ForRead`/`ForUpdate`/`ForDelete` so Custom Timeouts work |
| [AZR007](checks/AZR/AZR007_state_change_conf) | `StateChangeConf`from `github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry` must not be used — prefer a custom poller implementing `pollers.PollerType` driven via `pollers.NewPoller(...).PollUntilDone(ctx)` |

### AZD — Data Sources

| Rule | Description |
|------|-------------|
| [AZD001](checks/AZD/AZD001_data_source_empty_set_id) | Data sources must return an error when a resource cannot be found, not call `d.SetId("")` |
| [AZD002](checks/AZD/AZD002_data_source_mark_as_gone) | Data sources must return an error when a resource cannot be found, not call `metadata.MarkAsGone` |

### AZS — Schema & Typed SDK Models

| Rule | Description |
|------|-------------|
| [AZS001](checks/AZS/AZS001_typed_sdk_model_64bit_types) | Typed SDK model fields (tagged `tfschema`) must use 64-bit numeric types — `int64` not `int`/`int16`/`int32`, `float64` not `float32` — including slices, maps, pointers, named types, and aliases of them |
| [AZS002](checks/AZS/AZS002_schema_default_type_mismatch) | Schema `Default` values must match the declared `Type` — a `bool` default on a `TypeInt` schema only fails at plan time; named constants are resolved via the type checker |
| [AZS003](checks/AZS/AZS003_schema_allows_empty_block) | Optional/required `TypeList` blocks whose properties are all optional with no defaults allow `foo {}`, which can crash expand functions or cause spurious diffs — constrain with `AtLeastOneOf`/`ExactlyOneOf`, a `Required` property, or a `Default` |
| [AZS004](checks/AZS/AZS004_enum_validation_missing_values) | `validation.StringInSlice` with a hand-written list of SDK enum values must use the SDK's `PossibleValuesFor<Enum>()` helper instead — partial lists reject valid API values, and even complete lists go stale when the SDK adds new ones |
| [AZS005](checks/AZS/AZS005_resource_missing_data_source) | Registered resources must have a data source of the same name — checked across untyped plugin SDK maps, typed SDK slices and framework wrapped slices, including feature-flagged conditional registration |
| [AZS006](checks/AZS/AZS006_data_source_missing_properties) | Data sources must expose the properties of their same-named resource — compares recursively collected schema property names per registration flavour (untyped maps, typed `Arguments()`/`Attributes()`, framework `Schema()`) and reports resource properties absent from the data source |

### AZC — Clients & SDK Usage

| Rule | Description |
|------|-------------|
| [AZC001](checks/AZC/AZC001_client_missing_base_uri) | Azure SDK (track1 & kermit) clients must be created via `NewFoosClientWithBaseURI` with the resource manager endpoint explicitly specified, not `NewFoosClient(o.SubscriptionId)` |

### AZT — Acceptance Testing

| Rule | Description |
|------|-------------|
| [AZT001](checks/AZT/AZT001_acceptance_test_external_package) | Acceptance test files (resource, data source, action, ephemeral — incl. list and generated variants) must use an external `_test` package to prevent circular dependencies |
| [AZT002](checks/AZT/AZT002_credentials_from_environment) | Tests (`_test.go` files only) must not obtain credentials via `os.Getenv("ARM_CLIENT_ID"/"ARM_CLIENT_SECRET"/"ARM_CLIENT_SECRET_ALT")` — create an `azurerm_user_assigned_identity` with minimal permissions instead |

### AZN — Naming Conventions

_No rules yet — reserved for property naming convention rules (e.g. percentage properties using a `_percentage` suffix rather than `_in_percent`)._

### AZV — Validation

_No rules yet — reserved for missing/incorrect validation rules (e.g. string arguments without a `ValidateFunc`)._

## Ignoring Reports

When run via golangci-lint, all azproviderlint reports on a line can be ignored with a `//nolint:azproviderlint` comment at the end of the offending line or on the line immediately preceding it.

To ignore a specific check — leaving the others active, and working under any driver including the standalone binary — use a `//azignore:<Rule>` comment in the same positions. Multiple rules can be listed separated by commas:

```go
d.SetId(*read.ID) //azignore:AZR001

//azignore:AZG001,AZR003
err := client.Delete(ctx, id)
```
