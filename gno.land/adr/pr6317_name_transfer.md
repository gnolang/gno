# ADR: Transferring a registered name to a new address

Status: **proposed**, for discussion. No code yet.

## Context

Names are registered in `r/sys/users` through controller realms such as
`r/sys/namereg/v0`. At deploy time `r/sys/names` reads that registry to decide
who may deploy under `gno.land/{r,p}/<name>/...`. Today a name is tied to one
address for good:

- `UserData.addr` is set once by `registerUser` and never changed.
  `UpdateName` changes the *name* for an address; the old name stays in
  `nameStore` as an alias pointing at the same record.
- Names are never freed. `Delete` marks the record deleted but leaves its keys
  in `nameStore`, and `RegisterUser` refuses any key already there
  (`ErrNameTaken`). Deleting and registering again from a new address
  therefore fails, for controllers and GovDAO alike.
- An address whose record was deleted can never register again
  (`ErrDeletedUser`).

A user who rotates keys, moves to a multisig or loses a key cannot take their
name and namespace with them.

### Who controls deployed packages today

The keeper writes the deploying address into `gnomod.toml` as
`[addpkg] creator`. That stamp is also the identity `init()` runs as. Five
checks decide who may act on a path:

| Operation | Current rule | Location |
|---|---|---|
| New deploy | `names.IsAuthorizedAddressForNamespace`: the deployer's own address, or the deployer holds the **current** name | `checkNamespacePermission`, `keeper.go` |
| Redeploy a live **private** package (normal path, or parked for activation) | stamped creator **and** the namespace check above | `checkRedeployPermission` + `checkNamespacePermission`, `keeper.go`, `keeper_inert.go` |
| Park over another submitter's parked package (public or private path) | parked submitter only | park guard, `keeper.go` |
| Activate a parked package | namespace check; over a live private package, also parked creator == live creator | `EnablePackage`, `keeper_inert.go` |
| Reject a parked package | parked submitter, an approver, or the live package's stamped creator (public or private) | `RejectPackage` / `isLivePackageOwner`, `keeper_inert.go` |

Because the namespace check accepts only a **current** name, a user who
renames from `alice` to `alice2` today can no longer redeploy their own
private packages under `r/alice/*` at all.

Public packages can never be replaced, so the right to *replace* a live
package only matters for private ones. The live owner of a public package
does have one right: rejecting parked submissions at its path. This ADR gives
that right to the effective owner. Parked submissions follow their own rules,
whether the path is public or private.

`GetParams` replaces an empty stored `SysNamesPkgPath` with the default
`gno.land/r/sys/names` (`params.go`). Clearing the parameter therefore does
**not** turn namespace enforcement off; only pointing it at a path where
nothing is deployed does.

## Decision

A user record (its current name plus every alias it gathered through renames)
can be moved to a new address. The new address becomes the **effective owner**
of the record's namespaces, which brings two rights:

- it may deploy **new** paths under the **current** name only, which is
  today's rule;
- it may act as owner of **existing** packages under the current name **and**
  every alias.

Moving a record changes who is authorized. It never rewrites who authored
existing bytes: stamps on live and parked packages stay as they are until the
owner redeploys.

Every realm-level change below is **dormant until activation** (section 6).
Before activation the three realms behave exactly as they do today, so
replayed history executes the same way it did on the source chain.

### 1. `r/sys/users`: record identity and the move

- **Record identity.** Each record gets an `id uint64`, which is never
  reused, and a `rev uint64`.
  - Both live in new B+ trees (`idStore: id → *UserData`,
    `metaStore: address → {id, rev}`), **not** as new `UserData` fields
    written at registration. Code that runs before activation writes
    nothing new, so the storage and deposit charges of historical
    transactions do not change on replay.
  - Activation assigns ids to every existing record, deleted ones included,
    one id per record (not per alias), in `addressStore` key order.
    Registrations after activation take the next id.
  - Every rename, delete and transfer after activation increments `rev`.
    Incrementing an id or rev past its maximum is refused, never wrapped.
  - New read-only API: `ID()`, `Rev()`, `ResolveID(id) *UserData`.
