# ADR: gnoweb prices its gnokey commands at the chain's gas price

## Status

Proposed.

## Context

Every gnokey command gnoweb offers for copy-paste (the Actions page and the
`<gno-form exec>` command block) hardcoded
`-gas-fee 1000000ugnot -gas-wanted 1_000_000_000` (#6276). The fee is paid
in full whatever gas the call uses, so every copied call cost 1 GNOT.

The gas price is not a constant: tm2's `GasPriceKeeper.UpdateGasPrice`
moves it every block, and the node answers it on `auth/gasprice`
(`{"gas":"1000","price":"1ugnot"}` on mainnet and onyx today). gnokey's
default `-simulate test` runs the call at the consensus max gas before
broadcasting and, if the gas wanted is too low, stops there, charges
nothing and prints a suggested gas-wanted (#6313). Measured on onyx: a
wugnot Deposit uses ~1.36 M gas, a small addpkg ~2.6 M.

## Decision

- `ClientAdapter` gets `GasPrice(ctx)`, which queries `auth/gasprice`. The
  RPC client caches the answer for 30 s behind one mutex, failures included,
  and bounds the query at 2 s on a context detached from the request, so a
  down node costs one slow query per 30 s and one abandoned request cannot
  cache a failure for everyone.
- `-gas-wanted` is a constant 10,000,000: several times what common calls
  use, and a call that needs more is stopped by the simulation for free.
- `-gas-fee` is `ceil(gasWanted × price.Amount / price.Gas × 1.5)` in the
  price's denom (15,000 ugnot = 0.015 GNOT today). The 50% margin covers the
  price rising between page load and broadcast; at these amounts it costs a
  fraction of a cent, while an underpaying tx is rejected at CheckTx.
- When the price cannot be read, is zero, invalid, or the fee overflows, the
  command uses 1 ugnot per 1,000 gas, gno.land's usual price, and the page
  says the price is estimated. Omitting `-gas-fee` would leave a command that
  does not run, and failing the page would block reading a realm over a
  value that has a sane default. If the real price is higher, the node
  rejects the tx before charging anything.
- The price reaches templates the way `ChainId`/`Remote` do: `HelpData` and
  `RealmRenderContext` → markdown `GnoContext` → `CommandData.Gas`.
- `ui/command` shows a quiet line under the command: the fee in GNOT, the gas,
  the price it came from, the margin, that unused gas is not refunded, and
  that fast mode simulates first. The fast-mode command carries a `#` comment
  saying the same; the copy button already strips `#` lines.

## Alternatives considered

- **Simulate each call server-side** and offer the exact gas. Every Actions
  page would run one simulation per function on the node, with arguments
  the user has not typed yet, and a realm could make its own page expensive
  to render. gnokey already simulates at broadcast, with the real arguments.
- **A fixed lower fee** (e.g. `10000ugnot`). Cheaper today, but the price is
  dynamic: a busy chain raises it and the copied command would start failing
  with no hint why.
- **A gnoweb flag for gas-wanted.** Not needed while one value fits; easy to
  add later next to `-chainid`.

## Consequences

- A copied call costs 0.015 GNOT at today's price instead of 1 GNOT.
- Each realm and Actions page reads the gas price, at most one node query
  per 30 s per gnoweb process.
- A call needing more than 10 M gas fails the fast-mode simulation for free
  and prints the gas to ask for; in full-security mode (no simulation on
  `gnokey broadcast`) it is broadcast and charged the fee.
- gnoweb-only change: no chain or gnokey change, rolling deploy.
- `$help&json` still carries no gas parameters; agents reading it pick their
  own.
