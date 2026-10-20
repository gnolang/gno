# ADR: GRC20 v1 with native `uint256` amounts

## Status

Proposed. Builds on `gnovm/adr/prxxxx_native_int256.md` (the native
`int256`/`uint256` types).

## Context

GRC20 tokens (`gno.land/p/nt/grc20/v0`) count in `int64`. Tokens bridged
from Ethereum, and any token with 18 decimals and a realistic supply, need
256-bit amounts. The only option so far was `p/onbloc/uint256`, a mutable
pointer-shaped library type that does not fit the GRC20 security model
(values in realm state that other code can alias and mutate).

The VM now has native `uint256`. Its arithmetic wraps, as Go's does, so a
ledger built on it must use the checked helpers in `math/overflow`
(`Addu256`, `Subu256`, and the panicking `...p` variants).

Owner decisions, fixed before this work:

- chain coins (`chain.Coin`, the banker) stay `int64`;
- GRC20 gets a new package, `p/nt/grc20/v1`; v0 stays unchanged for what is
  already deployed;
- token arithmetic never wraps.

## Decision

### New packages

| Package | What it is |
|---|---|
| `p/nt/grc20/v1` | v0's API and security model with every amount, balance, allowance and the total supply typed `uint256`. |
| `r/nt/grc20reg/v1` | A separate registry for v1 tokens, same keying and wrappers as v0, `uint256` wrapper amounts. |
| `p/nt/treasury/v1` | `Balance.Amount` and payment amounts are `uint256`; `GRC20Banker` handles v1 tokens. |

`grc20/v1` behaviour, relative to v0:

- every add and subtract goes through `math/overflow`; Mint refuses an
  amount that would take the supply past `math.MaxUint256`
  (`ErrMintOverflow`), and an insufficient balance or allowance is detected
  by `Subu256`'s overflow flag rather than a separate comparison;
- `ErrInvalidAmount` is gone: a `uint256` cannot be negative;
- a zero balance or allowance is never stored (`setBalance`, and
  `Approve(..., 0)` removes the entry), so an absent key and zero are the
  same state and `KnownAccounts` counts non-zero balances only. v0 stores a
  `0` after a zero mint, transfer or approval;
- events carry `strconv.FormatUint256` decimals; `RenderHome` prints the
  supply in decimal; `FormatAmount`/`Token.FormatAmount` apply `decimals`;
- tellers, `IsCanonicalTeller`, the home binding of `CallerTeller`, token
  identity and the `NewToken` event are unchanged.

`treasury/v1`: chain coins are `int64`, so `NewCoinsPayment` now takes
`[]Balance` (`uint256` amounts), joins them per denom with a checked add, and
range-checks each total down to `[1, math.MaxInt64]`, panicking otherwise
(`ErrCoinAmountOutOfRange` names the rule). The check happens when the payment
is built, for the same reason v0 rejects non-positive amounts there: a
governance payment must fail before the vote, not at execution. The payment
then holds ordinary `chain.Coins`, so `Send` has nothing left to convert.
`CoinsBanker.Balances` widens the held `int64` amounts, which are never
negative, so the widening is exact.

### Dependants

Unversioned realms in `examples/` are edited in place: that is how the repo
has always evolved them (the `cur realm` migration changed all their APIs the
same way), and a deployed realm's code cannot change anyway.

