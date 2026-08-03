# ADR: commondao per-DAO identities + treasury

## Context

Follow-up to the commondao Council ADR. The Common DAO Spec
(`docs/CONSTITUTION.md`, Appendix) gives every DAO its own address
minted from the hosting realm (`:1534-1537`) and a treasury that only
passed proposals may move (`:1542-1543`); the parent-ownership clause
(`:1507`) grounds ancestor controls over descendant treasuries. This PR
adds both to `p/nt/commondao/v0` and its (quarantined) reference realm.

## Decisions

### Identity: derived addresses, not persisted realm values

Each DAO's treasury address is `chain.DerivePkgSubAddr(pkgPath,
"dao/<id>")` — the address of the realm sub-identity
`cur.Sub("dao/<id>")` (interrealm ADR `pr5890_realm_sub.md`). The
package stores only the derived `address` (`WithAddress`; sub-realm
tokens cannot be persisted by design), and never calls `Sub` itself:
realm-side executors mint the sub-identity at execution time from
their own live `cur`. Addresses are collision-free (DAO IDs are
monotonic, never reused) and stable across logic upgrades and
un-quarantining (the `pkgPath` const is the same either way). The
package also carries a `treasuryFrozen` flag (narrow
`SetTreasuryFrozen` mutator per the boundary rule; the handle never
leaves the realm) that the realm enforces — the package never moves
funds.

### Spend (`:1542-1543`)

`CreateTreasurySpendProposal(daoID, to, denom, amount)` — single-coin,
supermajority (the default passage rule; no spend-specific mandate),
available to every council like the other constitutional powers. All
spend preconditions live in
`Validate`, which runs at proposal creation and again inside `Execute`
immediately before the executor: a treasury frozen, drained, or
dissolved after the proposal passed fails it cleanly (`StatusFailed`,
coins untouched) instead of panicking the tx and stranding it Passed.
Concurrent over-committing spends therefore fail in execution order
(pinned by z_11_f). The host `Execute` wrapper mints the DAO's
sub-identity and passes it in; the executor sends from that sub via
`banker.NewBanker(BankerTypeRealmSend, sub)` — the only
banker type permitted for sub-identities; banker sends move bank-keeper
balances without invoking recipient code, so there is no reentrancy
vector on top of Execute's remove-before-run rule.

### Ancestor controls (clawback, freeze)

Generalized from the Core-DAO wording (`:245-248`) to the whole tree
under `:1507` — a deliberate plan choice. Both are decided at **simple
majority** and validated as **strictly proper** ancestry
(`target.Parent()` walked upward; self-targeting rejected), so a DAO
can never claw back or unfreeze itself. A target can never block an
ancestor's controls.

- **Clawback** sweeps the target's full balance at execution time to a
  **fixed, non-nameable destination: the target's parent** — funds move
  one step up the tree toward their origin. A nameable destination
  would let an ancestor drain a descendant out of the tree entirely,
  which `:245-253` does not authorize. Clawback stays valid against
  soft-deleted and frozen descendants (coins landing on a dead DAO
  remain rescuable — z_12_d). Root DAOs are unclawbackable by
  construction (no ancestor exists). Chaining through a hostile
  intermediary: freeze the middle first (its standing Passed spend then
  fails at execution — z_11_e), then claw leaf→middle→you, one level
  per proposal cycle.
- **Freeze** sets the per-DAO flag; it does not cascade (ancestors
  freeze each descendant explicitly), and only a proper-ancestor
  proposal can unfreeze — the frozen council cannot free itself
  (z_14_b). The flag lives on the target, so any live proper ancestor
  can unfreeze after the freezing ancestor dissolves (z_14_e).
  **Freeze blocks every governed path out of the treasury.** The flag is
  checked on *every* way the target's own council can move its funds, not
  just the treasury-spend kind: the realm's execution kind (arbitrary
  `ExecFunc` run as the DAO's own sub) is freeze-gated the same way, at
  proposal creation and again inside `Execute` (z_20_g rejects the
  create; z_20_h fails a standing passed execution proposal cleanly once
  the target is frozen, funds untouched). Without this, a frozen DAO
  could drain its own treasury through an execution proposal.

  **Known limitation — freeze is not containment against a DAO that has
  run an execution proposal.** Freeze is realm state consulted by realm
  code. An execution closure receives the DAO's `sub` and can mint
  `banker.NewBanker(BankerTypeRealmSend, sub)`; that banker is a plain
  struct with no realm reference, so it persists across transactions
  (intended chain behavior — see `chain/banker`'s package comment and
  `banker_persistence.txtar`) and `SendCoins` re-checks only
  `pkgAddr == from`. A closure that moves nothing but *retains* the
  banker leaves behind a permanent, unrevocable bearer capability over
  the treasury address, usable later with no proposal and no vote, and
  reaching the bank keeper without re-entering realm code — so the
  freeze flag never runs. It equally survives council replacement,
  deregistering the execution kind, and dissolution. Setting it up costs
  two supermajorities (register `execution`, then pass one proposal) and
  reaches only that one DAO's address (`pkgAddr` is pinned), but that is
  precisely the rogue-council case freeze exists for. Enabling the
  execution kind should therefore be understood as waiving the freeze
  guarantee for that DAO. Closing it needs a mechanism outside this
  realm's reach (a non-persistable capability shape, address rotation,
  or a bank-level account freeze); recorded here rather than papered
  over. What freeze does **not** block is an ancestor's
  action *on* the frozen target: clawback and dissolution stay valid
  against a frozen descendant (freeze→clawback is the intended flow), so
  the freeze gate is applied only to the target's self-initiated
  movement, never to clawback/dissolution. **Orphan
  rescue**: when every proper ancestor is dissolved the freezing
  authority class is extinct, so — and only then — the target's own
  council may pass an unfreeze on itself (never a freeze), restoring
  the constitutional default and unlocking its funds (z_14_f). Without
  this, dissolving a frozen descendant's whole ancestor chain locked
  its treasury forever. The invariant preserved: no live ancestor's
  freeze can ever be undone by the target.
- **No re-parenting invariant**: parent pointers are set only at
  construction (`WithParent`, used only by `createSubDAO`) and no
  re-parenting path exists. A future re-parent feature would be a
  clawback-authority risk and must revisit this ADR.

### Dissolution × treasury

Without a sweep, dissolving a funded DAO strands its coins (no council
remains to pass a spend). The dissolution executor sweeps the
remaining balance — sub-DAOs to the **parent** (fixed, not nameable:
the same place a clawback would put the funds), root DAOs to a
**destination required at proposal creation**. The asymmetry is
deliberate: root dissolution is the DAO's own council at supermajority
disposing of its own funds (authority-equivalent to a spend with an
arbitrary recipient), while a sub-DAO sweep is adjacent to ancestor
clawback power. Both shape rules are enforced at construction and are
state-independent (parents never change; checking funded-ness instead
would be bypassable by donating dust mid-vote). A frozen DAO can still
be dissolved. Orphans below a dissolved middle DAO stay dissolvable:
`CreateDissolutionProposal` hosts the proposal in the **nearest
non-dissolved ancestor** (z_13_d). If every ancestor including the
root is dissolved, the orphan can no longer be dissolved or clawed
back — but its own council keeps full spend power (spend proposals are
self-hosted) and, if frozen, the orphan-unfreeze rescue above applies,
so a **living** orphan's funds are never stranded. Coins deposited to a
DAO's address *after* it is dissolved are a separate, accepted case: a
deleted DAO rejects Propose, so only a live proper ancestor can recover
them (clawback-on-deleted); a dissolved **root** has no ancestor, so its
later deposits are unrecoverable by design (burned). Render warns on a
dissolved DAO, and specifically flags the unrecoverable case when no
live proper ancestor remains.

### Denom-flood gas-bomb (known limitation)

> **Partly resolved.** Non-gas balances now live in their own store keys
> (`tm2/adr/prxxxx_realm_denom_balance_keys.md`), so an account's own transactions
> no longer get more expensive as junk denoms accumulate — the gas-bomb below can
> no longer be aimed at a treasury from outside. Two caveats: the bank's
> single-denom read is Go-only, so a realm still has to call `banker.GetCoins`
> and pay O(number of denoms held); and the dissolution sweep now writes one key
> per denom rather than one blob, so it got *more* expensive. The analysis below
> still applies to every path that enumerates.

The bank keeper stores each account's coins as **one amino blob**, read
and rewritten whole on any movement (`GetAccount`/`SetAccount`), and store
gas is charged per byte of that blob (`ReadCostPerByte`/`WriteCostPerByte`
× `len(blob)`). So **every** touch of a treasury account is
O(number-of-denoms) — not just the clawback/dissolution sweep (which moves
the full multi-denom balance in one `banker.SendCoins`), but also a
single-coin spend, whose `Validate` reads the account (`GetCoins`) at both
propose and execute and whose executor reads and rewrites the source
account in `SendCoins`.

Coin denominations are realm-mintable without limit and any address can be
funded permissionlessly, so an attacker can dust a DAO's (public,
pre-derivable — sequential ID) address with enough distinct `pkgpath:name`
denoms that **any** fund-movement tx exceeds the per-tx/block gas limit —
after which spend, clawback, **and** dissolution all abort on out-of-gas
and can never complete, locking the DAO's funds and its governance
entirely. (A read-only balance query costs the same O(#denoms), so even
`treasuryBalance` render becomes expensive.)

There is no theft and no atomicity break (out-of-gas reverts the whole tx;
the proposal stays Passed and re-executable). **Correction to an earlier
draft: single-coin spends are NOT immune** — the whole-account-blob read
and write makes them the same order as the sweep, so the flood locks the
DAO completely rather than leaving spends as an escape hatch (verified: a
single-coin spend's account read/write is O(#denoms), within ~5% of a
full sweep). The root cause is a chain-wide missing per-account
denom-count cap, not a commondao-specific authorization flaw, and there is
no commondao-side mitigation for the read (the bank exposes no
single-denom account read — `GetCoins` materializes the whole account).
Mitigations therefore live at the chain level (per-account denom cap) or
in a follow-up that batches the sweep across executions / bounds it to a
denom allowlist. Recorded as a known limitation.

## Alternatives considered

- Amount-bearing clawback: rejected — full-balance sweep sidesteps the
  insufficient-balance case and matches "return to the origin
  Treasury".
- Executor-side precondition checks (deleted/frozen/balance in the
  executor instead of Validate): rejected — Validate already runs
  inside the same Execute call right before the executor, so
  duplicating the checks is dead code; one check point keeps the
  failure mode uniform (`StatusFailed` with a greppable reason).
- Multi-coin spend parameters: deferred — a `denom/amount` pair is
  CLI-friendly; several coins are several proposals.

## Consequences

- New realm surface: `CreateTreasurySpendProposal`,
  `CreateTreasuryClawbackProposal`, `CreateTreasuryFreezeProposal`,
  and a `destination` parameter on `CreateDissolutionProposal`
  (breaking; quarantined realm, no live state).
- Package surface: `WithAddress`, `Address()`, `SetTreasuryFrozen`,
  `IsTreasuryFrozen`; the getters (never the mutator) are mirrored on
  `ReadonlyCommonDAO`.
- Render: DAO pages show the treasury address, live balances
  (readonly banker), and a frozen warning. Treasury and dissolution
  proposals are always available to councils, so funds donated to any
  DAO (including the genesis DAO) stay governable.
- Validation and rendering read the stored address while executors
  spend from the freshly minted sub-identity; both flow from the single
  `newDAOOptions` construction funnel (`WithAddress(daoAddress(id))`)
  and z_15_a pins the equality on-chain.
- Treasury filetest families (z_11 spend lifecycle, z_12 clawback
  incl. the grandparent case pinning the fixed parent destination,
  z_13 dissolution sweeps + orphan rescue, z_14 freeze, z_15 address
  stability, z_10_d treasury rendering, z_20_g/z_20_h execution-kind
  freeze gate at create/execute) plus definition unit tests; no funds
  can move without a passed proposal — the `SendCoins` call sites are the
  three treasury executors above, and any arbitrary-execution closure the
  realm's execution kind runs (itself freeze-gated).
- The filetests run against the in-memory `TestBanker`, so
  `gno.land/pkg/integration/testdata/commondao_treasury.txtar` proves
  the production path end to end on a real node: a real bank send
  funds the derived dao/1 address, an overdrawn spend is rejected at
  creation, a passed spend debits the DAO account and credits the
  destination through the real bank keeper, and a parent clawback
  sweeps a funded child treasury DAO-to-DAO (source account emptied,
  parent credited). The test patches the genesis council to the test key; the sub-DAO
  itself is created through the proposal flow on the real node.
