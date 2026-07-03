# HLD: Realm Transaction Sponsorship (`runtime.PayGas` + `runtime.PayStorage`)

- **Status**: Draft
- **Authors**: @omarsy
- **Created**: 2026-03-22

## 1. Problem Statement

Today, every transaction on Gno requires the signer to hold gnot to pay gas fees. This creates three friction points:

- **Onboarding friction**: Gasless transactions today require off-chain co-signing infrastructure. Without it, users must acquire gnot to pay gas — adding friction to the first on-chain interaction.
- **No pay-in-any-token**: A DeFi user with USDC can't swap without first acquiring gnot. The chain's native token becomes a barrier, not a feature.
- **Session key overhead**: Temporary keys for games and interactive dApps require the dApp to fund each key with gnot — operational overhead that scales with users.

**Goal**: Allow realms to pay gas on behalf of users, enabling truly gasless transactions (0 gnot in user wallet) while maintaining network security.

**Prerequisite**: The user's account must already exist on-chain (created when they first received funds). `PayGas` solves "user has 0 gnot balance", not "user has never existed." Account creation remains a separate one-time step (faucet, initial transfer, etc.).

## 2. Current Fee Model: First-Signer Pays

In Gno's current fee model, the **first signer always pays the gas fee** for all signers ([ante.go:117-125](tm2/pkg/sdk/auth/ante.go#L117-L125)). While not designed as a sponsorship mechanism, multi-signer txs could theoretically be used for gas sponsorship — though this is not done in practice today:

```go
// fetch first signer, who's going to pay the fees
signerAccs[0], res = GetSignerAcc(newCtx, ak, signerAddrs[0])
// ...
// deduct the fees
if !tx.Fee.GasFee.IsZero() {
    res = DeductFees(bank, newCtx, signerAccs[0], ak.FeeCollectorAddress(ctx), std.Coins{tx.Fee.GasFee})
}
```

This means a sponsor can already pay gas for a user by being the first signer:

```
Signer 0: Sponsor account (pays gas)
Signer 1: User account (0 gnot, just signs the action)
```

### Limitations of the Current Approach

| Limitation | Detail |
|-----------|--------|
| **Requires off-chain co-signing infrastructure** | The sponsor must sign every transaction in real-time, requiring an always-online co-signing service — a centralized single point of failure |
| **No on-chain conditional logic** | The sponsor can inspect the transaction content off-chain before co-signing, but cannot run on-chain logic (check contract state, verify balances, enforce rate limits) before committing to pay |
| **Sponsor is a separate EOA account** | A specific account pays, not the realm itself — realms don't have private keys, so the sponsor must be a separately funded and managed account |

### Why `PayGas` Is Better

| First-signer (today) | PayGas (proposed) |
|----------------------|-------------------|
| Needs co-signing infrastructure | User sends tx alone |
| Sponsor must be online to sign | Realm pays automatically on-chain |
| No conditional logic before paying | Realm can collect USDC, check whitelists, etc. |
| Sponsor is a specific account | Realm's own balance pays |
| Centralized co-signer | Fully decentralized, no off-chain infra |

## 3. Design Overview

The design introduces three components:

1. **`runtime.PayGas(maxFee int64)`** — a native function callable by realms during execution to commit to paying gas (maxFee in ugnot)
2. **Gas credit window** — a bounded amount of gas that a 0-fee transaction can consume before `PayGas` is called, configured as a consensus parameter
3. **Validator-level opt-in** — each validator independently decides whether to accept 0-fee transactions into their mempool

### Core Principle

The realm decides **mid-execution** whether to sponsor gas. This allows the realm to run arbitrary validation logic (check balances, collect alternative tokens, verify whitelists) before committing to pay.

## 4. Transaction Flow

```
User (0 gnot)                    Validator (opt-in)                All Nodes
     │                                │                                │
     │  Send 0-fee tx                 │                                │
     │───────────────────────────────►│                                │
     │                                │                                │
     │                      1. Verify signature                        │
     │                      2. Simulate with credit gas                │
     │                         → PayGas() called?                      │
     │                         → Realm has funds?                      │
     │                      3. Accept / reject                         │
     │                                │                                │
     │                      4. Gossip to network                       │
     │                                │───────────────────────────────►│
     │                                │                                │
     │                                │                      DeliverTx:
     │                                │                      - Execute tx
     │                                │                      - PayGas() → settle
     │                                │                      - No PayGas → fail
     │                                │                                │
```

