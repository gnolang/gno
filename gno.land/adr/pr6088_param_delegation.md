# ADR: delegating one chain parameter from GovDAO to another realm

Stacked on the MsgRun allowlist work; see
[pr6088_msgrun_allowlist_and_inert_charging.md](./pr6088_msgrun_allowlist_and_inert_charging.md)
for `run_submitters` itself. This document covers only the delegation
mechanism.

## Context

`run_submitters` is an address allowlist gating `MsgRun`. Adding an address
requires a GovDAO supermajority, which is the wrong cost for routine
whitelisting: the decision is a membership judgement, not a policy change.

The goal is to let GovDAO hand that one job to another DAO, revocably, without
handing over anything else.

### The shape is forced, not chosen

Three constraints leave exactly one design.

**Parameter writes are physically confined to one realm.** The `sys/params`
stdlib checks its caller's package path in the VM and refuses anything that is
not `gno.land/r/sys/params`. The getters carry the same check. So the capability
cannot be granted to another realm at all — "delegating a parameter" can only
mean adding an authorized path inside `r/sys/params`.

**The allowlist cannot move out of the parameter store.** It is a parameter
because the ante handler needs it, and the ante cannot execute VM code:
`getGnoTransactionStore` is an unchecked type assertion on a context value that
`beginTxHook` installs *after* the ante, and CheckTx returns before that hook
runs at all, so calling it from an ante panics. Even if it were wired up, VM
execution in CheckTx is a per-node, non-consensus decision, so divergent local
state would produce divergent mempool admission. The rule in this tree holds
without exception: **parameters are read in the ante, realm state in the
keeper.** A realm-managed allowlist, which would have made this an ordinary
ownership question needing no new mechanism, is therefore not available.

**A delegate cannot be given a fast path through GovDAO instead.**
`isValidCall` admits only `prev.IsUser()` or the origin's own `/e/<addr>/run`
realm, so no realm can submit a GovDAO proposal on its own behalf. Lowering a
quorum would not help either: v3 has no voting period, no quorum and no
deadline — proposals stay open until members act — so "slow" means attention,
not a timer.

## Decision

### A named slot, not a registry

```gno
var runSubmittersMgr string // "" == not delegated
```

The first design was a registry mapping a parameter key to a delegate path,
with a compile-time allowlist of which keys could be delegated. It was
discarded, and the reason generalizes.

`r/sys/params` can never be redeployed — `AddPackage` refuses an occupied path
— and the stdlib gate names that exact path. So making a **new** parameter
delegatable already requires editing this realm's source, which means a chain
relaunch. A registry's runtime generality would therefore never be exercised
beyond the keys already blessed in source; the only freedom that matters is
"which realm, or none", per blessed key. A `var` says exactly that.

What the simpler shape removes, rather than merely simplifies:

- **No key string to mis-compose.** The realm and the stdlib each had their own
  key-composition function, only one of which validated. Operating on a fixed
  key removes the second source of truth.
- **No container to leak.** A registry wants an accessor, and returning an
  `avl.Tree` or `bptree` pointer would let any caller invoke its exported `Set`
  — borrow rule #2 fires and the write commits under this realm's authority.
  Self-registration as a delegate.
- **No key the author never considered.** A shape-based allowlist of
  `<module>:p:<name>` — which is what a registry naturally wants, since
  enumerating keys defeats the point — would have admitted
  `bank:p:restricted_denoms` (unilateral denom freeze) and
  `auth:p:unrestricted_addrs`.

Cost: a second delegated capability is roughly fifteen lines rather than one
allowlist entry. Both need a relaunch regardless.

### Not a capability object

Nothing is handed to the delegate. Authority is re-checked on every crossing
call, so clearing the slot takes effect immediately. The contrast is
`banker.NewBanker`, where authorization happens at construction and the
resulting object can be retained across transactions — an unrevocable grant.

### The gate, shared with valset

```gno
func assertDelegate(_ int, rlm realm, want, subject string) {
	if want == "" { panic("unauthorized: no delegate is configured for " + subject) }
	if !rlm.IsCurrent() { panic("unauthorized: rlm is not the caller's live cur") }
	if rlm.Previous().PkgPath() != want { panic("unauthorized: only " + want + " may write " + subject) }
}
```

`assertValsetCaller` now calls this, so the realm has one authorized-caller
shape instead of two that can drift.

Two properties are load-bearing rather than defensive.

