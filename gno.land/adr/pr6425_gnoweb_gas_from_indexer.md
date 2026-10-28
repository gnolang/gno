# ADR: gnoweb sizes the fast-mode command's gas from recent calls

## Context

#6422 gives every copied gnokey command `-gas-wanted 10_000_000`, priced at
the chain's gas price ×3. Unused gas is not refunded, so a call that uses
1 M pays for 10 M. An indexer (#6231) knows what recent calls used.

## Decision

- The fast-mode command asks for twice the most gas any of the function's
  last 20 successful calls used, rounded up to 100 k, within [1 M, 10 M].
  Fewer than 3 calls, no indexer or a failed read keep 10 M.
- The secure-mode commands (`-broadcast=false`, sign, `gnokey broadcast`)
  keep 10 M and #6422's fee: they do not simulate.
- The price is decided once, at 10 M (1 GNOT cap, fallback), then applied
  to the measured gas, so a page shows one price.
- Only a successful single-`MsgCall` transaction into the package is
  measured, with a Go identifier of at most 64 bytes for a function and a
  positive gas use. gnoweb re-checks every row: on onyx, the indexer's
  wugnot filter also returns `auth/create_session` rows (~1.36 M, typed
  `UnexpectedMessage`) that would pass for cheap Deposits, while real
  Deposits use 7 to 9 M.
- One read of a package's last 100 transactions, cached 5 minutes (state
  growth or a chain upgrade changes gas use; a path cannot be redeployed),
  30 s after a failure, matching the indexer breaker. A page never waits:
  the first one shows 10 M while the read fills the cache. Stale entries are
  served while one read refreshes them; at most 4 reads run at once; a full
  cache (1,024 packages) is cleared.

## Trust model

Indexer data is not consensus data. It can only lower the fast-mode gas:
gnokey simulates that command first, so too little gas fails before any fee
is charged. A wrong indexer, or a few cheap successful calls, cannot make a
secure-mode transaction fail and still pay, and no fee exceeds #6422's.

## Alternatives

- **Lower the secure-mode gas too**: a wrong measure would cost the fee of
  a failed transaction. **Wait for the first read**: a slow indexer would
  slow pages. **Simulate server-side**: rejected in #6422.

## Consequences

A typical fast-mode fee drops by 10 M over its measured gas, for at most one
indexer read per package per 5 minutes. `-gas-wanted auto` (#3703) stays
the exact answer once the arguments are known.
