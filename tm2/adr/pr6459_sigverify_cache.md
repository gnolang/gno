# Verified-signature cache for CheckTx and mempool recheck

## Context

With recheck on (the default), every pending transaction goes through the
ante handler once on admission and again after every committed block. The
ante handler verifies every signature each time. Measured on a gno.land app
with the real `CListMempool`: recheck costs about 0.9 ms per pending
transaction per block, linear in depth, and signature verification is about
31% of it. At the default mempool size of 5000 that is several seconds of
recheck per block, almost all of it repeating verifications that already
succeeded.

## Decision

Add `auth.SigVerifyCache`, set through `AnteOptions.SigVerifyCache`, and wire
one per app in gno.land (`sigVerifyCacheSize = 10_000`).

- **Key.** sha256 over the length-prefixed public key bytes, the exact sign
  bytes and the signature: every input of `VerifyBytes`. The sign bytes carry
  chain id, account number, the current stored sequence, fee, msgs and memo,
  so a hit means "this exact verification already returned true".
- **CheckTx only.** Consulted and filled only when the context mode is
  `RunTxModeCheck`. DeliverTx and simulate always verify, so the cache cannot
  change what a block commits.
- **Gas.** The lookup happens after `sigGasConsumer` has charged for the
  verification, so a hit costs the same gas as a miss.
- **Legacy rendering.** A signature that only verifies over the legacy sign
  bytes is accepted as before and never cached.
- **Bound.** Fixed capacity, FIFO eviction. An evicted entry costs one full
  verification on its next check, which is today's behaviour.

## Alternatives considered

- Key on the signature bytes alone. Measured no faster, and unsafe: in this
  ante handler the account number, sequence and chain id are enforced only
  through the signature, so a signature-only key would let a committed tx be
  replayed once it leaves the mempool dedup cache, a seen signature be moved
  onto another body, and a tx from another chain through CheckTx.
- Use the cache in DeliverTx too. Saves about 0.24 ms per delivered tx, but
  makes a consensus path depend on node-local state. Not worth it.
- LRU instead of FIFO. Recheck walks the mempool in order, so with a working
  set larger than the cache both degrade the same way; FIFO needs no list.
- `Recheck: false` or a sequence-only recheck. Larger design change, out of
  scope here.

## Consequences

- About 1 MB of node memory at the default size.
- A mempool configured larger than the cache (about 10,000) thrashes the
  cache and falls back to full verification, never to a wrong answer.
- The cache is per process and starts empty after a restart.