**Empty denies, and the check comes first.** A direct call from a user account
has an empty previous package path — that is precisely what `IsUserCall` tests
— so comparing it against an unset slot would compare `""` to `""` and admit
every account on the chain. `r/gov/dao`'s own allowlist has the mirror-image
fail-open, kept deliberately for genesis bootstrap; nothing here needs that.

**Matching is exact, never a prefix.** A sub-realm identity minted by
`cur.Sub(subpath)` presents the synthesized path `host#subpath`, so a single DAO
hosted by a multi-tenant realm can hold the capability — for example
`gno.land/r/nt/commondao/v0#dao/42`. This is what makes per-DAO delegation
possible at all, and it is unforgeable by construction: `#` is rejected in every
real package path, and `Sub` requires the minting realm's own path as the host,
so no other realm can present a `commondao/v0#…` identity. But the anchored
idiom `p == host || strings.HasPrefix(p, host+"#")` would hand the capability to
*every* DAO that host serves, and CommonDAO membership, while invite-gated, is
unlimited once invited.

### valset stays a compile-time const

It was tempting to make valset the first registry entry, since it already *is* a
delegation — one hardcoded realm authorized for one key family. It does not fit,
for two structural reasons.

The verbs differ: valset needs a whole-list set plus a boolean set, applied
atomically. And the policies are opposite in kind — this delegation is
*additive*, with GovDAO retaining full access, while valset is *exclusive*, with
`assertNotValsetKey` locking GovDAO out of the family entirely. One table cannot
mean both without an `exclusive` flag, and a vote-settable exclusive flag is a
mechanism for GovDAO to lock itself out of a parameter until a relaunch.

The loss would also be real: today re-pointing the valset writer requires a
relaunch, whereas a registry entry could be re-pointed by one supermajority at
an arbitrary realm, bypassing v3's valoper-existence check, its `KeepRunning`
opt-out and its execution-time pubkey re-resolution.

### Add, and remove only what you added

The delegate may add addresses, and may remove only addresses it added. GovDAO
retains everything, including the existing whole-list setter.

Add-only was the first choice and was wrong. A rogue delegate's worst act is
unbounded *adds* — which is what the allowlist exists to prevent — while
denial-by-removal is the lesser harm; and add-only leaves a delegate unable to
de-list its own mistake, which is an allowlist manager's core job.

Two bounds make that safe, and they close different holes.

**Grant-scoping.** An entry that predates the delegation is never removable by
the delegate, so a list GovDAO curated survives a rogue one.

**A non-empty floor.** `RemoveRunSubmitters` refuses any removal that would take
the list to zero, whatever the grant record says. An empty `run_submitters` means
the gate is *off* — anyone on the chain may `MsgRun` — so emptying it is not a
smaller version of removing one address, it is the unilateral revocation of the
entire restriction GovDAO voted for. Grant-scoping alone does not prevent that:
it holds only while at least one entry the delegate did not grant survives, and
GovDAO replacing the list wholesale can remove its own entries without touching
the grant record. The floor makes the invariant structural rather than emergent.
Only GovDAO can reach zero, through the whole-list setter, where the resulting
list is on the ballot.

That invariant did not hold in the first implementation, and the way it failed
is worth recording because the mechanism is not obvious. `UpdateSysParamStrings`
**dedupes on add**, so re-adding an address already on the list leaves the
parameter unchanged — but the first version recorded a grant for every argument
regardless. That let the delegate launder authority over entries it never
granted: read the list, re-add all of it (a no-op on the parameter, yet every
address now recorded as its own), then remove all of it — including whatever
predated the delegation. The floor above would refuse the last of those
removals, but only the last: everything up to it still goes through, leaving the
delegate holding the sole listed address and so sole authority over who may
`MsgRun`. Found by audit, with a runnable proof; the existing test missed it
because it attempted the removal without the priming add.

The fix is that the grant record follows the **parameter**, not the argument: an
address already present when `AddRunSubmitters` runs gets no grant. The test
that pins it performs the full three-step laundering sequence, and fails against
the unconditional version.

Provenance lives in realm state, not in the parameter, so the two can disagree —
genesis, `gnogenesis params set`, or a future direct keeper write produces
entries with no grant recorded. That direction is safe: an unrecorded address is
simply not removable by the delegate. The record is cleared whenever the
delegation changes hands, so a new holder never inherits authority over its
predecessor's grants.

