# Plan: Record account number and sequence in genesis tx metadata

## Context

PRs #5489 and #5411 enable hard fork genesis replay — historical txs are
re-executed on a new chain. Signatures cover `(chain_id, account_number,
sequence)`. Chain ID is handled by `PastChainIDs`. Account numbers and
sequences need to be preserved so signatures verify correctly during replay.

Account numbers and sequences are NOT accessible from Gno contracts — they
are purely internal to the auth module for signature verification.

## Obtaining signer info

### Account numbers — via RPC

Query `ABCIQuery("auth/accounts/{addr}")` per unique sender address,
batched via the RPC client's batch API. Account numbers are stable (never
change once assigned), so the latest state is valid for all historical txs.

The RPC client already supports batching (see `rpcclient.NewBatch()`).
With ~50k unique addresses on gnoland1, batched queries complete in seconds.

Direct DB access via `source_dir.go` is a future enhancement — currently
the dir source only reads from pre-exported `txs.jsonl` files and has no
DB or block store access.

### Sequences — single-pass with buffered brute-force

Genesis-mode txs DO increment sequences (they go through the ante
handler). An account that sent N genesis txs starts historical blocks
at sequence N, not 0.

Each sender's counter is initialized on their first successful historical
tx via brute-force: verify the signature with sequences in `[0, finalSeq]`
where `finalSeq` comes from the RPC account query at halt_height. This is
the absolute upper bound and handles any combination of genesis txs and
preceding failures. For most senders (no genesis txs, no prior failures),
it resolves to 0 immediately. For the admin account (e.g. 78 genesis txs),
it resolves to 78. At most `finalSeq` signature verifications per sender —
trivial for gnoland1's low tx counts.

After initialization, process subsequent txs in block order:

- **Successful tx**: if no pending failed txs from this sender, assign
  current counter as pre-tx sequence, increment counter.
- **Failed tx**: buffer it. Don't assign sequence yet.
- **Successful tx after failed txs from same sender**: brute-force time.
  The buffered failed txs plus this successful tx form a gap. Try verifying
  this successful tx's signature with each possible sequence in
  `[counter, counter + num_buffered_failures]`. The matching sequence is
  the pre-tx sequence for this successful tx. Work backwards to assign
  sequences to each buffered failed tx (ante-fail = no increment,
  msg-fail = increment).

Signature verification is offline — `GetSignBytes` needs chain_id,
account_number, sequence, fee, msgs, memo — all from tx bytes + metadata.

Degenerate case: sender's last tx is a failure with no subsequent success.
Query the account's sequence at halt_height via
`ABCIQueryWithOptions("auth/accounts/{addr}", height=haltHeight)` to avoid
racing with ongoing block production. The difference between the queried
sequence and the last known counter tells us how many trailing failures
consumed sequences.

### Sequence semantics: pre-tx

Store the **pre-tx** sequence in `SignerInfo.Sequence` — i.e., the account
sequence BEFORE this tx is processed. This is the value used in
`GetSignBytes` for signature verification.

- Successful tx: force-set pre-tx sequence, Deliver runs, ante handler
  verifies signature with that sequence, increments naturally.
- Failed tx: force-set pre-tx sequence, skip Deliver. If the failure was
  ante-fail (no sequence consumed), the next tx has the same pre-tx
  sequence. If msg-fail (sequence consumed), the next tx has pre-tx + 1.

## Data structure changes

### GnoTxMetadata — add fields

File: `gno.land/pkg/gnoland/types.go`

```go
type GnoTxMetadata struct {
    Timestamp   int64               `json:"timestamp"`
    BlockHeight int64               `json:"block_height,omitempty"`
    ChainID     string              `json:"chain_id,omitempty"`
    Failed      bool                `json:"failed,omitempty"`
    SignerInfo  []SignerAccountInfo  `json:"signer_info,omitempty"`
}

type SignerAccountInfo struct {
    Address    crypto.Address `json:"address"`
    AccountNum uint64         `json:"account_num"`
    Sequence   uint64         `json:"sequence"`  // pre-tx sequence
}
```

- `SignerInfo` is a slice — a tx can have multiple msgs with different signers.
- `Failed` marks txs with non-zero return code on the original chain. Not
  re-executed during replay, but included so sequence impact is tracked.
- `omitempty` for backward compatibility.
- Needs amino registration in `gno.land/pkg/gnoland/package.go`.

## Genesis replay loop changes

