# AZG008 - pointer dereferences must have a nil check

AZG008 reports `*x` where nothing proves `x` is non-nil. SDK response fields are pointers because the API can leave them out, so an unchecked `*props.Status` is a panic waiting for the right response.

The check reads the code around the dereference and accepts it when any of these is true:

- **It sits inside a nil check.** `if x != nil { *x }`, `x != nil && *x`, `x == nil || *x`, or the else branch of `if x == nil`.
- **An earlier nil check bailed out.** `if x == nil { return }` (or `continue`, `break`, `panic`, `os.Exit`, `t.Fatal`) before the dereference, even when that `if` has an else.
- **The value was just created.** `x := &T{}`, `new(T)`, `pointer.To(v)`, or `if x == nil { x = &T{} }` right before.
- **It came from a function that never returns nil.** The analyzer reads each function's return statements, decides which pointer results are never nil, and shares that across packages. So `helpers.ExpandStringSlice(...)` and every `expandFoo()` that ends in `return &out` count. A function with any `return nil` path does not.
- **It came from a call whose error or ok was checked.** `x, err := f(); if err != nil { return }`, the same with `ok`, and also `if x, err := f(); err != nil { return }`, `if err == nil { *x }`, and the else of `if err != nil`. If `err` is overwritten before the check, that no longer counts. A type assertion's `ok` never counts, since it says nothing about nil.
- **It was set in the struct literal.** `x := T{F: pointer.To(v)}` proves `*x.F`, nested `&U{...}` literals included.
- **It is a composite ID part.** `id.First` / `id.Second` from `commonids.NewCompositeResourceID` or `ParseCompositeResourceID`, when the matching argument was non-nil.
- **A `pointer.From` comparison proved it.** `if pointer.From(x) != "" { *x }` or `len(pointer.From(x)) > 0`.
- **An alias was checked instead.** `if v := x.F; v != nil { *x.F }`, `v := x.F` earlier in the block, or `m := r.Model` in an outer if's init all count for each other. An alias stops counting once either side is reassigned.

The check is strict about what a guard covers. `if m.Properties != nil` does not cover `*m.Properties.Name`. Reassigning the variable from an unknown source between the check and the dereference cancels the check. Bare pointer parameters are trusted by default, since the nil check belongs at the call site; turn on `include-parameters` to report them too. Fields reached through a parameter, like `*props.Status`, are always checked.

## The fix

Reports come with a suggested fix: `*x` becomes `pointer.From(x)`, and `string(*x)` on an enum pointer becomes `pointer.FromEnum(x)`. The import is added if missing. Both turn a panic into a zero value, which is what you want in flatten and read code. Review it elsewhere.

**Values being sent to Azure are a separate, opt-in pass.** `pointer.From(existing.Model)` passed to a PUT would send an empty body instead of panicking, so those dereferences never get a fix and are only reported with `requestbody: true`. Run the default pass first and apply its fixes, then run with `requestbody` on and add the missing nil checks by hand. The analyzer works out which functions send a request body by reading their code, not their names: a function "writes" if it uses `http.MethodPut`/`Patch`/`Post` or a `"PUT"`/`"PATCH"`/`"POST"` string, or calls something that does; a parameter is "sent" if it reaches `json.Marshal` or `xml.Marshal`, including inside a closure or through another call. `CreateOrUpdate`, its `ThenPoll` wrappers, and autorest's `WithJSON` all qualify on their own. The dereference is caught whether it is passed directly, copied into a local, put in a struct field or literal, or passed by address. The report says why there is no fix.

`fix-with: none` turns off fixes entirely. Dereferences that must stay addressable, `*x = v`, `(*x).F = v`, `&*x`, `(*x)++`, or calling a pointer-receiver method on `*x`, cannot take `pointer.From` and are AZG009's to report, as are implicit dereferences like `m.Properties.Name` with a nil `Properties`.

## Flagged Code

```go
d.Set("status", string(*props.Status))
```

```go
if model.Properties != nil {
	d.Set("name", *model.Properties.Name) // check covers Properties, not Name
}
```

```go
// reported only with requestbody: true, and without a fix: the value is a PUT body
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
out := expandThing(d) // expandThing always returns &result
d.Set("count", *out.Count)
```

## Options

| Option | Default | Effect |
|---|---|---|
| `include-parameters` | false | also report dereferences of bare pointer parameters |
| `tests` | true | check `_test.go` files |
| `requestbody` | false | also report dereferences whose value is sent as a PUT/PATCH/POST body; these never get a fix |
| `fix-with` | `pointer.From` | suggested-fix form: `pointer.From` or `none` |

Set via `-AZG008.<option>` on the CLI or a rule-name key in the plugin's golangci settings.

## Ignoring Reports

Sometimes the dereference is deliberate: an invariant guarantees the field, or a panic is the right failure. When run via golangci-lint, a `//nolint:azproviderlint` comment at the end of the line, or on the line before, ignores the report.

To ignore only this check on a line, use `//azignore:AZG008 - <reason>` in the same positions:

```go
d.Set("status", string(*props.Status)) //azignore:AZG008 - <reason>
```
