# ADR: commondao Execute re-entrancy latch

> **Superseded in part by `prxxxx_commondao_open_kinds_pkg.md` (decision 3).**
> This ADR concludes a package-level latch is *impossible* (a mutable `/p/`
> global is forbidden by the borrow rule) and therefore lands ONLY a
> realm-global latch. That "impossible" is about a package *global*: a later
> ADR added per-DAO `executing` / `proposing` **fields** on the `CommonDAO`
> object, which ARE writable (a field on the owning realm's object, not a
> package global) and give standalone `/p/` consumers same-DAO re-entrancy
> protection this realm-global latch never gave them. The realm-global latch
> below is NOT removed — it still enforces the one-executor-per-tx / cross-DAO
> ancestor-dissolution straddle that a per-DAO latch cannot. The two are
> complementary layers; see the realm's `reentrancy.gno` header and
> `prxxxx_commondao_open_kinds_pkg.md` decision 3. The historical text below is
> kept for the record.

## Context

`CommonDAO.Execute` (`commondao.gno`) checks `if dao.deleted` once at entry,
removes the proposal from active storage before running the executor
(remove-before-run), runs the executor, then unconditionally finalizes the
proposal (`StatusExecuted` + archive) with no re-check. Every executor in the
realm today is closed and realm-authored, and none re-enters `Execute`, so the
whole system has run under an emergent "one executor per transaction"
assumption that all prior treasury/proposal reviews relied on.

The planned arbitrary-execution proposal kind breaks that: it runs a
proposer-authored closure `func(_ int, sub realm) error` that holds the DAO's
`sub` and can call back into the realm — including `commondao.Execute(cross(sub),
anyDAO, pid)` (post-deadline Execute is permissionless, and `cross(sub)` yields
a fresh primary cur that can mint any DAO's sub — see
`pr6012_commondao_exec_scope.md` for the sub-identity and `cross(sub)`
semantics). This ADR lands the latch
first, as a prerequisite, so the assumption becomes a structural invariant
before any re-entrant executor exists.

### The concrete issue (verified)

