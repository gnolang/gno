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
`[addpkg] creator`. That stamp is also the identity `init()` runs as and the
address that paid the storage deposit. Five checks decide who may act on a
path:

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

### 1. `r/sys/users`: record identity and the move

- `UserData` gains `id uint64`, assigned once at registration and never
  reused, and `rev uint64`. Every rename, delete and transfer increments
  `rev`. New read-only accessors are `ID()` and `Rev()`, plus
  `ResolveID(id) *UserData`.
- Internal `(u *UserData) transferTo(newAddr address) error`:
  - `u` must be the **authoritative** record, meaning
    `addressStore.Get(u.addr) == u`. The read-only trees hand out copies
    (`makeUserDataSafe`), and moving a copy would leave the name keys
    pointing at the original. The same check is added to `UpdateName` and
    `Delete`, which have the same latent problem today.
  - It rejects a deleted record, an invalid `newAddr`, and
    `newAddr == u.addr`.
  - It checks recipient eligibility against the **raw** `addressStore`:
    any entry, live or deleted, is refused (`ErrAlreadyHasName` /
    `ErrDeletedUser`). `ResolveAddress` hides deleted records, so it cannot
    be used here.
  - In one transaction it removes the old address key, sets `u.addr`, writes
    the new key, increments `rev`, and emits
    `Transferred(id, name, from, to)`. No alias walk is needed, because every
    alias key points at `u`.
- `(u *UserData) TransferTo(_ int, rlm realm, newAddr address) error` is the
  controller entry point. Like `UpdateName`, it checks `rlm.IsCurrent()` and
  authorizes **`rlm.Address()`** against the controller whitelist.
- `IsAddressEligible(addr address) bool` is a read-only query of the raw
  store, so controllers can refuse an ineligible recipient early.
- `ProposeTransferUser(cur realm, id uint64, to address)` is GovDAO's path for
  recovering a lost or compromised key.
  - It uses `NewFilterByTier(memberstore.T1)`. That filter restricts
    **voting** to T1; any DAO member can still **create** the proposal.
  - It binds to the record `id`, **not** to `rev`. Holders cannot rename
    themselves in namereg/v0, but a thief holding the key could transfer the
    record back and forth between addresses they control. Every transfer
    increments `rev`, so binding to `rev` would let them invalidate the
    recovery vote.
  - When it runs, it calls `ResolveID(id)` and refuses a deleted record.
  - The description states that every private package under the record's
    names becomes controllable by `to`, and that the current holder loses
    that control.
- `users.ProposeUpdateName`, `namereg.ProposeNewName` and both
  `ProposeDeleteUser` functions bind to `id` and call `ResolveID` when they
  run.
  - The users versions look up by address when they run. After the address
    is reused, that lookup can land on a different record from the one
    voters saw.
  - The namereg versions keep the pointer they captured at creation. That
    still targets the right record, but the name shown in the proposal can
    go stale. Binding to `id` makes all four behave the same way.
- After a transfer, the old address has no entry in `addressStore`. It may
  register a new name like any fresh address, and gets a new `id`.

### 2. Controller trust

Any whitelisted controller can call `TransferTo` on any record, just as any
controller can already rename or delete any record. Once this ADR lands, a
controller can therefore also hand out control of private packages. Since
the T1 filter on `ProposeTransferUser` would otherwise be easy to route
around, `ProposeNewController` and `ProposeControllerAdditionAndRemoval` move
to the same T1 voter filter.

### 3. `r/sys/namereg/v0`: two-step transfer for users

The owner offers, and the recipient accepts by sending a transaction.
Accepting proves the recipient holds the key, so a mistyped address cannot
swallow a name.

- `pendingTransfers`: record `id` → `{to address, rev uint64, expiresAt int64}`.
  There is at most one offer per record, and a new offer replaces the old
  one. It is removed on accept or cancel. Stale offers are never iterated, so
  they cost nothing to leave behind.
- Who can call: `OfferTransfer`, `CancelTransfer` and `AcceptTransfer` must
  be **direct account calls not signed with a delegated session key**.
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
//   status "disabled": the verifier is not enabled yet
//   status "unowned":  not a registered name (address namespaces included)
//   status "deleted":  the name belongs to a deleted record
//   status "owned":    owner is the record's address; the name may be the
//                      record's current name or any of its aliases
//
// Pause does not change the answer. Pause controls operations, not identity.
func NamespaceOwner(namespace string) (owner address, status string)
```

### 5. Keeper: effective ownership

Every row of the table in Context, except new deploys, goes through one
resolver:

```go
// effectiveOwner returns who owns pkgPath's namespace for ownership-gated
// operations. stamped is the creator to fall back to (the live package's,
// or the parked one's when nothing is live).
func (vm *VMKeeper) effectiveOwner(ctx sdk.Context, st gno.TransactionStore,
    params Params, pkgPath string, stamped crypto.Address) (crypto.Address, error)
