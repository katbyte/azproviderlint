# AZP006 - a service's shared code belongs in helpers, not helper

AZP006 reports a service package named `helper`.

Code shared by a service's resources lives in `internal/services/<service>/helpers`. A few services use `helper` for the same thing, and a service that gains a `helpers` beside its `helper` leaves every reader guessing which one holds what.

## Flagged Code

```text
internal/services/automation/
  helper/
    automation_variable.go
  helpers/
    automation_variable.go
```

```text
this service has both a helper and a helpers package: move this one into helpers
```

A service with `helper` alone is reported too:

```text
a service's shared code belongs in a package named helpers, not helper
```

## Passing Code

```text
internal/services/automation/
  helpers/
    automation_variable.go
```

## What counts

The folder `internal/services/<service>/helper`, by its path. One report is made per package, on the package line of its first file by name.

Not reported:

- a `helper` folder anywhere else, such as a resource named helper under `resources`
- the package's tests, which add nothing to the report

## No fix

AZP006 only reports. Rename the folder and the package to `helpers` and update the imports, or move the files into the `helpers` the service already has.

## Ignoring Reports

Put `//azignore:AZP006 - <reason>` at the end of the reported package line, or on the line above it. The reason is required.

```go
package helper //azignore:AZP006 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
