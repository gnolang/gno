# A redeployed realm keeps its object clock

## Status

Proposed

## The problem

Only a private realm can be deployed over. `AddPackage` refuses to re-add
over a live public package and `EnablePackage` applies the same rule, so
both paths admit a second deployment at a path a private realm already
occupies.

Both build the machine with an empty `MachineOptions.PkgPath`, and
deliberately: loading the live `PackageValue` puts it in the object cache,
and `SetCachePackage` then panics on the package the run is about to
build. With no package on the machine, `runMemPackage` takes its
fresh-package branch, `PackageNode.NewPackage` calls `NewRealm`, and the
deployment runs against a realm record whose `Time`, `Storage` and
`Deposit` are all zero. The deploy path is the one path that never reads
the record persisted at that path. `fillPackage` reads it on load,
`GetRealmByID` for a cross-realm borrow, and `processStorageDeposit` and
`QueryStorage` read it by path.

Three things follow from that one blank record.

`Realm.assignNewObjectID` mints an ObjectID as `{PkgID, Time}` after
incrementing `Time`, so the second deployment's objects take the ids the
first deployment's objects hold. Two objects then carry one ObjectID at
different times, which is exactly what a finalized object's tx-stamped
`NewTime` is supposed to rule out.

`defaultStore.SetObject` writes an escaped object's value hash into the
merkleized `"main"` store under the bare ObjectID, and `DelObject`, its
only deleter, runs only for the ids the realm's own deleted marks name. A
realm that restarts its counter never names the ids it minted before, so
every escaped ObjectID above the restarted high-water mark keeps a hash in
the app hash under an id the realm will mint again.

`processStorageDeposit` prices the transaction from `RealmStorageDiffs()`
and locks the deposit against `rlm.Storage` and `rlm.Deposit`, both zero.
The record left by the redeploy accounts for the second deployment alone,
while the escrow at the realm's storage-deposit address holds both
deposits. The gap is unrecoverable: refunds are sized from the record, no
message frees a live realm's bytes, and the escrow address is a truncated
hash of the realm path rather than of a public key, so nothing can sign for
it either.

On a private realm redeployed with strictly smaller source, `Time` goes 22
to 4, and the escrow holds 1,077,000ugnot against a record accounting for
172,000ugnot.

## The decision

A deployment over a live package takes over the realm persisted at its
path.

`Machine.RunMemPackageOverRealm` takes that record and installs it on the
new `PackageValue` in place of `NewPackage`'s blank one. `Time` continues,
so every id the new objects mint sits above every id the path already
holds. `Storage` and `Deposit` carry the prior baseline, so the record
accounts for what the escrow holds. `RunMemPackage` is that call with no
record, which is every other deployment.

The keeper reads the record only on the branch where it has already
established that a package is live at the path: `pv != nil` in
`AddPackage`, `liveBlob != nil` in `EnablePackage`. Reading it in the VM
instead would charge a first deployment for a lookup of a record that
cannot exist.

A record that is not the path's own, or a run that does not save, is
refused with a panic beside the storable-type check. Neither is
detectable afterwards: ObjectIDs are minted off whichever counter
arrives, and the record writes itself back under its own path, so a
record from elsewhere both mints the package value off the wrong counter
and saves over the realm it came from.

### The package value keeps its reserved ObjectID

A package path resolves to one ObjectID. `ObjectIDFromPkgPath` returns
`{PkgID, 1}`, and seven sites read the package value there: `GetPackage`,
`pkgPathFromPkgID`, `isPkgPrivateFromPkgID`, `isPkgEphemeralFromPkgID`,
`PushFrameCall`'s receiver and closure borrows, and `DidUpdate`'s
immutable-package check. Four of them panic on anything else:
`pkgPathFromPkgID` and `isPkgPrivateFromPkgID` with their own message, and
`GetPackage` and `PushFrameCall`'s receiver borrow through a bare
`.(*PackageValue)`.

That id came out of the mint by arithmetic, the package value being the
first object a realm starting at `Time` 0 marks, so a continued counter
mints it anywhere but 1 and leaves the previous deployment's package value
answering every call at the id the path resolves to. The first query after
such a redeploy panics with a bounds error.

`assignNewObjectID` therefore stamps a realm's own package value with the
reserved id rather than minting one, and raises `Time` to at least that
id. For a realm starting from zero that is what the arithmetic already
produced. For a redeployment it replaces the object the path resolves to
and leaves the counter where it stood, so the rest of the deployment mints
above the high-water mark.

## Scope

Both redeploy paths, through the one branch they share. The fresh-package
branch of `runMemPackage` is where every deployment that does not hand the
machine a package builds one, which is what `AddPackage`'s private branch
and `EnablePackage`'s live-blob branch both do.

A first deployment is unchanged, down to its gas: nothing is live at the
path, so no record is read and the blank realm stands.

## What this does not do

Evict the previous deployment's objects. They stay in the base store,
unreachable, and their escaped hashes stay in the merkleized store. What
changes is that their ids are now inside the live realm's own range
instead of above it, so nothing mints over them, and the realm's `Storage`
and `Deposit` account for them, so an eviction landing later has a record
to size a refund from. No message evicts a live realm's objects, and
freeing them needs an answer to who receives the released deposit, which is
locked from the signer of each transaction that wrote the bytes and
released to one address. Both are open.

Because the previous deployment's bytes are still there, the deposit a
redeploy charges is the whole of what the new deployment adds, not the
difference between the two source versions. The creator pays for bytes
that exist, which is what the deposit is for. That those bytes should not
exist is eviction's problem, not this one's.