- **Authoritative-record guard.** The read-only trees hand out copies
  (`makeUserDataSafe`). A controller that calls a mutator on a copy writes
  name keys that point at the copy. This affects the exported
  `UpdateName`, `UpdateNameIgnoreCanonical` and `Delete` today, and would
  affect `TransferTo`.
  - The guard goes in the **internal** mutators (`updateName`, `delete`,
    `transferTo`), so every entry point is covered, including the
    canonical-bypass and governance paths.
  - After the nil and deleted checks, they require
    `addressStore.Get(u.addr.String()) == u`. Checking the id alone is not
    enough, because a copy carries the same id.
  - Before activation, verify that every name and alias points at its
    authoritative record. Any split left by an earlier mutation of a copy is
    repaired explicitly as part of the migration.
- **Internal `(u *UserData) transferTo(newAddr address) error`:**
  - it rejects a deleted record, an invalid `newAddr`, and
    `newAddr == u.addr`;
  - it checks recipient eligibility against the **raw** `addressStore`: any
    entry, live or deleted, is refused (`ErrAlreadyHasName` /
    `ErrDeletedUser`). `ResolveAddress` hides deleted records, so it cannot
    be used here;
  - in one transaction it removes the old address key, sets `u.addr`, writes
    the new key, moves the `metaStore` entry, increments `rev`, and emits
    `Transferred(id, name, from, to)`. No alias walk is needed, because every
    alias key points at `u`.
- `(u *UserData) TransferTo(_ int, rlm realm, newAddr address) error` is the
  controller entry point. Like `UpdateName`, it checks `rlm.IsCurrent()` and
  authorizes **`rlm.Address()`** against the controller whitelist.
- `IsAddressEligible(addr address) bool` is a read-only query of the raw
  store, so controllers can refuse an ineligible recipient early.
- **Recovery proposal:**
  `ProposeTransferUser(cur realm, from address, to address) dao.ProposalRequest`.
  - When the proposal is created, it resolves `from` to a live authoritative
    record and captures that record's `id`. When it runs, it calls
    `ResolveID(id)` and refuses a deleted record.
  - It binds to `id`, **not** to `rev`. Holders cannot rename themselves in
    namereg/v0, but a thief holding the key could transfer the record back
    and forth between addresses they control. Every transfer increments
    `rev`, so binding to `rev` would let them invalidate the recovery vote.
  - It uses `NewFilterByTier(memberstore.T1)`. That filter restricts
    **voting** to T1; any DAO member can still **create** the proposal.
  - The proposal text names the id, the name and address when the proposal
    was created, and that recovery follows this identity through later
    renames and transfers. It also states that every private package under
    the record's names becomes controllable by `to`, and that the current
    holder loses that control.
- **Other DAO proposals on records.** `users.ProposeUpdateName`,
  `namereg.ProposeNewName` and both `ProposeDeleteUser` functions keep their
  address parameters. After activation they resolve the address to an `id`
  when the proposal is **created**, and call `ResolveID` when it runs.
  - Today the users versions look up by address when they run. After the
    address is reused, that lookup can land on a different record from the
    one voters saw.
  - The namereg versions capture a pointer at creation, so they already hit
    the right record. Binding them to `id` makes all four behave the same.
  - Proposal text the DAO has already stored is not refreshed.
  - Proposals created before activation keep their current behaviour.
- After a transfer, the old address has no entry in `addressStore`. It may
  register a new name like any fresh address, and gets a new `id`.

### 2. Controller trust and the limits of T1

Any whitelisted controller can call `TransferTo` on any record, just as any
controller can already rename or delete any record. Once this ADR lands,
controllers also control who holds private packages, so they are trust roots
for both identity and packages.

