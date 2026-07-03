# ADR: Realm Transaction Sponsorship

## Status

Proposed

## Context

Every transaction on Gno requires the signer to hold gnot for gas fees and storage deposits. This creates onboarding friction — users must acquire gnot before their first interaction. Existing workarounds (first-signer co-signing) require off-chain infrastructure.

Other chains solve this differently: Ethereum uses EIP-4337 (Paymasters + bundlers + smart wallets), Solana has native fee payers, NEAR has access keys with gas allowances, Cosmos SDK has the feegrant module. Each requires either off-chain infrastructure or lacks on-chain conditional logic.

## Decision

Introduce two independent native functions that allow realms to sponsor transaction costs:

- **`runtime.PayGas(maxFee int64)`** — realm pays gas fees, capped at maxFee (ugnot)
- **`runtime.PayStorage(maxDeposit int64)`** — realm pays storage deposits, capped at maxDeposit (ugnot)

Plus a tx-level flag:

- **`Fee.SponsorStorage = true`** — defers storage deposits to end-of-tx for multi-message sponsorship

## Key Design Decisions

### 1. Mid-execution sponsorship

The realm decides **during execution** whether to sponsor. This allows running arbitrary validation logic (check balances, collect alternative tokens, verify whitelists) before committing to pay. No other chain does this without off-chain infrastructure.

**Why not upfront commitment?** Pre-registration (`RegisterGasSponsor`) or feegrant-style allowances can't run on-chain conditional logic. The realm needs to execute code to decide.

### 2. Credit window (`MaxGasCreditPerTx`)

A consensus parameter defines how much gas a 0-fee tx can consume before `PayGas` is called. Execution starts on "credit" — if `PayGas` is never called, the tx fails. CheckTx simulation validates this before mempool entry.

**Why a consensus param?** Validators must agree on the credit window size. It bounds the free execution validators absorb for invalid 0-fee txs.

### 3. Settlement inside cached context

Gas and storage settlement execute inside the cached message execution context. On tx failure, the cache is not written — the realm does NOT pay. This prevents griefing attacks where an attacker engineers a failure after sponsorship to drain the realm without it receiving anything in return.

**Why not outside the cache?** If settlement persisted on failure, an attacker could: call a realm that collects USDC then calls PayGas → trigger failure → USDC collection reverts but gas payment persists → realm loses gnot, attacker keeps USDC.

### 4. PayGas and PayStorage are independent

A realm can call either or both. Gas and storage are separate concerns with separate budgets. A DeFi realm might sponsor gas (users pay in USDC) but not storage. A gaming realm might sponsor both.

**Why not one function?** Gas and storage have different economics (gas = computation, storage = persistent state). Separate caps prevent one from consuming the other's budget.

### 5. Function creator must match payer

The function containing `PayGas`/`PayStorage` must be defined in the same realm that `CurrentRealm` identifies as the payer. This prevents cross-realm callback attacks where a malicious realm tricks another realm into calling PayGas via a passed closure.

**Checked via:** `callerFrame.Func.PkgPath == currentPkgPath` (frame inspection at call depth 2).

### 6. SponsorStorage tx flag

Storage diffs are per-message (cleared between messages by `ClearObjectCache`). Without a flag, PayStorage only covers messages after it's called. The `SponsorStorage` flag signals upfront that diffs should be accumulated across all messages and settled once at end-of-tx.

**Why a tx flag instead of always deferring?** Per-message storage deposit is the existing behavior. Changing it for all txs would be a breaking change. The flag is opt-in and backward compatible.

### 7. Gas price from existing auth module

The gas price for settlement comes from `auth.GasPriceKeeper.LastGasPrice()`, not a new consensus parameter. This reuses the existing gas price system and avoids configuration duplication.

## Alternatives Considered

| Alternative | Why not |
|-------------|---------|
| Cosmos feegrant module | No on-chain conditional logic. Can't collect USDC before sponsoring. |
| Off-chain relayer (EIP-2771 style) | Requires external infrastructure. Centralization risk. |
| Realm pre-registration | Can't run arbitrary logic before committing to pay. |
| Post-execution refund | User still needs gnot upfront. Not truly gasless. |
| Single PayGas covering everything | Gas and storage have different economics. Separate caps needed. |
| Always defer storage to end-of-tx | Breaking change for existing per-message behavior. |

## References

- Full HLD: `docs/design/realm-gas-sponsorship-hld.md`
- Implementation PR: gnolang/gno#5382
