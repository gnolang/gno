# A local onyx-1 node and indexer for gnoscope

gnoscope.com reads onyx through two public endpoints that enforce a per-IP rate
limit. Measured on the production instance over 24 hours on 2026-11-23: 277
upstream 429s, every one of them on onyx, none on mainnet. Mainnet has a second
indexer and a second RPC configured as fallbacks; onyx has neither.

This directory is a **proven recipe** for removing onyx from that dependency by
running the node and the indexer on the deployment host, bound to loopback.
Nothing here has been deployed. The recipe was built and exercised on lyra on
2026-11-23; what was proven and what was not is in
[`claude/reports/d-20261123-onyxnode.md`](../../claude/reports/d-20261123-onyxnode.md).

**Read the report's recommendation before running any of this.** The recipe
works. Whether it is the right thing to run is a separate question, and the
report argues a specific answer with numbers.

## Files

| File | What it is |
|---|---|
| [BUILD.md](./BUILD.md) | the two binaries: exact commits, build commands, checksums |
| [genesis.json](./genesis.json) | onyx-1 genesis, sha256 in [genesis.json.sha256](./genesis.json.sha256) |
| [config.toml.fragment](./config.toml.fragment) | every config key that differs from the defaults, annotated |
| [gnoland-onyx.service](./gnoland-onyx.service) | node unit, loopback only |
| [tx-indexer-onyx.service](./tx-indexer-onyx.service) | indexer unit, loopback only |
| [INSTALL.md](./INSTALL.md) | the ordered procedure, snapshot restore and genesis replay |
| [RESET.md](./RESET.md) | what to do when the testnet resets or is retired, and how to notice |
| [ENDPOINTS.md](./ENDPOINTS.md) | what else serves onyx publicly, and where the 429 comes from |
| [PROOF.md](./PROOF.md) | transcripts of real gnoscope queries answering from this stack |
| [gnoscope-networks.onyx.json](./gnoscope-networks.onyx.json) | the networks.json fragment |

## The three facts most likely to bite

1. **A stale local node silently poisons gnoscope.** The pool adopts its chain
   identity from the earliest configured endpoint that can prove one, and then
   excludes anything that disagrees. Local first means a dead local node
   defines onyx and locks the real endpoints out. [RESET.md](./RESET.md) opens
   with this.
2. **`v1.8.0-rc.1` cannot replay onyx from genesis.** Two binaries are needed
   for a genesis replay. The snapshot path avoids this entirely.
3. **The rate limit is a Cloudflare rule, not a tx-indexer setting.** Raising
   it is a change in `gnolang/infrastructure`, not a flag on a process.
   [ENDPOINTS.md](./ENDPOINTS.md) has the exact numbers.

## Ports

Node p2p 26756, node RPC 26757, indexer 8567, indexer metrics 8568. All
loopback. None collides with the deployment host's in-use list; the indexer's
own default, 8546, does.