File: `gno.land/pkg/gnoland/app.go`

```go
// For historical txs with signer metadata, force-set account state.
// Uses pre-tx sequence — the value the signature was signed with.
if metadata != nil && metadata.BlockHeight > 0 && len(metadata.SignerInfo) > 0 {
    for _, si := range metadata.SignerInfo {
        acc := cfg.acck.GetAccount(ctx, si.Address)
        if acc == nil {
            acc = cfg.acck.NewAccountWithNumber(ctx, si.Address, si.AccountNum)
        } else {
            acc.SetAccountNumber(si.AccountNum)
        }
        acc.SetSequence(si.Sequence)
        cfg.acck.SetAccount(ctx, acc)
    }
}

// Failed txs: pre-tx sequence already set. For msg-fail txs (ante passed,
// sequence was consumed on original chain), we need to increment.
// The export tool already determined this — if the NEXT tx from this sender
// has a higher pre-tx sequence, the increment happened. So just skip Deliver
// and let the next tx's force-set handle it.
if metadata != nil && metadata.Failed {
    continue // skip Deliver
}

res := cfg.baseApp.Deliver(stdTx, ctxFn)
```

For successful txs: force-set pre-tx sequence → Deliver → ante handler
reads sequence, builds sign bytes, verifies signature, increments to
pre-tx + 1 → SetAccount persists. The next tx's force-set then overrides
to whatever the next pre-tx sequence should be.

For failed txs: force-set pre-tx sequence → skip Deliver → next tx's
force-set overrides to the correct next pre-tx sequence (same if ante-fail,
+1 if msg-fail). The export tool already encoded this in the metadata.

### AccountKeeper changes

File: `tm2/pkg/sdk/auth/keeper.go`

```go
// NewAccountWithNumber creates an account with a specific account number,
// bypassing the auto-increment counter. Updates the global counter if
// the given number would cause future collisions.
func (ak AccountKeeper) NewAccountWithNumber(
    ctx sdk.Context, addr crypto.Address, accNum uint64,
) std.Account {
    acc := ak.proto()
    acc.SetAddress(addr)
    acc.SetAccountNumber(accNum)

    // Read global counter directly (don't call GetNextAccountNumber —
    // it has side effects: reads AND increments).
    stor := ctx.GasStore(ak.key)
    bz := stor.Get([]byte(GlobalAccountNumberKey))
    var currentNum uint64
    if len(bz) > 0 {
        amino.MustUnmarshal(bz, &currentNum)
    }

    // Update counter if our number would cause collisions.
    if accNum >= currentNum {
        bz = amino.MustMarshal(accNum + 1)
        stor.Set([]byte(GlobalAccountNumberKey), bz)
    }

    return acc
}
```

- Uses `ctx.GasStore` (not `ctx.Store`) to match existing keeper pattern.
- Uses amino marshal/unmarshal (not binary.BigEndian) to match existing
  encoding of the global counter.
- Does NOT call `GetNextAccountNumber` (it has side effects).

## Export tool changes

Changes are in `source_rpc.go` only. The dir source (`source_dir.go`) is
a stub that reads pre-exported files — it will inherit these fields when
fed a txs.jsonl generated by the RPC source.

### Including failed txs

Include ALL txs from blocks. Mark failed ones with `Failed: true`.
Don't filter by `IsErr()`.

### Single-pass buffered sequence tracking

Each signer is tracked independently. A tx with multiple signers updates
each signer's state separately. Brute-force is per-signer (linear, not
combinatorial — each signer's sequence is independent of others).

