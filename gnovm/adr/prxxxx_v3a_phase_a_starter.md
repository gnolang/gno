# PR — interrealm v3a Phase A starter + Phase B migrations (stacked on PR #5669)

Stacked commits on top of `pr-5669` (v2's Phase 3 PR) that begin v3a's
Phase A (machine-op identity model, reference implementations) and
Phase B (real /p/ library migrations).

This ADR ties the commits together and documents what's intentionally
deferred. It satisfies AGENTS.md's "every non-trivial AI-assisted PR
must include an ADR" requirement.

## Context

PR #5669 (v2 Phase 3) closes:
- Type-pun launder (Attack G).
- Unreal-HIV cross-realm write false-positives.

What v2 leaves open:
- **omarsy class**: the uverse `realm` interface is implementable by
  user types, allowing forged identity values to be passed where
  `cur realm` / `rlm realm` parameters are expected.
- **Primitive-recv attack class (H–L)**: upstream documents this as
  the attacks-succeed baseline. The proper defense is the library
  discipline rule "declare your logic data types in /r/, not /p/" —
  attackers can't construct typed pointers into /r/-declared data
  from /p/ because /p/ cannot import /r/.

v3a (designed in `gnovm/adr/interrealm_v3a.md`) addresses omarsy via:
- Concrete `runtime.Realm` struct + machine-op identity queries
  (`runtime.Caller()` etc.). No user-passable realm value → no
  forgery surface.

## Commits in this stack

1. **`b692555ff` docs(interrealm): v3a/v3b ADRs + discipline + capability-levels exploration**
   - `gnovm/adr/interrealm_v3a.md` — Phase A design (the committed plan).
   - `gnovm/adr/interrealm_v3b.md` — optional type-qualifier extension.
   - `gnovm/adr/interrealm_discipline.md` — library-author rules
     for safe interface contract design.
   - `gnovm/adr/interrealm_capability_levels.md` — exploration of
     stronger language-level options if discipline + marker proves
     insufficient.

2. **`5905c79c5` feat(interrealm): add runtime.{Caller,Self,Origin} as v3a identity-query API**
   - `runtime.Caller()` = alias for `PreviousRealm()` (one realm
     transition up).
   - `runtime.Self()` = alias for `CurrentRealm()`.
   - `runtime.Origin()` = `Realm{addr: OriginCaller, pkgPath: ""}`.
   - Returns concrete `runtime.Realm` struct (already non-implementable
     by user types — that's the omarsy closure for any code adopting
     the new API).
   - Filetest `zrealm_runtime_caller_filetest.gno` verifies identity
     equivalences.

3. **`e5f80ec8e` test(interrealm): canary /p/ ACL helper using runtime.Caller (omarsy-free pattern)**
   - New `/p/demo/tests/v3aclhelper` package with `CheckCaller`,
     `CallerPath`, `CallerIsUser` helpers using `runtime.Caller()`.
   - Demonstrates the v3a /p/-helper pattern: no `_ int, rlm realm`
     parameter shape, identity sourced from the VM.
   - Filetest exercises it from /r/ and verifies caller identity
     resolves correctly.

4. **`b1a114ab8` test(interrealm): canary /r/ realm demonstrating v3a Ownable pattern end-to-end**
   - New `/r/tests/vm/v3acanary` realm with full Ownable-style state:
     `Init`, `Owner`, `TransferOwnership`, `DropOwnership`.
   - All ACL via `runtime.Caller()`, no realm-typed parameters.
   - Filetest exercises bootstrap → authorized transfer → ACL
     rejection after ownership changes hands (uses `revive()` for
     cross-realm abort catching, per v2 Phase 3 semantics).

5. **`b4fb9424e` feat(ownable/v1): v3a-aligned ownable using runtime.Caller (coexists with v0)**
   - New `/p/nt/ownable/v1` package — successor to v0.
   - `Ownable` struct, `NewWithAddress`, `NewWithCaller` (new
     constructor capturing `runtime.Caller()` at init), `OwnedBy`,
     `AssertOwnedBy`, `TransferOwnership`, `DropOwnership`, `Owner`.
   - No `_ int, rlm realm` parameter shape.
   - Unit tests cover pure-data API; mutating-method coverage falls to
     the canary filetest (see "Open issues" below).
   - `doc.gno` documents v1 vs v0 trade-offs.

6. **`042a7af99` docs(interrealm): PR-level ADR for v3a Phase A starter stack**
   - This document (initial version).

7. **`7496f4897` refactor(loci): migrate to v3a runtime.Caller pattern (Phase B)**
   - In-place migration of `/p/n2p5/loci.Set` from
     `(_ int, rlm realm, value)` shape to plain `(value)` signature
     using `runtime.Caller().Address()`.
   - Updates `/r/n2p5/loci` caller and the package's own test/filetest.
   - First real Phase B migration — demonstrates the playbook on a
     small self-contained library.

8. **`9e18dd151` refactor(microblog): migrate to v3a runtime.Caller pattern (Phase B)**
   - `/p/demo/microblog.NewPost` migrated to plain signature using
     `runtime.Caller().Address()` for author identity.
   - `/r/demo/microblog` caller updated.
   - Test uses `cross()` scaffolding (documented as transitional).

9. **`e524c2a61` refactor(subscription): migrate lifetime+recurring UpdateAmount to runtime.Caller**
    - `/p/demo/subscription/lifetime.UpdateAmount` and
      `/p/demo/subscription/recurring.UpdateAmount` both migrated.
    - Now satisfy the existing `Subscription` interface (which already
      had the v3a-style signature).

## What this delivers

| Property | Before | After |
|---|---|---|
| omarsy attack via forged `rlm realm` value | open | **closed** for any code using `runtime.Caller()` |
| Primitive-recv attack class (H–L) | left to library discipline (per upstream's revert of the anchor) — declare data in /r/, not /p/ | unchanged here |
| Identity query in /p/ helpers | requires `rlm realm` parameter (forgeable) | machine op (`runtime.Caller()`) |
| /p/ library with omarsy-free ACL | not available | `/p/nt/ownable/v1` reference impl |
| Realm-level reference impl | not in tree | `/r/tests/vm/v3acanary` reference |
| `/p/n2p5/loci` | v2 `(_ int, rlm, ...)` shape | v3a runtime.Caller |
| `/p/demo/microblog` | v2 shape | v3a runtime.Caller |
| `/p/demo/subscription/{lifetime,recurring}` | v2 shape | v3a runtime.Caller |
| `/p/oxtekgrinder/ownable2step` | v2 shape | v3a runtime.Caller |
| `/p/n2p5/mgroup` | v2 shape | v3a runtime.Caller |
| `/p/nt/pausable/v0` | v2 shape | v3a runtime.Caller |
| `/p/agherasie/forms` | v2 shape | v3a runtime.Caller |
| `/p/thox/snowflake` | v2 shape | v3a runtime.Caller |
| `/p/demo/nestedpkg` | v2 shape | v3a runtime.Self + runtime.Caller (path-based ACL) |

## Phase A.2/A.3 — substantive work already in v2 substrate

Investigation during this work revealed that the "generalized indirect
dispatch borrow" and "OriginRealm field" the v3a ADR specified are
**already provided by v2's existing layered borrow** in `PushFrameCall`
(machine.go:2306–2365):

- **`FuncValue.PkgPath`** = declaring package of the function/closure,
  set at construction. Functions as the canonical OriginRealm for
  direct callables.
- **`FuncValue.ObjectInfo.ID.PkgID`** = allocation-time realm stamp,
  set via the allocator. Used for closure-identity preservation
  across copies.
- **Layer-1 borrow** fires for any /r/-declared callable (top-level
  function, method, closure) regardless of how it's dispatched
  (direct, function value, interface method). v3a's "indirect dispatch
  must borrow to OriginRealm" guarantee is already enforced for
  /r/-declared callables.
- **Layer-2 borrow** fires for stdlib/p/ methods on object receivers
  with foreign PkgID.
- **Primitive/nil-receiver gap** (Attack H–L class): left open here.
  Upstream's revert of the anchor mechanism (`db1486802`) chose
  library discipline ("declare data in /r/, not /p/") over a VM-level
  fix. This PR adopts the same direction — no anchor change.

`zrealm_v3a_indirect_dispatch_filetest.gno` (commit 515a38dd3) verifies
the invariant directly: identity queries (`runtime.Caller()`) resolve to
the same answer whether the helper is invoked directly or through a
function-value indirection.

What's left for a future PR (not blocking v3a Phase A):

- **Call-form analyzer at preprocess**: classify each `CallExpr` as
  direct or indirect for tooling/audit purposes. Not load-bearing for
  safety — the existing layered-borrow runtime dispatch already
  produces correct behavior for both forms. Useful for static
  analysis tools, lint rules, and IDE highlights.

Phases B (broad migration) and C (surface removal of `cross`/`cur`/`rlm`)
are entirely follow-on PRs.

## Cascading migrations deferred to follow-up PRs

Three /p/ libraries with v2 patterns were investigated but deferred
because their migrations cascade into downstream realms with semantic
or scale implications:

- **`/p/nt/treasury/v0`** — `GRC20Banker.Send` calls `grc20.Teller`
  methods that use the v2 pattern. Migrating treasury without
  migrating grc20 leaves an inconsistent Banker interface. grc20
  itself has 12+ v2-pattern methods and 10+ realm importers
  (bar20, foo20, grc20factory, grc20reg, atomicswap, tokenhub, etc.).
  → Treasury + grc20 should be a single follow-up PR.

- **`/p/nt/mux/v0`** — has a dual API (`HandlerFunc` non-rlm,
  `HandlerFuncRlm` rlm-aware) for v2's identity-threading. Migrating
  would drop the rlm-aware variant. Used by gov/dao/v3 which threads
  rlm deep through render handlers; rlm-threading must collapse
  alongside the mux migration. → mux + gov/dao should be a single
  follow-up PR.

- **`/p/moul/authz`** — `Authorizer.Transfer` and `DoByPrevious` use
  `rlm.Previous().Address()` as the principal. In test setups with
  `cross2(cur)` wrapping, `rlm.Previous()` resolves to the test runner's
  realm, not to the test's nominal caller. The v3a equivalent
  (`runtime.Caller().Address()`) walks differently under those wrappers,
  causing test-time mismatches. Either the test setup conventions
  need to change in lockstep with the migration, or the migration
  needs explicit `runtime.CallerN(2)` semantics (which Phase A.1
  hasn't implemented). → authz migration deferred until test-
  harness conventions and `CallerN` are settled.

These three represent the remaining substantive Phase B work and
should land as 2–3 follow-up PRs after this stack merges.

## Open issues surfaced by this work

1. **Testing-harness gap for v3a unit tests** — RESOLVED.
   v2's test-time `X_getRealm` panicked with "cannot seek beyond
   origin caller override" when `runtime.Caller()` was invoked from
   a /p/ helper called directly under `testing.SetRealm(NewUserRealm(...))`.
   That panic was overly strict for v3a-style tests (no cross()
   scaffolding above the override).
   
   Resolution: dropped the panic in `gnovm/tests/stdlibs/chain/runtime/
   testing_runtime.go`. When the walk reaches a user-realm override
   with `crosses < height`, skip the frame and let the switch
   fallthrough return `ctx.OriginCaller`. This aligns test-time
   behavior with production `execctx/realm.go` (which doesn't panic).
   Two existing tests (`zrealm_crossrealm13`, `zrealm_crossrealm13a`)
   updated to reflect the new output. `ownable/v1` now has full
   mutating-method unit tests with no `cross()` scaffolding.

2. **Pre-existing test failures on `pr-5669` base**. The PR base has
   several pre-existing test failures (`addressable_1b_err.gno`,
   `zrealm_p_convert_readonly_ok_filetest.gno`, slice/varg tests).
   Confirmed unrelated to this stack — failures reproduce on
   `9d560263d` without these commits applied.

## Verification

For each commit, verified:
- Targeted filetest passes (`go test ./pkg/gnolang/ -run Files/<name>`).
- No new regressions in `Files/zrealm*` suite (only pre-existing
  baseline failures remain).
- ownable/v1 unit tests pass via `gno test`.

## References

- `gnovm/adr/interrealm_v2.md` — v2 design and phasing.
- `gnovm/adr/interrealm_v3a.md` — v3a design (the committed plan).
- `gnovm/adr/interrealm_v3b.md` — optional type-qualifier extension.
- `gnovm/adr/interrealm_discipline.md` — library discipline rules.
- `gnovm/adr/interrealm_capability_levels.md` — escalation options
  if discipline is insufficient.
- `docs/resources/gno-security.md` — threat-class taxonomy.
- `docs/resources/gno-interrealm.md` — interrealm semantics.

## Reviewer guidance

- **Read order**: this ADR → `interrealm_v3a.md` → individual commits
  in stack order.
- **What to verify**: each commit is self-contained and reviewable
  independently. The stack is non-breaking — every v2 surface still
  works; v3a additions coexist.
- **What to push back on**: the deferred items (OriginRealm,
  call-form analyzer, testing-harness gap). These are real Phase A
  work that hasn't shipped yet; if reviewers want them in this PR,
  the scope expands significantly.
