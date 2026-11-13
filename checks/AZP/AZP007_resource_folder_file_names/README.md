# AZP007 - files in a resource folder must use the agreed names

AZP007 reports a file in a resource's or action's own folder that is not named like the rest.

Every such folder uses the same few file names, so the create of any resource is in `r_create.go` and its schema in `r_schema.go`. A file named anything else is somewhere nobody looks, and is usually shared code that belongs in the service's `helpers` package.

## Flagged Code

```text
internal/services/loadbalancer/resources/lb_probe/
  resource.go
  r_create.go
  r_read.go
  r_delete.go
  r_schema.go
  loadbalancer.go
```

```text
loadbalancer.go is not one of the file names a resource or action folder uses: move its code into one that is, or into the service's helpers package if it is shared
```

## Passing Code

```text
internal/services/loadbalancer/resources/lb_probe/
  resource.go
  r_create.go
  r_read.go
  r_delete.go
  r_schema.go
```

## What counts

The folders `internal/services/<service>/resources/<name>` and `internal/services/<service>/actions/<name>`, by their path. A file in one must be named one of:

| Name | Holds |
|---|---|
| `resource.go`, `data_source.go`, `list.go`, `action.go`, `ephemeral.go` | the type and its registration |
| `r_create.go`, `r_read.go`, `r_update.go`, `r_delete.go` | one function each |
| `r_schema.go`, `r_models.go` | the schema and the models |
| `r_expand.go`, `r_flatten.go` | expand and flatten functions |

Each file is reported on its package line.

Not reported:

- tests, which are named after the file they test; see [AZT008](../../AZT/AZT008_test_file_not_named_after_resource)
- generated files, named `_gen.go`
- folders inside a resource folder, such as `migration`

## No fix

AZP007 only reports. Move the code into the file it belongs in, or into the service's `helpers` package when other resources use it.

## Options

| Option | Default | Effect |
|---|---|---|
| `allow` | | file names accepted as well |

`allow` takes several values: a list in the golangci settings, or the flag repeated on the command line. They can also be comma-separated in one entry.

```yaml
          AZP007:
            allow: [helper.go, customizediff.go]
```

```bash
azproviderlint -AZP007 -AZP007.allow=helper.go -AZP007.allow=customizediff.go ./...
```

See the [root README](../../../README.md#options) for where the settings go.

## Ignoring Reports

Put `//azignore:AZP007 - <reason>` at the end of the reported package line, or on the line above it. The reason is required.

```go
package lb_probe //azignore:AZP007 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
