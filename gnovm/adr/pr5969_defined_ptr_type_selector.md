# Defined pointer types: empty method set, no pointer embedding

Issue: https://github.com/gnolang/gno/issues/5957

## Context

Selecting a method through a defined (named) pointer type such as
`type D1 *D2` made the preprocessor panic `should not happen` instead
of reporting a normal "missing field" error:

```go
type D2 struct{}
func (D2) Foo() string { return "m" }
type D1 *D2

var x D1
_ = x.Foo // GnoVM: panic "should not happen"; Go: x.Foo undefined
```

`DeclaredType.FindEmbeddedFieldType` searches the defined type's base;
for base `*D2`, `PointerType.FindEmbeddedFieldType` finds `Foo` and
returns a method trail (`VPDerefValMethod` etc.). The `trail[0].Type`
switch in `DeclaredType.FindEmbeddedFieldType` only handles
`VPInterface`/`VPField`/`VPDerefField`, so any method trail fell into
the panicking `default:` branch.

Per the Go spec, a defined type whose underlying type is a pointer has
an **empty method set**; only fields promote through it (the selector
shorthand `x.f` for `(*x).f` applies to selectors "denoting a field
(but not a method)"). So the correct verdict for a method match through
a pointer base is "not found".

A related manifestation: Go rejects embedding a type whose underlying
type is a pointer (`type S struct{ D1 }` → "embedded field type cannot
be a pointer") at declaration, while GnoVM accepted the declaration and
then panicked on selection.

Review surfaced two more panicking shapes with the same root cause,
both hitting the `default:` branch in `PointerType.FindEmbeddedFieldType`
instead:

```go
type C struct{ F int }
type B *C
type A *B
var x A
_ = x.F // Go: x.F undefined; the (*x).f shorthand applies once, and
        // (*x) is again a defined pointer type with no field F.

type BI interface{ M() string }
type AI *BI
var y AI
_ = y.M // Go: y.M undefined (type AI is pointer to interface)
```

Here the pointer's element is a defined type whose own base is a
pointer (returns a `VPDerefField`-headed trail: a second indirection)
or an interface (returns a `VPInterface`-headed trail), neither of
which the switch handled.

A second review pass surfaced two more:

```go
type S struct{ *I }  // I an interface. Go: "embedded field type
                     // cannot be a pointer to an interface";
                     // GnoVM: accepted, then panic on s.M.

var p *D1            // type D1 *D2
_ = p.A              // Go: p.A undefined (*D1 needs a second deref);
                     // GnoVM: panic — the root canonEmbeddedType strip
                     // turned *D1 into D1, entering the defined-
                     // pointer crossing one level too late.
```

## Decision

Three changes in `gnovm/pkg/gnolang/types.go`. (The fix was originally
written against the pre-BFS recursive lookup — per-type
`FindEmbeddedFieldType` methods — and re-ported onto the BFS lookup
introduced by #5721 when merging master.)

1. **`resolveEmbedNode` / `lookupShallowestEmbedded`**: the spine walk
   gains a `fieldsOnly` flag. A `*PointerType` node inside the spine is
   always a defined type's base (root and embedded-field pointers are
   stripped by `canonEmbeddedType` before the walk), so crossing one
   switches to fields-only: method lookups on subsequent declared types
   and interface bases no longer count (a defined type whose underlying
   type is a pointer has an empty method set), and a second crossing
   exposes nothing (the `(*x).f` shorthand applies once — this is the
   `type B *C; type A *B` case). The flag propagates through the BFS
   (`embedLookupEntry.fieldsOnly`, threaded alongside the per-level
   `structs` expansion state) so methods of types embedded *inside* the
   pointed-to struct — including via embedded interfaces — don't
   promote either, while their fields still do. With phase 1 filtering
   these out, `buildEmbeddedTrail` no longer receives winners it cannot
   represent, which is what previously tripped its
   `should not happen` panics.

2. **`fillEmbeddedName`**: reject embedded fields that are still of
   pointer kind after one `unwrapPointerType` — i.e. a defined type of
   pointer kind (`D1`) or a pointer whose element is of pointer kind
   (`*D1`) — with Go's message "embedded field type cannot be a
   pointer"; and reject a pointer whose element is of interface kind
   (`*I`) with "embedded field type cannot be a pointer to an
   interface". A literal `*T` unwraps to `T` and stays legal, and an
   alias of a pointer type resolves to the `*PointerType` spelling and
   stays legal, both matching Go. This runs on both struct construction
   paths (`doOpStructType` and `buildFieldTypesAST`).

3. **`canonEmbeddedType`**: `*T` where `T` is a defined type of pointer
   kind returns nil (exposes nothing) — any selection through `*D1`
   would need a second deref, which Go's selector shorthand never
   performs. This covers the lookup root (`var p *D1; p.A`); spine
   crossings were already covered by (1)'s second-crossing rule.

Fields still promote through a defined pointer type (`x.A` for
`(*x).A`), unchanged and covered by a regression test, including
fields of types embedded inside the pointed-to struct.

A side effect of (1): a defined pointer type no longer (incorrectly)
satisfies interfaces via its base's methods — `VerifyImplementedBy`
uses the same lookup, and Go agrees (`D1` implements nothing).

## Alternatives considered

- **Handle the offending trails in `buildEmbeddedTrail` and promote
  them** (make GnoVM more permissive than Go): rejected — Gno tracks
  Go semantics, and go/types type-checking at `AddPackage` already
  rejects such programs, so the VM must agree with the checker rather
  than diverge.
- **Only patch the panic sites in `buildEmbeddedTrail` to return
  not-found**: insufficient — phase 1 (`lookupShallowestEmbedded`)
  would still report such names as found (affecting
  `VerifyImplementedBy` and shallowest-depth/ambiguity resolution),
  and field-headed trails ending in a method don't reach the panicking
  branches at all. The filter belongs in the phase that defines the
  method set.
- **Declaration check in `validateStructFields`**: `fillEmbeddedName`
  was chosen instead because it runs exactly once per embedded field on
  both construction paths and has the source `FieldType` at hand.

## Consequences

- `x.Foo` through a defined pointer type reports
  `missing field Foo in main.D1` (consistent with the existing
  declared-type message, cf. `method39.gno`) instead of panicking.
- `type S struct{ D1 }` / `struct{ *D1 }` fail at declaration with
  Go's message.
- Programs that (unintentionally) relied on method promotion through
  defined pointer types now get errors; such programs already failed
  go/types type-checking on-chain, so nothing deployable breaks.
- Tests: `method47–50.gno` (renumbered from 40–43 after merging
  master, which added its own `method40–46.gno`), `struct64.gno`,
  `struct64b.gno`, `struct64c.gno` (alias-of-pointer embedding stays
  legal), `struct65.gno` (pointer-to-interface embedding rejected),
  `ptr12.gno` (nested defined pointers, from review), `ptr13.gno`
  (pointer to defined interface type), `ptr14.gno` (`*D1` root exposes
  nothing), `ptr15.gno` (defined pointer type does not satisfy an
  interface via its base's methods).