After activation, `ProposeNewController` and
`ProposeControllerAdditionAndRemoval` take the same T1 voter filter as
`ProposeTransferUser`. Within the current DAO implementation, these filters
limit who votes on those proposal types. They do **not** make T1 the final
authority:
- the wider electorate can still delete a record through either registry,
  which permanently freezes its packages under this design and defeats
  recovery;
- it can also remove controllers, and replace the DAO implementation with
  one that ignores the filters.

**Required before activation:** decide the voter policy for
- rename and delete proposals in both registries;
- `ProposeRegisterUser`;
- controller removal;
- replacing the DAO implementation.

These are security decisions this design depends on, not consistency
tweaks. Audit the existing controllers and any open proposals against the
chosen policy.

### 3. `r/sys/namereg/v0`: two-step transfer for users

The owner offers, and the recipient accepts by sending a transaction.
Accepting proves the recipient holds the key, so a mistyped address cannot
swallow a name.

- **Storage.** `pendingTransfers` is a B+ tree keyed by the record `id` in a
  fixed-width encoding: `id → {to address, rev uint64, expiresAt int64}`.
  - It is **not** a native Gno map. A persisted map is rebuilt and
    re-serialized in full on every write, so a growing map would make every
    call more expensive.
  - Each operation touches one key. There is at most one offer per record;
    a new offer replaces the old one, and accept or cancel removes it.
  - Stale offers still occupy deposited storage, so `PruneTransfer(id)` lets
    anyone remove one offer that has expired or whose `rev` is stale.
- **Who can call.** `OfferTransfer`, `CancelTransfer` and `AcceptTransfer`
  must be **direct account calls not signed with a delegated session key**.
  - The direct-call part is `cur.Previous().IsUserCall()`.
  - The no-session part needs `runtime.GetSessionInfo()` to return
    `isSession == false`. `IsUserCall` only checks how the call was made,
    not which key signed it, and a session whose `AllowPaths` covers
    namereg could otherwise give the name away.
  - Multisig accounts that sign transactions directly are supported. Wallets
    implemented as code realms are not; for those, GovDAO is the route.
- `OfferTransfer(cur realm, to address)`:
  - the registry must not be paused, and the caller must have a live record;
  - `to` must be valid, must not be the caller, and must pass
    `IsAddressEligible`;
  - it records the record's current `rev`, sets
    `expiresAt = time.Now().Unix() + 7*24*3600` (block time, deterministic),
    and emits `TransferOffered`.
- `CancelTransfer(cur realm)` can be called by the current holder of the
  record, even while the registry is paused.
- `AcceptTransfer(cur realm, id uint64)`:
  - the registry must not be paused, the caller must be `offer.to`, and the
    current block time must be strictly before `expiresAt`;
  - `ResolveID(id).Rev()` must equal `offer.rev`. Any rename, delete or
    transfer since the offer was made, by anyone (DAO, another controller, or
    a transfer away and back), invalidates it;
  - it deletes the offer and calls `TransferTo(0, cur, caller)`. If that
    errors, the transaction aborts and the offer survives.
- No fee and no cooldown by default. GovDAO can add a fee later, the same way
  it sets `registerPrice`.

### 4. `r/sys/names`: a strict ownership query

New deploys need no change. `resolveCurrentName` follows the record, so after
a transfer `IsAuthorizedAddressForNamespace` accepts the new address and
refuses the old one.

The keeper needs a query that `IsAuthorizedAddressForNamespace` cannot answer
safely: that function returns `true` for everyone until `Enable` runs, and
recognizes only current names.

```go
// NamespaceOwner reports who owns existing packages under namespace.
//
//   status "disabled":     the verifier is not enabled yet
//   status "address":      namespace is an address
//   status "unregistered": a name-shaped namespace nobody has registered
//   status "deleted":      the name belongs to a deleted record
//   status "owned":        owner is the record's address; the name may be
//                          the record's current name or any of its aliases
//
// paused reports the verifier's pause flag (only meaningful once enabled),
// so the keeper needs one call, not two. Pause does not change owner or
// status: pause controls operations, not identity.
func NamespaceOwner(namespace string) (owner address, status string, paused bool)
```

