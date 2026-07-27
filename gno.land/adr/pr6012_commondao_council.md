# ADR: commondao Council + spec-default tally

## Context

`p/nt/commondao/v0` and its (quarantined) realm predate the Common DAO
Spec (`docs/CONSTITUTION.md`, Appendix `:1467-1543`). Four confirmed
misalignments: grouping machinery the spec never asked for (the spec
models a DAO as Charter + Council + sub-DAO tree); a tally denominator
drawn from the whole member pool with abstentions diluting it; no
treasuries; and DAO instances as registry entries under one realm
address. This PR fixes the first two and simplifies aggressively; a
follow-up PR adds per-DAO `cur.Sub` identities and treasuries.

## Scope

Deferred to the treasury PR: per-DAO addresses via `cur.Sub(<dao-id>)`
and proposal-gated treasuries (`:1534-1543`) — nothing here conflicts
with that design (DAO IDs are monotonic and never reused; the module
path is fixed regardless of quarantine location; no banker code).
Out of scope: the Charter
(Purpose/Description) and Bylaws/Mandates data model with ancestor
amendment (`:1485-1496`), and the m-of-n multisig representation
(`:1539-1541`), an optional alternative whose m ≥ 3 floor deliberately
does NOT apply to council tallying. Ancestor council mutation
(`:1531-1532`) is the next follow-up, not merely out of scope: the
≥1-member creation rule exists precisely because this rescue path is
absent, the treasury PR built all its machinery (proper-ancestor
validation, ancestor-hosted proposals), and with treasuries live a
root DAO whose council goes silent strands its funds (sub-DAO funds
stay clawback-rescuable). Related tension recorded, no code change:
`:250-253` reserves assigned-but-unused funds absent a Constitutional
Amendment — an off-chain restriction the clawback power cannot
machine-check.

## Decisions

### Council is a plain `*addrset.Set`

Deleted: `MemberStorage`/`MemberGrouping`/`MemberGroup` (+options,
+tests), the `exts/storage` broker mirror, and the quarantined
`exts/definition` quorum helper — ~2,600 lines. The constitutional
primitive is exactly an address set; subdivisions are sub-DAOs, not
member groups. commondao deliberately does **not** depend on
`p/nt/groups` (role machinery is dead weight here; consumers needing
roles use groups directly, as boards does).

### Boundary-enforced mutability

`CommonDAO` keeps narrow exported mutators (`UpdateCouncil`,
`Dissolve`, `Propose`, `Vote`, `Execute`, `Withdraw`) and
the security rule is the groups-package idiom: **the hosting realm
never returns or accepts `*CommonDAO`**; reads cross realm boundaries
via recursive readonly views (`ReadonlyCommonDAO`, `ReadonlyProposal`).
The alternative (package-private mutators + `/p/`-hosted builtin
definitions) was rejected: its "no dangerous exported
mutators" invariant is unauditable (the dangerous surface includes
`Vote(member,…)` spoofing, `Children()` returning a mutable list, and
proposal-storage handles) and it would force realm-param
authentication sprawl inside a data library. The former leak points were
closed: `Get()` → ungated `GetView()`; creation entry points return
the DAO ID (and seed the council at creation); the owner/`GetOptions`
surface referenced here was later removed (see Consequences).
`ReadonlyProposal` flattens
`Title()`/`Body()` and **never exposes `ProposalDefinition`** — a
definition's `Executor()` is a bound method over realm state, so
exposing it would let any holder execute pending proposals under the
hosting realm's authority (borrow rule #2).

### Tally denominator (spec `:1521`)

Adopted reading: **D = electorate size − abstains** — everyone
entitled to vote minus those who abstained. Silence counts against
passage; abstention is deference. Two rival readings rejected:

- *Votes-cast-so-far*: the first YES passes anything instantly —
  incoherent with "decided immediately" (`:1522-1524`).
- *AtomOne hybrid* (deadline tally over YES+NO cast; prospective
  irreversibility for the immediate arms): its immediate-arm formulas
  are algebraically identical to ours, diverging only at the deadline
  (council of 9, 4Y/1N/4 silent: hybrid passes, ours dismisses).
  Rejected because (a) the spec defines one denominator that must
  serve both the immediate arms and the deadline; (b) with quorum
  deliberately absent, an electorate-derived denominator is the only
  apathy-safe default (under the hybrid a 100-member council passes a
  treasury spend on a single YES); (c) the Gno spec deliberately
  diverged from its AtomOne source (added "total", dropped "(no
  quorum requirement)", made both arms symmetric-immediate).