## 5. Detailed Design

### 5.1 Native Function: `runtime.PayGas(maxFee int64)`

A new native function available to realm code:

```go
// std package — available to all realms
// maxFee is denominated in ugnot
func PayGas(maxFee int64)
```

**Behavior:**

- Can only be called **once** per transaction — the first realm to call it becomes the gas sponsor
- Only **realms** can call it (not packages, not `main` in `MsgRun`)
- **Function creator must match payer**: the function containing the `PayGas` call must be defined in the same realm that `CurrentRealm` identifies as the payer. If they differ, `PayGas` panics. This prevents cross-realm callback tricks.
- `maxFee` is the maximum amount in ugnot the realm is willing to pay
- The gas limit is **derived**: `gasLimit = maxFee / MinGasPrice` (both are consensus params, deterministic across all nodes)
- On call: the system verifies the realm has sufficient gnot balance (`realm.balance >= maxFee`)
- If the realm cannot cover the cost, `PayGas` panics (tx fails, no state changes)
- Gas consumed **before** `PayGas` is called counts toward the derived gas limit

**Why `maxFee` instead of `limitGas`:** Realm authors think in cost (ugnot), not gas units. Using `maxFee` protects the realm's budget against gas price governance changes — if `MinGasPrice` doubles, the realm gets less computation but never overspends. No code change needed.

**Call Rules:**

| Rule | Rationale |
|------|-----------|
| Only realms can call it | Realms have addresses and balances; packages and `main` do not |
| Function creator == payer | The function containing `PayGas` must be defined in the same realm package as `CurrentRealm`. Prevents cross-realm callbacks, closures, or foreign methods from triggering `PayGas` on behalf of another realm. |
| Only once per transaction | Simplifies gas meter accounting, prevents stacking. Second call panics. |
| Must be called within credit window | If credit gas is exhausted before `PayGas`, tx fails |
| Realm must have sufficient balance | `realm.balance >= maxFee`, checked at call time to fail fast |

**Examples of what is allowed and rejected:**

```
✅ MsgCall{A.Swap} → Swap() calls PayGas       (creator=A, payer=A)
✅ MsgCall{A.Foo}  → A calls B.Bar() → PayGas   (creator=B, payer=B — B chose to pay)
✅ MsgRun{main}    → main calls A.Swap() → PayGas (creator=A, payer=A)
✅ Realm A method   → A.Router.Swap() → PayGas   (creator=A, payer=A — same realm)

❌ Package calls PayGas                          (packages have no balance)
❌ main() calls PayGas                           (main has no balance)
❌ Closure from A passed to B, calls PayGas      (creator=A, payer=B — mismatch)
❌ Foreign method: struct from A, method called in B → PayGas (creator=A, payer=B — mismatch)
```

**Edge cases:**

| Case | Behavior |
|------|----------|
| Derived gas limit < gas already consumed | `PayGas` panics immediately — the realm's budget is already exceeded at current gas price. Tx fails, no state changes. |
| `maxFee = 0` | `PayGas` panics — invalid argument. |
| `maxFee < 0` | `PayGas` panics — invalid argument. |
| Multi-message tx | All messages share a single gas meter. `PayGas` in any message covers gas for the **entire tx** (all messages). Messages are atomic — if any message fails, all state changes (including earlier messages) are reverted. |

**Note on security:** A realm that calls `PayGas` unconditionally in a public function is opting in to pay gas for any caller. Realm authors must validate callers (whitelists, token collection, etc.) **before** calling `PayGas` to avoid being drained.

### 5.2 Native Function: `runtime.PayStorage(maxDeposit int64)`

`PayGas` covers gas fees but **not storage deposits**. Storage deposits are charged when realm state changes (bytes written × storage price). A separate `PayStorage` function lets the realm opt into paying storage deposits with its own cap.

`PayGas` and `PayStorage` are **fully independent** — a realm can call either or both:

| Combination | Gas paid by | Storage paid by |
|-------------|------------|-----------------|
| Neither | User | User |
| PayGas only | Realm | User |
| PayStorage only | User | Realm |
| Both | Realm | Realm (truly gasless) |

