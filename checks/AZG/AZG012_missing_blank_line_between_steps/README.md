# AZG012 - separate the steps of a longer function body with blank lines

AZG012 reports an `if`, `for`, `switch` or `select` in a longer body that is run straight into the statement after it, with no blank line between.

A body of any length is read as a sequence of steps. A block ends one, and the statement after it starts the next: get something, check it, use it. A blank line between them is what shows the boundary. Without it a twenty-line loop body reads as one undifferentiated run.

## Flagged Code

```go
for _, decl := range file.Decls {
	fd, ok := decl.(*ast.FuncDecl)
	if !ok || fd.Body == nil {
		continue
	}
	fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
	if !ok || runs[fn] != nil {
		continue
	}
	rest := strings.TrimPrefix(fn.Name(), "TestAcc")
	if rest != "" && unicode.IsLower(rune(rest[0])) {
		pass.Reportf(fd.Name.Pos(), "...")
		continue
	}
	pass.Reportf(fd.Name.Pos(), "...")
}
```

## Passing Code

```go
for _, decl := range file.Decls {
	fd, ok := decl.(*ast.FuncDecl)
	if !ok || fd.Body == nil {
		continue
	}

	fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
	if !ok || runs[fn] != nil {
		continue
	}

	rest := strings.TrimPrefix(fn.Name(), "TestAcc")
	if rest != "" && unicode.IsLower(rune(rest[0])) {
		pass.Reportf(fd.Name.Pos(), "...")
		continue
	}

	pass.Reportf(fd.Name.Pos(), "...")
}
```

## What counts

Only a statement with a body ends a step: `if`, `for`, `switch`, `select`, or a bare block. A literal assignment that happens to close with a brace does not. Closures are checked like any other body, and so are the bodies of `case` clauses.

Not reported:

- a short body, which is one group: five statements or fewer, or ten lines or fewer, see [Options](#options)
- two one-line `if`s in a row, which read as a single set of guards
- a block followed by the next `case` or `default`
- anywhere a blank line is already there

A comment on the line after the block belongs to the next step, so the report, and the blank line, go before the comment.

## The fix

`-fix` inserts the blank line.

## Options

| Option | Default | Effect |
|---|---|---|
| `statements` | 5 | bodies with this many statements or fewer are one group |
| `lines` | 10 | bodies spanning this many lines or fewer are one group |

A body is left alone when either holds. Set with `-AZG012.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Put `//azignore:AZG012 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
report(rest) //azignore:AZG012 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
