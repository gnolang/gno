# ADR: `cur` binding follow-ups (post-#6193 review fixes)

## Status

Implemented on `fix/cur-binding-followups`. PR number pending.

## Context

#6193 made the crossing `cur` parameter a fixed binding. It added three
preprocess write rules — assignment, address-of, range assignment target — and
a runtime identity check in `installInheritedCur` for the routes the syntactic
enumeration missed. That landed as `3d9f4dc8e`.

A post-merge review (line-by-line, removed-behavior, cross-file, interrealm,
cleanup, altitude, conventions angles; each candidate run against both the
merged commit and its parent) found six issues that survive execution:

1. A same-scope `cur, x := ...` DEFINE assigns the existing parameter slot, but
   the `n.Op != DEFINE` carve-out skipped it. `zrealm_cur_reassign.gno` pins the
   plain `cur = ...` form as a preprocess error; the `:=` form compiled and
   rebound, caught only by the runtime check at call time.
2. `t.Run` seeds a crossing sub-test with the top-level test's `cur`
   (`gnovm/tests/stdlibs/testing/testing.gno`), so `helper(cross(cur), t)`
   followed by a crossing sub-test panicked and aborted the whole `gno test`
   run. Worked on the parent commit.
3. The write rules keyed on the name `cur` plus a realm static type, not on the
   declaration. Package-level `var cur realm` and named results
   `func F(x int) (cur realm)` compile on the parent commit and were refused
   with "cannot reassign the crossing `cur` parameter".
4. `installInheritedCur` ran a full frame walk per no-cross crossing entry
   without a gas charge, while `doOpEnterCrossing` charges
   `OpCPUSlopeEnterCrossing` per call frame for the same traversal.
5. The identity panic interpolates `fv.Name`, empty for a function literal —
   exactly the callback shape the message is written for.
6. `curUsesPreprocessOrigin`'s doc header still described the removed
   structural test and claimed it survives AST persistence.

## Decision

**Write rules are declaration-aware.** A new `isCrossingCurParam` resolves the
written name through `GetBlockNodeForPath` and asks whether the declaration is
a crossing `FuncDecl`/`FuncLitExpr` — the same test the cur-call provenance
check makes before forwarding a bare `cur`. "May be cur-called" and "may not be
written" now agree by construction. Package-level vars, named results, and
block-scoped shadows stay writable; the shadow still cannot be cur-called,
which `zrealm_cur_shadow_call.gno`, `_if`, and `_range` pin.

The assignment check moved after the DEFINE/ASSIGN branch so it runs for both:
a DEFINE that reuses a name already declared in the same block takes
`defineOrDecl`'s "already defined" branch, which assigns that slot and only
then sets the name's path. `zrealm_cur_shadow.gno` flips from an expected
rejection to a positive control.

**The identity check exempts the testing stdlib's dispatch.** `harnessSeedsCur`
skips the mismatch panic when the calling frame belongs to package `testing`.
The harness threads the top-level test's `cur` on purpose and has no way to
read the live cur, so the VM-side exemption is smaller than changing the
harness. The testing package is not reachable from chain code; the same
carve-out already exists for crossing declarations (`crossingFromTestFile`,
`TestingBasePkgPath`).

**The install walk stays unmetered, tracked separately.** The review found
that the frame walk `installInheritedCur` runs is not charged, while
`doOpEnterCrossing` charges `OpCPUSlopeEnterCrossing` per call frame for the
same traversal. Metering it changes gas for every no-cross crossing entry,
which is consensus-visible, and the VM has no version gate for gas changes
(zero `Hardfork` references in `gnovm/`). It is therefore deferred to ride
with the other gas-schedule work in a coordinated upgrade:
gnolang/gno-fixes#115. A `NOTE` at the call site points there.

**Diagnostics name anonymous callees.** `funcDisplayName` falls back to the
source location for a function literal and to `<func literal>` when there is no
source, so the panic never renders as a bare trailing dot.

**The stale header is rewritten** to state the pointer-identity semantics and
why they hold (`.cur` is only baked by an in-process preprocess of a
synthesized call).

## Alternatives considered

- **Keep the DEFINE carve-out** on the argument that the runtime check covers
  it. Rejected: the carve-out's rationale is "the name is new", which is false
  for a same-scope `:=`, and the rule's message plus its fixtures advertise a
  compile-time guarantee that did not hold.
- **Keep the name+type rules over-broad**, as the shadow fixture documented.
  Rejected: the deliberate over-rejection's stated rationale ("refusing a
  shadow costs nothing") does not extend to package-level vars and named
  results, which are ordinary code and were broken by it.
- **Change the harness to seed the live cur** instead of exempting it. The
  harness has no getter for the live cur and `t.cur` is deliberately the
  top-level test's; adding a native getter is a wider testing-stdlib change
  for the same observable behavior.
- **Charge or cache the walk in this change.** Rejected for the metering
  charge: it is a consensus-relevant gas change that belongs in a coordinated
  upgrade with the rest of the schedule, not in a correctness fix. Rejected
  for caching the topmost cur on the Machine: bigger change with staleness
  risk on frame pop/revive, and it would not cover deferred crossing calls,
  which never run `doOpEnterCrossing`.
- **Drop the runtime check** now that the known source route (the DEFINE gap)
  is closed. Kept: the enumeration has been incomplete twice already, and the
  check does not depend on it being complete.

## Consequences

- Gas: unchanged. The unmetered install walk is the one known gap and is
  tracked at gnolang/gno-fixes#115 to ship with the other gas-schedule work.
- `zrealm_cur_shadow.gno` becomes a positive control (a shadowing realm-typed
  `cur` may be rebound and addressed); the `zrealm_cur_shadow_{call,if,range}`
  fixtures keep pinning that it cannot be cur-called.
- New tests: `zrealm_cur_reassign_define.gno`, `zrealm_cur_closure.gno` and
  `zrealm_cur_other_legal.gno` (filetests),
  `gnovm/tests/stdlibs/testing/cur_subtest_test.gno` (the `t.Run` regression),
  and Go unit tests in `op_call_test.go` for the harness exemption and the
  display-name fallback.
- The identity check is now a backstop with no known reachable source route;
  it is intentionally kept.