The rule, integer math only (`TallyDefault`):

    D = |E| - abstain        E = electorate snapshot; only E's votes count
    supermajority   pass ⇔ D > 0 && 3*yes >= 2*D      (":1511 two thirds or more" ⇒ >=)
    simple majority pass ⇔ D > 0 && 2*yes > D          (":1510 more than half" ⇒ >)
    dismiss (both)       ⇔ 2*no > D
    undecided at deadline ⇒ Dismissed

The `D > 0` guard is load-bearing: electorates can be all-abstain, and
without the guard `3*0 >= 2*0` would pass proposals with zero YES
votes. Pass is checked before dismiss (spec clause order); under
`yes+no ≤ D` the pairs are mutually exclusive — supermajority:
pass∧dismiss ⇒ `yes+no > 7D/6 > D`; simple: ⇒ `yes+no > D`. Both
contradict the invariant, which the electorate snapshot guarantees.

### Electorate snapshot at Propose

Each proposal copies the council into an unexported electorate set at
`Propose`; `Vote` gates on it and `TallyDefault` counts only its
members' votes. A live denominator is unsound: a council update
executing mid-vote could inflate the numerator with removed members'
recorded votes, drive D negative (auto-passing everything with zero
YES), break the exclusivity invariant, and decide proposals with no
vote event to trigger evaluation. Consequences: members added
mid-vote vote on the next proposal; members removed or resigned
mid-vote stay in old snapshots (their silence counts against
passage). Snapshots cost O(council) per active proposal — bounded by
the proposal cap; finished proposals retain their voting records and
snapshots indefinitely (archival is future work).

### Early termination (spec default)

Proposals are re-evaluated after every recorded vote (including vote
changes — `AddVote` overwrites, and a NO→ABSTAIN flip shrinking D may
legitimately trigger a pass): a YES tally at the definition's
threshold passes immediately (the proposal stays in active storage
until executed), and a NO majority dismisses immediately. `:1522`'s
"decided immediately by a supermajority" is read as "at the decision
rule's threshold": a sub-DAO creation proposal (`:1504` simple
majority) also passes early at its own threshold — one denominator
rule, one early-decision rule, per-type thresholds. `Execute` runs
early-passed proposals at once — skipping the re-tally and deadline
gate but still validating — and the deadline path dismisses undecided
proposals before validation (a proposal that never passed archives as
Dismissed, never Failed). `Threshold()` is part of `ProposalDefinition`
itself — every proposal is decided by the constitutional rules, checked
at compile time; `Proposal.ExpectedOutcome()` serves rendering.
Statuses:
`StatusRejected` → `StatusDismissed`; `StatusFailed` is reserved for
validation/executor errors. Realm-side, past-deadline `Execute` is
permissionless (the tally is deterministic); pre-deadline execution of
early-passed proposals keeps the council-member gate (the
`AllowExecution` flag was later removed — see the superseding note in
Consequences).

### Built-in thresholds (quorum/plurality removal is required)

CouncilUpdate: supermajority (`:1529-1530`). SubDAO creation:
**simple majority** — `:1504-1505` grants it explicitly. Text and
dissolution: the supermajority default. The old quorum knobs and
plurality selection had no legitimate home: the Governing Documents
that may override the defaults are exhaustively listed (`:1494-1496`)
and realm code is not among them. (Non-spec-bound DAOs would regain
full freedom through an additive custom-tallier hook if a consumer
ever materializes; see Alternatives.)

### Lifecycle rules

- `UpdateCouncil(add, remove)` applies idempotent set operations
  (final set `(council ∪ add) \ remove`; overlap rejected), so
  concurrently passed updates merge deterministically in execution
  order and full council replacement in one proposal is legal. An
  update that would empty a non-empty council returns an **error** —
  never a panic, which would revert the transaction and strand the
  proposal Passed forever; the executor propagates it and the
  proposal fails cleanly.
- DAO creation requires ≥1 council member; `Resign` refuses the last
  member. (Ancestor governance — the spec's rescue path for empty
  councils — is deferred, so an empty council would be permanently
  inert.)
