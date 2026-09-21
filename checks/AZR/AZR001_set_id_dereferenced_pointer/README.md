# AZR001 - SetId must use a resource id formatter/parser

AZR001 reports `d.SetId(*read.ID)`: setting the Terraform ID straight from the pointer the Azure API returned.

Azure is not consistent about resource IDs. Casing and segment order vary between services, and sometimes between operations of the same service. If the ID Azure hands back goes straight into state, the same resource can end up with different IDs on different days, which shows up as spurious diffs and failed imports. So the provider builds its own IDs with a generated formatter and parser, and calls `d.SetId(id.ID())`.

To generate a formatter, parser, and validator for a new ID, add a `resourceids.go` file to the service package:

```go
//go:generate go run ../../tools/generator-resource-id/main.go -path=./ -name=Server -id={the value of the Resource ID}
```

then run `make generate`.

## Flagged Code

```go
d.SetId(*read.ID)
```

## Passing Code

```go
subscriptionId := meta.(*clients.Client).Account.SubscriptionId
id := parse.NewServerID(subscriptionId, resourceGroup, name)
d.SetId(id.ID())
```

In Read, Update, and Delete, parse the ID back out with the generated parser:

```go
id, err := parse.ServerID(d.Id())
if err != nil {
	return err
}
```

## Ignoring Reports

Put `//azignore:AZR001 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
d.SetId(*read.ID) //azignore:AZR001 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
