# ADR: establish receipt for the valoper register and rotation fees

## Status

Proposed (PR pending).

## The problem

`r/gnops/valopers` charges two fees, both read from sysparams: `node:valoper:register_fee` in `Register`, and `node:valoper:rotation_fee` in `UpdateSigningKey`. Each was enforced the same way:

```gno
if fee := sysparams.GetValoperRegisterFee(); fee > 0 {
	minFee := chain.NewCoin("ugnot", int64(fee))
	sentCoins := unsafe.OriginSend()
	if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
		panic(...)
	}
}
```

`unsafe.OriginSend()` reports the transaction's declared send **envelope**, not what this realm received. Only a direct `maketx call` on valopers credits the envelope to valopers' address. For a `maketx run`, the keeper sets `pkgAddr := caller` (`gno.land/pkg/sdk/vm/keeper.go`), so the coins move from the caller to the caller and never land anywhere at all — while `OriginSend()` still reports the full amount to the callee. `gno.land/adr/pr6062_payable_send_check.md` already states this: "MsgRun is exempt: the coins are moved from the caller to the caller, so nothing actually moves."

So the amount check verified *intent* and never *receipt*. An operator could attach `-send 1000ugnot` to a `maketx run` script that calls `Register`, satisfy the fee check, and keep the coins. No hostile script is required — the bypass is free by construction.

Reported by @D4ryl00 as a P2, pre-existing and explicitly deferred to its own change. The `UpdateSigningKey` instance was found while scoping that one.

### Why an address comparison would not have worked

The natural-looking fix — compare `OriginCaller` or `Previous().Address()` against something — cannot distinguish the two shapes. A run script's realm is **address-indistinguishable** from its user: inside `main(cur realm)` the ephemeral `gno.land/e/<addr>/run` realm's own address *is* the caller's EOA address. Every address-based guard already in this realm accepts a `maketx run`:

- `Register`'s squat guard (`unsafe.OriginCaller() != addr`) passes, because `OriginCaller` is the user.
- `UpdateSigningKey`'s `v.auth.AssertPreviousOnAuthList(0, cur)` passes, because it compares `rlm.Previous().Address()`, and that address is the operator's own.

Only the pkgPath differs between the two shapes, and `IsUserCall()` is the check that reads it.

This matters for reviewing the test: section 6 of the txtar is not redundant with the auth-list check. Reverting the rotation guard alone makes a `maketx run` rotation succeed.

### A second, independent way the fee check yielded nothing

`Coin.Amount` is an `int64` but the sysparam is a `uint64`, and `int64(fee)` wraps. At `fee >= 2^63`, `minFee.Amount` is negative and `sentCoins[0].IsLT(minFee)` is false for *every* payment. The realm reads the fee as configured, runs the receipt guard, and collects 1ugnot while reporting `payment must not be less than -1ugnot`.

Two ways to reach it, neither requiring malice:

- Governance can set any `uint64` through the generic `r/sys/params` factories. Setting a deliberately prohibitive fee to *stop* registrations would instead make them free.
- `valopers/proposal.ProposeNewMinFeeProposalRequest` takes an `int64` and stores `uint64(newMinFee)`, so a negative fee — the obvious way to write "disable the fee" — round-trips to 2^64-1.

## Severity

Latent, not a mainnet blocker. Stated up front because it determines urgency, and each point is verified rather than assumed:

- **Both fees default to `0`** (`r/sys/params/valoper.gno`) and nothing in the repo sets them. Only the *key* appears outside tests, in the proposal factory that would set it. `if fee := ...; fee > 0` is dead code today.
- **Raising either fee takes a GovDAO proposal**, so there is a deliberate, observable action between "latent" and "live".
- **Even live, the fee is economically inert.** valopers has no banker and no withdrawal path, so fees strand at the realm address. Bypassing one avoided a burn; it never moved anyone's funds.
- **Registration confers no consensus power.** `Register` ends at `validators.NotifyValoperChanged`, which only writes `valoperCache`; joining the effective valset needs a separate GovDAO vote. Free registration is registry spam, bounded by gas and storage deposit.

The reason to fix it now rather than later: the control silently does nothing the moment governance enables it, which is the documented plan post-transfer-enablement, and nothing would go red.

## Decision

### 1. Pair every `OriginSend` amount check with `IsUserCall`

`assertPaidCallIsDirect` panics unless `rlm.Previous().IsUserCall()`, and is called at the top of both `fee > 0` branches.

This is the established in-tree pattern: `r/sys/namereg/v1.Register` does exactly this for its anti-squatting payment, and its comment already carries the reasoning. `IsUserCall` is the only `PreviousRealm` shape where the envelope is guaranteed to have landed at this realm's address — it excludes intermediate code realms and user-run ephemeral realms alike.

### 2. The guard sits inside the `fee > 0` branch, not at the top of the function

While a fee is unset there is no payment to establish receipt of, and gating unconditionally would break the operator-authored `maketx run` flows the squat guard is happy to accept. The restriction appears exactly when, and only when, money is involved.

This also answers the objection the old `Register` godoc raised against `IsUserCall` — that it "would only block legitimate `maketx run` flows". Section 1 of the txtar pins that compatibility property.

