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
gated by the proposing DAO's `AllowTreasuryProposals` option (default
off, like every proposal-type flag). All spend preconditions live in
`Validate`, which runs at proposal creation and again inside `Execute`
immediately before the executor: a treasury frozen, drained, or
dissolved after the proposal passed fails it cleanly (`StatusFailed`,
coins untouched) instead of panicking the tx and stranding it Passed.
Concurrent over-committing spends therefore fail in execution order
(pinned by z_11_f). The executor mints the DAO's sub-identity and
sends via `banker.NewBanker(BankerTypeRealmSend, sub)` — the only
banker type permitted for sub-identities; banker sends move bank-keeper
balances without invoking recipient code, so there is no reentrancy
vector on top of Execute's remove-before-run rule.

### Ancestor controls (clawback, freeze)

Generalized from the Core-DAO wording (`:245-248`) to the whole tree
under `:1507` — a deliberate plan choice. Both are decided at **simple
majority** and validated as **strictly proper** ancestry
(`target.Parent()` walked upward; self-targeting rejected), so a DAO
can never claw back or unfreeze itself. A target's own Options can
never block an ancestor (the gate is on the proposing DAO only).

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
  (z_14_b).
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
root is dissolved, the orphan is unreachable by governance — no live
council has authority over it; documented, accepted.

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
  (breaking; quarantined realm, no live state). `Options` gains
  `AllowTreasuryProposals` (default off).
- Package surface: `WithAddress`, `Address()`, `SetTreasuryFrozen`,
  `IsTreasuryFrozen`; the getters (never the mutator) are mirrored on
  `ReadonlyCommonDAO`.
- Render: DAO pages show the treasury address, live balances
  (readonly banker), and a frozen warning; the genesis DAO enables
  treasury and dissolution proposals so donated funds stay governable
  (its options are otherwise immutable — the owner is the realm's own
  address).
- Validation and rendering read the stored address while executors
  spend from the freshly minted sub-identity; both flow from the single
  `newDAOOptions` construction funnel (`WithAddress(daoAddress(id))`)
  and z_15_a pins the equality on-chain.
- 25 new filetests (z_11 spend lifecycle/gates, z_12 clawback incl.
  the grandparent case pinning the fixed parent destination, z_13
  dissolution sweeps + orphan rescue, z_14 freeze, z_15 address
  stability, z_10_d treasury rendering) plus definition unit tests; no
  funds can move without a passed proposal — the only `SendCoins` call
  sites are the three executors above.
- The filetests run against the in-memory `TestBanker`, so
  `gno.land/pkg/integration/testdata/commondao_treasury.txtar` proves
  the production path end to end on a real node: a real bank send
  funds the derived dao/1 address, an overdrawn spend is rejected at
  creation, a passed spend debits the DAO account and credits the
  destination through the real bank keeper, and a parent clawback
  sweeps a funded child treasury DAO-to-DAO (source account emptied,
  parent credited). The test patches the genesis council to the test
  key and switches `AllowTreasuryProposals` on in `defaultOptions`
  (the realm option is variadic-only, unreachable from `maketx call`).
