# ADR: Cliff-and-vest grant vault

## Status

Proposed, with a v0 implementation under `examples/`. Unaudited.
Not wired to genesis, GovDAO, or account-level vesting.

## Context

gno.land already has three things people reach for when they say
"vesting", and none of them is a grant contract another person can fund.

**Account send-limits.** `tm2/pkg/std/vesting.go` locks coins on a
`BaseAccount`. The default schedule is linear from `StartTime` to
`EndTime`. The delayed schedule holds everything until `EndTime`.
The bank keeper enforces the lock on sends. `SetVesting` runs from
gnoland genesis tooling, not from a realm. A realm cannot read how much
of an account is locked ([gnolang/gno#6235](https://github.com/gnolang/gno/issues/6235));
the workaround people use is a view realm such as `r/moul/vesting/v0`.
Launch allocations were reconciled in
[gnolang/gno#5798](https://github.com/gnolang/gno/issues/5798) (closed):
4% at transferability, then 4% a month for 24 months, held by tm2
multisig accounts. That schedule is a genesis parameter. Re-implementing
it here would fork the lock the bank already enforces.

**Operation timelock.** `examples/quarantined/gno.land/p/thox/timelock`
queues a call and releases it after a delay. It is an OpenZeppelin-style
`TimelockController`, not a token curve. [gnolang/gno#2086](https://github.com/gnolang/gno/issues/2086)
asked about a timelock package in the same sense and did not land one.
The multisig factory next to this vault has its own proposal delay. A
vesting vault should not grow a second scheduler.

**Payment rails.** `p/nt/treasury/v0` routes a payment through a
`CoinsBanker` or a `GRC20Banker`. `r/gov/dao/treasury/v0` executes a
voted payment immediately. [gnolang/gno#503](https://github.com/gnolang/gno/issues/503)
mentioned milestone funding and was not built as a contract. None of
these hold a balance on a curve.

**Grants wishlist.** [gnolang/grants#7](https://github.com/gnolang/grants/issues/7)
is about how a grant program is run, not a vault. No issue in that
repository asks for a cliff contract. This vault is the piece a program
would call, not the program itself.

The factory follows the same shape as `p/nt/multisig/v0` and
`r/nt/multisig/v0`: namespace `nt`, version `v0`, `chain.Emit` for
state changes, `sanitize.InlineText` on every user string in `Render`,
and one sub-realm address per object (`gno.land/adr` notes on
`DerivePkgSubAddr` in the treasury sub-address work). A multisig safe
can be a grant admin without this package importing it. The admin is
an address. `Revoke` and `TransferAdmin` succeed when
`cur.Previous().Address()` is that address, so a safe that cross-calls
while it is current presents the safe's address. `r/nt/multisig/v0`
does not invoke arbitrary realms on execute, so the wiring is a later
executor. v0 does not take that dependency.

## Decision

Two packages:

- `gno.land/p/nt/vesting/v0` is the state machine. No realm value, no
  untrusted call. `VestedAmount` is a pure function of the terms and a
  unix second.
- `gno.land/r/nt/vesting/v0` is a permissionless factory. `CreateCoin`
  and `CreateGRC20` append a grant only after the funds have moved to
  `cur.Sub("g/<id>")` and the balance delta equals `total`.

One asset per grant. Native denoms are any bank-valid denom; ugnot is
the one the README shows. GRC20 is a registry key. The allowance spender
is the factory address, because `RealmTeller` binds the spender to the
realm that asks.

Curves:

- Linear (`1`): `floor(total * elapsed / duration)` after the cliff,
  and `total` at and after the end.
- Step (`2`): `floor(total * completedPeriods / periodCount)`, with
  `period` dividing `duration`.

`cliff == 0` vests from `start` (zero at the exact start second).
`cliff == duration` is a full unlock at the end. Before the cliff,
vested is zero; at the cliff, the linear or step amount accrued since
`start` unlocks at once.

`floor` favors the vault. The end shortcut returns `total`, so dust is
not stuck on a finished grant. `Revoke` sets `returned = total - vested`,
so `claimed + releasable + returned == total` after every successful
mutation. Multiplication uses `math/bits` 128-bit `Mul64`/`Div64` and
rejects a quotient that does not fit in int64.

Revoke is allowed once, and only on a revocable grant, and only by the
admin. Vested tokens stay claimable. Unvested tokens go back to the
admin. There is no pause.

Identity:

- Claims, revoke, beneficiary change, and admin transfer use
  `cur.Previous().Address()`.
- `CreateCoin` also requires `cur.Previous().IsUserCall()` and an
  `OriginSend` envelope of exactly one coin equal to denom and total.
  `IsUser()` would accept `maketx run`, which can consume the envelope
  before the function reads it. A code realm is not the entry realm
  that was paid, and the origin-send banker refuses to spend from the
  factory in that case.

Events, emitted by the package when the books change: `GrantCreated`,
`GrantClaimed`, `GrantRevoked`, `BeneficiaryChanged`, `AdminTransferred`.
The realm does not emit a second copy.

`Render("")` lists at most the latest 20 grants. `Render("<id>")` is one
grant. Paths longer than 64 bytes, or with an extra segment, render as
unknown. Labels reject control characters at creation, and `Render`
still runs `sanitize.InlineText` on every user string.

## Rejected designs

- **Extending `std.VestingAccount`.** Realms cannot set it, and genesis
  send-limits already have an operator. A third-party grant is a
  different balance.
- **Reusing `p/thox/timelock`.** That queue releases a call, not units
  of a token. Vesting math does not belong in a scheduler.
- **Holding every grant at the factory address.** One balance cannot
  prove that grant A did not spend grant B. Sub-addresses make the bank
  the ledger and make isolation a transfer rule.
- **Books updated after the send.** A recovered panic in a test, or a
  realm caller that revives, would leave the books ahead of the coins
  or the coins ahead of the books. `Claim` and `Revoke` check the
  balance, then update the books, then pay. The insufficient-balance
  path does not touch the books.
- **`RecoverExcess` / sweep.** A direct donation to a grant address is
  not in the books. Letting the admin withdraw "extra" plus a revived
  panic would let the admin take another grant's coins if the books
  were ever wrong. Extra coins stay stuck. That is the safer failure.
- **Pause.** An admin who can pause claims can freeze tokens that have
  already vested. Revoke of the unvested part is the whole of the
  admin's power over the funds.
- **`IsUser()` for coin funding.** It treats `gno.land/e/<addr>/run` as
  the user. The run script can spend `-send` first. `IsUserCall`
  (`pkgPath == ""`) is the check, paired with the origin-send banker.
- **Storing a `realm` in the grant.** The VM refuses a stored realm in
  the cases that matter, and a stored realm would not be the caller
  anyway. The admin and the beneficiary are addresses.
- **Two-step fund.** Create, then send. A create that does not pull the
  funds leaves an empty grant people will claim against. The create
  message moves the funds, or it does not append the grant.
- **Reading `GetCoins`.** Anyone can send an address a new denom.
  `GetCoins` walks every denom. `GetCoin` is the read.

## Threat model

What this vault is for: a person or a realm locks a known coin or a
known registered token for a known beneficiary, on a published curve.
The beneficiary can take the vested part. A named admin can take the
unvested part back, once, if the grant says so.

What it is not: a replacement for genesis locks, a mixer, or a safe
against a token contract the user chose that lies about balances.

Attackers considered: another user, a `maketx run` script, a published
realm, a registered GRC20 whose code we do not control, a clock that
steps backwards within what consensus will sign, and a caller who wants
the render page to inject markdown.

### Findings

| ID | Issue | Severity | Disposition |
| --- | --- | --- | --- |
| F1 | `IsUser()` would let a `maketx run` spend the coin envelope before `CreateCoin` observes it | High | Fixed. `CreateCoin` requires `IsUserCall`. `filetests/z_run_filetest.gno` shows a run cannot fund and can still claim, because the run address is the user. `filetests/z_coin_filetest.gno` shows a foreign realm cannot fund, including a user-shaped caller whose envelope was paid to the other package. |
| F2 | Authorizing with `OriginCaller` would let any realm in the call claim for the signer | High | Fixed. Every mutation uses `cur.Previous().Address()`. `filetests/z_auth_filetest.gno` sets `OriginCaller` to the beneficiary while the caller realm is the filetest, and the claim fails. |
| F3 | `total * elapsed / duration` overflows int64 near `MaxInt64` | Medium | Fixed. 128-bit multiply and an explicit range check. `TestMulDivFloor`, `TestVestedAmountHuge`. |
| F4 | Rounding up would pay the beneficiary more than the curve | Low | Floor. Dust is paid at the end, not before. `TestVestedAmountBoundaries`, `TestPropertyInvariants`. |
| F5 | Updating books before the balance check lets a failed claim stick | Medium | Balance is checked first. `TestDrainedGrantDoesNotChangeBooks` empties the sub-address and checks that claim and revoke leave `claimed` and `revoked` unchanged. |
| F6 | `GetCoins` on an attacker-chosen address is unbounded gas | Medium | Reads go through `GetCoin` or `BalanceOf` on one token key. |
| F7 | `Render` writes the label and the path into markdown | Medium | Control characters rejected at `New`. `Render` caps the path at 64 bytes, refuses extra segments, and runs `sanitize.InlineText` on user strings. Home lists at most 20 grants. `TestRender`, `filetests/z_render_filetest.gno`. |
| F8 | One shared balance lets grant A pay grant B's beneficiary | High | Each grant is `cur.Sub("g/<id>")`. `pay` refuses to send if that sub-address does not match the address stored at `Bind`. `TestCoinClaimRevokeAndIsolation`, `filetests/z_grc20_filetest.gno`. |
| F9 | A GRC20 `Transfer` that calls back into `Claim` could double-spend if the books moved first | Low | `busy` rejects re-entry. Canonical `grc20.Transfer` does not call out; the lock is defense in depth. `TestReentrancyGuard` sets `busy` from the same package. `filetests/z_reenter_filetest.gno` shows an outside realm's callback-shaped call does not move the books. A non-canonical teller is rejected with `IsCanonicalTeller`. |
| F10 | A backwards block time could un-claim | Medium | `vested` never returns less than `claimed`. After revoke, the frozen amount is what later claims use. `TestClockBackwardsDoesNotUnclaim`. |
| F11 | Admin pause could freeze vested tokens | Medium | Not implemented. Admin power is revoke of the unvested part, once. |
| F12 | Sub-address derived differently at spend time than at create | High | `Bind` stores `DerivePkgSubAddr`. `pay` recomputes `cur.Sub` and compares. A mismatch panics before the send. |
| F13 | Revoke page said the admin could revoke again | Low | Fixed. A revoked grant renders "already revoked". `TestRender`. |

Accepted, not fixed:

| ID | Limitation | Severity | Why it stays |
| --- | --- | --- | --- |
| A1 | Time is `time.Now().Unix()`, the block time validators sign. There is no drift buffer. | Info | A buffer would change who gets paid. Consensus time is the clock. |
| A2 | A realm caller that revives a panic after a GRC20 transfer can desync books from balances. An EOA `MsgCall` cannot revive. | Low | Coins that disagree with the books stay in the grant. There is no sweep, so the admin cannot collect them. The pre-check means the insufficient-balance path does not desync. |
| A3 | Coins sent straight to a grant address are stuck. | Info | See rejected `RecoverExcess`. |
| A4 | Grant count is unbounded. Each create is paid for by its caller. `Render` does not walk more than 20. | Info | A global cap would be a new admin. |
| A5 | A registered token can report balances that are not what `Transfer` moved. The canonical teller check only recognizes the standard teller. | Medium | The user picks the token key. The vault does not promise a token it does not implement. |
| A6 | `Bind` rejects a sub-address that collides with the admin, the beneficiary, or the factory. The hash makes that branch unreachable in practice. | Info | The check stays so a collision cannot spend. |
| A7 | This does not reveal or replace account-level vesting ([#6235](https://github.com/gnolang/gno/issues/6235), [#5798](https://github.com/gnolang/gno/issues/5798)). | Info | Different lock, different owner. |
| A8 | v0 is unaudited. | Info | Stated on the READMEs and on `Render`. |

## Tests

`gno test` on both packages, including the filetests under
`r/nt/vesting/v0/filetests/`:

- Boundary table for the curve: before start, at start, before the
  cliff, at the cliff, `cliff == 0`, `cliff == duration`, duration 1,
  at the end, long after the end, amounts near `MaxInt64`, and a tiny
  total where floor matters. Step curves included.
- Seeded property test (40 trials, 12 steps): after every claim,
  revoke, beneficiary change, and admin change,
  `claimed + releasable + returned == total` when revoked, vested is
  monotonic in time, and a claim never exceeds vested.
- Realm tests for coin create, exact-envelope rejection, isolation,
  revoke before the remainder is claimed, claim-all-then-claim, GRC20
  register / allowance / unregistered token, drained sub-address,
  re-entry guard, and render.
- Filetests for impersonation, `maketx run`, foreign-realm coin
  funding, two-grant isolation, render injection, and a hand-called
  re-entry.

`gno` has no coverprofile. A one-off OpExec line trace under-counts:
many statements never appear as a stmt-stack line, so that trace is not
a coverage number and is not kept. The branches that are deliberately
unreachable from tests are the defensive panics in F9, F12, and A6
(non-canonical teller on a real registry token, sub-address mismatch
after a honest `Bind`, grant-id overflow). Everything else above has a
named test.

## Consequences

Callers get a grant id and a grant address. They do not get a claim on
the factory's own coins. Integrators who want a multisig admin put the
safe's address in `admin` and, when an executor exists, cross-call
`Revoke` from that safe. Until then the admin is an account or any
realm that can cross-call directly.

Upgrading the curve of an existing grant is a new package version, not
a migration. v0 grants stay on v0 math.
