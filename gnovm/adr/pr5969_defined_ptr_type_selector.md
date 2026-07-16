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

## Decision

Three changes in `gnovm/pkg/gnolang/types.go`:

1. **`DeclaredType.FindEmbeddedFieldType`**: after the base search
   returns a trail, if the base is a `*PointerType` and the trail's
   *last* element is not a field path (`VPField`/`VPDerefField`),
   return not-found. The last element is checked (not `trail[0]`)
   because a method promoted via an embedded field inside the
   pointed-to struct yields a field-headed trail
   (`[VPDerefField, VPValMethod]`), and interface-promoted methods
   yield `[VPDerefField, VPInterface]` — all of which Go also rejects.
   `rcvr != nil` was rejected as the discriminator because interface
   method matches return a nil receiver.

2. **`PointerType.FindEmbeddedFieldType`**: when the element search
   returns a `VPDerefField`-headed trail (the element is a defined type
   that resolved the selector through its own pointer base — a second
   indirection) or a `VPInterface`-headed trail (pointer to interface),
   return not-found. Go promotes neither through a pointer. The
   `default:` panic stays for genuinely impossible trail heads.

3. **`fillEmbeddedName`**: reject embedded fields that are still of
   pointer kind after one `unwrapPointerType` — i.e. a defined type of
   pointer kind (`D1`) or a pointer whose element is of pointer kind
   (`*D1`) — with Go's message "embedded field type cannot be a
   pointer". A literal `*T` unwraps to `T` and stays legal, and an
   alias of a pointer type resolves to the `*PointerType` spelling and
   stays legal, both matching Go. This runs on both struct construction
   paths (`doOpStructType` and `buildFieldTypesAST`).

Fields still promote through a defined pointer type (`x.A` for
`(*x).A`), unchanged and covered by a regression test.

A side effect of (1): a defined pointer type no longer (incorrectly)
satisfies interfaces via its base's methods — `VerifyImplementedBy`
uses the same lookup, and Go agrees (`D1` implements nothing).

## Alternatives considered

- **Handle the method trails in the `trail[0]` switch and promote
  them** (make GnoVM more permissive than Go): rejected — Gno tracks
  Go semantics, and go/types type-checking at `AddPackage` already
  rejects such programs, so the VM must agree with the checker rather
  than diverge.
- **Only fix the panic site (return not-found from the `default:`
  branch)**: insufficient — `[VPDerefField, VPValMethod]`-shaped trails
  (promoted via embedded field) don't reach the `default:` branch and
  would still incorrectly resolve through the `VPField/VPDerefField`
  case.
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
- Tests: `method40–43.gno`, `struct64.gno`, `struct64b.gno`,
  `ptr12.gno` (nested defined pointers, from review), `ptr13.gno`
  (pointer to defined interface type).
