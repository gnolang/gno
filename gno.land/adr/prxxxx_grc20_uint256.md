# ADR: GRC20 v1 with native `uint256` amounts

## Status

Proposed. Builds on `gnovm/adr/pr6387_native_int256.md` (the native
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
  reason, law)`. It returns `impl/v0.NewUpgradeDaoImplRequest(cross(cur), d,
  "gno.land/r/gov/dao/impl/v1", text)`, so the request carries v0's checks and
  disclosure text. `d` is an `installedDAO`: v1's `*GovDAO` plus the law the
  proposal carries (see below). The `*GovDAO` is never handed to another realm
  except inside that proposal. That matters because its `PostCreateProposal`
  and `ExecuteProposal` trust the proxy and take no caller check. A holder of
  the pointer could reset a proposal's votes or deny it.
- Its own `NewUpgradeDaoImplRequest` grants `[impl/v1, <new>]`, as v0's grants
  `[impl/v0, <new>]`.
- **Proposal filters fail closed.** impl/v0 stays in AllowedDAOs after the
  install, and its builders (`NewAddMemberRequest`,
  `NewPromoteMemberRequest`) tag proposals with impl/v0's `FilterByTier`, a
  distinct type from impl/v1's. A type switch that knew only v1's type read a
  v0 T1 proposal as unfiltered, so T2 and T3 votes could carry a T1
  membership change. `voterTiers` enforces both types with v0's rules (T1 →
  T1 only; T2 → T1 and T2; T3 → every tier). It refuses any other non-nil
  filter, and any filter tier that is not T1, T2 or T3, with
  `ErrUnsupportedFilter` in `PreCreateProposal`, so such a proposal never
  exists. `PostCreateProposal` panics on the same condition rather than
  default to "everyone votes". That filter is the only place impl/v1 type
  switches on a value another realm may build. Executors go through the
  proxy's `Executor` interface.
- **The install proposal carries the law.** impl/v0 keeps its law in an
  unexported variable with no accessor, and v0 is not changed, so impl/v1
  cannot read it. Instead `NewInstallRequest` takes the law impl/v1 is to
  use:
  - It refuses a Supermajority outside [66.66, 100]. 66.66 is impl/v1's
    default, so an install can never weaken the DAO. A DAO that wants a weaker
    law lowers it with a v1 law-change vote after the install, in the open.
  - The proposal text states the law ("Law after the install: Supermajority
    X%"). The DAO votes on it under impl/v0's own law.
  - `installedDAO.PreExecuteProposal` applies the carried law on its first
    proxy-validated call. That is the only place a law is read, so impl/v1's
    first decision already uses it. There is no window at the default law.
  - A later `impl/v1.NewChangeLawRequest` stands until the implementation
    changes again. The proxy keeps a request's executor, so an install
    request can be submitted twice. Therefore every successful
    implementation change executed under an `installedDAO` clears the
    applied marker (re-application, rather than refusing resubmission). An
    implementation change is an executor built in impl/v0 or impl/v1 whose
    description starts "Replaces the govDAO implementation". The next
    decision then applies the carried law of whichever `installedDAO` is
    current, so a resubmitted request re-applies its law.
    `govdao_impl_v1_install_reuse.txtar` installs at 80%, lowers the law to
    66.66%, executes the kept request again, and shows two of three votes
    refused at 80%. Without the reset that last proposal passes.
  - The proposer's `reason` is escaped with `sanitize.InlineText` before the
    law line is appended. impl/v0 renders proposal descriptions raw, and Codex
    showed that a reason ending in `\n\n<!--` opened an HTML comment that hid
    everything after it, the law line included.
  - impl/v1's own renderers (`renderProposalPage`, `StringifyProposal`)
    pass every proposal description and executor description through
    `sanitize.BlockRich`, and titles through `InlineText`. Any realm can
    build a `ProposalRequest`, and impl/v0's builders stay usable under v1,
    so per-builder escaping is not enough. BlockRich keeps the markdown but
    escapes HTML-block openers, so a description cannot hide the metadata,
    executor realm and vote stats after it. v1's `NewUpgradeDaoImplRequest`
    and `NewWithdrawMemberRequest` also escape their reasons. While impl/v0
    is still active, its renderer stays raw (v0 is unchanged), which is why
    `NewInstallRequest` escapes its own reason. The gnoweb golden
    `blockrich-proposal-description-cannot-hide-grant.txtar` renders a full
    upgrade proposal page, and `govdao_impl_v1_install_reuse.txtar` checks a
    hostile v1 upgrade proposal on chain.
  - Tests: `govdao_impl_v1_install_law.txtar` covers v0 raised to 80%, a
    weaker install refused, and an install carrying 80% with that hostile
    reason, whose law line is still visible and whose reason is escaped. v1's
    first decision (two of three) is then refused at 80%, and a later v1 law
    change stands. The gnoweb sanitize golden
    `inline-html-comment-cannot-hide-trailing-notice.txtar` renders the same
    construction through the production goldmark chain and shows the law line
    in the HTML.

**Migration path.**

1. A member creates `impl/v1.NewInstallRequest(cur, reason, law)` as a
   proposal, with `law` set to the DAO's current law. The DAO votes on it under
   `impl/v0` and executes it. `dao.UpdateImpl` installs v1's GovDAO, carrying
   that law, with `AllowedDAOs = [impl/v0, impl/v1]`.
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

The install proposal in step 1 should carry the DAO's current law. Voters
check it against what impl/v0 enforces, since it cannot be read back.

Proposal statuses live on the implementation instance. A proposal still open
under `impl/v0` when the install executes cannot be voted on or executed
afterwards, as for any implementation change. So open proposals should be
closed before the install.

`govdao_impl_v1_tier_filters.txtar` checks the filters after the install: a
v0-built T1 membership proposal refuses T2 and T3 votes, cannot be executed
until T1 votes, and a proposal with an unknown filter is refused at creation.
`govdao_treasury_v1_upgrade.txtar` runs steps 1 to 3 end to end. After the
funds move, it pays 2^64 of a grc20 v1 token and 1200 ugnot from
`treasury/v1`, and has an ugnot amount above MaxInt64 refused when the payment
is built. `govdao_proposal_treasury_tokens_update.txtar` still exercises
`treasury/v0` (with v0 tokens it defines itself, since the demo tokens are v1
now).

### Rollout: new chains only

This change targets a future fresh chain. That chain's genesis config will
list `p/nt/grc20/v1`, `r/nt/grc20reg/v1` and `p/nt/treasury/v1` explicitly,
because `p/nt` paths cannot be added after genesis under namespace
enforcement. Existing networks are not upgraded in place.

The existing networks' scripts under `misc/deployments` are untouched. Each
network is built from its own release branch or tag: `chain/test13`,
`chain/topaz`, `chain/sapphire`, `chain/pearl` (their `VALIDATOR.md`), the
`chain/onyx` tag on `chain/mainnet`, and mainnet's release tag (its
`VALIDATOR.md` says never to build from `master`). Each is checked against a
pinned manifest. Onyx, pearl, sapphire, topaz and mainnet checksum-lock the
resolved package list (`work/packages.gen.txt`) and other build artifacts.

Moving wugnot to grc20 v1 and adding the `r/gov/dao/*/v1` realms changes what
master's `examples/` resolve to. `gno tool deplist -test-dep` on mainnet's
package set gives 95 packages on this branch against 89 at the base commit,
where the list still hashes to mainnet's locked `0f58…a960` (onyx locks the
same hash). The six new packages are:

- `p/nt/grc20/v1` and `r/nt/grc20reg/v1`, via `r/gnoland/wugnot`;
- `p/nt/treasury/v1`, `r/gov/dao/treasury/v1`, `r/gov/dao/impl/v1` and
  `r/gov/dao/treasury/test/v1`, via `./gno.land/r/gov/...`.

Running mainnet's `gen-genesis.sh` from this branch therefore stops at step 4
with `checksum mismatch for work/packages.gen.txt`. That is expected for a
launched network's script run from master, and it is the reason those
networks are built from their release branches: onyx's README says as much,
noting that "`master`'s already differs from what the network runs".

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
6. **Upgrade the existing networks in place,** or list the v1 packages in
   their genesis scripts. Their genesis is fixed and pinned, and the owner
   chose a fresh chain for this change.
7. **Export the impl/v1 singleton (a `GetInstance` without the loader
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
