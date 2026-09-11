# Preprocess a redeployed package at its newest index

## Status

Proposed

## The problem

`AddMemPackage` opens by taking the next package index counter and writing
`pkgidx:<n> = <path>`, unconditionally. `DeleteMemPackage` removes a path's two
mempackage blobs and nothing else. A redeploy over a live private package
therefore keeps the path's original index entry and appends a second one. Both
entries name the path, and both resolve through `GetMemPackage` to the
redeployed content.

That index is the order a node preprocesses packages in when it starts:
`VMKeeper.Initialize` drives `PreprocessAllFilesAndSaveBlockNodes` over
`IterMemPackage`, which walks the index from 1 upwards and de-duplicates
nothing. So the redeployed content is preprocessed twice, and the earlier of
those passes runs at the position the path held before the redeploy, below
everything deployed since.

If the redeployed content imports a package deployed after its own original
deploy, that pass asks for a block node of a package nothing has saved yet,
and `GetBlockNode` panics with `unexpected node with location <dep>:0:0`.
`Initialize` is called bare from `NewAppWithOptions`, with no recover anywhere
on the path, so the process dies. It dies on the first restart after the
redeploy and on every restart after that, while a node that has not restarted
keeps answering calls out of `cacheNodes`. The failure is keyed on restart
history, and the sanctioned upgrade path is a governance halt at which every
node restarts at once.

The trigger is ordinary maintenance: deploy a private realm, publish a
library, update the realm to use it. Two paths reach it, `AddPackage`'s
private-redeploy branch and `EnablePackage`'s live-blob branch, and both hand
the store the same sequence, `DeleteMemPackage(path)` then
`RunMemPackage(mpkg, true)`.

## The decision

The boot pass reaches a path once, at its highest index.

`IterMemPackage` reads through `indexedPackagePaths`, which walks 1..counter
and returns each path once, at the highest entry naming it. That entry is
where the path's current content was stored, so a redeploy is preprocessed
after everything it could import, and one pass covers it however many times it
has been redeployed.

The surviving entry has to be the newest one. Keeping the first would satisfy
every count, leave the position inverted, and preserve the panic exactly.

Moving a path down the order cannot invert the opposite edge, a package that
imports it being preprocessed first, because nothing can hold that edge. Only
a path whose live package is private can be re-added at all, a public package
can never be replaced so it can never become private, and the importer refuses
a private path with `ImportPrivateError` at every deploy, genesis included. A
flat order cannot satisfy both directions at once; it does not have to.

## Scope

The writer is untouched, so a redeploy still leaves one dead index entry
behind. Nothing reads it. `IterMemPackage` is the index's only reader:
`FindPathsByPrefix` ranges the `pkg:` keyspace in the merkleized store and
de-duplicates the `#allbutprod` sibling itself, and `NumMemPackages` reads only
the counter. Removing the dead entry needs a path-to-index reverse key, which
puts a gas-metered read and write on every `AddMemPackage`, moves `GasUsed` on
every `MsgAddPackage` and `MsgEnablePackage`, and so needs a coordinated
height. That is index hygiene rather than a fix for this panic, and it belongs
in its own change, argued on its own terms.

`NumMemPackages` therefore keeps reporting index entries ever written, above
the number of packages `IterMemPackage` yields. Its only consumer on the chain
is `Initialize`'s `> 0` gate, which is load-bearing beyond skipping work on an
empty store: `IterMemPackage` returns a nil channel when there is no counter,
and ranging a nil channel blocks forever.

Under the `inert` code submission policy a park never reaches
`AddMemPackage`: `AddPackage` returns at `AddInertPackage`, which writes the
`inert_pkg:` blob alone. Index entries are written on enable and on
permissionless deploy only.

`DeleteMemPackage` without a re-add is unchanged, and still leaves the path's
index entry over a missing blob. `IterMemPackage` skips it, because
`GetMemPackage` answers nil. That is the shape
`gno.land/adr/pr6088_msgrun_allowlist_and_inert_charging.md` §5d describes for
a parked submission over a live path.

## Alternatives considered

**Prune the stale entry when the path is re-added.** The write side of the same
fix, through a `pkgidxrev:<path>` key that names the entry to drop. It changes
`GasUsed` by 83,028 at the default gas config on every `AddMemPackage`, moves
four pinned integration gas numbers, and leaves a permanent reverse key per
path. It buys no reader anything: keep-newest on read already yields the same
order whether or not the writer pruned.

That gas delta needs a coordinated height even though the index keyspace sits
outside the app hash, because consumed gas does not stay node-local. Every
delivered transaction is charged to the block gas meter, and gnoland's
`EndBlocker` calls `auth.EndBlocker`, which calls
`GasPriceKeeper.UpdateGasPrice`: it reads `ctx.BlockGasMeter().GasConsumed()`,
computes the next price from it through `calcBlockGasPrice`, and writes the
result under the keeper's store key, which `NewGasPriceKeeper(mainKey)` binds
to the merkleized main store. A block whose transactions consume different gas
on two binaries can carry a different `LastGasPrice`, and where the old price
and the new one differ in whether they move at all, one binary writes and the
other does not.

Whoever takes the prune also has to relax the reader in the same change:
`indexedPackagePaths` treats a missing entry below the counter as corruption
and panics, which is what a freed entry would look like.

**Rebuild the index at boot from the `pkg:` keyspace.** The rebuild would be in
path order, not deploy order, so a package would be preprocessed before its
dependency whenever the dependency's path sorts after it.

**Rewrite existing indexes in a one-time migration.** A step to write, schedule
and verify, on chains whose nodes are already dying, producing exactly the order
keep-newest gives for free.

## Consequences

- No gas moves, because nothing here writes to a store. The change is one
  reader of unmerkleized keys, and `Initialize`'s store comes from
  `gno.NewStore` with a nil gas context, so the boot reads are unmetered as
  before.
- No chain state changes and no hash moves. A node with duplicate entries
  already written boots correctly the moment it runs this binary, with nothing
  to repair first. Nodes take the fix one at a time.
- A boot pass over a redeployed path costs one preprocess of it, not one per
  redeploy it has taken.
- One dead index entry per redeploy keeps accumulating in the base store,
  unread. It costs the counter's over-report and the key's bytes.

## What we checked

- A redeploy is preprocessed once, above a dependency deployed in between, and
  the boot pass completes. Driven through both flows that can redeploy,
  `MsgAddPackage` alone and park plus `MsgEnablePackage`.
- The same redeploy with the dependency deployed first, which has no position
  to invert, boots clean before and after.
- At store level, three index entries over two paths yield two packages, the
  re-added one last.
- A whole node, over its own persisted data dir: an integration script deploys
  the realm, publishes the library, redeploys the realm onto it, and restarts
  the node, which comes up and answers a call to the realm out of the library.
  On the unfixed reader the same script reaches
  `NewAppWithOptions` → `VMKeeper.Initialize` →
  `PreprocessAllFilesAndSaveBlockNodes` → `GetBlockNode` and panics
  `unexpected node with location gno.land/p/demo/invdep:0:0`. Only the passing
  direction is asserted: the integration harness runs the node in a goroutine
  with no recover, so a boot panic aborts the test binary rather than failing
  one script, and a start failure is reported with `ts.Fatalf`, which the
  script's `!` prefix cannot reach.
- The VM, keeper and integration suites pass, with every pinned gas number in
  the integration suite unchanged.

## AI assistance

Written with AI assistance (Claude Code): the boot panic was reproduced
test-first through both redeploy flows, and review cut the change from a
two-sided fix to this read-side one. The human author reviewed and owns the
change.
