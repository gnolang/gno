# ADR: `cur` binding follow-ups (post-#6193 review fixes)

## Status

Implemented on `fix/cur-binding-followups` (PR #6196).

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
   declaration, so a package-level `var cur realm` and a named result
   `func F(x int) (cur realm)` were refused with "cannot reassign the crossing
   `cur` parameter". The review offered two ways out: make the rules
   declaration-aware and keep those bindings writable, or refuse to declare a
   realm-typed `cur` anywhere but as a crossing function's first parameter. It
   settled on the second: one rule, "a realm-typed `cur` is always the
   parameter", so the write rules can key on name plus type with no scope
   resolution, and the corner in the #6196 review — `func F(_ realm) (cur realm)`,
   where a crossing declaration's unnamed first parameter lets a named result
   named `cur` masquerade as the binding — is refused by construction.
4. `installInheritedCur` ran a full frame walk per no-cross crossing entry
   without a gas charge, while `doOpEnterCrossing` charges
   `OpCPUSlopeEnterCrossing` per call frame for the same traversal.
5. The identity panic interpolates `fv.Name`, empty for a function literal —
   exactly the callback shape the message is written for.
6. `curUsesPreprocessOrigin`'s doc header still described the removed
   structural test and claimed it survives AST persistence.

## Decision

**A realm-typed `cur` is only ever the crossing parameter.** `checkRealmCurDecl`
refuses a realm-typed `cur` declaration at every declaration site that is not a
crossing function's first parameter:

- named results, in the `FuncTypeExpr` handler next to the existing
  first-parameter naming rules;
- `var` declarations and fresh `:=` declarations, in `defineOrDecl`. A `:=`
  whose name was reserved by that very statement (checked through the
  `NameSource` origin) declares a new binding; one whose slot came from an
  earlier declaration is a rebind and is left to the write rules;
- range DEFINEs, in the `RangeStmt` TRANS_BLOCK arm, where the element type is
  known (`[]realm` value, `map[realm]X` key, ...);
- for-init DEFINEs, under the `cur.loopvar` name `initStaticBlocks1` renames
  them to, so the source binding is caught even though no source can spell its
  synthesized name.

The first-parameter rules from #6193 are unchanged: a crossing function's first
realm parameter must be named `cur` (or be the unnamed `.arg` placeholder), and
no later realm parameter may be named `cur`.

**The write rules key on name plus resolved type.** With the invariant above, a
realm-typed `cur` on the left of a write, or under `&`, can only be the
parameter, so `isCrossingCurParam` and its `GetBlockNodeForPath` scope
resolution are gone; the assignment check keeps running after the
DEFINE/ASSIGN branch only so the name's path is set, and a fresh realm-typed
`cur` cannot reach it (it is refused by `defineOrDecl` first). The cur-call
provenance check keeps its own declaration test and additionally requires the
declared first parameter to be named `cur`; now that declarations are banned,
that arm is belt-and-braces rather than load-bearing.

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
source location for a function literal and to `<func literal>` when there is
no source, so the panic never renders as a bare trailing dot.

**The stale header is rewritten** to state the pointer-identity semantics and
why they hold (`.cur` is only baked by an in-process preprocess of a
synthesized call).

## Alternatives considered

- **Keep the DEFINE carve-out** on the argument that the runtime check covers
  it. Rejected: the carve-out's rationale is "the name is new", which is false
  for a same-scope `:=`, and the rule's message plus its fixtures advertise a
  compile-time guarantee that did not hold.
- **Keep the name+type rules declaration-aware** (resolve the written name and
  refuse only a crossing `FuncDecl`/`FuncLitExpr` parameter). Implemented first
  on this branch, then rejected in review: it leaves four shapes of realm-typed
  `cur` in the language that are not identity, and two checks had to agree by
  construction ("may be cur-called" / "may not be written") instead of by
  invariant. The declaration ban makes the two agree trivially.
- **Ban realm-typed `cur` only in function scopes**, leaving package-level
  `var cur realm` legal. Rejected: a package-level placeholder cannot be
  cur-called (the provenance check resolves its declaring block to the
  `PackageNode` and refuses), so it is inert — but the invariant buys the simple
  write rules, and in-tree occurrences were all test-file placeholders, which
  the rename improves anyway.
- **Carve `_test.gno`/`_filetest.gno` files out of the declaration ban.** The
  chain never preprocesses test files (`MPFProd` filters them at deploy through
  `AsRunnable`, and `IterMemPackage` yields production blobs only at boot), so
  the carve-out would be invisible on chain and would avoid renaming the
  placeholders. Rejected: a language rule that changes by file suffix is worse
  than a mechanical rename of misleading placeholders.
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
- Compatibility: every realm-typed `cur` declaration outside a crossing
  function's first parameter is now a preprocess error. In-tree production code
  has none; the occurrences were the nil-placeholder pattern in test files
  (`var cur realm` used as the ignored `rlm` argument of uassert/urequire
  helpers). 20 non-quarantined and 8 quarantined examples test files were
  updated: the placeholder was renamed to `rlm` where it was actually used
  (valopers, sys/params), and deleted where it was dead.
- On-chain: deploy and boot only preprocess production files, so test-file
  placeholders never reach a chain. A deployed production package that declared
  a realm-typed `cur` outside a crossing parameter would panic
  `PreprocessAllFilesAndSaveBlockNodes` at the next node restart; no such
  package is known, and the same deployed-code scan #6193 ran applies.
- The write rules are name+type only; `isCrossingCurParam` is deleted.
- Fixtures: the four `zrealm_cur_shadow*.gno` fixtures pinned the old
  shadow-stays-writable behavior; they are replaced by
  `zrealm_cur_decl_{var,local,result,range,forinit}.gno`, one per declaration
  site. `zrealm_cur_legal.gno` keeps the positive controls (non-realm `cur`,
  realm-typed bindings under other names) and `zrealm_cur_other_legal.gno` is
  rewritten for the same. `std_unsafe0.gno` is unchanged: its `cur` local is a
  `runtime.Realm` value, not the `realm` type.
- The identity check is now a backstop with no known reachable source route;
  it is intentionally kept.
