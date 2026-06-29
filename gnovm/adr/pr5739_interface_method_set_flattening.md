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
failures (both regressions vs the pre-flattening baseline, confirmed against a
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

### Alternative B — keep the pre-flattening representation (rejected)

Name embedded interfaces from the resolved type (the minimal fix that shipped
first on this branch). Sound, but only equalizes alias-vs-target; leaves
`interface{ Stringer } != interface{ Str() string }` diverging from Go.

## Consequences

- Interface identity matches Go across embedding, aliasing, multi-level,
  diamond, order, mixed embed+direct, and cross-package (exported and
  unexported) method sets.
- The embedded-interface branches retained in `FindEmbeddedFieldType` /
  `VerifyImplementedBy` still serve interface types decoded from pre-change
  persisted state (their `Methods` are unflattened); they are otherwise
  unreachable for freshly-constructed types.
- Consensus-breaking: `TypeID`s of anonymous interfaces that embed other
  interfaces change, and the `FieldType` serialization gains a field. Must
  land with a chain upgrade.

## Follow-up: migrate persisted types, then assume "all flattened"

The embedded-interface branches in `FindEmbeddedFieldType` / `VerifyImplementedBy`
are retained only because **decode does not re-flatten** — an interface persisted
by pre-flattening code comes back with embedded-interface entries in `Methods`,
and those branches keep satisfaction correct for it. This is a half-measure:
such a type's *identity* is already inconsistent after this change (its `TypeID`
moved from the embed-name form to the flattened form), so any chain carrying old
state across this upgrade already needs a migration.

That migration is the natural place to **re-flatten every persisted interface
type** (rewrite its `Methods` into flattened, `PkgPath`-stamped form). Once it
guarantees that no unflattened `InterfaceType` can be decoded, the "all
flattened" invariant holds at runtime and we can:

1. Drop the embedded-interface recursion in `FindEmbeddedFieldType` /
   `VerifyImplementedBy` (or replace it with a `debug`-gated assertion that
   `Methods` contains no `InterfaceKind` entry).
2. Treat a decoded unflattened interface as a hard error rather than handling
   it — it should no longer occur.

Do **not** make `VerifyImplementedBy` assume-flattened (drop/ignore the
recursion) *before* that migration exists: without it, a decoded legacy
interface yields silently wrong satisfaction (ignore) or halts the chain
(panic). Sequencing is: ship this PR with the recursion retained → migration
re-flattens persisted types on upgrade → then remove the recursion.

Also deferred (orthogonal): the pure-VM `flattenInterfaceMethods` conflict path
is a `panic`, relying on go/types rejecting same-name/different-signature embeds
upstream. If a VM path can ever reach construction without that gate, convert it
to a positioned preprocess error.
