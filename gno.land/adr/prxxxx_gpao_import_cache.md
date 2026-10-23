# ADR: gpao keeps fetched genesis imports across verifications

## Context

gpao verifies each submitted package in a fresh `verify-one` child process, so the budget can be enforced by killing it and nothing survives between attempts. The child resolves every `/p/` and `/r/` import from the node over `vm/qfile`: one query for a package's file list and one per file. Its cache lives and dies with the process, so every verification fetched its whole import closure again. A candidate importing `gno.land/r/sys/names` costs 228 queries over a 33-package closure, and the next candidate importing the same libraries pays the same again.

The chain bounds what may be cached. `AddPackage` refuses a path that holds a live public package, so a public package the node served once is the package it serves for the life of the chain. A live private package can be redeployed under its path, and a path the node does not serve may be enabled later.

What may be cached safely is not the same as what should be: under the inert policy anyone can get a package approved, pay its deposit, and then submit candidates importing it. A cache keeping every public package would let a submitter grow the daemon's memory at their own pace.

## Decision

1. The daemon owns an `importCache` of public packages added at genesis, kept for the life of the process. Genesis is read from the `[addpkg]` height in the package's `gnomod.toml`, which the keeper stamps with the submitting block's height over whatever the submitter wrote; genesis runs at height 0. Packages added after genesis, private packages and "not found" answers are never kept.
2. Before spawning the child, the daemon walks the candidate's production import closure (`walkImports`, shared with the child's `buildChainImports`), answering from the cache and fetching the rest over `vm/qfile`. It stops at the first transport fault.
3. The child's stdin carries the candidate's amino JSON, as before, then an `importSeed` as a second JSON document: the packages found and the paths the node answered it does not serve. Other drivers of `verify-one` write the candidate alone and keep working: the child then fetches every import itself. The child seeds its per-verification getter with them, so a warm genesis closure costs it no `vm/qfile` query. Anything it is not handed it still asks the node for, and it still classifies every miss (`classifyUnserved`), so the verdict logic is unchanged.
4. The walk runs after the child's clock starts, under `--prepare-budget`, and its queries run under that context, so a stalled node ends in unavailability rather than a hung daemon.
5. The parent queries through `o.client.RPCClient`, which `newOracle` builds from the same `cfg.remote` the child receives as `-remote`.

## Alternatives considered

- **Cache every public package.** It saves the most queries, but its size follows whatever submitters get approved and import.
- **Cache a fixed list of core namespaces.** The list differs per chain and a namespace owner can still publish without bound; genesis is a property the chain records for every package.
- **An LRU bound over every public package.** It caps memory but lets a submitter evict the genesis libraries every candidate needs.
- **A disk cache shared by children.** It survives restarts, but the process that compiles untrusted code would write state that later verifications trust, and a dev chain reset under the same chain ID would serve stale packages unless every hit were re-checked against `vm/qpkgmeta_json`.
- **Children report what they fetched back to the parent.** Data would flow from the untrusted process into the trusted one.
- **JSON-RPC batching without a cache.** It cuts round trips, not queries, and composes with this design for the cold misses.
- **A new RPC client per verification built from `cfg.remote`.** tm2's HTTP client creates its own `http.Transport` with no idle timeout, so the daemon would accumulate idle connections.
- **The parent owning all node I/O, the child getting none.** Cleaner, and it removes the child's network access, but it moves the parked/live classification across the process boundary. Left for a separate change.

## Consequences

- A cold closure costs the same number of queries as before, made by the parent instead of the child. A closure of genesis packages costs none once fetched: for `gno.land/r/sys/names` on a chain that deployed it at genesis, 228 cold and 0 warm, and 0 for another candidate sharing part of the closure.
- A package added after genesis costs what it does on master, every verification.
- The cache is bounded by the chain's genesis: all of `examples/` is about 2.6 MB of production source.
- The parent now parses import blocks of candidate and dependency source (`parser.ImportsOnly`).
- Tests that change an oracle's remote build it with `newTestOracleAt`, and `newTestOracle` defaults to a closed port, so no test can reach a node on the developer's machine.
