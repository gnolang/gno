# ADR: Realm Transaction Sponsorship

## Status

Proposed

## Context

Every transaction on Gno requires the signer to hold gnot for gas fees and storage deposits. This creates onboarding friction — users must acquire gnot before their first interaction. Existing workarounds (first-signer co-signing) require off-chain infrastructure.

Other chains solve this differently: Ethereum uses EIP-4337 (Paymasters + bundlers + smart wallets), Solana has native fee payers, NEAR has access keys with gas allowances, Cosmos SDK has the feegrant module. Each requires either off-chain infrastructure or lacks on-chain conditional logic.

## Decision

Two natives in `chain/runtime` let a realm sponsor a 0-fee transaction:

- **`runtime.PayGas(maxFee int64, rlm realm)`** — `rlm` pays the transaction's gas, up to `maxFee` ugnot.
- **`runtime.PayStorage(maxDeposit int64, rlm realm)`** — `rlm` pays the storage deposits for its own storage, up to `maxDeposit` ugnot.

Invalid arguments always panic. A valid call only takes effect in a 0-fee transaction while the chain's credit window is open (see the tm2 ADR), and is a no-op otherwise: in fee-paying transactions, at genesis, and outside a transaction (queries, `gno test`). A 0-fee transaction that never calls `PayGas` is rejected. The realm is charged only if the transaction succeeds.

## Key Design Decisions

### 1. Mid-execution sponsorship

The realm decides **during execution** whether to sponsor, so it can run arbitrary logic first: collect payment in another token, check a whitelist, rate-limit. Pre-registration or feegrant-style allowances cannot run that logic.

### 2. The realm is a capability argument

`rlm` must be the current realm, checked with `rlm.IsCurrent()`, the same capability `banker.NewBanker(bt, rlm)` requires to spend a realm's coins. A realm passes the `cur` of its crossing function, or threads it into a helper. The native also refuses anything that is not a top-level `/r/` realm: packages, MsgRun's ephemeral realm, and sub-realm tokens (`cur.Sub(...)`).

`rlm` is the last parameter because a function whose first parameter is a `realm` is a crossing function.

**Why not inspect the call stack?** An earlier version required the function calling `PayGas` to be declared in the current realm. That hard-codes the VM's frame layout, and it lets a confused deputy through: if a realm exports a plain helper that calls `PayGas` and also runs a caller-supplied callback, an attacker passes the helper as the callback and the realm pays. A helper now needs the realm's own `cur`.

### 3. Settle only on success

Settlement runs in gno.land's `EndTxHook`, only when the transaction succeeded: the gas debit is `ceil(gasUsed × LastGasPrice)`, capped at `maxFee`, sent from the realm to the fee collector. On failure everything reverts and the realm pays nothing.

**Why not charge on failure, like a normal fee?** A realm that collects a token and then calls `PayGas` would be drained: an attacker makes the transaction fail afterwards, the token transfer reverts, and the gnot is still taken. The realm cannot defend itself, since its own bookkeeping reverts too. The cost of this choice, failing sponsored transactions that nobody pays for, is covered in the tm2 ADR.

### 4. PayStorage covers the sponsor's own storage, in its own messages

The sponsor pays only for growth of its own storage, and only in a message whose entry is the sponsor itself: a `MsgCall` to it, or its `init` during `MsgAddPackage`. Growth anywhere else stays on the caller. The budget is per transaction.

Both limits are needed because a storage deposit is refunded to whoever later frees the bytes, and storage is charged to the realm that allocated it:

- Paying for every realm the transaction touches let another message grow a realm the attacker controls, then free it in a second sponsored transaction and collect the deposit. Measured: the signer gained 478,700ugnot across two 0-fee transactions.
- Paying for the sponsor's own storage in any message still let another realm keep an object the sponsor allocated (any constructor-style function), which lands in the sponsor's storage, then drop it and collect. Measured: +597,800ugnot.

