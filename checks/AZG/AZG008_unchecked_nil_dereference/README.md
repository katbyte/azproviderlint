# AZG008 - pointer dereferences must have a nil guard or use pointer.From

AZG008 reports `*x` where nothing nearby proves `x` is not nil.

Fields in SDK responses are pointers because Azure is allowed to leave them out. An unchecked `*props.Status` works until the day a response comes back without `status`, and then the provider panics. Either check for nil first, or use `pointer.From(x)`, which returns the zero value for a nil pointer.

## Flagged Code

```go
d.Set("status", string(*props.Status))
```

```go
if model.Properties != nil {
	d.Set("name", *model.Properties.Name) // the check covers Properties, not Name
}
```

```go
// reported, but with no fix offered: the value is a PUT body
client.CreateOrUpdate(ctx, id, *existing.Model)
```

## Passing Code

```go
d.Set("status", pointer.FromEnum(props.Status))
```

```go
if props.Status != nil {
	d.Set("status", string(*props.Status))
}
```

```go
if props.Count == nil {
	return 0
}
return *props.Count
```

```go
out := expandThing(d) // expandThing always ends in return &result
d.Set("count", *out.Count)
```

## What counts as a check

The rule reads the code around the dereference and accepts it when any of these holds:

- **It is inside a nil check.** `if x != nil { *x }`, `x != nil && *x`, `x == nil || *x`, or the else branch of `if x == nil`.
- **An earlier nil check left the function.** `if x == nil { return }` before the dereference, or `continue`, `break`, `panic`, `os.Exit`, `t.Fatal`.
- **The value was just created.** `x := &T{}`, `new(T)`, `pointer.To(v)`, or `if x == nil { x = &T{} }` before it.
- **It came from a function that never returns nil.** The rule reads every function's return statements, works out which pointer results are never nil, and shares that between packages. So `helpers.ExpandStringSlice(...)` and any `expandFoo()` that always ends in `return &out` count. One `return nil` anywhere and it does not.
- **It came from a call whose error or ok was checked.** `x, err := f(); if err != nil { return }`, the same with `ok`, `if x, err := f(); err != nil { return }`, `if err == nil { *x }`, or the else of `if err != nil`. If `err` is overwritten before the check, it no longer counts. A type assertion's `ok` never counts, since it says nothing about nil.
- **It was set in the struct literal.** `x := T{F: pointer.To(v)}` proves `*x.F`, nested `&U{...}` literals included.
- **It is half of a composite ID.** `id.First` / `id.Second` from `commonids.NewCompositeResourceID` or `ParseCompositeResourceID`, when the matching argument was not nil.
- **A `pointer.From` comparison proved it.** `if pointer.From(x) != "" { *x }` or `len(pointer.From(x)) > 0`.
- **An alias was checked instead.** `if v := x.F; v != nil { *x.F }`, `v := x.F` earlier in the block, or `m := r.Model` in an outer if's init all count for each other, until either side is reassigned.

A check covers exactly what it names. `if m.Properties != nil` does not cover `*m.Properties.Name`. Reassigning the variable from an unknown source between the check and the dereference cancels the check.

Bare pointer parameters (`func f(x *T) { *x }`) are trusted by default, since the nil check belongs at the call site. Turn on `include-parameters` to report them. Fields reached through a parameter, like `*props.Status`, are always checked.

## The fix

`-fix` turns `*x` into `pointer.From(x)`, and `string(*x)` on an enum pointer into `pointer.FromEnum(x)`, adding the import if needed. Both turn a panic into a zero value. In read and flatten code that is what you want. Elsewhere, read the fix before applying it.

**No fix is offered when the value is being sent to Azure.** `pointer.From(existing.Model)` passed to a PUT would send an empty body instead of panicking, which is worse. Those dereferences are reported with a message explaining why there is no fix, and need a hand-written nil check. Set `requestbody: false` to leave them out of a run you plan to apply with `-fix`.

The rule decides which functions send a request body by reading what they do, not what they are called. A function sends a write if it uses `http.MethodPut`, `MethodPatch`, or `MethodPost`, or a `"PUT"`, `"PATCH"`, or `"POST"` string, or calls something that does. A parameter is the body if it reaches `json.Marshal` or `xml.Marshal`, including inside a closure or through another call. `CreateOrUpdate`, its `ThenPoll` wrappers, and autorest's `WithJSON` all qualify on their own. The dereference is caught whether the value is passed directly, copied into a local, put in a struct field or literal, or passed by address.

`fix-with: none` turns fixes off entirely. Some dereferences cannot take `pointer.From` at all because the result must stay addressable: `*x = v`, `(*x).F = v`, `&*x`, `(*x)++`, or a pointer-receiver method on `*x`. Those, and implicit dereferences like `m.Properties.Name` with a nil `Properties`, are AZG009's to report.

## Options

| Option | Default | Effect |
|---|---|---|
| `include-parameters` | false | also report dereferences of bare pointer parameters |
| `tests` | true | check `_test.go` files |
| `requestbody` | true | report dereferences whose value is sent as a PUT/PATCH/POST body. These never get a fix |
| `fix-with` | `pointer.From` | what the fix suggests: `pointer.From` or `none` |

Set with `-AZG008.<option>` on the CLI or under the rule name in the golangci settings; see the [root README](../../../README.md#options).

## Ignoring Reports

Sometimes the dereference is deliberate: an invariant guarantees the field, or a panic is the right way to fail. Put `//azignore:AZG008 - <reason>` at the end of the line, or on the line above it. The reason is required.

```go
d.Set("status", string(*props.Status)) //azignore:AZG008 - <reason>
```

Under golangci-lint, `//nolint:azproviderlint` in the same place also works, but it silences every azproviderlint check on that line.
