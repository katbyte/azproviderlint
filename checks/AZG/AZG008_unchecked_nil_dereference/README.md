# AZG008 - pointer dereferences must have a reachable nil guard

The AZG008 analyzer reports explicit pointer dereferences in value context — `*props.Status`, `string(*props.Status)` — where nothing establishes the pointer is non-nil. SDK response fields are pointers precisely because the API may omit them, so an unguarded dereference is a latent panic. No generic linter catches this: staticcheck's SA5011 needs a contradictory nil check to exist, and nilness-style analyses only flag provably-nil values. Everything AZG008 reports is auto-fixable with `pointer.From`; dereferences that need a hand-written guard instead — implicit selector dereferences, write targets — are the sister check [AZG009](../AZG009_nil_dereference_requires_guard)'s.

A dereference counts as guarded when any of these proves the exact dereferenced chain non-nil:

- an enclosing `if`/`for`/`case` condition containing `x != nil`, including via `&&` (`x != nil && *x > 0`) and short-circuit `||` (`x == nil || *x == 0`), or the else branch of a pure-`||` nil condition
- an earlier statement in an enclosing block reading `if x == nil { ... }` whose body ends in `return`, `break`/`continue`, `panic`, `os.Exit`, or a `Fatal*` call
- every assignment to the variable being a provably non-nil source: `&T{...}`, `new(T)`, `pointer.To(...)`, a `flag` package constructor, an immediately-invoked func literal whose every return is one of these, or a call to a function proven never to return nil in that position (every return statement supplies a non-nil source or a proven local, propagated across packages as analysis facts, so `helpers.ExpandStringSlice` and the `expand*` helpers returning `&x` count; a function with any `return nil` path does not) — including a field set that way inside the composite literal that built the struct (`x := T{F: pointer.To(v)}` proves `x.F`, nested `&U{...}` literals included)
- an earlier `if x == nil { x = &T{} }` default-init
- the variable coming from `x, err := f(...)` or `x, ok := f(...)` with an `if err != nil` / `if !ok` exit (as a following statement or with the call as the if's init), the body of `if err == nil` / `if ok`, or the else branch of `if err != nil` / `if !ok` — the conventional Go contracts under which the other results are valid; a companion reassigned before the check no longer counts, and a type assertion's or map lookup's ok never does
- `id.First` / `id.Second` of a `commonids.NewCompositeResourceID` / `ParseCompositeResourceID` result, when the matching argument is itself non-nil
- `pointer.From(x) != ""` or `len(pointer.From(x)) > 0` in the enclosing condition
- the variable aliasing a chain that satisfies any of the above (`payload := existing.Model` after an `existing.Model == nil` early return)
- a nil check on an alias of the chain (`if v := x.F; v != nil { *x.F }`, `v := x.F` earlier in the block, or `m := r.Model` in an enclosing if's init); an alias goes stale once either side is reassigned

Guards are matched by path equality on the whole chain — `if m.Properties != nil` does not cover `*m.Properties.Name`. A reassignment from an unknown source between guard and dereference invalidates the guard, including guards in enclosing scopes (`if x != nil { x = f(); *x }` is reported). Dereferences of bare pointer parameters are trusted by default (the nil check belongs at the call sites); field dereferences through a parameter are always in scope. Chains containing calls or index expressions are out of scope. `_test.go` files are checked by default; set `tests` to false to skip them.

The report carries a suggested fix: `string(*x)` of a string-kinded enum pointer becomes `pointer.FromEnum(x)`, other dereferences become `pointer.From(x)`, and the pointer import is added when missing. Both change behaviour from panic to zero value — the desired semantics in flatten/read paths, but review before applying elsewhere. `fix-with: none` reports without fixes. A dereference whose value becomes the body of a PUT/PATCH/POST — passed to a function that serialises that parameter while performing a write, directly or via a local that flows into one — is always reported without a fix. That verdict is derived, not matched by name: a function writes when it references a `net/http` write method (or a `"PUT"`/`"PATCH"`/`"POST"` literal) or calls one that does; a parameter is serialised when it reaches `encoding/json`/`encoding/xml` marshalling, including inside a closure, or a callee that serialises it. Facts carry both across packages, so go-azure-sdk's `req.Marshal(input)`, autorest's `WithJSON(v)`, and every `*ThenPoll` wrapper qualify on their own behaviour. `pointer.From(existing.Model)` as a PUT body would send an empty request instead of panicking, so those sites need a hand-written nil check. Dereferences whose pointee must stay addressable (`*x = v`, `(*x).F = v`, `&*x`, `(*x)++`, a pointer-receiver method `(*x).M()`) are reported by AZG009 instead.

## Flagged Code

```go
d.Set("status", string(*props.Status))
```

```go
if model.Properties != nil {
	d.Set("name", *model.Properties.Name) // guard covers Properties, not Name
}
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

## Options

| Option | Default | Effect |
|---|---|---|
| `include-parameters` | false | also report dereferences of bare pointer parameters |
| `tests` | true | check `_test.go` files |
| `fix-with` | `pointer.From` | suggested-fix form: `pointer.From` or `none` |

Set via `-AZG008.<option>` on the CLI or a rule-name key in the plugin's golangci settings.

## Ignoring Reports

A dereference can be deliberate — an invariant guarantees the field, or a panic is the correct failure mode. When run via golangci-lint, reports can be ignored with a `//nolint:azproviderlint` Go code comment at the end of the offending line or on the line immediately preceding it.

To ignore only this check on a line — leaving any other azproviderlint checks active — use a `//azignore:AZG008 - <reason>` comment instead, in the same positions:

```go
d.Set("status", string(*props.Status)) //azignore:AZG008 - <reason>
```