- `Dissolve` dismisses every in-flight proposal (reason "DAO
  dissolved") before soft deleting; deleted DAOs reject `Propose`,
  `Vote`, `Execute`, and `Resign`. Dissolution is terminal.
- `Propose` is capped (`MaxActiveProposals`, default 32, floor 1) —
  every active proposal carries a snapshot. `CapExempt` definitions
  (council updates) bypass the cap but are bounded to one active
  exempt proposal per creator: a member who fills the cap with junk
  must never be able to block their own removal.
- Membership validation was removed from council-update proposals by
  design (idempotent semantics): adding an existing member or
  removing a stranger is a legal no-op, so concurrent updates cannot
  invalidate each other.
- `Execute` removes the proposal from active storage **before** any
  definition code (`Validate`, `Tally`, the executor) runs, so a
  re-entrant `Execute` from inside an executor finds no proposal and
  cannot run it twice.

## Alternatives considered

- Live electorate with vote pruning inside every council-mutating
  executor: strictly more code, path-dependent outcomes; rejected for
  the snapshot.
- Keeping `IsQuorumReached`/plurality for the built-ins: not
  permitted by the Governing-Documents rule above.
- `/p/`-hosted built-in definitions with package-private mutators:
  rejected by a 2/3 review vote for boundary-enforced mutability (see
  "Boundary-enforced mutability" above).
- A `CustomTallier` escape hatch (custom tally logic evaluated at the
  deadline) and `CustomizableVoteChoices` shipped in early revisions:
  removed by a design-simplicity review as consumer-less speculation —
  custom choices also persisted a per-proposal choice tree. A follow-up
  review folded `Threshold()` into `ProposalDefinition` itself; a future
  custom tallier is reintroduced additively as an optional interface
  checked before the default rules.

## Consequences

- Net −2,600+ lines; float-free deterministic tallying; the tally,
  snapshot, and boundary rules are pinned by a package test suite
  (14-case tally boundary table, snapshot gating both directions,
  early-termination lifecycle, cap + exemption, would-empty-fails-
  cleanly) and a realm filetest suite (92 at branch HEAD) including
  render goldens that pin both escaping of user-controlled text
  (names, bodies, vote reasons) and correct formatting of the built-in
  proposals' markdown bodies, plus a dissolved-parent sub-DAO
  rejection. The treasury follow-up ADR adds more on the same branch.
- Breaking API changes throughout (quarantined realm; not deployed to
  any chain — no live state exists).
- The realm's crossing functions deliberately do **not** open with
  `cur.IsCurrent()` guards: per `gnovm/adr/interrealm_v2.md`, crossing
  functions do not require the check because the runtime ensures it is
  always true — the guard exists for the `(_ int, rlm realm)`
  *non-crossing* borrowing pattern, which this realm never uses (its
  one authority relay, `dao.Execute(id, cur)`, passes the realm's own
  genuine `cur`). AGENTS.md's blanket always-guard rule predates this
  distinction and should be reconciled with the interrealm ADR.
- The realm's `New` accepts description and members parameters. The
  invite check deliberately uses `unsafe.OriginCaller()` (invitations
  target EOAs) — confirmed intended, pre-existing behavior. `New` takes
  a description so the Charter (`:1485`) has a creation-time slot.
- **Superseded — see `pr6012_commondao_ownership_rescope.md`.** This PR
  originally shipped an owner/`Options` host layer (`AllowListing`,
  `AllowRender`, `AllowChildren`, `AllowExecution`, an owner-tunable
  cap, `GetOptions`/`UpdateOptions`/`TransferOwnership`). A later design
  review on the same branch removed it entirely: authority now follows
  only the spec's two axes (a DAO's own council; its parent chain), the
  sole surviving host bit is an opt-in `listed` flag, and creation
  gating moved to an origin-keyed `creators` set. The paragraphs below
  and any earlier reference to `GetOptions`/`UpdateOptions`/owner gates
  describe the removed layer.
- Sub-DAOs of a dissolved parent are currently un-dissolvable (their
  dissolution proposal is hosted in the deleted parent, which rejects
  proposals); the treasury PR's nearest-live-ancestor rescue covers
  this, and no funds are at stake before treasuries exist.