### 3. Check `IsCurrent()` before `Previous()`

AGENTS.md requires it in crossing functions, and every other `(_ int, rlm realm)` helper in this tree already does it (`authorizable`, `sys/params/delegate`, `gov/dao/types`, `sys/validators/v0/cache`). Both call sites pass their live `cur`, so this is defense-in-depth against a future caller threading a stale or stashed realm value: `Previous()` reads the `prev` field verbatim, so on a non-live token it answers for the wrong frame.

### 4. Bound the fee before the `int64` conversion

`minFeeCoin(fee uint64)` panics with `ErrFeeParamOutOfRange` when `fee > math.MaxInt64`. Failing closed is the point: an unrepresentable fee must block the paid path, not open it.

The bound sits at the consumption site rather than at `ProposeNewMinFeeProposalRequest`, for two reasons. It covers both routes to a bad value (the proposal factory *and* the generic `r/sys/params` factories), and `ProposeNewMinFeeProposalRequest`'s signature is preserved for historical-replay compatibility — gnoland-1's `set_minfee.gno` MsgRun calls it, so adding a panic there would change replay behavior for an input no longer under our control.

### 5. Remove a false claim from `Register`'s godoc

The old comment argued `IsUserCall` was unnecessary because "fees are validated against `banker.OriginSend` in a way that's symmetric to `IsUserCall` via direct comparison". There was no such comparison. That rationale is what would have carried this through another review.

## Alternatives considered

**A true receipt check.** Track cumulative fees in realm state and require the realm's own balance to have risen by `fee` since the last call. This works for any caller shape, including a realm on an operator's auth list. Rejected: it puts new persisted state and a banker into a genesis realm, and it lets anyone pre-pay another operator's fee by donating to the realm address. Not worth it while both fees are 0 and collected fees are unwithdrawable anyway.

**Gate on `IsUser()` instead of `IsUserCall()`.** Wrong in exactly the way this class of bug is usually wrong: `IsUser()` accepts `maketx run` ephemeral realms, which is the shape being excluded. See `docs/resources/effective-gno.md § Verifying inbound Coin payments`.

**Clamp the out-of-range fee to `MaxInt64` instead of panicking.** Same practical outcome (nobody can pay), but the failure reads as a bizarre payment demand rather than a misconfiguration, and it silently accepts a param the realm cannot represent.

**Reject negative fees in `ProposeNewMinFeeProposalRequest`.** Covers the likeliest trigger but not the generic `r/sys/params` path, and changes replay behavior for a signature deliberately frozen. Rejected in favour of the consumption-site bound.

**A shared `p/` helper for the `IsUserCall` + `OriginSend` pairing.** This is now the fifth in-repo hand-rolled copy of that pattern (namereg, boards' `checkAnonFee`, atomicswap, disperse, valopers). Worth doing, but out of scope here; noted as follow-up work.

## Consequences

**A realm on an operator's auth list can no longer drive a paid rotation.** `AddToAuthList` takes a plain `member address`, so a realm can sit on an operator's auth list. Once `rotation_fee` is nonzero, this guard blocks that shape, because it is not a user call.

An EOA rotation bot is unaffected, and that is the shape the auth-list docs describe ("HSM-bound or rotation-bot"), so nothing documented breaks. A realm-mediated *paid* rotation would. The alternative that preserves it is the true receipt check rejected above. This is recorded in the `assertPaidCallIsDirect` godoc as a KNOWN LIMIT so it is a visible decision rather than an accident.

**An out-of-range fee now blocks registration instead of making it free.** Both are failures, but the new one is loud and reversible by the same governance action that caused it.

**No behavior change while both fees are 0**, which is their state today and in genesis.

## Validation

`gno.land/pkg/integration/testdata/valopers_fee_receipt.txtar`, six sections on a real chain:

1. fee unset — a `maketx run` registration is still accepted (the property the guard must not break)
2. governance raises `register_fee` to 1000ugnot
3. the `maketx run` bypass is refused, and valopers holds nothing
4. underpaying a direct call is still caught by the amount check
5. the legitimate direct call succeeds **and the realm's balance actually rises**
6. the same for `rotation_fee`, throttle dropped so the rotation is not height-blocked

Balance goes 0 → 1000 → 1500 across the paid sections, which is the assertion the old check could not make.

Each guard was reverted individually and the corresponding section confirmed red: reverting the `Register` guard prints `UNREACHABLE: registered without paying`, reverting the `UpdateSigningKey` guard prints `UNREACHABLE: rotated without paying`. The second is the one worth noting — it proves the auth-list check does not independently block that shape.

`TestValopers_Register/unrepresentable_fee_fails_closed` covers the fee bound, and was likewise confirmed red against the unbounded conversion.

Green: `gno test` over `examples/gno.land/r/gnops/valopers/...`, the `valopers` and `params_valset` integration suites (20 scripts), and `gno fmt -diff` on the changed files.

## AI assistance

Written with AI assistance (Claude Code). The bypass and the fee-param wrap were each reproduced before being fixed, and every guard was re-reverted afterwards to confirm its test fails without it. The human author reviewed and owns the change.