| Realm | Decision | Public API |
|---|---|---|
| `r/demo/defi/foo20` | migrated to grc20 v1 + grc20reg v1 | `uint256` amounts |
| `r/demo/defi/grc20factory` | migrated | `uint256` amounts, including `initialMint` and `faucet` |
| `r/tests/vm/test20` | migrated (atomicswap's test token) | exports v1 `*Token`/`*PrivateLedger` |
| `r/demo/defi/atomicswap` | GRC20 swaps resolve through grc20reg v1 | `NewCustomGRC20Swap` takes a v1 `*Token`; coin swaps unchanged |
| `r/demo/disperse` | GRC20 dispersal uses grc20factory's v1 tokens | `DisperseGRC20` takes `[]uint256`; the string form parses amounts strictly with `strconv.ParseUint256` (it used to ignore parse errors); ugnot dispersal unchanged |
| `r/gnoland/wugnot` | migrated; converts at the ugnot boundary | `uint256` amounts, `Withdraw(uint256)` |
| quarantined packages | stay on v0 | unchanged |

For `maketx call`, a `uint256` parameter takes the same canonical decimal
string as an `int64` one, so command lines keep working (and now accept values
above 2^63). Realm callers must pass `uint256`, and `qeval` prints
`(N uint256)`.

wugnot is backed one to one by ugnot, so its supply is bounded by
`math.MaxInt64`. Deposit widens the sent amount; Withdraw takes a `uint256`
and, after the balance check, range-checks it to `int64` before sending
coins. The second check cannot fail while the invariant holds; it is there so
the conversion can never truncate silently.

### GovDAO treasury: stays on v0

`r/gov/dao/treasury/v0` and `r/gov/dao/impl/v0` are not changed:

- `r/gov/dao` versions its parts by path (`impl/v0`, `memberstore/v0`,
  `treasury/v0`) behind the unversioned proxy, and moves to a new
  implementation by a governance vote (`dao.UpdateImpl`). Moving the
  treasury to v1 the repo's way means a new `r/gov/dao/treasury/v1` realm
  plus an `impl/v1` that imports it, installed by a vote; editing
  `treasury/v0` in place would change the API of a versioned, deployed realm
  under its own version;
- a new treasury realm has a new address, and the funds stay at
  `treasury/v0`'s address until a v0 payment proposal moves them;
- the treasury's GRC20 token list is empty by default and set by
  governance, so nothing is lost today.

The consequence is that the DAO treasury handles v0 tokens only, and the demo
tokens and wugnot now are v1 tokens. `govdao_proposal_treasury_tokens_update`
supplies its own v0 tokens for that reason. A `treasury/v1` + `impl/v1` pair is
the follow-up when the DAO needs v1 tokens.

### Genesis scripts

`p/nt/grc20/v1` and `r/nt/grc20reg/v1` are listed next to their v0 entries in
the `misc/deployments/*/gen-genesis.sh` sets that list v0. wugnot pulls them in
as dependencies anyway, but `p/nt` paths cannot be added after genesis under
namespace enforcement, so they are explicit, as v0 is.

## Alternatives considered

1. **Change v0 in place to `uint256`.** Breaks every deployed token and its
   callers; ruled out by the owner.
2. **One registry for both versions.** `grc20reg/v0.Get` returns a v0
   `*Token`; holding both would need an interface or two getters, changing
   v0's API. Two registries keep each one typed.
3. **`uint256` amounts behind an `int64` API in wugnot.** wugnot never needs
   more than `int64`, but a v1 token's teller, registry wrappers and
   `Balance` all speak `uint256`; converting back at every entry point would
   add range checks with no benefit. The conversion lives at the one place
   amounts really are `int64`: the coins.
4. **Range-check coin payments in `CoinsBanker.Send` instead of at
   construction.** Simpler data, but an out-of-range payment would pass a vote
   and fail at execution, which v0 already avoids for non-positive amounts.
5. **Extend `uassert`/`ufmt` with 256-bit cases.** Useful, but those are
   shared `p/nt/.../v0` packages outside this change. Tests compare natively
   and print with `strconv` instead.

## Consequences

- Breaking API change for realm callers of foo20, grc20factory, test20,
  wugnot, atomicswap's `NewCustomGRC20Swap`, and disperse's `DisperseGRC20`.
- A token's v0 and v1 identities are unrelated: same symbol in the same realm
  can exist in both registries.
- Gas: a non-zero `uint256` is a boxed value in avl; storing it costs a few
  more bytes than an `int64`. Balances of zero cost nothing, since they are not
  stored.

## Key files

| File | Role |
|---|---|
| `examples/gno.land/p/nt/grc20/v1/token.gno` | the checked ledger |
| `examples/gno.land/p/nt/grc20/v1/README.md` | differences from v0 |
| `examples/gno.land/r/nt/grc20reg/v1/grc20reg.gno` | the v1 registry |
| `examples/gno.land/p/nt/treasury/v1/banker_coins.gno` | `uint256` to `int64` coin payments |
| `gno.land/pkg/integration/testdata/grc20_v1_uint256.txtar` | arguments, cap and persistence across a restart |
