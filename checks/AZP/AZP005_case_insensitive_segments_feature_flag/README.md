# AZP005 - do not set the case-insensitive segments feature flag

AZP005 reports any assignment to `features.TreatUserSpecifiedSegmentsAsCaseInsensitive`.

The case-insensitive ID comparison feature is not finished. Too many other pieces still need to change before it works without causing more problems than it fixes. Until that work is done the flag must not be set or exposed anywhere.

## Flagged Code

```go
features.TreatUserSpecifiedSegmentsAsCaseInsensitive = true
```

## Passing Code

Remove the assignment.

## Ignoring Reports

Put `//azignore:AZP005 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
features.TreatUserSpecifiedSegmentsAsCaseInsensitive = true //azignore:AZP005 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
