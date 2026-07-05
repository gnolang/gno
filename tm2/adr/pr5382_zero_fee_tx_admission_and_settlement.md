# ADR: Tendermint2 support for 0-fee (realm-sponsored) transactions

## Status

Proposed (part of PR #5382, "realm transaction sponsorship")

## Context

PR #5382 lets a realm pay a user's gas and storage from its own balance,
enabling 0-fee ("gasless") transactions. The realm-facing design and the
settlement logic live in gno.land and are covered by
`gno.land/adr/pr5382_realm_transaction_sponsorship.md`. This ADR records the
**Tendermint2-layer** decisions, because the feature required consensus-param,
ante-handler, baseapp, mempool-config, and `std.Tx` changes that the tm2 layer
must own and keep deterministic.

The core tension: whether a tx pays a fee is normally known before execution,
but *realm sponsorship is decided during execution* (a realm calls
`runtime.PayGas` only after running its own logic). So tm2 must (a) admit and
meter a tx that carries no fee, (b) let the VM tell the SDK, mid-execution, that
a realm committed to pay, and (c) reject any 0-fee tx where no realm did so —
identically on every validator, at both CheckTx and DeliverTx.

## Decision

Add a bounded, consensus-enforced **gas credit window** for 0-fee txs, a
per-validator mempool **opt-in**, a dedicated **admission mode** that executes
the tx during CheckTx, and a **success-only settlement hook**; carry sponsorship
state on the in-process `sdk.Context` (never on the wire).

## Key design decisions

### 1. Credit window as a consensus parameter (`Block.MaxGasCreditPerTx`)

A new `BlockParams.MaxGasCreditPerTx` (default `0` = feature disabled) sizes the
gas meter for a 0-fee tx before any realm pays. It is validated `>= 0` and
`<= Block.MaxGas` (`tm2/pkg/bft/types/params.go`), so one sponsored tx can never
be sized larger than a whole block. Being a consensus param, it is identical on
every validator and doubles as a chain-wide kill switch.

### 2. Two-gate model: consensus enforcement vs. local admission policy

`MaxGasCreditPerTx > 0` (consensus) enables the credit window and the
"PayGas-was-called" enforcement in **every** mode. A separate per-validator
`AppConfig.AllowZeroFeeTxs` (`tm2/pkg/sdk/config/config.go`, `[application]`
section — alongside `MinGasPrices`) gates only whether *this* validator admits
0-fee txs into *its* mempool. DeliverTx behaviour does not depend on the opt-in,
so block validation stays deterministic even for validators that reject 0-fee
txs locally.

### 3. `RunTxModeCheckExecute` for mempool admission

Ante-only CheckTx cannot know whether a realm will call `PayGas`. A new
`RunTxMode` runs the tx's messages during CheckTx (to validate sponsorship) while
persisting only the ante's account-sequence increment to `checkState` and
discarding the message writes; it verifies signatures normally. `RunTxModeSimulate`
was rejected for this because its throwaway cache discards the sequence bump,
capping a sender to one in-flight sponsored tx per block. Rechecks fall back to
the cheap ante-only `RunTxModeCheck`.

### 4. `PayGas`-was-called enforcement in `runTx`, all modes

A 0-fee tx that never calls `PayGas` has no payer and is rejected in `runTx`
(shared by Check/CheckExecute/Deliver), so a block proposer cannot force-include
a free tx that skips sponsorship. Whether a realm called `PayGas` is read from
the in-process `PayGasInfo.MaxFee > 0`, not from the wire.

### 5. Settlement via a success-only `EndTxHook`; `GasMeter.SetLimit`

tm2 exposes an `EndTxHook(ctx, result) error` invoked **only on tx success**;
gno.land implements the actual debiting. On failure everything reverts and the
realm pays nothing (see Alternatives). `store.GasMeter` gains `SetLimit`, used by
`PayGas` to *shrink* the credit window to what `maxFee` affords (never raise it).

### 6. `std.Fee.SponsorStorage` and the `ValidateBasic` relaxation

`std.Fee` gains a `SponsorStorage bool` (proto field 3, `omitempty`) that defers
storage deposits to end-of-tx so one `PayStorage` covers a multi-message tx.
`Tx.ValidateBasic` accepts a canonical zero fee (previously rejected) so 0-fee
txs pass structural validation; the ante rejects `SponsorStorage=true` on a
normal fee-paying tx so the mistake surfaces at submission.

### 7. Sponsorship state on the in-process `sdk.Context`, not on `Result`

`Result` must stay wire-compatible with `abci.ResponseDeliverTx`, so
`PayGasInfo`/`PayStorageInfo` (and `txCaller`/`sponsorStorage`) live on the
in-process `sdk.Context` as shared pointers, freshly allocated per tx in `runTx`.

### 8. Reported `GasWanted` for 0-fee txs is the credit window

The ante reports `GasWanted = MaxGasCreditPerTx` for a 0-fee tx (not the
client-supplied value) so the mempool packs blocks against real worst-case gas.

## Consequences

- **Determinism is preserved.** Classification (0-fee, credit window,
  PayGas-enforcement) is consensus-param-driven and mode-independent; `GasUsed`
  is not part of `LastResultsHash`, so settlement metering choices don't fork the
  app hash.
- **The `EndTxHook` contract now means "settle on success."** A downstream tm2
  embedder that sets this hook must implement settlement (and its own failure
  semantics); an embedder that *forgets* it would run sponsored txs for free.
- **`GasMeter.SetLimit` weakens the meter's fixed-limit invariant.** Every meter
  holder can now resize the limit; the "only shrink" rule is enforced by the
  `PayGas` caller, not the type. Alternate meter implementations must replicate it.
- **The universal `sdk.Context` carries feature-specific fields.** Convenient but
  couples a reusable layer to one feature; correctness relies on per-tx allocation.
- **Opt-in raises a validator's CheckTx cost.** `CheckExecute` runs the full VM
  (up to `MaxGasCreditPerTx`) per first-time 0-fee tx, with no per-account
  admission rate limit in v1 (deferred). Rechecks are ante-only to avoid
  per-block re-execution.
- **A `MaxGasCreditPerTx` change is latent** until node restart (memoized at
  InitChain), despite being governance-tunable.
- **Open item:** sponsored-tx compute is charged to the block gas meter, which
  feeds the dynamic gas-price update, so gasless-tx load raises the price normal
  users pay while contributing nothing to the fee market. Whether that is
  intended is unresolved and should be adjudicated before enabling on a live
  chain.

## Alternatives considered

- **Upfront feegrant / pre-registration** (payer declared before execution).
  Rejected: the realm must run conditional on-chain logic (collect an alt-token,
  check a whitelist) before deciding to sponsor.
- **Charge the sponsor for gas on failure** (parity with normal fee txs, which
  keep their fee on failure). Implemented, then reverted: it reintroduces a
  griefing attack where an attacker engineers a post-`PayGas` failure so the
  realm's message-side collection reverts but its gnot is still taken. The atomic
  all-or-nothing rule (realm pays only when the whole tx commits) is safer; the
  residual free-execution-on-failure is bounded by the credit window and only
  reachable by a block proposer (who can waste their own block anyway).
- **`RunTxModeSimulate` for admission.** Rejected: it discards the ante sequence
  increment (one in-flight sponsored tx per account per block) and required a
  signature-verification override; `RunTxModeCheckExecute` avoids both.
- **Settlement inside tm2 baseapp.** Rejected: fee/coin logic is app-specific;
  tm2 provides the hook and the meter primitive, gno.land owns the debiting.
- **Carrying sponsorship state on `Result`.** Rejected: `Result` must stay
  wire-compatible with `abci.ResponseDeliverTx`.

## References

- gno.land ADR: `gno.land/adr/pr5382_realm_transaction_sponsorship.md`
- HLD: `docs/design/realm-gas-sponsorship-hld.md`
- Files: `tm2/pkg/bft/types/params.go`, `tm2/pkg/bft/abci/types/types.go`,
  `tm2/pkg/sdk/auth/ante.go`, `tm2/pkg/sdk/baseapp.go`, `tm2/pkg/sdk/types.go`,
  `tm2/pkg/sdk/context.go`, `tm2/pkg/sdk/config/config.go`,
  `tm2/pkg/store/types/gas.go`, `tm2/pkg/std/tx.go`.
