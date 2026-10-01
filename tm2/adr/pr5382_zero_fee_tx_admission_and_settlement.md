# ADR: Tendermint2 support for 0-fee (realm-sponsored) transactions

## Status

Proposed (part of PR #5382, "realm transaction sponsorship")

## Context

PR #5382 lets a realm pay a user's gas (and its own storage deposits), enabling
0-fee transactions. The realm-facing design lives in gno.land and is covered by
`gno.land/adr/pr5382_realm_transaction_sponsorship.md`. This ADR records the
Tendermint2 decisions: the consensus parameter, ante handler, baseapp, mempool
configuration and `std.Tx` changes.

The core tension: whether a tx pays a fee is normally known before execution,
but sponsorship is decided during execution, when a realm calls
`runtime.PayGas`. So tm2 must (a) admit and meter a tx that carries no fee,
(b) let the VM tell the SDK mid-execution that a realm committed to pay, and
(c) reject any 0-fee tx where no realm did, identically on every validator.

## Decision

A consensus-enforced **gas credit window** for 0-fee txs, a per-validator
mempool **opt-in**, an **admission mode** that executes the tx during CheckTx,
and a **success-only settlement hook**. Sponsorship state lives on the
in-process `sdk.Context`, never on the wire.

## Key design decisions

### 1. Credit window as a consensus parameter (`Block.MaxGasCreditPerTx`)

`BlockParams.MaxGasCreditPerTx` (default `0` = disabled) sizes the gas meter of
a 0-fee tx. It is the tx's whole gas budget: `PayGas` may shrink it to what
`maxFee` buys, never raise it. It is validated `>= 0` and `<= Block.MaxGas`, so
one sponsored tx is never larger than a block, and it doubles as a chain-wide
kill switch.

### 2. Two gates: consensus enforcement vs. local admission policy

`MaxGasCreditPerTx > 0` (consensus) enables the credit window and the
"PayGas was called" enforcement in every mode. `AppConfig.AllowZeroFeeTxs`
(`[application]`, next to `MinGasPrices`) only decides whether *this* validator
admits 0-fee txs into *its* mempool. DeliverTx does not depend on it.

With the window closed a 0-fee tx is rejected in every mode, as on master.
`Tx.ValidateBasic` now accepts an empty fee because amino encodes any zero coin
as `""`, so a `0ugnot` fee always arrives as `Coin{}`; the ante restores the
rejection right after `ValidateBasic` unless the window is open, so the order of
checks, and the error a block records, match master. A sponsored tx is also not admitted before
the first block, when gno.land's genesis ante funds unknown signers.

### 3. `RunTxModeCheckExecute` for mempool admission

Ante-only CheckTx cannot know whether a realm will call `PayGas`. First-time
CheckTx of a 0-fee tx runs its messages in `RunTxModeCheckExecute`, verifying
signatures normally, persisting only the ante's sequence increment to
`checkState` and discarding message writes. `RunTxModeSimulate` was rejected:
its throwaway cache drops the sequence bump (one in-flight sponsored tx per
account per block). Rechecks stay ante-only, so pending 0-fee txs are not
re-executed every block.

### 4. "PayGas was called" is enforced in `runTx`, all modes

A 0-fee tx that never calls `PayGas` has no payer and fails in `runTx` (shared by
Check, CheckExecute and Deliver), so it never changes state. It is read from the
in-process `PayGasInfo.MaxFee > 0`. It is checked after the messages run, so a
proposer can still include one: it fails, but burns up to the credit window of
block gas that nobody pays for (see Open below).

### 5. Settlement via a success-only `EndTxHook`; `CommitTxHook`; `GasMeter.SetLimit`

`EndTxHook(ctx, result) error` runs only on success, and gno.land implements the
debit there. A returned error fails the tx. It runs in both `RunTxModeDeliver`
and `RunTxModeCheckExecute`: at admission it is a dry run whose writes land in
the discarded cache, and it is what rejects an insolvent sponsor or an
over-budget storage commitment before the tx is gossiped. Simulate (RPC gas
estimation) does not run it.

Because of the dry run, the hook must not write outside the cache-wrapped
store. Committing gno.land's transaction store does, so it moved to a new
`CommitTxHook(ctx)`, called only for a successful delivered tx.

Settlement events are appended to `result.Events` after the hook, and a failed
tx reports no events. `result.Events` feeds `LastResultsHash`.

`store.GasMeter` gains `SetLimit`, used by `PayGas` to shrink the limit. An
infinite meter reports `Limit() == 0`; `PayGas` leaves it alone, so source-gas
replay is not refused.

### 6. Sponsorship state on the in-process `sdk.Context`

`PayGasInfo` is allocated per tx in `runTx` and shared by pointer with the VM.
It is not on `Result`, which must stay wire-compatible with
`abci.ResponseDeliverTx`. Storage sponsorship is gno.land's own state and does
not touch tm2.

### 7. Reported `GasWanted` for 0-fee txs is the credit window

The ante reports `GasWanted = MaxGasCreditPerTx` for a 0-fee tx, so the mempool
packs blocks against the real worst case rather than the client's value.

## Consequences

- **Sponsorship makes `gasUsed` consensus state.** `GasUsed` is not in
  `LastResultsHash`, but the sponsor's debit is `ceil(gasUsed × price)`, so any
  difference in gas between nodes changes a bank balance and forks the app hash
  on the next block, not only when it flips an out-of-gas outcome. (Master
  already writes block gas into state through the gas price update, but only
  coarsely.)
- **`EndTxHook` now means "settle on success".** Breaking for tm2 embedders; an
  embedder that forgets to settle runs sponsored txs for free.
- **`GasMeter.SetLimit` weakens the meter's fixed-limit invariant.** The
  "only shrink" rule is enforced by `PayGas`, not the type. Breaking for
  out-of-tree meters.
- **Opting in raises a validator's CheckTx cost.** CheckExecute runs the VM, up
  to the credit window, for every first-time 0-fee tx, under the mutex that
  CheckTx shares with consensus. A tx rejected at admission consumes no sequence
  and can be resent with new bytes, so this is free to repeat. There is no
  per-account admission limit yet.
- **Opening the window without a gas price makes sponsorship inert.** Nothing
  cross-validates `MaxGasCreditPerTx` against `auth.initial_gasprice`; `PayGas`
  then panics "gas price not set". Loud, and an operator configuration issue.
- **Hardfork replay of a sponsored tx does not reproduce the source debit.**
  Settlement recomputes it from the replay's gas and the target chain's gas
  price. Unreachable today: no chain has sponsored txs.
- **A `MaxGasCreditPerTx` change is latent until restart** (memoized at
  InitChain).
- **Feature off is master, except for gas.** With the window closed every tx
  follows master's checks in master's order. Only the stdlib grew: a tx that
  loads `chain/runtime` uses about 3.5K more gas.
- **Open: failing sponsored txs.** A sponsored tx that fails at delivery charges
  nobody, while its gas still counts against the block and feeds the gas price.
  Admission closes the deterministic routes (the settlement dry run), but two
  remain. Admission and delivery see different state, and a realm can
  condition `PayGas` on the difference: on `ChainHeight()`, or on a one-shot
  allowance several pending txs all pass. And a proposer can include 0-fee txs
  that were never admitted, since DeliverTx cannot apply local admission
  policy. Either way a tx burns up to the credit window at delivery for free.
  Settlement cannot fix it: no realm called `PayGas`, so there is no one to
  charge. The mitigation has to act at delivery, e.g. an on-chain penalty on
  the signer of a failing 0-fee tx; an admission limit alone does not cover
  the proposer. Required before enabling the window on a public chain.

## Alternatives considered

- **Upfront feegrant / pre-registration.** Rejected: the realm must run
  conditional logic before deciding to sponsor.
- **Charge the sponsor on failure.** Implemented, then reverted: it lets an
  attacker drain a realm by failing the tx after `PayGas` (see the gno.land ADR).
- **`RunTxModeSimulate` for admission.** Rejected: it drops the sequence bump
  and needed a signature-verification override.
- **Settlement inside tm2 baseapp.** Rejected: coin logic is app-specific; tm2
  provides the hooks and the meter primitive.
- **Carrying sponsorship state on `Result`.** Rejected: wire compatibility.

## References

- gno.land ADR: `gno.land/adr/pr5382_realm_transaction_sponsorship.md`
- Files: `tm2/pkg/bft/types/params.go`, `tm2/pkg/bft/abci/types/types.go`,
  `tm2/pkg/sdk/auth/ante.go`, `tm2/pkg/sdk/baseapp.go`, `tm2/pkg/sdk/types.go`,
  `tm2/pkg/sdk/context.go`, `tm2/pkg/sdk/config/config.go`,
  `tm2/pkg/store/types/gas.go`, `tm2/pkg/std/tx.go`.
