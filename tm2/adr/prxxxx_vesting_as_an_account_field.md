# PRXXXX: Vesting as a field on BaseAccount, not an account type

## Status

Proposed. Supersedes the guard added in the preceding commit, which reported the
account type this change deletes.

## Context

Vesting was three account types embedding a shared `BaseVestingAccount`, and the
bank decided whether to apply a lock by asserting a stored account to
`std.VestingAccount`. That assertion was the entire enforcement contract, and it
was silent when it failed.

It did fail. `BaseVestingAccount` carried a schedule but had no vesting maths, so
it did not implement the interface and its schedule locked nothing. That was
unreachable, but the same shape produced a reachable fault one layer up: gno.land
swapped its account prototype to `*GnoAccount`, while genesis kept building
`*ContinuousVestingAccount` from a plain `std.BaseAccount`. So a vesting account
was not a gno.land account. Two consequences, both verified:

- It does not implement `std.AccountUnrestricter`, so while ugnot is a restricted
  denom `canSendCoins` finds no whitelist bit and refuses every transfer. The
  vesting lock is never even consulted.
- `applyUnrestrictedAddrs` does an unchecked `acc.(*GnoAccount)`, so listing a
  vesting address in `UnrestrictedAddrs` panics the chain during InitChain with a
  bare interface-conversion message.

The natural configuration -- vest an account, and whitelist it so it can move
funds during the token lock -- was a boot panic.

Both faults have the same cause. Which type is stored decided whether a rule
applied, and two different layers each believed they owned that decision.

## Decision

One account type. `BaseAccount` gains a `VestingSchedule` field, the schedule
gains the vesting maths, and `std.Account` gains `GetVesting`, `SetVesting` and
`LockedCoins`. Enforcement reads the field.

Deleted: `VestingAccount`, `BaseVestingAccount`, `ContinuousVestingAccount`,
`DelayedVestingAccount`, both constructors, `SpendableCoins`,
`upgradeVestingAccount`, the `upgraded` flag threaded through `subtract`, and
`vestingScheduleOf`.

Adding a method to `std.Account` costs nothing: `SetAddress` is implemented once
in the whole tree, on `*BaseAccount`, so every account type already inherits its
implementation by embedding and no independent implementer exists.

`VestingSchedule.Type` now selects the curve at read time rather than selecting a
type at genesis, so it can no longer disagree with how an account actually vests.
`upgradeVestingAccount` is gone because there is nothing left to upgrade: an
elapsed schedule already locks nothing, so collapsing it saved nothing and cost a
write on a path that could still fail.

### Why this does not move the app hash

genproto2 emits a struct field only when it marshals to non-zero length, and a
zero `VestingSchedule` marshals to nothing. An account without vesting therefore
encodes byte-for-byte as it did before the field existed. `json:",omitempty"`
does the same for JSON, so query output is unchanged too -- and the two are
independent, so the binary encoding does not depend on the tag.

Measured, not assumed: the integration txtar suite, whose goldens contain account
query output, passes unchanged.

### Migration

Removing the three registered types means state holding one stops decoding. Only
genesis ever created them, and a fork re-parses the text balances file, so a fresh
or re-genesised chain is unaffected. This is safe because the chain is
pre-mainnet; on a live chain it would need a migration.

## Alternatives considered

**Embed `ContinuousVestingAccount` inside `GnoAccount`.** Every account would then
satisfy the vesting interface with an empty schedule, and `upgradeVestingAccount`
would see it as fully vested and collapse it to a plain `BaseAccount` on the first
debit -- silently dropping the attributes. It also hardcodes one curve.

**Put the field on `GnoAccount` instead of `BaseAccount`.** tm2's bank cannot read
a gno.land type, so it would need an interface -- reintroducing the assertion that
caused both faults.

**Keep the types and fix the genesis path.** Addresses the panic, leaves the
mechanism that produced it.

## Consequences

A realm's derived address can now hold a vesting schedule, because it is an
ordinary account like any other. That makes a vesting treasury expressible, which
it was not before.

Session accounts inherit the field. They hold no coins, so a schedule on one is
inert; nothing sets one.

Nothing outside genesis can write a schedule: `applyBalance` is the only caller of
`SetVesting`, and it validates first. `NewSessionAccount` builds from a prototype
and sets only address, pubkey, number and master.

`Validate` no longer requires a start time for a cliff, since nothing reads one.
The genesis text format still writes and requires three fields, so `Parse` and
`String` remain symmetric; the relaxation only widens what is accepted.

`subtract` treats a nil account as one holding no account-tier coins. That was
already true and is now load-bearing in one more place, so
`subtractCoinsUnrestricted` reads the account rather than passing nil. Getting
this wrong failed two existing tests, which is how it was caught.
