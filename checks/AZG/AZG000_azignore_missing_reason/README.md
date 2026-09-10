# AZG000 - azignore directives must give a reason

AZG000 reports `//azignore:` comments that do not say why.

A suppression with no reason tells the next reader nothing. Is this a deliberate exception, a false positive, or something someone meant to come back to? Write it down:

```go
//azignore:<Rule>[,<Rule>...] - <reason>
```

The reason is any text after the rule list. The `-` is optional, so `//azignore:AZR001 deliberate subset` works too. Rule names are always letters then digits, so the reason cannot be mistaken for a rule even if it starts with a dash or contains commas.

A directive without a reason still suppresses its target check. This report is the only thing you see, so one problem gives one message rather than two.

## Flagged Code

```go
d.SetId(*read.ID) //azignore:AZR001

//azignore:AZG001,AZR003
err := client.Delete(ctx, id)

d.SetId(*read.ID) //azignore:AZR001 -
```

## Passing Code

```go
d.SetId(*read.ID) //azignore:AZR001 - legacy resource, ID formatter tracked in #1234

//azignore:AZG001,AZR003 combined form obscures the retry loop here
err := client.Delete(ctx, id)
```

## Ignoring Reports

`//azignore:AZG000` is deliberately not honoured, otherwise a bare directive could hide the report about itself by adding `AZG000` to its own list.

Under golangci-lint, `//nolint:azproviderlint` on the line still works. If bare directives are fine by your project's policy, turn the check off in the plugin settings:

```yaml
linters:
  settings:
    custom:
      azproviderlint:
        settings:
          disable: [AZG000]
```
