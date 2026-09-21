# AZT001 - acceptance tests must use a _test package

AZT001 reports acceptance test files declared in the service package (`package compute`) instead of the external test package (`package compute_test`).

The acceptance test framework imports the service packages. If an acceptance test lives inside the service package and imports the framework, that is an import cycle. Putting the tests in a `_test` package avoids it.

A file counts as an acceptance test when its name ends in `_resource_test.go`, `_data_source_test.go`, `_action_test.go`, or `_ephemeral_test.go`, including the `_list`, `_identity`, and `_gen` variants such as `_resource_list_test.go` or `_resource_identity_gen_test.go`. The suffix has to match exactly, so unit tests that only happen to contain the word `resource` (`storage_queue_resource_manager_id_test.go`, `parse/resource_group_assignment_test.go`) are left alone.

## Flagged Code

```go
// in example_resource_test.go
package compute
```

## Passing Code

```go
// in example_resource_test.go
package compute_test
```

## Ignoring Reports

Put `//azignore:AZT001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
package compute //azignore:AZT001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