```go
type signerState struct {
    accNum       uint64
    finalSeq     uint64                // from RPC query, absolute upper bound
    seq          uint64                // current pre-tx sequence counter
    initialized  bool                  // true after first brute-force resolves starting seq
    pendingFails []pendingFailedTx     // buffered failed txs for this signer
}

signerStates := map[crypto.Address]*signerState{}

// Query account info on first encounter (lazy). Cache in signerStates.
// Could also be batched in a pre-pass over all txs to collect unique
// signers first — either approach works.
func getOrCreateSignerState(addr crypto.Address) *signerState {
    if ss, ok := signerStates[addr]; ok {
        return ss
    }
    acc := queryAccountAtHeight(client, addr, haltHeight)
    ss := &signerState{}
    if acc != nil {
        ss.accNum = acc.AccountNumber
        ss.finalSeq = acc.Sequence
    }
    // If acc is nil, account was never created on-chain (all txs were
    // ante-fails). Defaults: accNum=0, finalSeq=0.
    signerStates[addr] = ss
    return ss
}

// For each tx in block order:
signers := stdTx.GetSigners()
sigs := stdTx.GetSignatures()

if successful {
    for i, signer := range signers {
        ss := getOrCreateSignerState(signer)

        if !ss.initialized || len(ss.pendingFails) > 0 {
            // Brute-force to find this tx's pre-tx sequence.
            // First time: range is [0, finalSeq] to account for genesis txs.
            // Subsequent: range is [ss.seq, ss.seq + len(pendingFails)].
            lo := ss.seq
            hi := ss.seq + uint64(len(ss.pendingFails))
            if !ss.initialized {
                lo = 0
                hi = ss.finalSeq
            }

            resolvedSeq := bruteForceSignerSequence(
                stdTx, i, sigs[i], ss.accNum, lo, hi, chainID)

            // Assign sequences to buffered failed txs (cosmetic/audit-only).
            if len(ss.pendingFails) > 0 {
                assignFailedTxSequences(ss.pendingFails, ss.seq, resolvedSeq)
                ss.pendingFails = nil
            }
            ss.seq = resolvedSeq
            ss.initialized = true
        }

        signerInfos[i].Sequence = ss.seq
        ss.seq++
    }
} else {
    // Buffer this tx for each signer independently
    for _, signer := range signers {
        ss := getOrCreateSignerState(signer)
        ss.pendingFails = append(ss.pendingFails, ...)
    }
}

// After all blocks: resolve trailing failures.
for signer, ss := range signerStates {
    if len(ss.pendingFails) == 0 {
        continue
    }

    if !ss.initialized {
        // Never had a successful tx. ss.seq is still 0.
        // finalSeq includes genesis txs + any historical failures that
        // consumed sequences. Infer the genesis offset:
        //   startingSeq = finalSeq - (number of pending failures that consumed)
        // Since we can't know which failures consumed without an anchor,
        // cap consumed at len(pendingFails). The genesis offset is the rest.
        consumed := ss.finalSeq - ss.seq // includes genesis
        if consumed > uint64(len(ss.pendingFails)) {
            // Genesis txs account for the excess. Set starting seq to
            // skip past them.
            ss.seq = ss.finalSeq - uint64(len(ss.pendingFails))
            consumed = uint64(len(ss.pendingFails))
        }
        // Assign cosmetically: at most len(pendingFails) consumed.
        // Individual assignments are audit-only (failed txs are skipped).
        assignTrailingFailedTxSequences(ss.pendingFails, ss.seq, consumed)
    } else {
        // Initialized: ss.seq is correct. consumed = seqs since last known.
        consumed := ss.finalSeq - ss.seq
        assignTrailingFailedTxSequences(ss.pendingFails, ss.seq, consumed)
    }
}
```

### Account numbers

`ABCIQuery("auth/accounts/{addr}")` per unique sender, batched.
Cache results — each address is queried only once.

## Optional flag: SkipHistoricalSigVerification

A `SkipHistoricalSigVerification` flag may be supported for faster replay
during development/testing. When set, signature verification is skipped
for historical txs (BlockHeight > 0), eliminating the need for correct
SignerInfo. The metadata is still collected and exported regardless.

This flag is NOT recommended for production use — signature verification
provides security guarantees that the replayed txs are authentic and
unmodified from the original chain.

## Out of scope (separate PRs)

- VFS optimization (fast fsync on macOS).
- `getBeginBlockLastCommitInfo` panic at InitialHeight — needs fix for
  missing validator sets.
- Lenient `GenesisTxResultHandler` for replay-with-modifications.

## Verification

- Tests:
  1. Successful tx replay with force-set pre-tx sequence.
  2. Account creation for missing accounts with correct number.
  3. Failed ante-tx: sequence unchanged for next tx.
  4. Failed msg-tx: sequence incremented for next tx.
  5. Brute-force recovery with mixed ante/msg failures.
  6. Global account counter updated by NewAccountWithNumber.
  7. Degenerate case: trailing failures resolved via RPC at halt_height.
  8. Multi-signer tx.
  9. Account that only appears in failed txs (never succeeds).
  10. Sender whose first tx is a failure.
  11. Sequence 0 edge case (brand-new account's first tx).
- Run: `go test ./gno.land/pkg/gnoland/ -run TestGenesis -v`
- Run: `go test ./gno.land/pkg/integration/ -run txtar`
