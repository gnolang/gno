# ADR: commondao proposal system

## Context

The proposal system is the package's governance core: `Propose` /
`Vote` / `Execute` / `Withdraw` in
`examples/gno.land/p/nt/commondao/v0/commondao.gno`, the default tally
`TallyDefault` in `record.gno`, the status machine and
`ProposalDefinition` interface in `proposal.gno`, and the nine realm
definition types (text, self/ancestor council-update, sub-DAO, dissolve,
treasury spend/clawback/freeze, and set-proposal-kind). It was reviewed
against a frozen invariant set, converged by a three-reviewer audit that
verified every invariant, the type matrix, and the tally math directly
against source.

The review found **no MAJOR correctness or security defect**. The tally math
is already exhaustively pinned by `TestTallyDefaultProperty` (every
electorate ≤ 8 × every yes/no/abstain split × both thresholds, cross-checked
against an independent reference). The gaps were coverage-only, now closed
(see Consequences). This ADR records the deliberate interpretations so they
are not re-litigated.

## Decision — deliberate interpretations

- **Silence counts as opposition.** The tally denominator is `D =
  |electorate snapshot| − abstains`, i.e. the *electorate size* minus
  abstains, not *votes cast* minus abstains (`record.gno`). A member who
  neither votes nor abstains still sits in `D`, so silence weighs against
  passage. This is the chosen reading of the Constitution's default rules;
  it makes early passage require real turnout, not just a quorum of the
  present.
- **The electorate is snapshotted at Propose and is immutable.** Both the
  Vote-path early termination and the Execute-deadline re-tally use
  `p.Electorate()`, never the live council. Members added after Propose vote
  on the *next* proposal; members removed or resigned afterward remain in the
  electorate (their silence still counts). Council churn cannot swing an
  in-flight proposal.
- **A Passed proposal never expires.** Early termination sets `StatusPassed`
  and leaves the proposal in active storage; `Execute` runs the
  `StatusPassed` branch with no deadline check, so a decided proposal is
  executable before *or* after its voting deadline.
- **Post-deadline finalization is permissionless; pre-deadline is
  council-only.** The package `Execute`/`Vote`/`Withdraw` carry no caller
  check — authorization lives in the realm wrappers (`public.gno`), which
  gate pre-deadline execution on council membership and open it once the
  deadline passes (the tally is then deterministic). Withdraw is
  creator-only at the realm layer.
- **Only self council-update is CapExempt.** A DAO's own council-removal must
  never be blockable by its own full active-proposal cap, so
  `councilUpdatePropDefinition` bypasses the cap (bounded instead to one
  active exempt proposal per creator). Ancestor council-update, being hosted
  in the ancestor's proposal storage, is subject to the ancestor's cap and is
  correctly NOT exempt.
- **Withdraw on a dissolved DAO returns `ErrProposalNotFound`, not
  `ErrDAOIsDeleted`.** `Withdraw` has no explicit `deleted` guard, but
  `Dissolve` dismisses every in-flight proposal before soft-deleting, so a
  dissolved DAO has an empty active store and `Withdraw` fails by
  construction. Safe; the asymmetry is intentional (not worth a redundant
  guard).
- **ID overflow is unreachable.** `genID.TryNext` overflows only at 2⁶⁴
  proposals; there is no seed and no public path to approach it, so
  `ErrOverflow` is a code-inspection invariant with no test.
- **Status-transition legality is structural.** The remove-before-run
  ordering plus the per-status gates make illegal transitions unreachable
  (a finished proposal is not in active storage, so it cannot be re-executed
  or re-voted); this is enforced by construction rather than a dedicated
  transition-matrix test. The one path that retires a *Passed* (decided but
  unexecuted) proposal without executing it is `Dissolve`, which dismisses
  every in-flight proposal — active and passed — into a DAO that then rejects
  all execution (`ErrDAOIsDeleted`); the "Passed never downgrades" rule thus
  holds except under dissolution, which is safe by construction.

## Consequences

- **Coverage gaps closed (test-only; no production code changed):**
  - `z_7_g`, `z_12_f`, `z_14_g` — behavioral pins that sub-DAO creation,
    treasury clawback, and treasury freeze pass by **Simple** majority: a
    5-member council splitting 3 YES / 2 NO passes Simple (`2·3 > 5`) but not
    Super (`3·3 < 2·5`). Previously these were pinned only by unit assertions
    on the returned constant; every lifecycle filetest used a single-member
    council where the two thresholds are indistinguishable. Mirror of
    `z_11_h` (spend = Super). Mutation-verified.
  - `z_9_f` — the Execute-deadline re-tally uses the electorate snapshot: a
    3-member proposal with one YES, then two resignations, is dismissed at
    the deadline (snapshot `D=3`), not passed (live `D=1`). Mutation-verified.
  - `TestVotingRecordReVoteSameChoice` — a same-choice re-vote is idempotent
    (counter stays 1). Mutation-verified against a skip-decrement-when-
    unchanged mutant, which the pre-existing different-choice test misses.
- The exhaustive `TestTallyDefaultProperty` already pinned the D-math (both
  boundaries and the `D≤0` guard); four candidate gaps (in-flight council
  change, late vote, CapExempt-per-creator, deadline-dismissal) were found
  already pinned and needed no new tests.
- Package + realm suites and the treasury txtar remain green.
