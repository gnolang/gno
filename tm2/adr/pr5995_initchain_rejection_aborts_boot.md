# PR5995: A Genesis the App Rejects Stops the Node

## Context

An ABCI app refuses a genesis by setting `ResponseInitChain.Error`.
`InitChainSync` passes that response through as a value and returns no Go
error, and `Handshaker.ReplayBlocks` checked only the Go error. So a node whose
app refused its genesis finished the handshake, saved the height-0 ABCI
responses as if the genesis had been applied, and kept running with no app
state.

gno.land's `InitChainer` refuses a genesis this way whenever `loadAppState`
fails: an `InitialHeight` that disagrees with the app state, a rejected
`SignerInfo`, an `AppState` of the wrong type. `gnogenesis fork test` replays a
genesis through an in-memory node to check it boots, and it printed `PASS` on
such a genesis. Its only guard compared the transactions delivered with the
ones expected, which is `0 < 0` on a genesis carrying none.

A refusal also leaves records behind. Before calling `InitChain`,
`ReplayBlocks` aligns the state record to the genesis `InitialHeight` and saves
it, and `LoadStateFromDBOrGenesisDocProvider` has already persisted the genesis
doc. On the next boot the persisted doc wins over the file, so an operator who
corrects `genesis.json` and restarts on the same data dir boots the old doc
again.

## Decision

### Honour the response error in the handshake

`ReplayBlocks` returns `InitChain rejected the genesis: <error>` when
`res.IsErr()`, before the height-0 responses are saved. The fix sits where
every app's response passes, so no app has to panic to be heard.

### Let a corrected genesis replace an unapplied one

The height-0 ABCI responses now mean that `InitChain` succeeded.
`genesisApplied` probes that key with `Has`. When it is absent and the file
names the same chain id and app hash, `resetToGenesisDoc` rebuilds the state
record from the file and replaces the persisted doc. Once the key exists the
persisted doc stays authoritative, as before. A failed probe fails the boot:
read as "not applied", it would rebuild the state record of a committed chain.

### Narrow the fork test's guard

`execTest` now receives the handshake error from `NewInMemoryNode` and never
reaches the guard on a refused genesis. The guard stays as a tripwire for an
`InitChain` loop that skips a deliverable transaction.

### Pass the default genesis state by value

`NewDefaultGenesisConfig` built `AppState` as a `*GnoGenesisState`, which
`loadAppState` refuses as an invalid type. The refusal used to be ignored; it
now stops boot, so the default config passes a value.

## Alternatives Considered

**Keep the fork test's guard as the detector.** It counts transactions, and
the refusals it names happen before the transaction loop. Rejected.

**Panic in `InitChainer` instead of returning the error.** The valoper
coverage check already does, and a panic does stop boot. Fixing the handshake
covers every app at once and keeps `ResponseInitChain.Error` meaningful.
Rejected.

**Gate the reset on an empty block store.** An earlier version of this branch
did, and it also reset a genesis whose `InitChain` had succeeded before the
first block committed. The height-0 key changes exactly when `InitChain`
succeeds. Rejected.

## Consequences

- A node whose app refuses the genesis exits during the handshake with the
  app's message, and `gnogenesis fork test` reports it as a failure.
- An operator can correct `genesis.json` and restart on the same data dir, as
  long as the chain id and app hash are unchanged.
- A data dir where an older binary already booted past a refusal carries the
  height-0 key, so the corrected file is ignored there until it is wiped.
- The reset leaves the refused genesis's validator and consensus-parameter
  records at its old `InitialHeight`. Only RPC reads those heights.
- A data dir whose genesis was applied before the height-0 save existed,
  gnolang/gno#1941, has no height-0 key. It reaches the reset only when its
  persisted genesis doc has no app state, which `saveGenesisDoc` began writing
  in gnolang/gno#5684, after it.
- `gnoland start` reads the genesis through `StreamingGenesisProvider`, whose
  loader drops `initial_height` and whose app state path runs neither the
  `InitialHeight` comparison nor `validateSignerInfo`. It boots a genesis the
  fork test refuses, from height 1; this change leaves that path as it is.

## Key Files

| File | Role |
|------|------|
| `tm2/pkg/bft/consensus/replay.go` | `ReplayBlocks` stops on `res.IsErr()` |
| `tm2/pkg/bft/node/node.go` | `genesisApplied`, `resetToGenesisDoc` |
| `contribs/gnogenesis/internal/fork/test.go` | `execTest` guard |
| `gno.land/pkg/gnoland/node_inmemory.go` | `NewDefaultGenesisConfig` |
