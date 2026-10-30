# ADR: gnoweb prices its gnokey commands at the chain's gas price

## Context

The gnokey commands gnoweb offers for copy-paste (the Actions page and the
`<gno-form exec>` command block) hardcoded `-gas-fee 1000000ugnot
-gas-wanted 1_000_000_000` (#6276): every copied call paid 1 GNOT, whatever
gas it used. The gas price is not a constant: tm2 moves it every block and
the node answers it on `auth/gasprice` (`1ugnot/1000gas` on mainnet and onyx
today). gnokey's default fast mode simulates before broadcasting and stops a
call that would run out of gas, charging nothing (#6313). A wugnot Deposit
uses 7 to 9 M gas on onyx; a small addpkg ~2.6 M on a local chain. An
optional indexer (#6231) knows what recent calls used.

## Decision

- **Gas.** Both commands ask for 10,000,000 gas by default: several times
  what common calls use. With an indexer, the fast-mode command asks for
  twice the most gas any of the function's last 20 successful calls used,
  rounded up to 100 k, within [1 M, 10 M]; fewer than 3 calls, no indexer
  or a failed read keep 10 M. The secure-mode commands (`-broadcast=false`,
  sign, `gnokey broadcast`) always keep 10 M: they do not simulate.
- **Fee.** `ceil(gas × price.Amount × 3 / price.Gas)` in the price's denom:
  30,000 ugnot (0.03 GNOT) for 10 M today. tm2 moves the price by at least
  one unit per busy block, so at 1 ugnot per 1,000 gas the first busy block
  doubles it; ×3 covers two in a row. Once the amount is large, a full block
  raises it about 4.3% (compressor 10, target 70%), and ×3 covers about 26
  full blocks. A fee that is too low is rejected at CheckTx, for free.
- **Price.** Decided once, at 10 M: a ugnot fee above 1 GNOT (the old
  hardcoded fee), or a zero, invalid or overflowing price, falls back to
  1 ugnot per 1,000 gas, and the note says "an estimated … (the chain's
  price was unavailable)". The same price is applied to the measured gas,
  so a page shows one price and no fee exceeds the 10 M one.
- **Indexer data.** Only a successful single-`MsgCall` transaction into the
  package counts, with a Go identifier of at most 64 bytes for a function
  and a positive gas use. gnoweb re-checks every row: on onyx, the
  indexer's wugnot filter also returns `auth/create_session` rows (~1.36 M,
  typed `UnexpectedMessage`) that would pass for cheap Deposits.
- **No page waits.** The price is fetched lazily (a `sync.OnceValue` in the
  render context, called only by a `<gno-form exec>` or an Actions page
  with functions) and cached stale-while-revalidate for 30 s; only its first
  read waits, bounded at 2 s, and a failed query keeps the last good price.
  Indexer reads never block a page: the first view shows 10 M while one
  background read of the package's last 100 transactions fills a cache kept
  5 minutes (30 s after a failure), with at most 4 reads in flight and
  1,024 packages, cleared when full.
- **Note.** Each mode has its own two-line note under the command, which
  follows the mode switch and is never copied: the fee, the gas (and where
  a measured gas comes from), the price, the ×3, no refund, and whether the
  mode simulates.

## Trust model

Indexer data is not consensus data. It can only lower the fast-mode gas,
which gnokey simulates before charging anything. A wrong indexer, or a few
cheap calls made to drag the measure down, cannot make a secure-mode
transaction fail and still pay.

## Alternatives considered

- **Simulate each call server-side** for an exact gas: one node simulation
  per function per page view, with arguments the user has not typed yet,
  and a realm could make its own page expensive to render. gnokey already
  simulates at broadcast, with the real arguments.
- **A fixed lower fee** (e.g. `10000ugnot`): cheaper today, but a busy chain
  raises the price and the command fails with no hint why.
- **Lower the secure-mode gas too**: a wrong measure would cost the fee of a
  failed transaction. **Wait for the first indexer read**: a slow indexer
  would slow pages. **Poll the indexer**: load for packages nobody views.

## Consequences

- A copied call costs 0.03 GNOT at today's price instead of 1 GNOT, less in
  fast mode when the function's recent calls are cheaper.
- At most one `auth/gasprice` query per 30 s and one indexer read per
  package per 5 minutes; none for pages without a command.
- The fallback is in ugnot: on a chain with another fee denom the fallback
  tx is rejected at CheckTx, which charges nothing. `auth/gasprice` does not
  expose the node's own `min_gas_prices`; a node with a higher floor also
  rejects at CheckTx.
- In secure mode `gnokey broadcast` does not simulate: a call needing more
  than 10 M gas is broadcast and charged the fee.
- gnoweb-only change. `-gas-wanted auto` (#3703) in gnokey stays the exact
  answer; gnoweb can print it once that gnokey release is widespread.