```go
// std package — available to all realms
// maxDeposit is denominated in ugnot
func PayStorage(maxDeposit int64)
```

**Behavior:**

- Same call rules as `PayGas`: only realms, function creator must match payer, once per tx
- `maxDeposit` caps the total storage deposit the realm will pay in ugnot
- If total storage deposits exceed `maxDeposit`, the tx fails (storage budget exceeded)
- If `PayGas` was already called, `PayStorage` must be called by the **same realm**
- Can be called independently of `PayGas` (realm pays storage but not gas, or vice versa)

**Storage deposit deferral:** Storage deposits are settled at the **end of the transaction** (in `endTxHook`), not per-message. This means:
- In multi-message txs, `PayStorage` covers storage for ALL messages — including messages executed before `PayStorage` is called
- The gno transaction store accumulates `RealmStorageDiffs()` across all messages, and settlement runs once after all messages complete
- If the tx fails, all storage deposits revert (inside cached context)

**Combined usage for truly gasless transactions:**

```go
func DoWork(cur realm) string {
    runtime.PayGas(1000000)    // realm pays gas up to 1 gnot
    runtime.PayStorage(500000) // realm pays storage up to 0.5 gnot
    // ... work that modifies state ...
    return "done"
}
```

Without `PayStorage`, the caller pays storage deposits even when `PayGas` is active. A user with 0 gnot calling a realm that modifies state would fail on the storage deposit.

### 5.3 Gas Credit Window

A **consensus parameter** that defines the maximum gas a 0-fee transaction can consume before `PayGas()` must be called:

```go
// In ConsensusParams
type GasParams struct {
    MaxGas            int64
    MaxGasCreditPerTx int64 // NEW — credit window (e.g., 500,000)
    MinGasPrice       Coin  // NEW — used for PayGas settlement and gas limit derivation
}
```

**Invariant:** `MinGasPrice` must be > 0 when `MaxGasCreditPerTx` > 0. Otherwise `PayGas` gas limit derivation (`maxFee / MinGasPrice`) is a division by zero.

- Applies only to transactions with `GasFee == 0`
- Credit gas counts toward the block gas limit
- If `PayGas()` is not called before the credit is exhausted, the transaction fails
- Tunable via governance — can be increased if realms need more setup gas, or decreased if abuse occurs
- **Recommended initial value: `500,000` gas** — must be large enough to support multi-message patterns (e.g. `Approve` + `TransferFrom` + `PayGas`). A single cross-realm call can consume 50-100k gas, so multi-msg txs with 2-3 calls before `PayGas` need 300-500k gas of credit.
- **Default value: `0`** — feature is disabled until activated by governance. Setting `MaxGasCreditPerTx = 0` acts as a kill switch. Requires a coordinated chain upgrade so all nodes understand the new consensus param before activation.

### 5.4 Validator-Level Opt-In

Each validator configures locally whether to accept 0-fee transactions:

```toml
# config.toml — per-validator configuration
[mempool]
allow_zero_fee_txs = false          # default: false (conservative)
```

**Behavior by role:**

| Role | Opt-in | Opt-out (default) |
|------|--------|-------------------|
| **CheckTx** | Verify signature, then **simulate** the tx with credit gas. Only accept into mempool if `PayGas` is called and realm has funds. | `GasFee == 0` → reject immediately |
| **Block validator** | Executes DeliverTx normally (no special handling) | Executes DeliverTx normally (no special handling) |

**Key insight**: Simulation happens at CheckTx time, **before** the tx enters the mempool. This means only validated 0-fee txs are gossiped to other nodes — no network-wide spam. Block validation (DeliverTx) is the same for all validators. A validator that doesn't accept 0-fee txs into their own mempool will still validate blocks containing them.

### 5.5 Gas Meter Mechanics

During execution of a 0-fee transaction, the gas meter operates in two phases:

```
Phase 1: Credit Phase
  - GasMeter initialized with limit = MaxGasCreditPerTx
  - Gas consumed from credit budget
  - No payer yet

Phase 2: Realm-Sponsored Phase (after PayGas called)
  - Gas limit derived: maxFee / MinGasPrice
  - GasMeter limit updated to derived gas limit
  - Gas already consumed in Phase 1 carries over
  - Realm balance is the backing

Settlement (post-execution):
  - Actual cost = gasUsed × MinGasPrice
  - Deducted from realm's gnot balance
  - On success: state + settlement committed
  - On failure: everything reverted (realm does NOT pay)
```

