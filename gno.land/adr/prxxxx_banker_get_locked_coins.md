# ADR: Realms can read how much of a balance is vesting-locked

## Context

`banker.GetCoins` returns an account's total coins, vesting-locked coins
included. That is what the bank stores. A realm that displays the number is
showing a figure the holder may not be able to transfer. On a chain where
genesis accounts vest, that is the common case.

`std.Account.LockedCoins(blockTime)` already answers the question, and the
bank uses it in `SubtractCoins` and in the `bank/spendable` query. Neither
path is reachable from realm code. A schedule is written once, at genesis,
and no message changes it, so the data is public and still invisible to a
realm. Embedding the genesis table, or asking the user to paste a schedule,
is what realms do instead. [gnolang/gno#6235](https://github.com/gnolang/gno/issues/6235)
asks for the locked amount, not the schedule.

## Decision

`chain/banker` gains a read-only `GetLockedCoins(addr) chain.Coins`.

`SDKBanker` implements it as the call that already exists:

```go
acc := bnk.vmk.acck.GetAccount(bnk.ctx, addr)
if acc == nil {
    return nil
}
return acc.LockedCoins(bnk.ctx.BlockTime())
```

No new keeper method and no new state. The native binding
`bankerGetLockedCoins` has the same parameters and the same `(denoms,
amounts)` return as `bankerGetCoins`, and it is charged with that native's
gas row. The account read is metered by the store, as `GetCoins`' balance
reads are.

The value is the schedule's remainder, unclamped. Fees can debit coins the
schedule still calls locked, so a locked amount can exceed the balance.
`SubtractCoins` and the spendable query clamp the difference at zero; a
realm that wants a spendable figure does the same. Returning the unclamped
remainder is what keeps the realm and the transfer check on one number.

An address with no account, and an account with no schedule, both return
nothing. Every banker type may call it. It writes nothing.

`gno test`'s banker has no account store, so `GetLockedCoins` is empty
there. A genesis vesting account is covered by the txtar, and the curve
over block time by `TestGetLockedCoinsTracksTheSchedule`.

## Alternatives considered

**Return the schedule** (`original`, `start`, `end`, `kind`). A realm could
then project a future date. It would also pin `VestingSchedule`'s shape into
the VM's public surface, so a later curve would be a breaking change. The
locked amount is the figure a transfer enforces. The issue asks for that
and not the schedule.

**Return the clamped spendable set.** That is what `bank/spendable` serves
off chain. It mixes the balance and the schedule into one number, and a
realm that also calls `GetCoins` could not recover the lock. The lock is
the missing input; the clamp is a `max(0, balance-locked)` the realm can do.

**A `GetLockedCoin(addr, denom) int64`.** It would match `GetCoin`'s cost
story for one denom. The issue's sketch returns `Coins`, the same type as
`GetCoins`, and a schedule's denom count is fixed at genesis rather than
grown by a third party. A single-denom reader can wait.

## Consequences

- Realm code can tell a locked balance from a spendable one without a copy
  of the genesis table. `r/moul/vesting/v0` is not in this tree; it keeps
  taking the schedule from its caller until someone updates that realm.
- Native gas for the new binding equals `bankerGetCoins`. Store gas is one
  account read. `native_bench_output.txt` holds no samples for it, so
  re-running the fitter against that checked-in snapshot drops the row;
  `BenchmarkNative_Banker_GetLockedCoins_*` and its `NATIVE_SPECS` entry in
  `gen_native_table.py` are wired so the next reference-hardware calibration
  run fits it directly.
- Adding a stdlib method and native binding is a consensus-breaking change:
  on a running chain it ships in a coordinated upgrade rather than a rolling
  one. `banker.gno` is committed into genesis as stdlib source, so the pinned
  multistore hash in `TestAppHashCrossrealm38` moves. The crossrealm38
  scenario does not read a lock. Gas rows are not committed state.
- Widening `chain/banker.Banker`, `chain/banker.BankerInterface`, and
  `execctx.BankerInterface` breaks out-of-tree mocks and chain embedders that
  implement them directly; all in-tree implementors (`SDKBanker`,
  `RunSDKBanker` by embedding, `TestBanker`, the calibration mock, and test
  mocks) are updated here.
- `gno test` does not model vesting. A filetest that calls `GetLockedCoins`
  sees an empty set.
