# PR5169: Block Backup & Restore

## Context

gno.land nodes had no mechanism for block-level backup or state restoration from
archived blocks. The existing `tx-archive` tool operates at the transaction level
only — it preserves transaction content but discards consensus state (block headers,
validator commits, block metadata). This makes it unsuitable for scenarios requiring
full chain reconstruction with validator set verification.

Operators needed a way to:
- Back up a running node's block history efficiently
- Bootstrap new nodes from archived blocks without syncing from peers
- Resume interrupted backups without re-downloading everything
- Verify block integrity during restoration (validator commit signatures)

Issue: #1827. Built on top of #3946 (Norman's initial WebSocket streaming work).
Alternative approach to #4950.

## Decision

Implement a three-layer block backup/restore pipeline:

### 1. WebSocket Streaming Endpoint

A new RPC method `BackupBlocks(startHeight, endHeight)` streams full blocks over
WebSocket. One `ResultBackupBlock` response per block, with a final `Done: true`
sentinel.

WebSocket was chosen over HTTP because:
- Blocks can be very large; streaming avoids buffering entire ranges in memory
- No need for pagination logic or Content-Length estimation
- Natural backpressure via the WebSocket write channel

The endpoint validates height ranges and returns errors for out-of-bounds requests.

### 2. Archive Format (tar + zstd)

Blocks are stored in chunked, compressed archives:

- **Format**: tar archive compressed with Zstandard (zstd)
- **Chunk size**: 100 blocks per file
- **Naming**: `{019d-zero-padded-height}.tm2blocks.tar.zst`
- **Block encoding**: Amino-marshaled `*types.Block` as individual tar entries
- **Metadata**: `info.json` tracking version, start/end heights

Design choices:
- **tar**: widely supported, efficient for bundling, streamable
- **zstd**: faster than xz/gzip at comparable compression ratios; decompression
  speed matters more than compression ratio for restore operations
- **100-block chunks**: balance between too many small files and too-large archives;
  enables granular resume points
- **Atomic writes**: blocks written to `next-chunk.tar.zst`, renamed to final name
  only after flush — prevents partial files on crash
- **Directory lock**: `flock`-based `blocks.lock` prevents concurrent writers
- **Writer poisoning**: after any write error, subsequent writes fail immediately
  to prevent silent data corruption

### 3. Restore Command

`gnoland restore --backup-dir <path> [--end-height N] [--skip-verification]`

Restore flow:
1. Initialize node (genesis, data dir, proxy app)
2. Determine start height from `blockStore.Height() + 1`
3. Open backup via `backup.WithReader()` (iterator pattern)
4. For each block pair (N, N+1):
   - Verify N+1's `LastCommit` against N's header using validator set (unless `--skip-verification`)
   - Call `blockExec.ApplyBlock()` — full ABCI execution (BeginBlock, DeliverTx, EndBlock)
   - Batch-save to blockStore every 1000 blocks

The two-block buffering is required because Tendermint2's commit model stores block
N's commit inside block N+1's `LastCommit` field.

### 4. Backup Tool

`tm2backup` is a standalone binary in `contribs/` that:
- Connects to a node's WebSocket RPC
- Sends a `backup` JSON-RPC request with height range
- Streams blocks into the archive writer
- Supports resuming via `info.json` state

## Key files

| File | Role |
|------|------|
| `tm2/pkg/bft/rpc/core/backup.go` | `BackupBlocks` WebSocket RPC endpoint |
| `tm2/pkg/bft/rpc/core/routes.go` | Route registration (WS-only) |
| `tm2/pkg/bft/rpc/core/types/responses.go` | `ResultBackupBlock` response type |
| `tm2/pkg/bft/backup/writer.go` | Archive writer (tar+zstd, chunking, atomic writes) |
| `tm2/pkg/bft/backup/reader.go` | Archive reader (iterator pattern) |
| `tm2/pkg/bft/backup/util.go` | Metadata (`info.json`), file naming, chunk alignment |
| `tm2/pkg/bft/blockchain/reactor.go` | `Restore()` — block verification and application loop |
| `tm2/pkg/bft/store/store.go` | `SaveBlockWithBatch()` for batched persistence |
| `tm2/pkg/bft/node/node.go` | `Node.Restore()` delegation to blockchain reactor |
| `gno.land/cmd/gnoland/restore.go` | CLI command, flag parsing, node initialization |
| `contribs/tm2backup/main.go` | Backup tool binary |

## Relationship to tx-archive

| Aspect | Block backup (this PR) | tx-archive |
|--------|----------------------|------------|
| **Granularity** | Full blocks (headers, commits, metadata, txs) | Transactions only |
| **Consensus state** | Preserved (validator sets, commits) | Lost |
| **Verification** | Validator commit signatures can be checked | No consensus verification |
| **Use case** | Full node bootstrap, disaster recovery | Transaction replay, migration |
| **Speed** | Faster restore (blocks pre-validated) | Slower (must re-propose/commit) |

They are complementary: block backup for infrastructure/ops, tx-archive for data migration.

## Known Limitations

1. **N+1 block requirement**: the last block in a backup cannot be committed during
   restore because block N's commit lives in block N+1's `LastCommit`. Restore
   effectively stops at `endHeight - 1`.

2. **WebSocket-only**: the streaming endpoint requires WebSocket; HTTP clients cannot
   use it directly.

3. **Full re-execution**: restore calls `ApplyBlock()` for every block, which
   re-executes all transactions through GnoVM. This is correct but slow for large
   chains. The bottleneck is GnoVM execution, not I/O.

## Consequences

- Operators can back up and restore gno.land nodes from block archives
- New nodes can be bootstrapped without peer syncing
- Interrupted backups/restores can be resumed efficiently
- Block integrity is cryptographically verified during restore (by default)
- The archive format is portable and uses standard compression (tar+zstd)
- The `tm2backup` tool is decoupled from the node binary, enabling independent versioning
