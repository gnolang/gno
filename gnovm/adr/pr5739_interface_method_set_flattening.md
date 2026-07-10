# Interface method-set flattening and cross-package unexported-method identity

## Context

GnoVM derives an anonymous interface's `TypeID` from its method list. Before
this change, an *embedded* interface was kept as a single named entry in
`InterfaceType.Methods` (e.g. `interface{ Stringer }` stored a method-list
entry named `Stringer` whose type was the embedded interface). That made the
embedded-interface *name* — including an alias spelling — part of the
`TypeID`, so GnoVM diverged from Go:

- `interface{ Stringer }` was not identical to `interface{ Str() string }`.
- `interface{ SAlias }` (where `type SAlias = Stringer`) was not identical to
  `interface{ Stringer }`.

**Precise behavior on master (verified):** master named the embed entry from the
*resolved* type — `fillEmbeddedName` does `case *DeclaredType: ft.Name = ct.Name`
— and never flattened. So the two divergences are not symmetric:

| comparison | master | Go |
|---|---|---|
| `interface{ SAlias }` vs `interface{ Stringer }` | **identical** (alias resolves to the same `Stringer` entry name) | identical |
| `interface{ Stringer }` vs `interface{ Str() string }` | **distinct** (entry named `Stringer` vs method `Str`) | identical |

So on master the alias case was already correct; only embed-vs-explicit
diverged. No naming policy — resolved or spelled — can fix embed-vs-explicit,
because the embed is still one named entry rather than its methods. Only
flattening removes the embed name from identity.

Go computes interface identity from the **flattened method set**; embedding
contributes methods, not a name. PR #5739 therefore flattens embedded
interfaces into their constituent methods at type construction
(`flattenInterfaceMethods`, called from `doOpInterfaceType` and
`staticTypeFromAST`), so the embed/alias spelling no longer leaks into the
`TypeID`.

## The cross-package unexported-method problem

Flattening exposed a latent representational gap. An unexported interface
method's identity in Go is `(pkgpath, name)` — `p.sec` and `q.sec` are
distinct methods, and a type outside `p` cannot satisfy an interface
containing `p.sec` (the "sealed interface" pattern). GnoVM encodes that
package qualification **once**, on the containing `InterfaceType.PkgPath`,
used by `FieldTypeList.TypeIDForPackage` and by interface-satisfaction gating
in `VerifyImplementedBy` / `FindEmbeddedFieldType`.

When flattening hoists an unexported method out of an interface defined in
package `P` into an anonymous interface in package `Q`, that single-pkgpath
encoding re-qualifies the method to `Q`. A `/challenge` pass reproduced two
failures (both regressions vs the baseline before this change, confirmed against a
real-Go oracle):

1. **Identity over-collapse.** `interface{ p.Sec }` (with unexported `p.sec`)
   became identical to `interface{ sec() int }` declared in `q` — Go treats
   them as distinct.
2. **Satisfaction bypass (security-relevant).** A type declared in `q` with a
   `sec()` method was accepted as satisfying `p.Sec`, bypassing the sealed-
   interface mechanism. Go correctly rejects it.

Over-collapse is the dangerous direction (type confusion / access-control
bypass), and it is consensus-relevant in a VM.

## Decision

Make flattening fully Go-faithful by recording each method's **origin
package** alongside the method, instead of relying on the enclosing
interface's single `PkgPath`:

