# AZG013 - functions must not always call each other

AZG013 reports two or more functions that call each other on every path, so none of them can return.

It compiles, and calling any of them runs until the program is out of stack. It usually comes from copying a helper onto another type: its `r.template(data)` now means the new type's `template`, and that one calls the helper.

## Flagged Code

```go
func (r MsSqlExecuteJobAction) template(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
`, MsSqlExecuteJobAction{}.msSqlJobStepTestBasic(data))
}

func (r MsSqlExecuteJobAction) msSqlJobStepTestBasic(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s
`, r.template(data))
}
```

```text
MsSqlExecuteJobAction.template always calls MsSqlExecuteJobAction.msSqlJobStepTestBasic, which always calls MsSqlExecuteJobAction.template again, so neither returns
```

## Passing Code

```go
// a way out before the call
func countDown(n int) int {
	if n == 0 {
		return 0
	}

	return countBack(n)
}

func countBack(n int) int {
	return countDown(n - 1)
}
```

## What counts

A call is always made when it is reached whatever the function is given: it sits in the function's own statements before the first `if`, `switch`, `for` or `return`, or in what that statement asks or returns. In `if a() {`, `a` is always called. So is `b` in `return fmt.Sprint(b())`.

The loop is reported once, at the call in the function that comes first in the file.

Not reported:

- a call inside an `if`, `switch`, `for` or `select`, or after one
- the right side of `&&` and `||`
- a call in a function written in place, and `defer` and `go` calls
- anything after `panic`, or after a call to `Exit`, `Fatal`, `FailNow` or `Skip` in another package
- a call through an interface or a variable
- functions in different packages
- one function that always calls itself; staticcheck reports that as SA5007

## No fix

AZG013 only reports. One of the functions is calling the wrong thing: point it at what it was copied from, or copy that as well.

## Ignoring Reports

Put `//azignore:AZG013 - <reason>` at the end of the reported line, or on the line above it. The reason is required.

```go
	return other(v) //azignore:AZG013 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