### 5.6 Settlement

After transaction execution completes **successfully**:

1. Calculate `actualCost = gasUsed × MinGasPrice`
2. Deduct `actualCost` from the realm's gnot balance via `BankKeeper.SendCoins(realmAddr, feeCollectorAddr, actualCost)`
3. State committed (including settlement)

**On failure after `PayGas`**: The realm does **not** pay. All state is reverted, including the settlement. This protects realm authors from griefing attacks where an attacker engineers a failure after `PayGas` to drain the realm's gnot while the realm gets nothing (e.g., USDC collection reverted but gnot deducted).

**Why this is safe**: CheckTx simulation already validated that `PayGas` is called and the realm has funds. Failed txs in DeliverTx only occur on state divergence (another tx changed conditions between CheckTx and DeliverTx) — the same race condition that exists for all txs today. This is rare and bounded.

**Gas price for 0-fee txs:** Normal txs derive gas price from `GasFee / GasWanted`. For 0-fee txs, this is `0/0` (undefined). Instead, the **consensus `MinGasPrice`** is used for settlement (see `GasParams` in Section 5.2).

**Settlement context:** Settlement executes **inside** the cached message execution context (in `vm/keeper.go`, at the end of message execution). It is committed together with all other state changes only when the tx succeeds. On failure, everything reverts — message state and settlement. See Section 8.5.

**VM → SDK communication:** The `PayGas` native function populates a `PayGasInfo` struct on the execution context:

```go
type PayGasInfo struct {
    RealmAddr crypto.Address // the realm that called PayGas
    MaxFee    int64          // maxFee argument in ugnot
}
```

This is set by the GnoVM during execution and read by the SDK settlement logic in `runTx`. The execution context already flows between the VM and SDK layers, so no new plumbing is needed — just a new field on the context.

### 5.7 CheckTx Simulation

Validators that opt-in **simulate** 0-fee txs at CheckTx time, before accepting them into the mempool. This ensures only validated 0-fee txs are gossiped to the network.

**How it works:** For 0-fee txs, CheckTx runs the full VM execution (not just the ante handler) using the existing `RunTxModeSimulate` path with a cached context:

```
CheckTx for 0-fee tx:
  1. Verify signature
  2. Simulate execution with credit gas (MaxGasCreditPerTx)
  3. PayGas() called and realm has funds? → accept into mempool
  4. PayGas() not called or insufficient funds? → reject
```

**What state it runs against:** The validator's `checkState` — the latest committed state. This is the same state normal CheckTx runs against.

**Latency impact:** Simulation cost per tx is bounded by `MaxGasCreditPerTx` (credit window gas). With `MaxGasCreditPerTx = 500,000`, simulating one 0-fee tx costs at most 500k gas of CPU — a fraction of a block.

**CheckTx vs DeliverTx divergence:** State can change between CheckTx simulation and DeliverTx (other txs committed in between). If a 0-fee tx passes CheckTx but fails in DeliverTx (e.g., realm balance drained by another tx), it is a **valid failed transaction** — no consensus violation. This is the same divergence that exists for any tx today (e.g., user balance changes between CheckTx and DeliverTx).

**Implementation:** Modify CheckTx to run full simulation for 0-fee txs. See Section 8.8.

## 6. Realm Usage Examples

### 6.1 Basic Gas Sponsorship

```go
package myapp

import "chain/runtime"

func DoSomething() {
    // Realm pays gas for all callers, up to 500 ugnot
    runtime.PayGas(500)

    // ... realm logic ...
}
```

### 6.2 Conditional Sponsorship (Collect USDC)

The user sends a single transaction with two messages — approve and swap — using multi-`MsgCall`:

```
Tx {
    Msgs: [
        MsgCall{ Caller: user, PkgPath: "gno.land/r/demo/usdc",  Func: "Approve",
                 Args: [swapRealmAddr, amount] },
        MsgCall{ Caller: user, PkgPath: "gno.land/r/demo/swap",  Func: "Swap",
                 Args: [amount] },
    ],
    Fee: { GasWanted: 0, GasFee: 0 },
}
```