**Revocation does not sweep.** Addresses the delegate added remain. A sweep
would make the executed effect invisible at vote time, and would silently remove
nothing whenever the grant record had drifted, leaving an operator believing
cleanup had happened. GovDAO's whole-list setter is the bounded reset, and it
shows voters the exact resulting list.

## Consequences

### The delegate holds an unmetered per-transaction knob

`run_submitters` is read by the ante handler on every transaction, before
`auth.SetGasMeter` installs the per-tx meter — so in DeliverTx the decode is
charged to a passthrough bounded by remaining *block* gas rather than the
sender's `GasWanted`, and in CheckTx to an infinite meter with no block bound,
repeated on every mempool recheck. `GetParams` bech32-decodes every entry.

So list length is an unmetered chain-wide constant, and the delegate can grow
it. It is bounded by `maxAddressListLen` (1000), added separately after this was
found; the bound applies here for free because `UpdateSysParamStrings` re-sets
the whole list and re-enters `WillSetParam` and `Params.Validate`, so both the
cap and bech32 validation cover a delegate's additions.

This is the strongest argument against `run_submitters` being the first
delegated key. `pkg_approvers` is read only in the keeper, costs nothing per
transaction, and gates a higher-volume judgement workflow — but it grants an
*irreversible* power, since enabling a package runs `init()` and materializes a
realm while `DisablePackage` is still a stub. Neither is obviously safer;
`run_submitters` is implemented because it is the one that was asked for.

### Delegating to a DAO trusts more than that DAO's council

Registering `host#dao/N` trusts the host realm's deployed executor code as well.
CommonDAO's `Execute` chooses the operative DAO from the proposal definition —
`op := daoID`, overridden when the definition implements `Funded` — so the host
*can* mint any of its subs. Today that is safe for this purpose, because
`FundingDAOID` is implemented only by treasury and dissolution definitions,
which use the sub solely for banker sends, and the arbitrary-code `execution`
kind does not implement it. That is a property of the deployed code, not a
language guarantee.

**Delegate to a root DAO.** A sub-DAO's council can be rewritten from above via
an ancestor council-update proposal, and its treasury frozen — and
`executionPropDefinition.Validate` fails execution while the treasury is frozen,
so an ancestor's freeze is a kill switch on the delegated capability. A root DAO
has no proper ancestor.

### No new proposal kind is needed, and adding one would be costly

The deployed `r/nt/commondao/v0` resolves kinds by name against a compile-time
catalog and deliberately refuses a foreign kind by value. But the catalog
already contains an opt-in `execution` kind that runs a caller-supplied closure,
registerable by a supermajority. So a DAO reaches this capability today via
`execution` plus a small policy realm — needed because a closure cannot be
encoded in a transaction, and because `CreateExecutionProposal` requires the
calling realm's own address to sit on the DAO's council.

Adding a genuinely new kind would require redeploying the realm at a new path,
which changes `cur.Sub` derivation and would strand every existing DAO treasury.

### The gate must not panic at a DAO

A panic crossing back into CommonDAO aborts the transaction, and for a proposal
that has already passed, every retry aborts identically — permanently stuck.
`IsRunSubmittersDelegate` and `RunSubmittersGrantedBy` are exposed as pure
predicates so a policy realm can check before acting. The write functions still
panic on refusal, which is correct for a direct caller; a DAO-facing wrapper
should pre-check.

## Testing

Sixteen Gno tests in `examples/gno.land/r/sys/params/delegate_test.gno`, and an
end-to-end txtar. Mutations verified against the unit tests: dropping the
empty-slot check, anchored-prefix matching, unscoped removal, the `IsCurrent`
frame check, the arming guard, and the grant-laundering guard each fail exactly
the test that covers them.

`run_submitters_delegate.txtar` is the wiring test: it delegates through a real
GovDAO proposal, has the policy realm grant MsgRun rights to an account GovDAO
never named, confirms that account can actually run, then revokes and confirms
the former delegate is refused. The **refusals** are what distinguish "the gate
authorized this caller" from "there is no gate"; every write assertion would
pass with the gate deleted.

Two different refusals are needed, and this took two attempts to get right. The
"no delegate is configured" cases fire when the slot is EMPTY, so they survive
the path comparison in `assertDelegate` being deleted — the whole txtar passed
with any realm able to write any delegated parameter. Only a caller that holds
no delegation, attempted while ANOTHER realm holds one, exercises the comparison
itself. The `rogue` realm exists for that and nothing else, and deleting the
comparison now fails on `only gno.land/r/test/paramsgate may write`.

