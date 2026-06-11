# ADR: Depth-based shadowing for promoted struct fields and methods

## Status

Proposed (AI-assisted fix; found via differential testing against the Go
toolchain).

## Context

Go resolves a selector `x.f` to the field or method "at the shallowest
depth in T where there is such an f" (Go spec, "Selectors"). If there is
more than one `f` at that shallowest depth, the selector is ambiguous and
illegal. A field or method promoted from a shallower embedded field
therefore **shadows** a same-named one promoted from a deeper embedded
field.

`StructType.FindEmbeddedFieldType` (`gnovm/pkg/gnolang/types.go`) drives
all field/method selection and interface-satisfaction checks. Its previous
implementation walked embedded fields and treated *any* second match as a
conflict, with no notion of depth. This caused a divergence from Go,
confirmed against the Go toolchain:

**Valid code wrongly rejected (depth shadowing not applied).** When a
method was promoted at depth 1 and a same-named method was also reachable
at depth 2+, gno reported a "conflict" and concluded the type did not
implement the interface. Example: `T` embeds `Shallow` (with `M`) and `Mid`
(embedding `Deep`, with `M`); Go resolves `T.M` to `Shallow.M`, but gno
rejected `var i I = T{}` at preprocess with "T does not implement I
(missing method M)". This is the dangerous case: on-chain `addpkg`
type-checks with go/types first (which accepts the program), so such a
package passes the deployment gate and then panics when the VM
preprocesses it.

## Decision

Apply Go's depth-based shadowing rule in the embedded-field search of
`StructType.FindEmbeddedFieldType`:

- Search direct (depth-0) fields first; a direct match shadows anything
  promoted and returns immediately.
- Among embedded fields, keep the candidate with the **shortest trail**
  (trail length is the promotion depth). A strictly shallower candidate
  shadows a deeper one; two candidates at the same shallowest depth mark
  the selector **ambiguous** (no match).

The shared recursion-guard set is kept unchanged (it both prevents cycles
and, as a side effect, deduplicates a type reachable through two embedded
paths — see "Out of scope" for the one case where that matters). The
function's return contract (`trail, hasPtr, rcvr, field, accessError`) is
unchanged, so runtime selector navigation (`OpSelector` → `GetValueAt`) and
all existing callers are unaffected for the cases they already handled;
only the resolution of multi-candidate promotions at different depths
changes, to match Go.

## Alternatives considered

- **Full breadth-first method-set computation** (à la `go/types`
  `lookupFieldOrMethod`), or extending the return contract to propagate
  ambiguity-with-depth: more faithful, and would also fix the diamond case
  below, but a larger and riskier change to a hot, widely-used function.
  Rejected in favor of the minimal depth-comparison change that closes the
  dangerous (deploy-gate-bypass) false-negative.
- **Per-branch copies of the recursion-guard set** to detect the diamond:
  rejected. It fixes the diamond but introduces a regression — when one
  embedded sibling is *internally* ambiguous and another has a clean match
  at the same depth, the per-branch search loses the sibling's ambiguity
  and wrongly accepts (Go rejects). Keeping the shared set preserves the
  correct rejection for that case (covered by
  `embed_method_shadow_internal_ambiguous`).

## Out of scope (pre-existing, unchanged by this fix)

This fix closes the common **disjoint-subtree** shadowing case (a method/field
promoted from one embedded subtree shadowing a same-named one in an unrelated
subtree at a greater depth). It does **not** fix the broader, pre-existing
class of promotion bugs caused by the **shared recursion-guard set**, which is
a graph visited-set where Go's promotion semantics require a per-path (tree)
traversal. Differential fuzzing against the Go toolchain over random
multi-embedding shapes shows gno (both before and after this fix) diverges
from Go on a large fraction of them. The dominant cases:

- **Shared-type pruning (false reject).** When the same type is reachable
  through two embedded paths at *different* depths, the shared set marks it
  visited on the first (often deeper) path and prunes the second, so the
  legitimate *shallower* promotion is lost and gno rejects a program Go
  accepts. Example: `T0`/`T1` both define `M`; `T2{T1;T0}`, `T3{T2;T0}`,
  `T4{T3;T1}` — Go resolves `T4.M` to `T1.M` (depth 1), gno rejects with
  "does not implement". This is the **dangerous direction** (go/types accepts,
  so the package clears the on-chain `addpkg` gate and then panics at VM
  preprocess). Unchanged from upstream.
- **Diamond ambiguity (false accept).** When the *same* embedded type is
  reachable through two distinct paths at the *same* depth (`T` embeds `A` and
  `B`, both embedding `Base`), the shared set deduplicates `Base`, so gno finds
  one path and accepts (running `Base.M`) where Go rejects as an ambiguous
  selector. Gated on-chain by go/types, so it only affects `gno run` / direct
  VM use. Unchanged from upstream.

A complete fix needs the per-path / depth-aware-ambiguity algorithm noted under
Alternatives (each embedded branch explored with its own cycle-guard, plus
propagation of the depth at which a subtree becomes ambiguous), which would
subsume this change. It is a larger rewrite of a hot, widely-used function and
is deliberately left as a follow-up; this ADR's change is the minimal,
zero-regression step that closes the disjoint-subtree case without touching the
return contract.

- **Cross-package inaccessible promotion.** When an embedded member
  promoted from a *foreign* package is inaccessible (unexported, different
  `PkgPath`), `FindEmbeddedFieldType` returns an access error and aborts the
  whole search, instead of skipping that candidate so an accessible
  same-named member at another depth can resolve. Go skips foreign-package
  unexported members entirely, so gno wrongly rejects e.g. a type that
  embeds a foreign `B` (with unexported `m`) and a local `Inner` (with its
  own `m`) where Go resolves to `Inner.m`. Present identically on upstream
  (the `// XXX make test case and check against go` marker predates this
  change) and orthogonal to depth shadowing; left to a follow-up.

## Consequences

- Field/method selection and interface satisfaction now match Go for
  depth-shadowed promotions. In particular, valid programs that passed the
  on-chain go/types gate no longer panic at VM preprocess.
- The ambiguous-selector and missing-method diagnostics still read
  "missing method"/"missing field" rather than Go's "ambiguous selector";
  both reject, so spec compliance holds, but the message could be improved
  in a follow-up (the resolver returns only "no match", not the reason).
- The change is a localized comparison (no extra allocation; the
  recursion-guard set is threaded as before), so there is no measurable
  cost beyond the unchanged walk. Preprocess is gas-metered on-chain;
  embedding breadth/depth is already bounded by the pre-existing
  `MaxEmbedDepth` / field-count caps.
- Tests: `embed_method_shadow0/1` (shadowing accepted; selection resolves
  to the shallow member; multi-level) and `embed_method_shadow2`
  (pointer-embedded shadowing) pin the fix — they panic at preprocess on
  the old code. `embed_method_shadow_ambiguous` (same-depth siblings),
  `embed_method_shadow_fieldmethod` (field vs method at the same depth),
  and `embed_method_shadow_internal_ambiguous` (an internally-ambiguous
  sibling alongside a clean one) are parity/regression guards: all were
  already rejected by the old logic, so they pass on the old code too; they
  are kept for go/types parity (via the `TypeCheckError` directive) and to
  guard the new same-depth comparison against the per-branch-copy
  regression described in Alternatives.