The swap realm collects USDC and pays gas:

```go
package swap

import (
    "std"
    "gno.land/r/demo/usdc"
)

func Swap(usdcAmount int64) {
    caller := std.GetOrigCaller()

    // Collect USDC from user (approved in the first MsgCall)
    usdc.TransferFrom(caller, std.CurrentRealm().Addr(), usdcAmount)

    // Now realm pays gas from its gnot balance, up to 1000 ugnot
    runtime.PayGas(1_000)

    // ... perform swap logic ...
}
```

**Execution flow:**
1. `MsgCall[0]`: `usdc.Approve()` — runs on credit gas, no `PayGas` called
2. `MsgCall[1]`: `swap.Swap()` — collects USDC, then calls `PayGas(1_000)` → realm takes over gas payment (up to 1000 ugnot)
3. Settlement: total gas consumed across both messages is deducted from the swap realm's gnot balance

### 6.3 Whitelist-Based Sponsorship

```go
package premium

import "chain/runtime"

var whitelist = map[std.Address]bool{}

func Register(addr std.Address) { /* admin only */ }

func Action() {
    caller := std.GetOrigCaller()
    if !whitelist[caller] {
        panic("not whitelisted")
    }

    runtime.PayGas(300)

    // ... do work for whitelisted user ...
}
```

## 7. Security Analysis

### 7.1 DoS Vectors and Mitigations

| Attack | Mitigation |
|--------|------------|
| **Mempool spam with 0-fee txs** | Validator opt-in + CheckTx simulation filters invalid txs before mempool entry. Rate limiting can be added later. |
| **Force nodes to simulate** | Only opt-in nodes simulate at CheckTx. Non-opt-in validators reject 0-fee txs immediately (zero cost). |
| **Drain realm's gnot balance** | Realm author's responsibility — use whitelists, rate limits, or collect payment before `PayGas` |
| **Exhaust credit window for free** | Credit gas is bounded by `MaxGasCreditPerTx`. Counts toward block gas limit. CheckTx simulation filters txs that don't call `PayGas`. |
| **Call PayGas from sub-realm** | Allowed — the calling realm explicitly consents by calling `PayGas`. Realm authors must validate before calling. |
| **Unconditional PayGas in public function** | Realm author bug — any caller can drain the realm's gnot. Realm authors must guard `PayGas` with validation logic. |
| **Call PayGas multiple times** | Only first call is valid; subsequent calls panic |
| **Trigger failure after PayGas** | Realm does NOT pay on failure — all state reverts. Free execution is bounded by CheckTx simulation (only valid txs enter mempool). |

### 7.2 Validator Cost Analysis

```
Validator opts out:   GasFee == 0 → reject immediately (zero cost)
Validator opts in:    CheckTx simulation up to MaxGasCreditPerTx gas (bounded)
DeliverTx:            same as any tx (realm pays on success)
```

Validators who don't opt in bear zero additional cost. Opt-in validators accept the simulation cost at CheckTx time, bounded by the credit window.

## 8. Implementation Touch Points

### 8.1 Consensus Params

- **File**: `tm2/pkg/bft/types/params.go`
- **Change**: Add `MaxGasCreditPerTx int64` and `MinGasPrice Coin` to `GasParams`

### 8.2 Native Function

- **File**: `gnovm/stdlibs/std/native.go` (or equivalent)
- **Change**: Register `runtime.PayGas(maxFee int64)` native function
- **Behavior at call time**:
  1. Verify caller is a realm (not a package or `main`)
  2. Verify `PayGas` not already called in this tx (check `PayGasInfo` on context)
  3. Check `realm.balance >= maxFee`
  4. Derive gas limit: `maxFee / MinGasPrice`
  5. Verify derived limit >= gas already consumed (panic otherwise)
  6. Update gas meter limit to derived value
  7. Set `PayGasInfo{RealmAddr, MaxFee}` on execution context
  - **Note**: Realm balance is NOT debited at call time — final `gasUsed` is unknown. Debit happens at settlement (Section 8.5) after execution completes.

### 8.3 Ante Handler

