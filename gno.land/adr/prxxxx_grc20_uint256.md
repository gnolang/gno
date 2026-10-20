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
- `RenderHome` escapes the token name and symbol with `sanitize.InlineText`.
  `validName` admits markdown (a name like `[Claim reward](https://evil.example)`
  is valid display text), and v0 prints it raw, so any v0 token page, including
  `grc20factory`'s, can carry a live phishing link. v1 escapes it. The v1 render
  paths of `grc20reg/v1` (`sanitize.InlineText`), `grc20factory` (path
  segments) and `p/nt/treasury/v1` (token keys via `sanitize.TableCell`,
  payment lines via `sanitize.InlineText`) do the same. **Follow-up:** v0
  (`grc20/v0` `RenderHome`, and `p/nt/treasury/v0`'s render) has the same
  injection and is left unchanged here.

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

### GovDAO: a v1 treasury and implementation, installed by a vote

`r/gov/dao` versions its parts by path (`impl/v0`, `memberstore/v0`,
`treasury/v0`) behind the unversioned proxy, and switches implementation by a
governance vote (`dao.UpdateImpl`). The uint256 treasury follows that
convention; v0 is not changed.

| Realm | What it is |
|---|---|
| `r/gov/dao/treasury/v1` | `treasury/v0` on `p/nt/treasury/v1` and `grc20reg/v1`: uint256 balances and payments, grc20 v1 tokens, coin payments range-checked to int64. Its own address and funds. |
| `r/gov/dao/impl/v1` | `impl/v0` with its treasury proposal builders (`NewTreasuryPaymentRequest`, `NewTreasuryGRC20TokensUpdate`) pointed at `treasury/v1` and `p/nt/treasury/v1`. Same member store (`memberstore/v0`) and builders otherwise. |
| `r/gov/dao/treasury/test/v1` | `treasury/test/v0`'s tests, ported. |

`impl/v1` differs from `impl/v0` in two more places:

- `GetInstance` (loader-only in v0) is replaced by `NewInstallRequest(cur,
  reason)`. It returns `impl/v0.NewUpgradeDaoImplRequest(cross(cur), govDAO,
  "gno.land/r/gov/dao/impl/v1", reason)`: v0's checks and disclosure text, with
  v1's own singleton as the new DAO. The `*GovDAO` is never handed to another
  realm. That matters because its `PostCreateProposal` and `ExecuteProposal`
  trust the proxy and take no caller check. A holder of the pointer could reset
  a proposal's votes or deny it.
- Its own `NewUpgradeDaoImplRequest` grants `[impl/v1, <new>]`, as v0's grants
  `[impl/v0, <new>]`.

**Migration path.**

1. A member creates `impl/v1.NewInstallRequest` as a proposal. The DAO votes
   on it under `impl/v0` and executes it. `dao.UpdateImpl` installs v1's
   singleton with `AllowedDAOs = [impl/v0, impl/v1]`.
2. The v0 treasury's funds are moved by an ordinary v0 payment proposal:
   `impl/v0.NewTreasuryPaymentRequest` with a v0 coins payment to
   `treasury/v1`'s address. It is voted on and executed under `impl/v1`. The
   executor is v0's, and `treasury/v0.Send` accepts it because `impl/v0` is
   still allowed. Any v0 GRC20 balances stay in `treasury/v0`, which keeps
   working for them.
3. `impl/v1.NewTreasuryGRC20TokensUpdate` lists the v1 tokens.
   `impl/v1.NewTreasuryPaymentRequest` pays from `treasury/v1`.
4. Once nothing is left at `treasury/v0`, a later `UpdateImpl` can drop
   `impl/v0` from `AllowedDAOs`.

Proposal statuses live on the implementation instance. A proposal still open
under `impl/v0` when the install executes cannot be voted on or executed
afterwards, as for any implementation change. So open proposals should be
closed before the install.

`govdao_treasury_v1_upgrade.txtar` runs steps 1 to 3 end to end. After the
funds move, it pays 2^64 of a grc20 v1 token and 1200 ugnot from
`treasury/v1`, and has an ugnot amount above MaxInt64 refused when the payment
is built. `govdao_proposal_treasury_tokens_update.txtar` still exercises
`treasury/v0` (with v0 tokens it defines itself, since the demo tokens are v1
now).

### Genesis scripts

**Mainnet's inputs are untouched.** `misc/deployments/mainnet.gno.land/
gen-genesis.sh` declares its package set final and checksum-locks its launch
artifacts, including the resolved package list (`work/packages.gen.txt`).
Editing it would break the rebuild, and cannot deploy anything on the running
chain anyway. Packages for a running chain go through its upgrade process, not
through genesis inputs.

The token migration changes mainnet's resolved list even with the script
untouched. `gno tool deplist -test-dep` on mainnet's unchanged set gives 95
packages; at the base commit it gave 89, and the base list hashes to the
script's locked `0f58…a960`. The six new ones are `p/nt/grc20/v1` and
`r/nt/grc20reg/v1` (imported by `r/gnoland/wugnot`), and `p/nt/treasury/v1`,
`r/gov/dao/treasury/v1`, `r/gov/dao/impl/v1` and `r/gov/dao/treasury/test/v1`
(matched by `./gno.land/r/gov/...`). Running the script on this branch stops at
step 4 with `checksum mismatch for work/packages.gen.txt`. Mainnet genesis
would not silently miss v1; it does not build at all from this branch. That
is inherent to moving wugnot, a mainnet-listed realm, to grc20 v1. Whether
mainnet takes it, and how, is the owner's decision.

The other networks' scripts (onyx, pearl, sapphire, topaz, test13) list
`p/nt/grc20/v1`, `r/nt/grc20reg/v1` and `p/nt/treasury/v1` next to the v0
entries, because `p/nt` paths cannot be added after genesis under namespace
enforcement. They resolve the same set without the explicit lines, so the
lines only document the intent. All of these networks have past genesis
times, and every one except test13 checksum-locks its package list, so their
rebuilds change for the same reason mainnet's does.

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
5. **Edit `treasury/v0` and `impl/v0` in place.** It would change the API of
   versioned realms under their own version. It also cannot happen on a live
   chain, where the code is immutable.
6. **Export the impl/v1 singleton (a `GetInstance` without the loader
   check), or install a fresh `NewGovDAO()`.** The first hands out a pointer
   whose methods trust the proxy. The second installs an instance the
   proposer could have kept a pointer to. `NewInstallRequest` avoids both.

## Consequences

- `p/nt/ufmt` formats `int256`/`uint256` (`%d`, `%v`, `%T`), and
  `p/nt/uassert` compares them (`Equal`, `NotEqual`, `Empty`). Both changes
  only add cases.
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
| `examples/gno.land/r/gov/dao/treasury/v1/treasury.gno` | the GovDAO uint256 treasury |
| `examples/gno.land/r/gov/dao/impl/v1/impl.gno` | `NewInstallRequest` |
| `gno.land/pkg/integration/testdata/grc20_v1_uint256.txtar` | arguments, cap and persistence across a restart |
| `gno.land/pkg/integration/testdata/govdao_treasury_v1_upgrade.txtar` | the install, the funds move and v1 payments |