Two pre-existing coverage gaps were closed while factoring the shared gate:
`assertValsetCaller` had no test in this realm, and neither did
`assertNotValsetKey` — the guard whose own comment says it stops a GovDAO
supermajority from writing validator-set state through the generic factory.

## Still open

1. **`pkg_approvers` is the natural second client**, and the comparison above
   should be settled deliberately rather than by which one was asked for first.

2. **The grant record grows without bound across GovDAO resets.** Nothing
   reconciles it with the parameter, so a delegate that adds a thousand
   addresses, has GovDAO reset the list, and adds a thousand more leaves two
   thousand stale entries. Each add is paid for by the delegate, so this is
   self-funded rather than free, and it grants no authority the delegate does
   not already have — but a reconciliation pass, or dropping the record on a
   whole-list set, would be tidier.

3. **No policy realm ships.** The txtar's `paramsgate` is deliberately ungated
   so that a refusal there is unambiguously the params-side gate. A real one
   needs its own governance, and the closure it hands to
   `CreateExecutionProposal` must capture its values *at propose time* — the
   closure body is frozen when the proposal is created, but anything it reads
   from mutable realm state at execution time is an effect the council approved
   without seeing.

4. **Provenance can drift from the parameter**, and nothing reports it. A
   `Render` on this realm — it has none today — would make both the slot and the
   grant record auditable without a query per key.

Written with AI assistance (Claude Code). The design in this document is the
second one; the first is described under "A named slot, not a registry" and was
discarded for the reasons given there.

## The whole-list path, and why the proposer must be on the list

An empty `run_submitters` means the MsgRun gate is off. So the dangerous state is
not an empty list but a **non-empty one that names nobody able to create a GovDAO
proposal**, and that state cannot be undone: creating a proposal needs MsgRun,
because a `ProposalRequest` carries an `Executor` and `MsgCall` cannot build one
from string arguments, while `isValidCall` refuses realm intermediaries. The vote
that armed the gate would be the last vote.

Two changes close the ways to get there in one step.

`vm:p:run_submitters` is reserved from the nine generic
`NewSysParam*PropRequest` factories, exactly as `node:valset:*` already is, and a
dedicated `ProposeSetRunSubmitters` owns the whole-list path. Reserving the key
keeps the number of ways to write this parameter at one, rather than adding the
same check to nine siblings and hoping the tenth remembers.

That constructor requires the proposer's own address to be on any non-empty list
it proposes. GovDAO refuses a proposal from a non-member, so if the proposal
exists its author is a member -- and they just signed the transaction that
created it, so the address demonstrably holds a key. That is what makes this
stronger than requiring the list merely to *name* a member: any member may enrol
an arbitrary address, so a named member proves neither that anyone holds the key
nor that they will use it.

It is checked at proposal creation, not in the executor. Nothing in
`r/gov/dao` recovers from an executor panic, so a check that fires at execution
turns a passed proposal into one that can never be executed. At creation the
refusal reaches a person who can still fix the list.

**What this is not.** It is a floor, not an invariant, and three gaps are worth
stating plainly rather than discovering later:

- It decays. The proposer may resign from GovDAO afterwards, and nothing
  re-checks the list when they do. It rules out arriving at a dead list in one
  vote; it cannot rule out drifting into one.
- It is not global. `gnogenesis params set vm.run_submitters` and the
  `RealmParams` loader write the keeper directly, bypassing this realm entirely,
  so a memberless list is still reachable at genesis.
- It does not bound capture. A member who lists only themselves satisfies the
  rule and then holds the sole MsgRun capability on the chain.

**Recovery, if it happens anyway.** Voting and executing an already-created
proposal are both `MsgCall`-able (`MustVoteOnProposalSimple`,
`ExecuteProposal`), and proposals never expire. So a "reset `run_submitters`"
proposal created *before* the gate is armed is a permanent one-shot rescue that
survives a total MsgRun lockout. Worth doing on any chain that arms this gate.

**The durable fix, deliberately not taken here.** All of this exists because
proposal *creation* is MsgRun-only, and that is an argument-marshalling
accident rather than a security property -- `isValidCall` already admits
`MsgCall`. Making one narrow proposal `MsgCall`-constructible, through a
GovDAO-controlled registry of proposal-factory realms, would remove the coupling
instead of guarding one instance of it, and would cover the genesis paths no
realm-side rule can reach. That is a change to `r/gov/dao` and needs its own
review.