- **File**: `tm2/pkg/sdk/auth/ante.go`
- **Change**: When `GasFee == 0` and `MaxGasCreditPerTx > 0`:
  1. Skip fee deduction (no fees to deduct)
  2. Set gas meter with `limit = MaxGasCreditPerTx` (credit window)
  3. Continue to message execution (do NOT skip like normal CheckTx)

### 8.4 Gas Meter

- **File**: `tm2/pkg/store/types/gas.go`
- **Change**: Allow gas meter limit to be updated after creation (for when `PayGas` raises the limit from credit window to derived limit). Add a `SetLimit(newLimit int64)` method or similar.

### 8.5 Settlement

- **File**: `gno.land/pkg/sdk/vm/keeper.go`
- **Change**: At the **end** of message execution (not at PayGas call time), if `PayGasInfo` is set on the context, deduct `gasUsed × MinGasPrice` from the realm's balance. This runs inside the cached execution context — on success it commits with all state, on failure it reverts with everything. See Section 5.5.

### 8.6 Validator Config

- **File**: `tm2/pkg/bft/config/config.go`
- **Change**: Add `AllowZeroFeeTxs bool` to mempool config (default: `false`)

### 8.7 CheckTx for 0-Fee Txs

- **File**: `tm2/pkg/sdk/baseapp.go` (in `CheckTx` / `runTx`)
- **Change**: When `GasFee == 0` and `AllowZeroFeeTxs == true`:
  1. Switch mode from `RunTxModeCheck` to `RunTxModeSimulate` so message execution is NOT skipped
  2. Run full tx (ante handler + messages) with credit gas meter
  3. After execution, check if `PayGasInfo` is set on context
  4. If set and realm has funds → accept into mempool
  5. If not set → reject (tx would fail in DeliverTx anyway)

## 9. Alternatives Considered

