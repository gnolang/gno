# A query that evaluates a vesting schedule, instead of returning it

## Context

An account with a vesting schedule has three interesting numbers — what it
holds, what is still locked, and what it may transfer today — and the chain
serves none of them together.

| query | returns | what it omits |
|---|---|---|
| `auth/accounts/{addr}` | the account, schedule included | what the schedule evaluates to |
| `bank/balances/{addr}` | `GetCoins(addr)` | that any of it is locked |

`queryBalance` is one line — `amino.MarshalJSONIndent(bh.bank.GetCoins(ctx, addr), …)`
— so it reports a vesting account's full balance with no hint that most of it
cannot move. A caller who wants the spendable figure must fetch the account,
fetch the current block time, and evaluate `std.VestingSchedule` themselves.

That last step is the problem. The curve is not a one-liner to restate:

- Two variants. `VestingContinuous` is linear between `StartTime` and
  `EndTime`; `VestingDelayed` vests nothing until `EndTime` and everything
  after. `VestedCoins` short-circuits differently for each.
- `VestingContinuous` is the **empty string**, and the field is `omitempty`, so
  a continuous schedule serialises with no `type` key at all. A reimplementation
  that keys off a present-and-equal-to-something check gets every continuous
  account wrong.
- The linear branch multiplies in `big.Int` because `amount * elapsed`
  overflows int64 for large grants — precisely the grants where being wrong
  matters most.
- The result must be clamped. A schedule can lock more than the account still
  holds, once an unrestricted transfer has spent into the locked portion;
  `BankKeeper.SubtractCoins` clamps at zero for exactly this, and a naive
  subtraction reports a negative.

So every off-chain reimplementation — wallet, indexer, explorer — is a place
where the number drifts from what the keeper actually enforces. The failure is
quiet: the caller sizes a transfer by its own arithmetic and the chain rejects
it, or worse, the caller under-reports and the user believes funds are locked
that are not.

## Decision

Add `bank/spendable/{addr}`, alongside `balances` and `supply`, returning:

```json
{
  "coins":      "10000000000ugnot",
  "locked":     "9598991185ugnot",
  "spendable":  "401008815ugnot",
  "block_time": "1789228670"
}
```

`block_time` is a **string**: amino renders an int64 as a quoted JSON number,
the same trap `querySupply` already documents. A client author writing a parser
from this block would otherwise get a type error on the one field that is not a
Coins string.

It calls `acc.LockedCoins(ctx.BlockTime())` — the same function
`SubtractCoins` calls — over `bh.bank.GetCoins(ctx, addr)`, the same set
`queryBalance` returns, and clamps per denom the way `SubtractCoins` does. One
implementation of the curve, and the endpoint is a caller of it.

`block_time` is echoed because the answer is a function of it. A continuous
schedule releases coins every second, so the same account queried a moment later
reports a different `spendable`; without the timestamp a caller cannot tell a
stale answer from a current one, nor reproduce the arithmetic.

### Why `bank` and not `auth`

The schedule is a field on the account, so `auth` looks like the natural home,
and the first version of this change put it there as a sub-path of
`queryAccount`. That was wrong, for a reason worth recording because it is not
obvious from where the data lives:

**The rule being mirrored is deliberately tier-agnostic.** `SubtractCoins` reads
each denom through `GetCoin`, which routes to whichever tier holds it, precisely
so a schedule over a `/`-prefixed genesis denom is enforced — and `gno.land`
accepts such schedules. But `auth` can only see the account object, whose coins
field holds the gas denom alone. An `auth`-side endpoint therefore reports a
split-tier denom in `locked` and silently omits it from `spendable`, telling a
caller nothing can move while a transfer would be allowed. That is the same
under-reporting the endpoint exists to prevent, reappearing for the denom class
a client is least likely to hard-code.

`bank` has both halves already: it imports `auth` (`keeper.go:10`), `BankKeeper`
holds an `AccountKeeper` (`keeper.go:45`), and `bankHandler` holds a
`BankKeeper`. The move cost no wiring — no `NewHandler` signature change, no new
keeper, no new dependency.

### Why the chain rather than the indexer

The obvious alternative is to put this in the indexer and avoid shipping node
binaries. Two things argue against it.

It is not a consensus change, so the cost being avoided is smaller than it
looks. `BaseApp.Query` is read-only; its result never enters a block, reaching
neither the AppHash nor `LastResultsHash`. Nodes on different binaries can serve
different query paths without diverging. This is a rolling upgrade, not a fork.

