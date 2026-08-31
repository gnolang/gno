# ADR: commondao ownership re-scope

## Context

The realm carried an owner/`Options` layer bolted onto a spec whose only
control primitives are the Council and the parent→sub-DAO tree
(`docs/CONSTITUTION.md`, Appendix). Three independent design reviews found it
incoherent, and one fault was a defect:

1. **Root owner controlled the whole sub-tree.** `isOwner` resolved via
   `TopParent()`, so a sub-DAO's `Options` answered to the *root's* owner key —
   a second authority axis contradicting the spec's parent-controls-child model
   (`:1507`), which the treasury clawback/freeze/dissolution design correctly
   follows one level at a time.
2. **Flags leaked into governance; one silently vetoed a passed proposal.**
   `AllowChildren` was re-checked in `subDAOPropDefinition.Validate()`, which
   runs at execution — so an owner flipping it off *after* a sub-DAO proposal
   passed forced that decided council vote to `StatusFailed`, contradicting the
   `Options` docstring's own claim that governance "cannot be switched off."
   `AllowExecution` (owner delays decided proposals) and the owner-set proposal
   cap (owner pins it to 1, no council override — the "jammed cap" trap) were
   milder versions.
3. **No renounce; a lost owner froze `Options` forever.** The genesis DAO,
   owned by the realm's own uncallable address, ran fine — proof the owner was
   optional scaffolding.

Governance and treasury were never owner-gated, so the blast radius was
bounded; this was a spec-fidelity and design-clarity problem.

## Decision

Reduce authority over a DAO to exactly the spec's two axes — the DAO's own
Council and its parent chain — and confine the host layer to *how this realm
presents a DAO in its own UI*.

- **Remove the owner concept entirely.** Deleted the `ownership` and `options`
  bptrees, `getOwnership`/`getOptions`, `isOwner`/`assertIsOwner`/`hasOwnership`,
  and the exported `GetOptions`/`UpdateOptions`/`IsOwner`/`TransferOwnership`.
  `RenounceOwnership` is moot — there is nothing to renounce. Fault #1 and the
  dual-authority model die by construction.
- **Remove `AllowChildren`.** Sub-DAO creation is a `:1504` Council
  simple-majority right; the vote to pass `CreateSubDAOProposal` *is* the
  decision. Deleting the flag also removes the execution-time veto (fault #2's
  defect).
- **Remove `AllowExecution`.** Early-passed proposals are always executable
  (pre-deadline: a Council member; post-deadline: permissionless), consistent
  with the early-termination rule "decided ⇒ final" (`:1522-1524`). The dropped
  "respect the full voting window even after early-pass" policy has no spec
  basis.
- **Fix the proposal cap at `DefaultMaxActiveProposals = 32`.** The cap is a
  storage-safety bound (every active proposal stores an O(council) electorate
  snapshot), not a governance parameter; the realm no longer tunes it, which
  removes the jammed-cap trap. `CapExempt` council-updates still bypass the cap,
  so a full queue can never block a member's own removal. The package keeps
  `SetMaxActiveProposals` as a library seam; the realm simply never calls it.
- **Remove `AllowRender`.** All DAO state is public on-chain, so a "rendering
  not enabled" page is pointless friction; pages always render.
- **`Options` → a single `listed` bit.** The only surviving host concern is the
  home-index listing (a spam surface), kept opt-in (default off) and flipped by
  `SetListed(cur, daoID, bool)`, gated on Council membership of *that DAO* like
  `Resign` — listing is cosmetic and reversible, so a single member may toggle
  it. `IsListed` reads it. `Options` struct, `defaultOptions`, and `options.gno`
  are deleted.
- **Invite / creation gating re-homed to a `creators` set, keyed on the
  transaction origin.** With `ownership` gone, the repeat-creation skip moves to
  a dedicated `creators` bptree marked when an invite is consumed. It keys on
  `unsafe.OriginCaller()` — matching how invites are issued and consumed
  (invitations target EOAs) — rather than the immediate caller. This also closes
  a latent vector in the prior caller-keyed skip, where any realm that created
  one DAO became an unlimited DAO factory for un-invited origins.

## Alternatives considered

- **Keep a slim per-DAO owner for listing** (+ `RenounceOwnership`): rejected —
  it re-introduces a second authority principal and an ownership lifecycle for a
  single cosmetic bit that the DAO's own Council can set directly.
- **Council-tunable cap** via a proposal type: rejected — adds a proposal type
  and a hard storage ceiling to tune a number that 32 already covers; the cap is
  a chain-storage bound, not a governance knob.
- **Listing via a simple-majority proposal** rather than a single-member setter:
  a 2/1 review split; the setter won as proportionate to a reversible UI toggle
  and to avoid consuming a (now fixed-32) proposal slot for cosmetics.
- **Keying the creators set on the caller** (preserving the exact prior
  behavior): rejected — it preserves the unlimited-factory vector; origin-keying
  is more spam-resistant and was verified not to break the multi-DAO
  builder-realm test (its origin is a constant EOA across creations).

## Consequences

- **Removed exported realm surface:** `GetOptions`, `UpdateOptions`, `IsOwner`,
  `TransferOwnership`, and the `Options` type. **Added:** `SetListed`,
  `IsListed`. **Changed:** `New` (invite-skip re-keyed to origin via `creators`;
  no owner assignment); `Execute` (dropped the `AllowExecution` gate, kept the
  pre-deadline Council-member check); `CreateSubDAOProposal` (dropped the
  `AllowChildren` gate); `createDAO` (dropped its `owner` parameter). All
  breaking — the realm is quarantined, no live state.
- **Treasury invariants untouched:** all three reviews verified clawback,
  freeze, dissolution, and orphan-rescue key only off parent pointers and
  `IsDeleted`/`IsTreasuryFrozen` — never off owner/`Options`. The treasury
  integration txtar passes unchanged.
- **Package cleanup:** removed the now-unused `TopParent()` whole-tree
  resolver — the mechanism behind fault #1 (root-owner-controls-subtree). No
  authority path ever jumps to the root: control resolves only via one-level
  `Parent()` and proper-ancestor walks (`assertIsProperAncestor`). A later
  three-reviewer pass over the ownership+parent model confirmed the authority
  map collapses to exactly two axes (a DAO's own Council and its ancestor
  chain) with no residual owner axis.
- **Tests:** deleted 10 filetests whose sole intent was a removed behavior
  (owner transfer, per-DAO flag toggling, owner cap-tuning + the jammed-cap
  trap, the `AllowChildren`/`AllowExecution` gates); rewrote `z_6_g` to pin
  permissionless past-deadline execution without the removed flag; added
  `z_16_a`/`z_16_b` for `SetListed` (member toggles on/off; non-member
  rejected). The multi-DAO-per-invite tests (`z_4_a`, `z_15_a`) pass unchanged
  under the `creators` re-home. Settings-page goldens shrink from the four-flag
  Options table to a small Info table (Listed, Max active proposals; a
  Proposal-kinds row was later added by the proposal-kind registry).