DAO H has two passed, post-deadline proposals: `P_exec` (execution kind) and
`P_dissH` (self-dissolution; root DAOs host their own). Finalizing `P_exec`
runs its closure, which calls `Execute(cross(sub_H), H, P_dissH)`. The nested
call sweeps H's treasury, `setListed(false)`, and `H.Dissolve()` → sets
`H.deleted = true`. `P_exec` (already removed at the remove-before-run step) is
not dismissed. The nested call returns; H's outer frame resumes on a now-deleted
H and finalizes `P_exec` as Executed + archived — violating dissolution's
terminality contract ("a deleted DAO rejects executions; nothing remains
pending"), with the closure holding H's authority after H is swept and dead.
Not sequential-equivalent: dissolve-first ⇒ the entry guard rejects `P_exec`;
exec-first ⇒ no dissolution without nesting. Two other adversarial lenses found
the treasury paths safe only by a fragile textual read-adjacent-to-send
invariant, and a "conditional-execution free option" (nest another DAO's
Execute, observe in-tx, panic-to-revert) with no in-protocol weaponization but
real external-composition reach.

## Decision

A **global re-entrancy latch on the realm's public `Execute`**: a realm module
`var executing bool`, raised on entry and lowered by defer; a nested `Execute`
(while `executing`) panics, which — thrown across the `cross()` boundary the
re-entrant call arrived through — aborts the whole transaction.

- **Realm-side, not package.** A mutable `/p/` global is forbidden: writing one
  panics `invariant violation: DidUpdate called on external-realm object
  without prior readonly check` (verified empirically). The realm legally holds
  mutable module state. Consequence: the package cannot self-enforce
  non-re-entrancy; any realm embedding `/p/nt/commondao` and running re-entrant
  executors must guard its own `Execute` gateway, as this realm does. The
  realm's public `Execute` is the sole caller of `dao.Execute`, so one guard
  there is complete for this realm. **Note for the future execution kind:** the
  package `CommonDAO.Execute` is exported and unguarded, and `CommonDAO.Parent()`
  returns a live `*CommonDAO`, so the realm-side latch covers re-entry only via
  the public gateway. The arbitrary-execution kind must ensure its closures
  cannot obtain a `*CommonDAO` handle and call the package `Execute` directly —
  that path would bypass this latch.
- **Global (one bool), not per-DAO.** `Execute` is latched; `Vote`, `Create*`,
  `Withdraw`, `Resign` are not — no executor re-enters them, legitimate
  executors need them (sub-DAO creation, the dissolution sweep), and a proposal
  executor acting as its DAO in *another* DAO (e.g. casting a council vote)
  rides those paths. Sequential (non-nested) `Execute` calls in one tx remain
  allowed; only nesting is blocked.
- **`enterExecute` / `leaveExecute` helpers** hold the check so the guard is
  unit-testable without a full crossing/DAO setup.

## Alternatives considered

- **Per-DAO latch + target-check** (a set of executing DAO ids; block same-DAO
  re-entry at entry, plus `assertNotExecuting(target)` at dissolution and
  clawback). Investigated in depth and found VIABLE and safety-equivalent *for
  the straddle*: the headline gap (sub-DAO dissolution/clawback is hosted in
  the ancestor, so a latch keyed on the *executing* DAO misses the descendant
  being deleted under the ancestor's Execute) is closable by checking the
  *target's* latch, and the target-check is complete because treasuries are
  id-keyed, `Dissolve()` does not cascade, and the outer frame reads no
  cross-DAO state after its closure. Rejected anyway because (1) the only
  capability it preserves over global is nesting an `Execute` of a *different*
  DAO, which re-opens the free-option and which nothing needs, and (2) it is
  worse programmer UX: a set plus ~10 scattered, non-centralizable target
  checks and a standing rule that every future DAO-mutating executor must
  carry one — reinstating the fragile, non-local invariant the latch exists to
  remove. Global is one bool with nothing to remember.
- **Post-executor `dao.deleted` re-check instead of a latch.** Fixes only the
  dissolution case, not the general re-entrancy class. Too narrow.
- **Package-level latch.** Impossible (mutable `/p/` global forbidden, above).

## Consequences

- `public.gno` `Execute` gains `enterExecute(); defer leaveExecute()` after
  `assertCurrent`. New file `reentrancy.gno` holds the latch and helpers.
- No existing flow changes: no current executor nests `Execute`, so all package
  and realm filetests and all commondao txtars stay green. The flag is
  raised/lowered within a single `Execute`; it cannot leak across transactions.
  On a normal return the defer lowers it; on a rejected re-entry the panic
  aborts the transaction and the aborted tx's realm writes (including the flag)
  are never committed — cross-tx safety rests on that commit-gate rollback, not
  on the defer.
- A unit test (`reentrancy_test.gno`) pins the mechanism directly: a nested
  `enterExecute` panics, `leaveExecute` clears the latch. Mutation-verified:
  removing the guard fails exactly that test. The full end-to-end chain (a
  re-entrant closure → `Execute(cross(sub), …)` → the latch panic → an
  unrecoverable, whole-tx abort) was confirmed in post-review with a throwaway
  probe kind. (Since landed: `z_20_i` is that green end-to-end pin,
  mutation-verified, added with the execution kind.) At this ADR's time it
  was not landed as a green regression test because the latch
  panic is unrecoverable (so a `_test.gno` cannot assert it green) and filetests
  cannot register a custom kind; the green end-to-end pin lands with the
  arbitrary-execution kind, which exposes a public `Create*` a filetest
  `// Error:` can drive.
- "One executor per transaction" is now a structural, local, machine-checked
  invariant rather than an emergent property.
