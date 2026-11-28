# ADR: Bounding The Proposer Replay Behind The `validators` Route

## Context

The `validators` RPC route takes any height from 1 to the current one and
returns the validator set at that height, proposer priorities included.
`state.LoadValidators` finds the last stored set at or below that height and
replays one round of proposer selection per height between them. A set is
stored only where it changed or every 100000 heights, so one request can
replay 99999 rounds, each a pass over every validator. Nothing was kept
between requests, so a client asking for heights far from a stored set paid
that replay again on each request.

## Decision

The route loads through `state.ValidatorsCache`, held by the RPC
`Environment`. The cache returns what `LoadValidators` returns. While it
replays, it keeps a copy of the set every 1000 rounds, keyed by the stored
height and the round count, and a later load resumes from the last kept copy
below its height. The cache remembers the DB it replayed from. A load that
replays from another DB first drops every kept copy, so a copy from one DB
never answers a load on another. The node gives its RPC environment one state
DB, so the node never drops them. Once a stretch between two stored sets has
been replayed, a load in it replays at most 1000 rounds. A mutex serialises
replays, so concurrent requests for a stretch nobody has replayed yet replay it
once.

Resuming from a kept copy must not rescale the priorities a second time, so
`ValidatorSet.AdvanceProposerPriority` runs the rounds alone and
`IncrementProposerPriority` calls it after its rescale.

Consensus, block execution and every stored byte stay as they were:
`LoadValidators` keeps its result and its callers, and the cache lives only in
the RPC environment.

## Alternatives considered

- *Cache results keyed by height*: a client asking for a new height each time
  never hits it.
- *Store a set more often*: changes what every node writes, and leaves the
  heights already written as costly as before.
- *Refuse heights far from a stored set on the route*: the bound is closed,
  and a client loses the validator set of most past heights.
- *Resume with `IncrementProposerPriority`*: its rescale was a no-op on every
  replayed set tried, which is evidence and not a proof, so the replay would
  rest on an invariant nothing states.

## Consequences

- The first load in a stretch costs what it cost before; later loads in it
  cost at most 1000 rounds where they cost up to 99999.
- Memory grows with the stretches clients ask for, at most one kept set per
  1000 heights of chain, 99 sets per fully replayed stretch. A kept set of 100
  validators takes about 7.5 KB, so 10 million heights would hold about 75 MB
  until the node restarts.
- `LoadValidators` in block execution still replays from the stored set on
  every block; its height comes from the chain, not from a client.