- Add `FieldType.PkgPath` — the defining package of an unexported method
  (empty for exported methods and for legacy/non-flattened entries, which
  fall back to the enclosing interface's `PkgPath`).
- `flattenInterfaceMethods` stamps the origin package on each hoisted (and
  directly-declared) unexported method, and deduplicates on `(pkgpath, name)`
  so that two same-named unexported methods from different packages coexist.
- `FieldTypeList.Less` and `FieldTypeList.TypeIDForPackage` key/qualify on the
  per-method `PkgPath` when set.
- `VerifyImplementedBy` gates unexported access against the method's origin
  package, not the enclosing interface's package.

This is fully faithful to Go, including the pathological case of two distinct
same-package sealed interfaces with identical method sets.

### Cost

`FieldType` is amino-serialized (`gnolang.proto` / `pb3_gen.go`), so adding
`PkgPath` changes the persisted shape of the type representation — a
serialization-format change requiring a coordinated chain upgrade, on top of
the `TypeID` changes flattening already implies. The new field is appended
(proto field 5) and defaults empty, so legacy-decoded entries behave exactly
as before.

## Alternatives considered

### Alternative A — do not hoist unsafe embeds (rejected, recorded for trace)

Flatten only when identity is preserved: hoist exported methods and
unexported methods whose origin package equals the enclosing interface; for an
embed carrying cross-package unexported methods, keep it as a sub-interface
entry and reuse the existing recursive satisfaction path.

- **Pro:** sound (no over-collapse, no bypass), small diff, **no serialization
  change** — `FieldType` shape is untouched.
- **Con:** conservative — it *over-distinguishes* in one case: two distinct
  sealed interfaces *in the same package* with identical method sets, embedded
  from a *third* package, are treated as distinct where Go treats them as
  identical. Over-distinction is the safe direction (a legitimate equality is
  missed; satisfaction can never be forged), deterministic across nodes, and
  the triggering case is exotic.

This was rejected only because we chose full Go-fidelity over avoiding the
serialization change. If the serialization/migration cost is later judged not
worth the pathological case, Alternative A is the drop-in fallback: it closes
the same security hole with no persisted-shape change. This ADR exists so that
trade-off can be revisited.

### Alternative B — keep the embed-entry representation (rejected)

Name embedded interfaces from the resolved type (the minimal fix that shipped
first on this branch). Sound, but only equalizes alias-vs-target; leaves
`interface{ Stringer } != interface{ Str() string }` diverging from Go.

## Consequences

- Interface identity matches Go across embedding, aliasing, multi-level,
  diamond, order, mixed embed+direct, and cross-package (exported and
  unexported) method sets.
- Consensus-breaking: `TypeID`s of anonymous interfaces that embed other
  interfaces change, and the `FieldType` serialization gains a field. Must
  land with a chain upgrade (same release class as #5737: coordinated
  upgrade / fresh genesis).
- Runtime **assumes flattened** interfaces. The legacy embedded-interface
  branches in `FindEmbeddedFieldType` / `VerifyImplementedBy` are dropped; an
  `InterfaceKind` entry in `Methods` (only possible by decoding
  bytes persisted before this change) is a **hard error** (`panicUnflattened`),
  enforced where the concern lives: ungated at the **decode boundary**
  (`fillType`, reached from both type-entry decode and object loads — store
  bytes are external input, so this check stays in production), and under
  `-tags debugAssert` at the interior sites (method resolution, satisfaction,
  `TypeID`), which may assume the invariant on a validated store.

## Follow-up: same-spelled unexported members must not shadow each other

Review (davd-gzl, omarsy) showed that once two same-named unexported methods
from different packages legally coexist (which flattening enables), every
name-keyed lookup with first-match-wins semantics becomes order-dependent and
wrong: `interface{ sec() int; ifaceext.Sec }` failed selector resolution, its
own `I`-to-`I` assignment ("main.I does not implement main.I"), and concrete
satisfaction (a type with its own `sec` plus a promoted `ifaceext.sec`).

Two layers were fixed:

- **Static lookups** (`StructType`/`InterfaceType`/`DeclaredType`
  `FindEmbeddedFieldType`): a name match whose unexported gate fails is a
  *distinct identity*, not an error — skip it and keep scanning (struct
  fields fall through to embedded fields; declared-type direct methods fall
  through to the base). `accessError` is reported only when the search ends
  with no accessible match; the `trail != nil ⇒ accessError == false`
  contract is clamped once in the `findEmbeddedFieldType` dispatcher (callers
  check `accessError` before `trail`).
- **Runtime dispatch** (`getPointerToFromTV`, `VPInterface` case): the
  resolver used the *dynamic type's* package as `callerPath`, which picks the
  wrong member when identities collide (e.g. an interface method call
  resolving to a same-named unexported *field* of the receiver). Selector ops
  now pass the executing package (`m.Package.PkgPath`) — statically gated to
  equal the method's origin for unexported selectors. The exported
  `GetPointerToFromTV` keeps the dynamic-type fallback (debugger,
  collision-free by construction there).

Pinned by `iface_embed_sel_order.gno` (order-independent selector),
`iface_embed_same_name.gno` (self-assignment), and
`iface_embed_field_shadow.gno` (field vs promoted method, static + runtime),
each verified against real Go first.

Diagnostics qualify stamped methods by origin package, so such an interface no
longer prints as `interface {sec func() int; sec func() int}` with two
indistinguishable entries. `FieldTypeList.string` qualifies the method list, and
`FieldType.diagName` qualifies the method named by every `VerifyImplementedBy`
error and by the duplicate-method panic in `flattenInterfaceMethods`. A message
therefore never names a bare `sec` beside an interface that prints two distinct
ones. Directly-declared unexported methods are unstamped and print unchanged,
which is why `diagName` qualifies only on a stamp while `idName` qualifies every
unexported name.

## Rollout: state persisted before this change is unsupported (decided)

Every `InterfaceType` is born one of three ways: **preprocess (AST)**,
**runtime (`doOpInterfaceType`)**, or **decode**. The first two always flatten;
decode faithfully reproduces stored bytes. So an unflattened interface can
reach runtime in exactly one situation: decoding bytes persisted **before
this change**.

Tolerating those bytes (the recursion branches this PR originally kept) was a
half-measure: the type's *identity* already moved with this change, so
resolution/satisfaction would "work" against a type whose equality, type-keyed
store entries, and hashes are silently split from post-change construction.
Silent-wrong loses to loud-fail on a chain, so the branches are **dropped**
and a decoded unflattened interface **panics** with an actionable message.

A network carrying state persisted before this change has two supported paths, both of which
make unflattened bytes impossible (the panic is then a dead invariant check):

- **Fresh genesis / tx replay** (how gno.land testnets regenesis): all state
  is reconstructed through the new VM, flattened by construction. The only
  pre-fork audit is source acceptance — historical txs must still preprocess
  (this PR tightens one case: cross-package unexported-method selection).
- **Re-flatten migration** (only if in-place state must survive): walk the
  type table, flatten every `InterfaceType` (stamping origin `PkgPath`),
  remap old→new `TypeID`s in all object refs, merge newly-colliding entries,
  and verify no `InterfaceKind` entry remains.

Conflict handling (same-name, different-signature embedded methods) needs no
follow-up: `flattenInterfaceMethods` `panic`s, but during preprocessing that is
wrapped into a positioned `*PreprocessError` (the same idiom as other type
errors, e.g. "struct has no field"), so it surfaces as a matchable `// Error:`.
go/types also rejects it as a `// TypeCheckError:`. Both are pinned by
`iface_embed_conflict.gno`, matching Go's "duplicate method" compile error.