`users.ResolveName` returns nil for both "never registered" and "deleted".
`NamespaceOwner` tells them apart with `users.IsNameTaken`: a name that is
taken but resolves to nil belongs to a deleted record. An alias resolves with
`isCurrent == false`, which still counts as `owned`.

### 5. Keeper: effective ownership

Every ownership-dependent decision goes through one resolver. A sender who
is authorized for some other reason never triggers it: for example, the
parked submitter or an approver rejecting a parked package.

```go
// effectiveOwner returns who owns pkgPath's namespace for ownership-gated
// operations. stamped is the live package's creator, or empty when nothing
// is live. An empty owner means nobody holds owner standing.
func (vm *VMKeeper) effectiveOwner(ctx sdk.Context, st gno.TransactionStore,
    params Params, pkgPath string, stamped crypto.Address) (owner crypto.Address, paused bool, err error)
```

| `NamespaceOwner` result | Effective owner |
|---|---|
| `owned` | the record's address |
| `deleted` | none: owner-gated writes are refused (frozen) |
| `unregistered` | none, as today: the namespace check already refuses everyone once names is enabled |
| `address` | the address itself if nothing is live, or if it is also the stamped creator (today's combined rule). A legacy package whose creator differs from its address namespace gains no new rights |
| `disabled`, or the configured names realm is not deployed | the stamped creator, as today |
| the call errors (missing function, wrong return shape, panic) | **none**: the owner branch is refused. An error never falls back to the creator |

Applied per operation:

- **Replacing a live private package.** The submitter must be the effective
  owner. This **replaces** the namespace check in all three places it runs
  for this case:
  - the normal path (`checkNamespacePermission`, `keeper.go` AddPackage);
  - the inert park path (`keeper.go`, inert branch);
  - `EnablePackage` (`keeper_inert.go`), which checks the parked creator.

  Missing any one of them would leave alias packages updatable on one path
  and blocked on another. The namespace check still runs, unchanged, when
  nothing is live at the path.
- **Pause.** Today pause is enforced *inside*
  `IsAuthorizedAddressForNamespace`. Dropping that call on the replacement
  path would therefore drop pause as well. The keeper refuses every
  submission, park and activation when the resolver reports `paused` and
  names is enabled, including ones authorized by ownership. Before `Enable`
  there is no pause, as today.
- **Parking.** A park is a deploy, so it first passes the deploy gate that
  applies to its path: the namespace check if nothing is live there, or the
  effective-owner check if a private package is live. The park guard then
  applies as well: if another submission is already parked at the path, the
  new submitter must be that parked submitter, as today, or the effective
  owner. Ownership does not put the owner's address on the old bytes. The
  owner's submission replaces them and carries its own stamp and `init()`
  identity.
- **Activating.** Over a live private package, the parked creator must be the
  effective owner *when the package is activated*, because the name may have
  moved since it was parked. With nothing live, activation keeps today's
  namespace check on the parked creator. A first deploy parked by the old key
  before a transfer therefore fails to activate afterwards, and the new owner
  can reject it.
- **Rejecting.** `RejectPackage` first accepts the parked submitter (who can
  always withdraw their own bytes) or an approver, without calling the
  resolver. Only a sender relying on owner standing reaches it, and an error
  refuses only that branch. Deletion, pause and resolver errors never take
  away the submitter's or the approver's right to reject. Owner standing
  applies whether or not anything is live.
- **`QueryPackageMeta`.** It evaluates the new blockers for **both** first
  deploys and live redeploys, using its throwaway store:
  - pause;
  - failed namespace authorization;
  - deleted ownership;
  - owner mismatch;
  - a resolver error.

  A first deploy whose creator has lost the namespace must not report only
  "waiting for an approver". `ReasonOwnerMismatch` is reworded from "can
  never be enabled" to "currently blocked by owner mismatch", since the
  mismatch can now go away.

Implementation constraints:

- **Store.** The resolver takes the caller's transaction store.
  `QueryPackageMeta` passes its throwaway store. It must not use the
  context-bound store that `callRealmBool` fetches.
- **Ordering.** It runs before `init()` and before storage-deposit
  accounting, under the same read-only contract as `callRealmBool`.
  `NamespaceOwner` must stay a pure read.
- **Genesis replay.** During `IsGenesisReplay`, the resolver and every
  ownership gate this ADR adds are skipped. The live-redeploy waivers stay
  as they are, and the park guard and the owner branch of `RejectPackage`
  are waived too.
  - `pr5511_chain_upgrade_genesis_replay.md` marks failed transactions
    `Failed` and skips them, so only transactions the source chain accepted
    are re-delivered.
  - This is necessary but not sufficient. The upgrade replaces the sys
    realms' source in their original genesis transactions
    (`contribs/gnogenesis/internal/fork`), so all of history runs against
    the **new** realm code. Section 6 keeps that code dormant until
    activation for this reason.
- **Gas.**
  - Owner-gated paths (live private redeploy, park over a park, activate
    over a live package, rejection by owner standing) pay for a metered
    realm call instead of a string compare.
  - Replacing or activating over a live public package is still refused
    before that call.
  - Measure with identical workloads, with warm and restarted stores, and
    with growing registry and offer indexes.

**Accepted risk: the stamped-creator fallback.** It applies when names is
disabled or not deployed. Governance can reach that state by pointing
`SysNamesPkgPath` at an undeployed path, which hands control back to the
stamped creator, possibly a compromised old key. We accept this. That same
state already lets *any* address deploy into *any* namespace
(`checkNamespacePermission` returns nil), so the fallback is not the weakest
point. Making an empty parameter mean "disabled" would be a separate change.

### 6. Activation

The sys realms are public (none sets `private = true`), so they cannot be
replaced in place. `r/sys/users` holds the registry state and is not
versioned, so this change ships as a **chain upgrade** using genesis replay.
That has one consequence the code must respect: history replays against the
new realm code.

- **Dormant until activated.** Every new realm behaviour stays switched off
  until a single activation transaction runs. That covers:
  - id and rev bookkeeping;
  - the authoritative guard;
  - binding proposals to an id;
  - T1 filters on controller proposals;
  - transfers.

  Before activation, the realms' code paths must produce the same results,
  storage and events as the source chain. Otherwise strict replay fails: a
  historical controller proposal passed by T2/T3 votes would be refused
  under a T1 filter, and a proposal created before its target was
  registered would fail an id lookup.
- **The activation transaction.** It is appended after the replayed history
  in the upgrade's genesis file, and gated at height 0 in the same way as
  `AddControllerAtGenesis`. It:
  - assigns ids;
  - checks and repairs the authoritative records;
  - reserves the namespaces the audit names;
  - sets the activation flag.
- **Open proposals keep their rules.** GovDAO stores a proposal's voter
  filter when the proposal is created, so a proposal still open at
  activation keeps the voters it was created with.
- **Ids must stay stable across future upgrades.** Once ids appear in
  transactions and proposals, a later replay-based upgrade must not change
  anything that renumbers them. That rules out reordering or removing
  pre-registrations, or adding records before the activation point.
  - New reservations, and dropping `test1`, go into transactions appended
    after history.
  - Before boot, compare the source chain's and the replayed chain's
    id → record mappings, deleted records included.
- **Strict replay.** Run with `StrictReplay = true` over the full source
  history. `GasReplayMode = "source"` only changes gas metering; it does not
  waive realm authorization failures or storage-deposit limits. Do not skip
  historical transactions that newly fail just to get the chain to boot.
- **Keeper and transfers go live together.** The keeper rule takes effect
  for live transactions as soon as the chain starts. Transfers become
  callable only after activation, so a transferred namespace is never left
  without anyone able to redeploy its packages.

PRs can still be split for review: realms (sections 1–4 and the activation
step), keeper (section 5), the approver changes (`PkgCreator`, see
Consequences), and the audit tool or report. They merge behind the same
upgrade.

### Authorization matrix after this ADR

Namespace states:
- **current**: the record's current name;
- **alias**: a name the record has renamed away from;
- **deleted**: a name belonging to a deleted record;
- **unregistered**: a name-shaped namespace nobody holds;
- **address**: an address namespace;
- **off**: names disabled, or the configured names realm is not deployed.

| | current | alias | deleted | unregistered | address | off |
|---|---|---|---|---|---|---|
| **Effective owner** | record's address | record's address | none | none | the address, if nothing is live or it is the stamped creator | stamped creator |
| New deploy | owner | nobody | nobody | nobody | that address | anyone |
| Replace live private | owner | owner | nobody | nobody | owner | owner |
| Park | the deploy gate that applies (new deploy, or replace live private); if something is already parked there, also its submitter or the owner | same | same | same | same | same |
| Activate (nothing live) | parked creator must pass "New deploy" | same | same | same | same | same |
| Activate over live private | parked creator must be the owner | same | refused | refused | same as current | same as current |
| Reject parked | submitter, approver, owner | same | submitter, approver | submitter, approver | submitter, approver, owner | submitter, approver, owner |

When names is enabled and paused, every deploy, park and activation is
refused, in every column except off, where pause has no effect today either.
Rejection is never paused.

## Alternatives considered

- **Old and new owner can both redeploy** (stamped creator *or* name owner).
  This keeps today's behaviour for every existing package, but a stolen old
  key could still replace private realm code after the user has moved, which
  is the main case transfer is meant to fix. Rejected.
- **Ownership follows only the current name.** After a transfer, packages
  under aliases would stay with the stamped (old) creator, so recovering from
  a stolen key would leave those packages with the thief. Rejected.
  Limiting aliases to *existing* packages means no new paths can be added
  under an alias. A user could still deploy a private placeholder before
  each rename and keep upgrade rights in several namespaces. In namereg/v0
  every rename needs a governance vote, which limits this. The ADR accepts
  that trade-off in exchange for being able to recover aliases.
- **The stamped creator owns address and unregistered namespaces.** This
  would give a legacy creator rights the namespace check denies them today,
  for example on a package deployed before `Enable` under someone else's
  address namespace. Rejected; those namespaces keep today's combined rule.
- **Move only the current name** and leave the aliases on the old address.
  This splits one identity's history across two records. Rejected.
- **Merge into a recipient that already has a name.** That needs merging two
  records and picking which name stays current. Rejected; the recipient must
  be fresh.
- **Transfer controlled only by GovDAO.** Routine key rotation would need a
  vote. Rejected; GovDAO keeps a T1-voted path for recovery.
- **Offers keyed by address.** If an address is transferred away and later
  registers a different name, an old offer would move the new record. Offers
  bind to `id` and `rev` instead.
- **A DAO transfer that moves the name but not package control.** That needs
  a per-record flag and a second ownership rule. Rejected in favour of
  T1-voted proposals whose text states what they grant.
- **Rewriting the stamped creator when ownership moves.** The stamp is also
  the `init()` identity. Rewriting it would attribute bytes to an owner who
  never submitted them. Rejected.
- **New fields on `UserData`, written at registration.** These would change
  storage charges for every historical registration on replay, and could
  push signed deposit ceilings over their limits. Rejected in favour of
  side indexes filled in at activation.

## Consequences

- **Rights move to the name holder.** Anyone who deployed a private package
  under a name they do not hold loses control of it to the record's holder.
  Once `names` is enabled, that can only have happened through genesis or
  before `Enable`. The pre-registered system names in
  `namereg/v0/preregister.gno` (`sys`, `gov`, `gnoland`, `gno`, `nt`, `demo`,
  `archive`, `x`, `gnolang`, `test1`) are the main exposure.
  - **Required as part of the upgrade proposal:** list every live private
    package in a namespace that is, or could become, a registered name. That
    includes `unregistered` namespaces, because a later registration (or a
    DAO rename onto that name) would take their packages over.
  - Each case where owner and creator differ must be explicitly accepted.
  - The activation transaction reserves every unregistered namespace that
    holds legacy packages for an agreed owner.
  - Under the new rule the owner can redeploy directly, so no separate
    re-stamping route is needed. The audit decides who gains control.
- **Renamed users regain control of their old namespaces.** Today a rename
  locks the user out of their own private packages under the old name. With
  aliases recognized, they can redeploy them again. New deploys under an
  alias stay refused.
- **Deleting a record freezes its packages, with no way back.** Transfer
  refuses deleted records, registration cannot reclaim their names, and
  there is no undelete. Their packages stay frozen unless a later upgrade
  adds a recovery mechanism.
- **T1 voters gain a new power.** They can hand control of a record's private
  packages to another address. That is the price of key recovery. See
  section 2 for why this is not a hard T1 boundary.
- **A stolen key can now evict the owner.** Today a thief who holds a
  user's key shares the name with them. With transfer, the thief can offer
  the record to an address they control and accept straight away, locking
  the user out. Only the T1 recovery proposal can undo that. This is the main
  cost of user-led transfer; see the open question on a delay.
- **Address namespaces never move.** `gno.land/r/<address>/...` stays with
  that address. A user recovering from a stolen key leaves those packages
  with the thief. The docs should steer users who care about recovery
  towards deploying under a name.
- **Address-keyed state outside the registry does not move.** `ownable`
  owners, admin lists and balances in other realms stay with the old address.
  The docs must say so.
  - The one exception is a private realm the new owner redeploys. A private
    redeploy re-runs `init()` with fresh package globals as the new stamped
    creator (`pr6164_redeploy_continues_realm.md`), so an `ownable` owner set
    in `init()` follows the redeployer. That is existing redeploy behaviour,
    not something this ADR adds.
- **Deposits stay with the realm's accounting.** A transfer neither refunds
  deposits nor attributes them to the new holder.
  - A private redeploy keeps the previous deployment's objects and their
    deposit (`pr6164`). It charges for the new objects and recovers nothing
    from the old, unreachable ones.
  - Released storage is refunded to the caller passed to
    `processStorageDeposit`: the message sender on a normal deploy, and the
    parked creator on activation. While ugnot is restricted, refunds go to
    `StorageFeeCollector` instead.
  - There is no ledger of who deposited what. Moving control can therefore
    change who is able to free storage and collect refunds that earlier
    writers paid for. This is the same for any private redeployer today.
- **The CLA still applies.** The new address must have a valid
  `r/sys/cla` signature before it can deploy, park or redeploy
  (`checkCLASignature` runs on all three paths). A transfer does not carry
  the old address's signature over.
- **`test1` is pre-registered to a key whose mnemonic is public**
  (`DefaultAccount_Address` in `gno.land/pkg/integration`). The
  pre-registration list has no chain-id gate. On a production chain, anyone
  can already deploy under `r/test1/*`, and with this ADR anyone could also
  transfer the record away. The activation transaction should deal with
  `test1` on production chains, without changing ids that are already in
  use.
- **Off-chain consumers** (indexers, gnoweb `/u/<name>`) that cache
  address ↔ name need to handle `Transferred`.
- **Approvals bind source, not submitter.**
  - Content hashes exclude `[addpkg]`. If a new owner parks the same bytes
    under their own stamp, a standing approval for those bytes still
    matches.
  - `PkgHeight` catches a re-park in a later block, but submissions in the
    same block share a height.
  - `init()` runs as the parked creator, so approvals need a way to pin it.
    Add an optional `PkgCreator` field to `MsgEnablePackage`:
    - append it after the existing fields, so Amino field numbers do not
      change;
    - reject a supplied value that does not match the parked creator, before
      type-checking or execution;
    - skip it during genesis replay.
  - The same release must:
    - update approval producers (`contribs/gpao`) to keep and sign the
      reviewed submission's creator;
    - make `QueryPackageMeta` report the **parked** submission's creator and
      height separately from the live package's.

    Approvals that leave `PkgCreator` out keep today's source-only meaning.

## Testing

- **`r/sys/users`:**
  - aliases follow the transfer;
  - the old address no longer resolves, can register again, and gets a new
    `id`;
  - live and deleted recipients are refused;
  - a deleted user cannot transfer;
  - a non-controller is refused;
  - copies from both read-only trees are refused by `TransferTo`,
    `UpdateName`, `UpdateNameIgnoreCanonical` and `Delete`, without changing
    any index;
  - `canonicalStore` and the exact-name reservations are unchanged;
  - `rev` increments on rename, delete and transfer, and refuses overflow;
  - the event is emitted;
  - the T1 proposal works end to end, survives a DAO or controller rename and
    repeated holder-led transfers while the vote is open, and is refused if
    the record is deleted;
  - an id-bound rename or delete proposal does not touch a different
    identity after address reuse;
  - before activation, every path produces the same results, storage and
    events as the current code.
- **`r/sys/namereg/v0` filetests:**
  - offer → accept;
  - accept by the wrong address;
  - expiry exactly at `expiresAt`;
  - cancel, including while paused;
  - self-offer;
  - calls that fail `IsUserCall`;
  - **session-signed** offer and accept;
  - the recipient registers a name between offer and accept;
  - a DAO rename, delete or transfer during an open offer invalidates it;
  - transferring away and back invalidates an old offer;
  - an offer made, then the address reused, cannot be accepted against the
    new record;
  - pruning expired and stale offers;
  - gas stays flat as the offer index grows.
- **`r/sys/names`:** `NamespaceOwner` for each status: before `Enable`, while
  paused (unchanged answer), an address namespace, an unregistered name, an
  alias, a deleted record, and a current name.
- **Keeper:** every cell of the authorization matrix, plus:
  - a resolver error refuses the owner branch, while the submitter and the
    approver can still reject;
  - `QueryPackageMeta` uses its throwaway store and reports the new blockers
    for first deploys and live redeploys, before and after transfer, rename,
    deletion, pause and resolver failure;
  - parking and approving public paths;
  - a first deploy parked by the old key before a transfer cannot be
    activated afterwards, and the new owner can reject it;
  - a new owner without a CLA signature is refused;
  - an address namespace stays with its address after the record moves;
  - a legacy package whose creator differs from its address namespace gains
    no rights;
  - `PkgCreator`: left out, matching, mismatched, replaced within the same
    block, and older messages on replay.
- **Upgrade:**
  - full-history strict replay with the new code dormant, including
    successful historical T2/T3 votes on controller proposals, proposals
    created before their target was registered, and any history of
    mutations on copies;
  - the activation transaction assigns the expected ids and its storage cost
    is accounted for;
  - the id → record mapping matches between source and replayed chains;
  - an A → B → C upgrade, where chain B has transfers and id-bound
    proposals, keeps every id stable.
- **Integration `.txtar`:** register → deploy a private realm → offer/accept
  → redeploy from the new key → old key refused, for both the current name
  and an alias.
- **Commands:** `go test ./gno.land/pkg/sdk/vm/ -run Gas`,
  `go test ./gno.land/pkg/integration/ -run TestTestdata`, and the examples
  suite.

## Open questions

- Offer expiry: is 7 days right?
- Should acceptance take effect only after a delay, during which GovDAO can
  freeze it? A delay does not stop a thief who holds the same key, but it
  gives the T1 recovery path time to act before packages change hands.
- Should transfers cost a fee, or have a cooldown?
- Deleted records: should freezing be the end state, or should a later
  mechanism let governance assign a deleted record's packages to a new owner
  without bringing the name back?
- The voter policies listed in section 2 must be settled before activation.