And the indexer would have to reimplement the curve, which is the whole problem
above. It would also need current account state and block time, neither of which
it holds — so it would end up calling `auth/accounts` and doing the arithmetic
anyway, buying an HTTP shape rather than independence.

### Why not more fields on the account's `vesting` object

The shape a reader reaches for first is `auth/accounts/{addr}` returning the
locked amount inside `vesting`, next to the schedule that produced it. It is
ruled out twice over.

`VestingSchedule` is **persisted**: `AccountKeeper.SetAccount` writes
`amino.MarshalAny(acc)` into the store, and the type has generated binary
marshalling (`std/pb3_gen.go`). A new field changes the bytes written for every
vesting account, so the IAVL hash changes and nodes diverge — a state-breaking
consensus change, against a query that needs no fork at all. It would also mean
regenerating `pb3_gen.go`.

And the values are not storable even in principle: `locked` is a function of
block time, so a stored copy is stale the moment it is written. The enrichment
has to happen at query time regardless, which is what this endpoint is.

## Consequences

**`GetCoins` walks the split-tier prefix for a caller-supplied address**, so its
cost rises with the number of denoms that address holds — which a third party can
grow by sending it new ones, the hazard AGENTS.md calls out. This is the same
exposure `queryBalance` already accepts on the same unmetered query path, and it
is the price of reporting a set that agrees with what transfers enforce. An
`auth`-side version would have been O(1) with an attacker-independent cost; that
was the one real argument for the rejected placement, and it loses to
correctness.

An address that has never transacted has no account. The endpoint answers with
zeros rather than an error: it is the honest answer for an address that holds
nothing, and it spares every caller a special case.

Nodes must upgrade to serve the path; one that has not returns
`unknown bank query endpoint`. Clients should treat that as "not available here"
rather than as an error about the account.

**"Spendable" means transferable, and nothing more.** Locked coins can still
leave the account as gas fees and storage deposits — `std/vesting.go` is explicit
that "not spendable" means not transferable. A caller must not present this
number as "what you can spend on fees". Nor does it account for transfer
restrictions unrelated to vesting, which are enforced elsewhere; on a chain with
restricted transfers an account can have a positive `spendable` and still be
unable to send.

The clamp is load-bearing for the wire format, not merely for presentation:
`Coins.MarshalAmino` is `Coins.String()` and **never errors**, so an unclamped
negative does not fail on the server. It marshals cleanly as `"-990ugnot"`, is
served, and fails in the client's decoder.

### Tests

Eight, added to `handler_test.go` beside the existing query-handler tests, each
on its own store because the keeper's store is not safe for concurrent use:

- mid-schedule, the only point where a wrong curve is visibly wrong — at either
  end every implementation agrees
- **a schedule over a split-tier denom**, which is the case that decides the
  module this belongs in
- the endpoint agrees with `std.VestingSchedule.LockedCoins` across eight
  offsets spanning before, during and after the window, so the two are pinned to
  the same function rather than to each other's arithmetic
- a delayed schedule is a cliff, with a guard asserting that a continuous
  schedule would disagree **at the instant the subtests assert** — compared
  anywhere else the guard would not protect them
- the clamp, where the schedule locks more than the balance
- an account with no schedule, which is almost every account
- an address with no account at all
- a malformed address is an error, not an empty answer

Checked by ablation:

- reading the account tier instead of `bank.GetCoins` — i.e. reverting to the
  `auth` placement — fails **only** the split-tier test, which is what makes that
  test the evidence for the module choice
- removing the clamp fails only the clamp test
- reading a fixed time instead of `ctx.BlockTime()` fails the mid-schedule, cliff
  and agreement tests, leaving the no-schedule and unknown-account cases green —
  neither reaches the `LockedCoins` call
- asserting the continuous-curve value at `end-1` fails the cliff subtest

## Alternatives considered

**Extend `auth/accounts/{addr}` with a `spendable` sub-path.** Built first, then
moved; see "Why `bank` and not `auth`".

**Return the figures from `auth/accounts` as extra fields.** Breaks every
existing parser of that response, and the concrete account type is gno.land's
`GnoAccount`, so tm2's auth handler would have to mirror a struct it does not own.

**Do it client-side in gnokey.** Cheapest, and worth doing anyway for the CLI's
own output, but it solves it for one client. Wallets, explorers and the indexer
would each reimplement the curve, which is the drift this is meant to close.

## Follow-ups

- `BankKeeper.SpendableCoins`, called by both `SubtractCoins` and this query, so
  the clamp has one implementation rather than two spellings of the same
  arithmetic.
- A `gnoclient` method and a txtar, so the wire path is exercised end to end;
  nothing in the tree calls this yet.
