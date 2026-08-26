# PRXXXX: Vesting schedules — report the unenforceable account type, warn on implausible genesis times

## Status

Proposed

## Context

An audit of the vesting code found the arithmetic correct and the enforcement
correctly placed, but three things that read as safe and are not. None is
exploitable on the current code. All three are the kind that stay quiet until
someone relies on them.

### The enforcement contract is a type assertion

`BankKeeper.SubtractCoins` decides whether to apply a lock by asserting the
stored account to `std.VestingAccount`. That is the whole contract: an account
that does not implement the interface has no lock, whatever fields it carries.

`BaseVestingAccount` carries a `VestingSchedule` but has no vesting maths of its
own, so it does not implement the interface. A bare one is therefore unlocked.
Verified end to end before the change: an account holding a schedule with
nothing vested sent its entire balance.

Two things advertised it as a supported stored type. It is amino-registered in
`tm2/pkg/std/package.go`, so it can be decoded from state; and the bank's
account-tier invariant validated its schedule, which means a bare one with a
well-formed schedule was reported healthy while locking nothing.

It is not reachable today. The only construction sites are inside the two real
constructors, always embedded; `auth.GenesisState` carries no account list; and
every `SetAccount` caller either writes back an account it read or builds one of
the two concrete types.

### A schedule's times are never compared against the chain's clock

`VestingSchedule.Validate` checks that the end time is after the start time and
nothing else. A schedule that already ended when the chain starts is valid and
vests everything immediately. One written in milliseconds instead of seconds
lands tens of thousands of years out and never vests. Both are silent, and both
look ordinary in a genesis file.

Rejecting either would be wrong: replaying an old genesis onto a fork is
legitimate, and its schedules are legitimately in the past.

### Two fields do not mean what they appear to

`DelayedVestingAccount.GetStartTime()` returns a hardcoded zero while the
embedded schedule keeps whatever genesis configured, so the two disagree
whenever a start time was supplied — which genesis requires, and `Validate`
enforces, even though delayed vesting ignores it.

`VestingSchedule.Type` selects which account type genesis builds and is never
read again; the stored Go type is what decides how coins vest.

## Decision

### Report the unenforceable type instead of validating it

The account-tier invariant now reports a bare `BaseVestingAccount` as broken,
and `vestingScheduleOf` no longer returns a schedule for it. Reporting is the
right response because the fault is the account type: the schedule cannot be
made to matter by being well formed.

`BaseVestingAccount` is left without the three methods on purpose. There is no
correct answer for a type that does not know whether it vests linearly or at a
cliff, so supplying one would turn a reported fault into a silent guess. Two
compile-time assertions record which types the contract covers, and a test pins
that this one does not.

Amino registration is left alone. It is what makes a bare one decodable, but
removing it would churn generated code for a type that is never encoded on its
own, and the invariant now catches the case that registration allows.

### Warn about implausible genesis times, do not reject them

`gnogenesis verify` now warns when a vesting schedule ends at or before the
genesis time, or more than a hundred years after it. That bound is well past any
real token schedule and well short of what a milliseconds-for-seconds mistake
produces.

The operator tool is the right home. It has the genesis time, it runs before the
chain does, and a warning there leaves fork replay working. The horizon is
computed by adding to the genesis time rather than subtracting from the end time,
because the end time is an operator-supplied `int64` that can sit anywhere in the
range while the genesis time is a real date.

### Say what the rest means

Comments now record that the lock blocks transfers but does not reserve a
balance — fees and storage deposits debit through the unrestricted path and never
consult a schedule — that delayed vesting ignores its start time, and that `Type`
is not read after genesis.

## Alternatives considered

**Give `BaseVestingAccount` the three methods.** Makes it enforceable, and picks
a vesting curve for a type with no basis to pick one. A wrong lock is worse than
a reported fault.

**Drop it from amino registration.** Removes the decode path, churns generated
code, and moves the codec's type table for a type nothing encodes.

**Reject implausible times in `Validate`.** `Validate` has no notion of now, and
an error would break the legitimate fork-replay case.

**Warn at boot instead.** The InitChainer would have to be careful to keep the
warning off consensus state, and it fires when it is already too late to edit the
file.

**Change fees to respect the lock.** A consensus change well outside an audit
follow-up. Documented rather than changed; worth deciding deliberately before
mainnet, since a holder reads "vested" as "reserved".

## Consequences

A bare `BaseVestingAccount` in state now breaks an invariant instead of passing
as healthy. Nothing produces one, so no existing chain state is affected.

Making that type enforceable later fails two tests, which is the point: it should
be a deliberate change, and the invariant would need updating alongside it.

`gnogenesis verify` gains output on genesis files with unusual vesting times. It
still exits zero; only the text changes.

The `VestingAccount` interface declares six methods and production uses two,
`LockedCoins` and `GetVestingCoins`. The four unused getters are what force
`GetStartTime()` to return a value that disagrees with the stored field. Left
alone here to keep this change to the audit findings, but narrowing the interface
would remove the need for that method to exist at all.
