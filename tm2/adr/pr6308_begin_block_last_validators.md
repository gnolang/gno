# Take BeginBlock's last validators from the state instead of reloading them every block

## Status

Proposed, in PR #6308.

## Context

`getBeginBlockLastCommitInfo` (`tm2/pkg/bft/state/execution.go`) builds the
`LastCommitInfo` of every `RequestBeginBlock`: the validators of
`block.Height-1`, in order, each flagged by whether its precommit is in
`block.LastCommit`. It loaded that set with `LoadValidators(stateDB,
block.Height-1)` on every block.

`LoadValidators` is not a plain read. `saveValidatorsInfo` stores a full
validator set only when the set changes and every `valSetCheckpointInterval`
(100000) blocks; every other height stores a reference to the last change. A
load for any other height takes the nearest stored set and calls
`IncrementProposerPriority(height - lastStoredHeight)` to reconstruct the
proposer priorities at `height`, so its cost is linear in the distance from
the last stored set and in the number of validators. Across a checkpoint
window the per-block cost climbs from nothing to up to 100000 increments.

Measured with pprof on a gno.land mainnet node (v1.5.0, 8 vCPU, 6 validators)
while it synced from genesis, `LoadValidators` under
`getBeginBlockLastCommitInfo` was 49% of CPU at height ~80000 (80000 blocks
past the checkpoint at 0) and 13% at ~116000 (16000 past the one at 100000).
A micro-benchmark of `IncrementProposerPriority(k)` on n validators runs at
2.6–3.6 ns per step per validator (Apple M-series), linear in both. None of
the work reached the app: the commit info reads addresses, powers and
positions, and proposer priorities are the one thing the replay computes.

The set is already in memory. For the state `ApplyBlock` is given,
`state.LastValidators` are the validators of `block.Height-1`, the set that
signed `block.LastCommit`: `updateState` sets `LastValidators` to the previous
`Validators`, which was the previous `NextValidators` by copy, and `saveState`
stores that same `NextValidators` at `LastBlockHeight+2`, the set
`LoadValidators` later returns. Members and order are identical; validator
sets are ordered by address at construction and `IncrementProposerPriority`
changes priorities and the proposer, never the order.

## Decision

When `state.LastBlockHeight == block.Height-1`, use `state.LastValidators`;
otherwise load by height as before. The equality holds on the `ApplyBlock`
path, where `state.ValidateBlock` has already required
`block.Height == state.LastBlockHeight+1`, and it is exactly the condition under
which `state.LastValidators` is the set for `block.Height-1`. It fails on
the `ExecCommitBlock` path: the restart handshake replays the blocks the app
is missing against the latest state (`consensus/replay.go`, `replayBlocks`),
so there `LastValidators` belong to a later height and the load stays.

`RequestBeginBlock` is byte-identical to before, so app hashes do not change.
The existing sanity check that the commit has one slot per validator is kept.

## Alternatives considered

- Lower `valSetCheckpointInterval`, for example to 1000. It bounds every load
  that remains (RPC `/validators`, evidence validation, the restart replay)
  and is backward compatible, since `LoadValidators` already falls back to
  the last change height when a checkpoint is missing. But it keeps a load,
  and so a disk read and a replay, on every block, and on data written by
  older binaries the new checkpoint heights are absent, so historical loads
  fall back to replaying from the last change height instead of from the
  old checkpoint. Worth doing on its own merits; not needed for the hot path.
- Cache the last loaded set in `BlockExecutor`. More state to keep correct,
  when the state the executor is handed already carries the set.
- Make `ExecCommitBlock` take the exact pre-block state so that the load
  can go entirely. The replay path has no historical state to hand it
  (its own TODO says so); out of scope.

## Consequences

- The per-block cost of the commit info is gone on the sync and consensus
  paths. Estimated from the profiles above, a sync from genesis spends
  about a third less CPU on average and about half less near the end of
  each checkpoint window; this is an estimate, not a measured resync.
- No change to what is stored, to the ABCI requests, or to any RPC answer.
- `getBeginBlockLastCommitInfo` now relies on the invariant that
  `state.LastValidators` is the set for `state.LastBlockHeight`, which
  `updateState` and `MakeGenesisState` maintain; the test
  `TestGetBeginBlockLastCommitInfo_LastValidators` pins both paths.

## Context the AI agent worked from

This change was prepared with an AI coding agent. It profiled the node
(`prof_laddr`, `go tool pprof`), read `execution.go`, `store.go`,
`validator_set.go` and `consensus/replay.go` at v1.5.0 and master to
establish the lineage and ordering of `LastValidators` and where
`ExecCommitBlock` is called from, benchmarked the replay in isolation, and
checked the test against master's code (it panics there, as the stored set
is deleted before the in-memory path is exercised). The speedup is derived
from the profile shares, not from a timed resync.