The package value's own bytes are counted twice. It is saved as a created
object at the reserved id, so its full size enters the storage diff, while
the inherited `Storage` still counts the copy it overwrote.

## Alternatives considered

**Refuse to deploy over a live realm.** One guard, and it closes the
object-identity and the deposit problem together by making both
unreachable. It also stops every historical redeploy transaction from
replaying, so forking a chain that has one in its history needs that
transaction patched out. The exemption that admits a private redeploy was
added deliberately, so withdrawing it is a separate decision.

**Evict the previous object graph and charge the difference.** The
complete fix, and out of reach here. Eviction has to walk an unbounded
`oid:<pkgid>:` keyspace with no reverse-reference index, and
`processStorageDeposit`'s release branch pays a single address, the caller
it was handed or `StorageFeeCollector` while ugnot is restricted, with no
per-depositor ledger to split it among the signers who funded the bytes.

**Read the record inside the VM, on every deployment that saves a realm.**
Eight lines, no new parameter, and no caller can forget it. It also
charges every `/r/` `AddPackage` for one base-store read of a key that is
absent on all but a redeploy: 59,000 gas, `ReadCostFlat`, which moves the
exact-gas pins in `gnokey_gasfee.txtar` and
`addpkg_testfile_restart_gas.txtar` and the suggested gas the documented
simulate-then-broadcast workflow reports to users. A redeploy fix should
not reprice every deploy, so the read moved to the callers that had
already answered the question. The regression tests, one per path, are
what a third caller would break.

**Save the new package value as an update of the old one** rather than as
a created object at the same id. That would charge its true size
difference instead of counting its bytes twice. `saveUnsavedObjects` does
not recurse into an updated object's children, on the grounds that they
were persisted through the created list, and on a redeployment none of
them were: every child is new. Sizing the diff correctly while keeping
the created walk means reading the previous package value to learn its
size, which puts a second object at the reserved id in the object cache.

**Hand the machine the live package.** `MachineOptions.PkgPath` set to the
path loads the live `PackageValue` through `fillPackage`, which already
carries the persisted realm. It also puts that package in the object
cache, and `SetCachePackage` then panics on the package the deployment
builds. That is why both call sites pass an empty path.

## Consequences

This moves consensus, and it does not repair the state a past redeploy
left behind. Every node has to change binary at the same point, and a
chain that has already redeployed a private realm needs a state repair
this change does not carry.

A validator on the old binary computes different values for a block that
redeploys a private realm:

- the ObjectIDs the deployment mints, and therefore the `oid:` keys it
  writes in the base store,
- the escaped-object hash keys it writes into the merkleized `"main"`
  store, which is the app hash,
- the realm record's `Time`, `Storage` and `Deposit`,
- the storage deposit charged, and therefore the creator's and the
  escrow's balances,
- the gas of that block's transaction, which now includes the read of the
  realm record.

Replaying the transaction stream under this binary does not produce a
clean store. The first deployment's objects are still written, the
redeploy still leaves them unreferenced, and their escaped hashes are
still in the merkleized store. No transaction can remove a raw `oid:`
key, so clearing them needs eviction rather than replay.

Upgrading a chain in place needs state repair that is not in this change.
Such a chain holds a realm record whose `Time` sits below ids the path
already minted, so the next redeploy continues from the lowered counter
and mints over the orphans between the two high-water marks. Raising it to
the true high-water mark means scanning that realm's `oid:` keyspace,
which is the same walk eviction needs. A replay from genesis does not
carry that problem: the record it computes is the one the fixed rules
produce.

A chain with no private redeploy in its history is unaffected.

The deposit numbers move as follows, on a realm whose first deployment
accounts for 9,050 bytes and whose second is strictly smaller source.
Before: the second deployment is charged 172,000ugnot for 1,720 bytes, the
record accounts for 1,720 bytes and 172,000ugnot, and the escrow holds
1,077,000ugnot, of which 905,000 is unaccounted and unrecoverable. After:
the second deployment is charged 172,900ugnot for the 1,729 bytes it adds,
the record accounts for 10,779 bytes and 1,077,900ugnot, and the escrow
holds 1,077,900ugnot. The 900ugnot difference in the charge is the larger
`NewTime` values in the encoded objects.

## What was checked

- `TestAddPackageRedeployKeepsTheRealmObjectClock` and
  `TestEnablePackageRedeployKeepsTheRealmObjectClock` pin a monotonic
  counter on each path, and each ends by evaluating a function the
  redeployed source declares, so a counter fix that stopped the redeploy
  from taking effect fails them.
- `TestEnablePackageRedeployKeepsTheRealmObjectClock` also pins that no
  escaped-object hash key survives above the live realm's high-water mark.
- `TestRedeployedPrivateRealmEscrowMatchesItsRecord` pins the escrow
  against the realm's own record, having first pinned that the two agreed
  after one deployment.
- `TestRunMemPackageOverRealmRefusesAForeignRecord` pins both
  preconditions on the new entry point, by the message each panic gives.
- `./gnovm/...` and `./gno.land/...` pass, the integration suite included.
  The exact-gas pins in `gnokey_gasfee.txtar` and
  `addpkg_testfile_restart_gas.txtar` are untouched, which is the check
  that a first deployment's gas did not move.
- `cd examples && go run ../gnovm/cmd/gno test ./...` passes.

## AI assistance

Written with AI assistance (Claude Code): the defect was reproduced
test-first on both deploy paths, and the reserved-ObjectID half came out
of the control those tests carry, which reddened when the realm record
alone was carried forward. The human author reviewed and owns the change.
