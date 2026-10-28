# ADR: gnoweb prices its gnokey commands at the chain's gas price

## Context

The gnokey commands gnoweb offers for copy-paste (the Actions page and the
`<gno-form exec>` command block) hardcoded `-gas-fee 1000000ugnot
-gas-wanted 1_000_000_000` (#6276): every copied call paid 1 GNOT, whatever
gas it used. The gas price is not a constant: tm2 moves it every block and
the node answers it on `auth/gasprice` (`1ugnot/1000gas` on mainnet and onyx
today). gnokey's default fast mode simulates before broadcasting and stops a
call that would run out of gas, charging nothing (#6313). Measured on onyx: a
wugnot Deposit uses ~1.4 M gas, a small addpkg ~2.6 M.

## Decision

- `-gas-wanted` is a constant 10,000,000: several times what common calls use,
  and a call that needs more is stopped by the simulation for free.
- `-gas-fee` is `ceil(10M × price.Amount × 3 / price.Gas)` in the price's
  denom: 30,000 ugnot (0.03 GNOT) today. tm2 moves the price by at least one
  unit per busy block, so at 1 ugnot per 1,000 gas the first busy block
  doubles it; ×3 covers two in a row. Once the amount is large, a full block
  raises it about 4.3% (compressor 10, target 70%), and ×3 covers a burst of
  about 26 full blocks. A fee that is too low is rejected at CheckTx, for free.
- A ugnot fee above 1 GNOT (the old hardcoded fee) means the price is
  unusable, like a zero, invalid or overflowing one. The command then uses the
  fallback, 1 ugnot per 1,000 gas, and the note says "estimated … (the chain's
  price was unavailable)". Omitting `-gas-fee` would leave a command that does
  not run; failing the page would block reading a realm over a value with a
  sane default.
- The price is fetched lazily: the render context carries a `sync.OnceValue`,
  called only by a `<gno-form exec>` or by an Actions page with functions.
- `ClientAdapter.GasPrice` caches stale-while-revalidate: once a price exists
  it is returned at once, and after 30 s one background query refreshes it. A
  failed query keeps the last good price and is retried after 30 s. Only the
  first read waits, bounded at 2 s. No caller holds a lock across the query.
- A quiet note under the command gives the fee, gas, price, the ×3 headroom, that unused
  gas is not refunded, and that fast mode simulates first.

## Alternatives considered

- **Simulate each call server-side** for an exact gas. One node simulation per
  function per page view, with arguments the user has not typed yet, and a
  realm could make its own page expensive to render. gnokey already simulates
  at broadcast, with the real arguments.
- **A fixed lower fee** (e.g. `10000ugnot`). Cheaper today, but the price is
  dynamic: a busy chain raises it and the command fails with no hint why.

## Consequences

- A copied call costs 0.015 GNOT at today's price instead of 1 GNOT.
- At most one `auth/gasprice` query per 30 s per gnoweb process, and none for
  pages without a command.
- The fallback is in ugnot: on a chain with another fee denom the fallback tx
  is rejected at CheckTx, which charges nothing.
- `auth/gasprice` does not expose the node operator's own `min_gas_prices`; a
  node with a higher floor rejects the tx at CheckTx, also for free.
- In secure mode `gnokey broadcast` does not simulate: a call needing more
  than 10 M gas is broadcast and charged the fee.
- gnoweb-only change: no chain or gnokey change.