What remains is whatever the sponsor's own messages run. Everything they add to the sponsor's storage is charged to it, whoever's code adds it, for example an object it hands to a hook or callback that another realm keeps. And if the sponsor lets callers free storage it sponsored, they receive that deposit. Within one transaction a free in the sponsor's storage is first netted against what it locked earlier in the transaction, whoever paid for the freed bytes. So a sponsored function must not run code it does not control, and `PayStorage`'s doc comment says so. Charging only the sponsor's own code would need the VM to track which realm's code produced each byte; it does not.

Storage in other realms is the user's to pay. A sponsor that wants to cover it can send the user gnot in the same transaction; the deposit is then the user's, which matches who receives the refund.

### 5. No balance pre-checks

`PayGas` and `PayStorage` only record the commitment. Settlement is authoritative: the bank checks every debit before writing, and CheckTx admission dry-runs settlement, so an insolvent sponsor never enters the mempool. Pre-checks also read the realm's whole balance with `GetCoins`, whose cost is set by whoever sends the realm new denoms.

### 6. Gas and storage may be sponsored by different realms

A shared gas paymaster may call `PayGas` while the app realm calls `PayStorage` for its own storage. Each realm is charged only its own commitment, from its own balance.

### 7. Gas price from the auth module

The gas limit (`maxFee` divided by the gas price, capped by the credit window) and the settled cost use the auth module's dynamic `LastGasPrice`, the price normal transactions are checked against. No new parameter.

### 8. No `Fee.SponsorStorage` flag

An earlier version had a tx flag that deferred storage settlement to the end of the transaction, so that `PayStorage` could cover messages before its call. It was removed: after #6173 the signed fee is rendered as `{amount, gas}`, so the flag was not covered by signatures and anyone relaying the tx could flip it, and signing it would make Ledger refuse the transaction. A commitment already covers the later messages of its transaction, so a sponsor's message just goes first.

## Alternatives Considered

| Alternative | Why not |
|-------------|---------|
| Cosmos feegrant module | No on-chain conditional logic. Can't collect a token before sponsoring. |
| Off-chain relayer (EIP-2771 style) | Requires external infrastructure. Centralization risk. |
| Realm pre-registration | Can't run arbitrary logic before committing to pay. |
| Post-execution refund | User still needs gnot upfront. Not truly gasless. |
| Single PayGas covering storage too | Gas is burned, storage deposits are refundable; they need separate payers and separate rules. |
| PayStorage covering any realm | Drainable through refunds (decision 4). Safe support needs per-sponsor deposit tracking in realm state. |
| Charging the sponsor on failure | Reintroduces the drain in decision 3. |

## Consequences

- A `PayGas` commitment covers the gas of every message in the transaction, including ones the sponsor did not call; size `maxFee` accordingly.
- Realm authors must gate sponsorship (whitelist, payment, rate limit). An unconditional `PayGas` in a public function pays for anyone.
- Measured cost: a GRC20 approve + transferFrom paymaster transaction uses about 6M gas (`sponsorship_usecase_test.go`), so the credit window must be at least that for the motivating use case.
- The two native gas entries are measured: `payGas` charges 1097 + 35448·len(pkgPath)/1024 gas and `payStorage` 1091 + 35576·len(pkgPath)/1024, about 1.9K for a 24-byte realm path and 10K at the 256-byte pkgpath limit. Both natives match the pkgpath regexp twice (`IsRealmPath`, then `IsGnoRunPath` inside `DerivePkgCryptoAddr`) before hashing, which costs 2.2x `chain.packageAddress` per byte; the placeholder rows copied from it charged about half at the length limit. The benches use the regexp's worst case, one-letter segments (`gno.land/r/a/a/...`). Measured on an Apple M1 Pro (data in `gnovm/cmd/calibrate/sponsorship_bench_m1pro_arm64.txt`); like the rest of the table, they need re-measuring on the reference hardware before deployment.
- Adding the natives changes the `chain/runtime` stdlib committed at genesis, so the genesis app hash changes, and every transaction that loads `chain/runtime` uses about 3.5K more gas, sponsored or not.
- A sponsor's functions cannot be batched: a second `PayGas` or `PayStorage` in the same transaction panics.

## References

- tm2 ADR: `tm2/adr/pr5382_zero_fee_tx_admission_and_settlement.md`
- Implementation PR: gnolang/gno#5382
