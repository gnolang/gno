# ADR: `gnoland start` keeps a genesis file's initial height

## Status

Proposed.

## Context

A genesis carries its initial height twice. `GenesisDoc.InitialHeight` is the
top-level one the consensus engine and the block store follow, and
`GnoGenesisState.InitialHeight` sits inside `app_state` so tooling can read it.
`gnogenesis fork generate` writes both, and `applyInMemoryAppState` refuses a
genesis where they differ.

`gnoland start` reads genesis through `LoadStreamingGenesisDoc` instead, which
decodes top-level fields by name. It never named `initial_height`, so the field
was logged as unknown and dropped, and a hardfork genesis started its chain at
height 1. `applyStreamingAppState` had no mismatch check either.

Adding only the check would refuse every hardfork genesis, since the height it
compares against was always 0. So the loader and the check change together.

## Decision

`decodeTopLevelField` reads `initial_height` into the doc, on a cold cache and a
warm one, and both loaders call `validateInitialHeight`.

`applyStreamingAppState` panics on a mismatch or an unreadable app-level height,
as `enforceStrictReplay` does. An app refuses a genesis through
`ResponseInitChain.Error`, and the handshake reads only the Go error, so a
returned error boots a chain on whatever was written before it. Measured with a
returned error: top-level 1000, app-level 7, a chain at height 1000 with no
accounts, where the code before this change boots the full state at height 1.

## Alternatives considered

- **Return the error, as `applyInMemoryAppState` does.** It reads the same in
  both loaders, and it needs the handshake to read `ResponseInitChain.Error`
  first, which gnolang/gno#5995 proposes. Once that lands, both panics can become
  returns.
- **Read the field only on a first boot.** It keeps a bare-number field from
  stopping a restart, adds a mode to the loader for a chain nobody has named, and
  still opens a fresh node at the file's height.

## Consequences

A chain whose genesis sets no height sees no change: at the same base, a node
restarted onto this binary carries on from its height with the same app hash.

Three cases change for a genesis that sets one:

- A chain booted at height 1 on a file carrying the field keeps going on restart,
  since the persisted genesis doc wins, while a fresh node on that file opens at
  the file's height rather than at 1.
- A file carrying only the app-level height used to replay and now panics.
- A top-level height that amino cannot read, a bare number for one, stops the node
  at load on every boot, as any other unreadable top-level field does.