| Alternative | Why not |
|-------------|---------|
| **First-signer pays (existing in Gno)** | Requires off-chain co-signing infrastructure. Sponsor must be online. No on-chain conditional logic. Centralization risk. See [Section 2](#2-existing-mechanism-first-signer-pays). |
| **Cosmos feegrant module** | No conditional logic — can't collect USDC before sponsoring. Requires pre-registration per grantee. |
| **Realm pre-registration (`RegisterGasSponsor`)** | Can't run arbitrary logic before committing to pay. Doesn't support the "collect payment then sponsor" use case. |
| **Off-chain relayer (Ethereum EIP-2771 style)** | Requires external infrastructure. Centralization risk. Adds latency. |
| **Post-execution refund** | User still needs gnot upfront. Not truly gasless. |

## 10. Design Decisions

The following questions were raised during design and have been resolved:

1. **`PayGas` covers all gas consumed in the tx** — including gas consumed before `PayGas` was called (credit phase). This simplifies the model: the realm pays for the entire tx, not a partial slice. The derived gas limit (`maxFee / MinGasPrice`) is the total cap.

2. **`gasUsed > derived gas limit`** — the tx fails with out-of-gas when the derived limit is reached, same as normal txs.

3. **Derived gas limit < gas already consumed** — `PayGas` panics immediately. The realm's budget at current gas price is already exceeded.

4. **`PayGas` takes `maxFee` (ugnot), not `limitGas` (gas units)** — realm authors think in cost, not gas units. The gas limit is derived from `maxFee / MinGasPrice`. This protects the realm's budget against gas price governance changes: if `MinGasPrice` doubles, the realm gets less computation but never overspends.

5. **Multi-message txs** — supported. The first realm to call `PayGas` across all messages becomes the sponsor. All messages are atomic (all-or-nothing).

6. **Gas price for 0-fee txs** — uses consensus `MinGasPrice` parameter. See Section 5.5.

7. **CheckTx for 0-fee txs** — runs full simulation (not just signature check). Only accepts into mempool if `PayGas` is called and realm has funds. See Section 5.6.

## 11. Open Questions

1. **Should there be a minimum gnot balance for realms to be eligible?** Could prevent edge cases where realm balance drops between `PayGas` call and settlement. Counter-argument: `PayGas` already checks balance at call time, and settlement happens in the same block.

2. **Mempool rate limiting for 0-fee txs**: Per-sender and per-block caps (e.g., `max_zero_fee_txs_per_block`, `max_zero_fee_txs_per_sender`) can be added later if spam becomes a problem. Not needed for initial implementation.

## 12. Future Direction: Proposer Fee Share

In the initial design, validators who opt-in to 0-fee txs absorb the simulation cost (credit window gas) as a public good. This works but doesn't create a strong economic incentive to opt-in.

### The Idea

Proposers who include successful `PayGas` txs in a block earn a **bonus fee** — a percentage of the realm's gas payment:

```
Realm pays 1000 ugnot via PayGas
  → 900 ugnot → fee collector (distributed to all validators normally)
  → 100 ugnot → proposer bonus (reward for processing 0-fee txs)
```

### Why This Matters

| Without fee share | With fee share |
|-------------------|----------------|
| Opt-in is altruistic — validator absorbs simulation cost for no extra reward | Opt-in is profitable — validator earns bonus for every successful 0-fee tx |
| Few validators opt-in → gasless txs have high latency | More validators opt-in → gasless txs included faster |
| Larger credit window = more risk for validators | Larger credit window = more opportunity for validators |

### Design Sketch

A new consensus parameter:

```go
type GasParams struct {
    MaxGas               int64
    MaxGasCreditPerTx    int64
    MinGasPrice          Coin
    ProposerFeeSharePct  int64 // NEW — e.g., 10 (= 10%)
}
```

At settlement:
1. `actualCost = gasUsed × MinGasPrice`
2. `proposerBonus = actualCost × ProposerFeeSharePct / 100`
3. `feeCollectorAmount = actualCost - proposerBonus`
4. Send `proposerBonus` to proposer address
5. Send `feeCollectorAmount` to fee collector

### Why Not Now

- The base `PayGas` mechanism should be proven first
- Fee share adds complexity to settlement
- The incentive model needs economic analysis (what % makes opt-in rational?)
- Can be added via governance without changing the `PayGas` API

This is a **backward-compatible enhancement** — realms don't need to change their code. Only the settlement logic and a new consensus param are needed.

## 13. Future Direction: Account Abstraction Roadmap

PayGas is pillar 1 of Gno's account abstraction strategy. It solves gas — the most visible friction point. Two complementary pillars can follow, each independent but amplified when combined with PayGas.

### Pillar 1: Gas Abstraction (This HLD)

**Status:** This proposal.

Users transact with 0 gnot. Realms pay gas, optionally collecting payment in other tokens. No off-chain infrastructure needed.

### Pillar 2: Session Keys (Realm-Level)

**Status:** Future — enabled by PayGas, no additional protocol changes needed.

Session keys are temporary, scoped keys that sign transactions on the user's behalf. The user authorizes once, then the app acts without wallet popups — critical for games, trading, and any interactive dApp.

**How other chains do it:**

| Chain | Approach | Infra needed |
|-------|----------|-------------|
| NEAR | Protocol-native FunctionCall access keys (one contract, method whitelist, gas allowance) | None |
| Ethereum | Smart contract wallet validator modules (per-function, per-argument policies) | Bundler + Paymaster + EntryPoint + smart wallet |
| Starknet | Native AA account contracts with session validator + admin blocklist | None (all accounts are contracts) |
| Sui | zkLogin ephemeral keys (OAuth → ZK proof → ephemeral key, no permission scoping) | ZK prover + salt service |

**Gno's approach:** Session keys managed at the **realm level**, similar to NEAR's simplicity but with PayGas enabling gasless sessions:

```go
// Realm manages session keys — no protocol changes needed
var sessions = map[std.Address]Session{}

type Session struct {
    UserAddr  std.Address
    Expiry    int64
    MaxCalls  int
    CallsUsed int
}

// User calls this once with their main key
func RegisterSession(sessionKey std.Address, expiry int64, maxCalls int) {
    caller := std.GetOrigCaller()
    sessions[sessionKey] = Session{caller, expiry, maxCalls, 0}
}

// Session key holder calls this — has 0 gnot
func GameAction(action string) {
    sess := sessions[std.GetOrigCaller()]
    if sess.Expiry < std.GetTimestamp() { panic("session expired") }
    if sess.CallsUsed >= sess.MaxCalls { panic("call limit reached") }
    sessions[std.GetOrigCaller()] = Session{
        sess.UserAddr, sess.Expiry, sess.MaxCalls, sess.CallsUsed + 1,
    }

    runtime.PayGas(200) // realm pays gas for session key holder

    // ... execute action as sess.UserAddr ...
}
```

**Why this works without protocol changes:** The realm enforces scoping (expiry, call limits, function restrictions). PayGas removes the gas requirement. The session key is a regular Gno account — no smart wallets, no new key types. The trade-off vs Ethereum: less granular (no per-argument policies in the protocol), but far simpler.

**What the user experiences:**
```
1. Sign once to create a session (main key)
2. Play the game / trade / interact (session key, no popups, no gas)
3. Session expires automatically
```

### Pillar 3: Flexible Authentication (Protocol-Level)

**Status:** Future — requires protocol changes, independent of PayGas.

Today the ante handler only accepts ed25519/secp256k1 signatures. To support social login (Google, Apple), passkeys (WebAuthn/P-256), or realm-delegated authentication, the signature verification layer needs extension.

**Possible approaches:**
- Add new key types to the ante handler (P-256 for passkeys, RS256 for OAuth JWTs)
- Allow realms to register custom signature verifiers
- Realm-delegated auth: a realm vouches for the caller, ante handler accepts it

**This is the most complex pillar** — it changes the security boundary (ante handler) and the account model. Each approach has significant security implications that need their own HLD.

**Combined with PayGas + session keys:** A user authenticates via Face ID (pillar 3), gets a session key (pillar 2), and transacts without gas (pillar 1). Full account abstraction — no wallet, no gnot, no popups.

### How the Pillars Combine

```
┌─────────────────────────────────────────────────────────┐
│                   User Experience                        │
│  "Sign in with Google → use the app → never see gas"    │
└────────┬──────────────────┬──────────────────┬──────────┘
         │                  │                  │
    ┌────▼────┐       ┌─────▼─────┐      ┌────▼─────┐
    │ Pillar 1│       │ Pillar 2  │      │ Pillar 3 │
    │ PayGas  │       │ Session   │      │ Flexible │
    │         │       │ Keys      │      │ Auth     │
    │ 0 gnot  │       │ No popups │      │ No wallet│
    │ needed  │       │ per action│      │ setup    │
    └─────────┘       └───────────┘      └──────────┘
     This HLD          Realm-level        Protocol
                       No protocol        change
                       changes            needed
```

Each pillar delivers value independently:
- PayGas alone: gasless UX (this HLD)
- Session keys alone: no wallet popups (but user needs gnot)
- Flexible auth alone: social login (but user needs gnot + popups)

Combined, they deliver full account abstraction. PayGas is the foundation because it removes the most universal friction (gas), enables session keys to work without funding, and is the simplest to build.

## 14. Comparison with Other Chains

| Chain | Approach | Free execution? | On-chain conditions? | Infra needed? |
|-------|----------|----------------|---------------------|---------------|
| **Gno (this proposal)** | Credit window + `PayGas()` | Bounded (credit window) | Yes (arbitrary realm logic) | None |
| Ethereum EIP-4337 | Paymaster pre-deposit | No | Yes (Paymaster contract) | Bundler + alt mempool |
| Solana | Native fee payer field | No | No (off-chain only) | Co-signing backend |
| NEAR | Access keys + NEP-366 | No (prepaid allowance) | Limited (method scoping) | Relayer for meta-txs |
| Cosmos SDK | feegrant module | No | Limited (msg type filter) | None |
| Sui | Sponsored tx (dual-sig) | No | No (off-chain only) | Gas pool service |

**Gno's differentiator**: Ethereum EIP-4337 achieves similar on-chain conditional sponsorship via Paymasters, but required a complex solution — bundlers, alt mempool, EntryPoint contract, and smart contract wallets — because:

- "Someone pays before any opcode" is deeply embedded in the EVM's gas model — changing it risks breaking billions of dollars of deployed contracts
- EIP-4337 was designed to avoid consensus changes, working within existing constraints
- While EIP-7702 (Pectra, 2025) begins closing the EOA gap, the gas model constraint remains

Gno takes a protocol-level approach: `PayGas` is a native function supported by consensus param changes, gas meter modifications, proposer simulation, and settlement logic. Simpler than EIP-4337's full stack, but not trivial. The advantage is that Gno is designing the VM, gas model, and consensus layer together — no backward compatibility constraints with existing contracts. The result: regular accounts submit standard transactions, no smart wallets or bundlers required.