```

| `NamespaceOwner` status | Effective owner |
|---|---|
| `owned` | the record's address |
| `unowned` | the stamped creator, as today |
| `deleted` | none: ownership-gated writes are refused (frozen; recovery is a governance matter) |
| `disabled`, names realm not deployed, or `SysNamesPkgPath` empty | the stamped creator, as today |
| the call errors (missing function, wrong return shape, panic) | **none**: the write is refused. An error never falls back to the creator |

Applied per operation:

- **Redeploying a live private package.** The submitter must be the effective
  owner. This **replaces** the namespace check for this one case, so an owner
  can update private packages under an alias as well as under the current
  name.
- **Pause.** Today pause is enforced *inside*
  `IsAuthorizedAddressForNamespace`. Dropping that call on the redeploy path
  would therefore drop pause as well. The keeper adds an explicit check:
  when names is deployed and `names.IsEnabled() && names.IsPaused()`, it
  refuses every submission, park and activation, including ones authorized
  by ownership. Before `Enable` there is no pause, as today.
- **Parking.** A park is a deploy, so it first passes the deploy gate that
  applies to its path: the namespace check if nothing is live there, or the
  effective-owner check if a private package is live. The park guard then
  applies as well: if another submission is already parked at the path, the
  new submitter must be that parked submitter, as today, or the effective
  owner. Ownership does not put the owner's address on the old bytes. The
  owner's submission replaces them and carries its own stamp, deposit and
  `init()` identity.
- **Activating.** Over a live private package, the parked creator must be the
  effective owner *when the package is activated*, because the name may have
  moved since it was parked. `redeployBlockedReason` mirrors this, and
  `ReasonOwnerMismatch` is reworded from "can never be enabled" to
  "currently blocked by owner mismatch", since the mismatch can now go away.
- **Rejecting.** Allowed for the parked submitter (they can always withdraw
  their own bytes), an approver, or the effective owner of the namespace
  whether or not anything is live there. This keeps the existing reason for
  the live owner's standing (a stranger's parked blob blocks the owner) and
  extends it to the namespace owner.

Implementation constraints:

- **Store.** The resolver takes the caller's transaction store.
  `QueryPackageMeta` passes its throwaway store. It must not use the
  context-bound store that `callRealmBool` fetches.
- **Ordering.** It runs before `init()` and before storage-deposit
  accounting, under the same read-only contract as `callRealmBool`.
  `NamespaceOwner` must stay a pure read.
- **Genesis replay.** During `IsGenesisReplay` (every InitChain-delivered
  transaction, not just height 0), the resolver is **not called**, so replay
  never needs a query API that did not exist yet. The live-redeploy checks
  stay waived as they are today.
  - The park guard and the owner branch of `RejectPackage` are waived too.
    Otherwise a cross-address re-park or reject that ownership allowed after
    the upgrade would fail on replay.
  - This assumes replayed history contains only transactions the source
    chain accepted. Confirm that against the genesis-replay ADR
    (`pr5511_chain_upgrade_genesis_replay.md`) before implementing.
- **Gas.**
  - Ownership-gated paths (live private redeploy, park over a park, activate
    over a live package, reject) pay for a metered realm call instead of a
    string compare.
  - Replacing or activating over a live public package is still refused
    before that call. Rejecting a parked submission at a live public path
    may still need it.
  - Measure with identical workloads, with both warm and restarted stores.

On the `disabled` / unset row: falling back to the stamped creator could hand
control back to a compromised old key if governance cleared
`SysNamesPkgPath`. We accept that. Clearing it already lets *any* address
deploy into *any* namespace (`checkNamespacePermission` returns nil), so the
fallback is not the weakest point in that state.

### Authorization matrix after this ADR

Namespace states:
- **current**: the record's current name;
- **alias**: a name the record has renamed away from;
- **deleted**: a name belonging to a deleted record;
- **address**: an address namespace;
- **off**: names realm disabled, not deployed, or `SysNamesPkgPath` unset.

| Operation | current | alias | deleted | address | off |
|---|---|---|---|---|---|
| New deploy | owner | nobody | nobody | that address | anyone |
| Redeploy live private | owner | owner | nobody | stamped creator | stamped creator |
| Park | the deploy row that applies (new deploy, or redeploy live private), **and** if something is already parked there: its submitter or the owner | same | same (owner: nobody) | same (owner: stamped creator) | same (owner: stamped creator) |
| Activate over live private | parked creator must be owner | same | refused | parked creator must be stamped creator | same as address |
| Reject parked | submitter, approver, owner | same | submitter, approver | submitter, approver, stamped live creator | same as address |

"owner" means the record's current address; in the address and off columns
it is the stamped creator. When names is deployed, enabled and paused, it
blocks every deploy, park and activation. The off column keeps today's
behaviour, where pause has no effect. Rejection is never paused.

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
  the `init()` identity and the deposit payer. Rewriting it would
  misattribute bytes and deposits the new owner never consented to.
  Rejected.

## Consequences

- **Rights move to the name holder.** Anyone who deployed a private package
  under a name they do not hold loses control of it to the record's holder.
  Once `names` is enabled, that can only have happened through genesis or
  before `Enable`. The pre-registered system names in
  `namereg/v0/preregister.gno` (`sys`, `gov`, `gnoland`, `gno`, `nt`, `demo`,
  `archive`, `x`, `gnolang`, `test1`) are the main exposure.
  - **Required as part of the upgrade proposal:** list every live private
    package in a namespace that is, or could become, a registered name. That
    includes namespaces that are `unowned` today, because a later
    registration (or a DAO rename onto that name) would take their packages
    over.
  - Each case where owner and creator differ must be explicitly accepted.
  - Every unowned namespace that has legacy packages must be **reserved**
    for an agreed owner (pre-registered) before the rule is turned on.
  - Under the new rule the owner can redeploy directly, so no separate
    re-stamping route is needed. The audit decides who gains control.
- **Renamed users regain control of their old namespaces.** Today a rename
  locks the user out of their own private packages under the old name. With
  aliases recognized, they can redeploy them again. New deploys under an
  alias stay refused.
- **Deleting a record freezes its packages.** Ownership-gated operations on
  private packages under a deleted record's names are refused until
  governance acts.
- **T1 voters gain a new power.** They can hand control of a record's private
  packages to another address. That is the price of key recovery. Adding a
  controller now also needs a T1 vote.
- **Address-keyed state outside the registry does not move.** `ownable`
  owners, admin lists and balances in other realms stay with the old address.
  The docs must say so.
- **Off-chain consumers** (indexers, gnoweb `/u/<name>`) that cache
  address ↔ name need to handle `Transferred`.
- **Approvals bind source, not submitter.**
  - Content hashes exclude `[addpkg]`. If a new owner parks the same bytes
    under their own stamp, a standing approval for those bytes still
    matches.
  - `PkgHeight` catches a re-park in a later block, but submissions in the
    same block share a height.
  - `init()` runs as the parked creator. An approval that needs a specific
    `init()` identity must therefore bind the parked creator explicitly.
    That calls for an optional `PkgCreator` field on `MsgEnablePackage`,
    which is in scope for the keeper PR.

## Rollout

`r/sys/users`, `r/sys/names` and `r/sys/namereg/v0` are public packages
(none sets `private = true`), so they cannot be replaced in place.
`r/sys/users` holds the registry state and is not versioned. The change
therefore ships as **one coordinated chain upgrade**: realm changes, keeper
rule and migration audit together. Transfers must not be callable before the
keeper rule is live; otherwise a transferred namespace would leave its
private packages with no one able to redeploy them.

PRs can still be split for review: realms (sections 1–4), keeper
(section 5), and the audit tool or report. They merge behind the same
upgrade.

## Testing

- **`r/sys/users`:**
  - aliases follow the transfer;
  - the old address no longer resolves, can register again, and gets a new
    `id`;
  - live and deleted recipients are refused;
  - a deleted user cannot transfer;
  - a non-controller is refused;
  - a copy from the read-only tree is refused by `TransferTo`,
    `UpdateName` and `Delete`;
  - `canonicalStore` and the exact-name reservations are unchanged;
  - `rev` increments on rename, delete and transfer;
  - the event is emitted;
  - the T1 proposal works end to end, survives a rename by the thief while
    the vote is open, and is refused if the record is deleted;
  - an id-bound rename or delete proposal does not touch a different
    identity after address reuse.
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
    new record.
- **`r/sys/names`:** `NamespaceOwner` for each status: before `Enable`, while
  paused (unchanged answer), an address namespace, an alias, a deleted
  record, and a current name.
- **Keeper:** every cell of the authorization matrix, plus:
  - an error from the names call refuses the write;
  - `QueryPackageMeta` uses its throwaway store and reports the reworded
    reason;
  - parking and approving public paths;
  - an old approval after a re-park by a new owner;
  - genesis replay of a history containing cross-address parks and private
    redeploys, before and after the upgrade.
- **Integration `.txtar`:** register → deploy a private realm → offer/accept
  → redeploy from the new key → old key refused, for both the current name
  and an alias.
- **Commands:** `go test ./gno.land/pkg/sdk/vm/ -run Gas`,
  `go test ./gno.land/pkg/integration/ -run TestTestdata`, and the examples
  suite.

## Open questions

- Offer expiry: is 7 days right?
- Should transfers cost a fee, or have a cooldown?
- Deleted records: should freezing be the end state, or should there be a
  governance path that assigns a deleted record's packages to a new owner
  without bringing the name back?
- Should `ProposeRegisterUser`, `ProposeUpdateName` and `ProposeDeleteUser`
  also move to T1 voters, for consistency with controller admission?
