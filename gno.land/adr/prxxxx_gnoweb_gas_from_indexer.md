# ADR: gnoweb sizes a command's gas from the function's recent calls

## Context

#6422 gives every copied gnokey command `-gas-wanted 10_000_000`, and the
fee is that gas at the chain's price, ×3. Unused gas is not refunded, so a
call that uses 1.4 M pays for 10 M. When gnoweb has an indexer (#6231), the
gas each recent call used is known.

## Decision

- When an indexer is configured, `-gas-wanted` for a function is twice the
  most gas any of its last 20 successful calls used, clamped to
  [1 M, 10 M]. The fee is worked out from that gas exactly as in #6422
  (price ×3, 1 GNOT cap, fallback price).
- Only a successful transaction holding one message, a `MsgCall` whose
  package path and function match exactly, is measured. Each row is checked
  again in gnoweb: the indexer's filter has let unrelated transactions
  through (#6411). A multi-message transaction is skipped because its gas
  cannot be split between its calls.
- Fewer than 3 matching calls, no indexer, or an indexer error or timeout
  keeps 10 M and today's note. So does a size at the ceiling.
- ×2 covers a call whose cost grows with state or arguments (a longer list,
  a first-time storage write) beyond what recent calls showed. The 1 M floor
  keeps a cheap call's command from failing on a small variation.
- The note says where the gas comes from: "for up to 2.8 M gas (its last 20
  calls used at most 1.4 M)".
- One `RecentByPackage` read (the last 100 transactions) covers every
  function of a package, so the Actions page makes one. It is cached per
  package for 5 minutes and served stale while one background read
  refreshes it; a failed read keeps the last answer. At most 1,024 packages
  are kept. Only a page with a command reads it: the Actions page with
  functions, or a `<gno-form exec>`, through a `func(pkgPath, fn)` in the
  render context.
- A page waits for a package's first read at most 2 s, the bound #6422 uses
  for the first gas price read, and several commands on one page share that
  wait. The read itself may run up to 10 s and serves the next page.

## Trust model

Indexer data is not consensus data. The indexer can only lower the gas below
10 M, never raise it, and so never makes a command cost more than #6422's.
A wrong or hostile answer yields a command that runs out of gas: gnokey's
default fast mode simulates first and stops it before anything is charged.

## Alternatives considered

- **Simulate each call server-side.** Rejected in #6422: one node
  simulation per function per view, with arguments the user has not typed.
- **The median or a percentile.** Lower, but a call whose cost varies would
  fail more often; the maximum of recent calls is the case already seen.
- **One read per function.** The indexer filters by package; per-function
  reads would multiply queries for the Actions page.

## Consequences

- With an indexer, a typical call's copied fee drops by the ratio of
  10 M to its measured gas, e.g. 0.03 to 0.008 GNOT for 2.76 M gas.
- At most one indexer read per package per 5 minutes, none for pages
  without a command, and none without an indexer.
- A function rarely called among its package's last 100 transactions keeps
  10 M.
- `gnokey maketx call -gas-wanted auto` (#3703) simulates with the real
  arguments and stays the exact answer; this is an estimate for a command
  copied before the arguments are known.
- gnoweb-only change: no chain or gnokey change.
