# AZS004 - enum validation must use the SDK's possible-values helper

AZS004 reports `validation.StringInSlice([]string{...}, ...)` where the hand-written list is made of an SDK enum's constants.

A hand-written list is a copy of something the SDK already knows. If it is missing a value, users cannot set something the API accepts. If it is complete today, it is wrong the day the SDK adds a value. When the SDK ships a possible-values helper, use it.

The report tells you which case you are in:

- **Missing values**: the message names them.
- **Complete list**: switch to the helper.
- **Extra values** that are not in the enum at all: the message names them. Typos should be fixed. Deliberate extras should be appended to the helper's result rather than kept in a hand-written list.

## Flagged Code

```go
validation.StringInSlice([]string{
	string(virtualmachines.VirtualMachinePriorityTypesLow),
	string(virtualmachines.VirtualMachinePriorityTypesRegular),
	// missing VirtualMachinePriorityTypesSpot ("Spot")
}, false)
```

## Passing Code

```go
validation.StringInSlice(virtualmachines.PossibleValuesForVirtualMachinePriorityTypes(), false)
```

## What counts

A type is treated as an enum only when its package exports a possible-values helper: `PossibleValuesFor<Enum>()` in go-azure-sdk, or `Possible<Enum>Values()` in older SDKs. A named string type with a couple of convenience constants is not reported.

The older SDKs' helpers return a slice of the enum type, which `StringInSlice` will not take. The advice then names a conversion: `validation.StringInEnumSlice(cdn.PossibleTransformValues(), false)` when the validation package has that generic wrapper (azurerm's `internal/tf/validation` does), otherwise `pointer.FromEnumSlice(pointer.To(cdn.PossibleTransformValues()))` from go-azure-helpers.

`StringInSlice` is matched by name and signature in any package whose path is or ends in `validation`, so both the plugin SDK's `helper/validation` and provider wrappers of it count. Plain string literals in the list count toward coverage. Lists with non-constant elements, or with constants from more than one enum, are skipped because they cannot be proven incomplete.

## Options

| Option | Default | Effect |
|---|---|---|
| `allow-missing-values` | false | do not report lists that leave out enum values (deliberate subsets) |
| `allow-extra-values` | false | do not report lists that add values not in the enum (deliberate supersets) |

A list that matches the enum exactly is always reported, since the helper is a pure win there. Set options with `-AZS004.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Deliberately supporting only part of an enum is a fine reason to suppress this. Put `//azignore:AZS004 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
ValidateFunc: validation.StringInSlice([]string{ //azignore:AZS004 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
