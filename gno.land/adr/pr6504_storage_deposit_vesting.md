# ADR: The storage deposit respects the vesting schedule

## Context

`VMKeeper.lockStorageDeposit` charged the deposit with
`bank.SendCoinsUnrestricted`, which skips the transfer restriction, the
session spend limit and the vesting schedule. The comment beside it gave one
reason, the transfer restriction; the session limit was then re-applied by hand
with `auth.CheckAndDeductSessionSpend`, and the schedule was not.

The refund does not go back to the payer. `processStorageDeposit` pays it to
the caller of the message that frees the storage, or to the storage fee
collector while `ugnot` is a restricted denom. So an account whose coins are
all locked could pay a deposit, and a second address with no schedule could
free the same bytes and receive about the same amount as spendable coins. The
schedule then locks nothing.

`tm2/adr/pr6095_vesting_as_an_account_field.md` recorded that fees and
storage deposits bypass the schedule, and asked for a deliberate decision on it
before mainnet. Fees are consumed and never come back. A deposit comes back, to
an address the caller chooses.

## Decision

The bank gains `SendCoinsVestingChecked`: a debit through `SubtractCoins`,
which enforces the schedule and nothing else, then `AddCoins`. It skips the
transfer restriction and the session limit exactly as the unrestricted send
does. `lockStorageDeposit` charges through it. The refund stays on
`SendCoinsUnrestricted`: it debits the realm's escrow address, which carries no
schedule. Gas stays on `SendCoinsUnrestricted` too.

A call, a run or a package deploy now fails when its deposit exceeds the
payer's unlocked balance, with the bank's `vesting locked coins error`. A
package enable fails the same way, its deposit being charged to the parked
package's creator rather than to the approver.

The check applies while `ugnot` is restricted too, when the refund goes to the
fee collector and nothing leaks. Keeping one rule for both periods means a
deposit never depends on a parameter governance can flip.

Genesis replay is exempt: while `auth.IsGenesisReplay` holds, the charge goes
through `SendCoinsUnrestricted` as before. A fork that replays a chain's history
through its genesis must reproduce what that chain accepted, and a deposit paid
from locked coins before this check existed would otherwise fail; under the
default `PanicOnFailingTxResultHandler` the fork would not boot. The redeploy
rule in `VMKeeper.AddPackage` waives itself on replay for the same reason. The
predicate also holds for a fresh chain's own genesis transactions, which the
operator writes and which run once. A replay never runs a transaction its
source chain refused once the backup marks it failed, and `gnogenesis fork` and
the standard format of `tx-archive backup` both mark it, so the waiver cannot
turn a refusal under this check into a success. A backup `tx-archive` took
before it wrote the flag carries none.

## Alternatives considered

**Call `SendCoins`.** It enforces the schedule, but it also applies the
transfer restriction, which would refuse the deposit of every account not on
the token-lock whitelist while `ugnot` is restricted, and it deducts the
session limit a second time.

**Check the schedule in the VM keeper.** It would copy the locked-amount
arithmetic out of `SubtractCoins`, and the two copies could drift.

**Refund the payer instead of the releaser.** It would also close the path,
but it changes who earns the refund, needs a record of payers per realm, and
migrates existing deposits. Out of scope for this fix.

## Consequences

- Transactions from a vesting account that paid a deposit from locked coins now
  fail. This changes DeliverTx results, so every node takes it at one height.
- Gas is unchanged for every accepted deposit, with or without a schedule: the
  extra account read `SubtractCoins` makes for a vesting payer is a cache hit.
  `addpkg_testfile_restart_gas.txtar` pins the exact gas of a deploy that pays
  a deposit, and passes unchanged.
- Every write that grows realm storage now comes out of unlocked coins, a
  governance proposal included. On `gnoland-1` the 3,262,417 continuous
  schedules started with 4% unlocked and unlock to 2028-09-12; the genesis
  GovDAO T1 member held about
  31.7 GNOT spendable at the last block read, and gains about 1.3 GNOT a day.
- Locked coins still pay gas. A schedule caps what leaves to another address;
  it does not reserve a balance.
