# ADR: gnopie — httpie-inspired CLI for gno.land

## Context

Interacting with gno.land chains currently requires composing `gnokey maketx`
commands by hand: you must know the RPC endpoint, chain ID, gas parameters,
function signatures, and argument types. This friction makes casual exploration
and scripting painful.

`httpie` (the HTTP client) solved a similar problem for REST APIs by providing
a natural verb-based syntax, auto-discovery, and sensible defaults. gnopie
applies the same philosophy to gno.land.

Key design goals:

1. **Zero config for reads.** Auto-discover RPC and chain ID from the domain's
   gnoweb page via `<meta name="gnoconnect:*">` tags — no config file needed
   to browse a realm.
2. **Verb-based ergonomics.** GET, EVAL, READ, INSPECT, CALL, RUN mirror the
   mental model of HTTP verbs applied to on-chain resources.
3. **httpie-style smart dispatch.** The default `GET` verb routes automatically:
   realm paths call `Render`, function expressions go to EVAL, symbol names go
   to READ, network domains go to INSPECT.
4. **Auto-gas.** CALL/RUN simulate first, then broadcast with an estimated gas
   + configurable buffer (default 20%) so users never need to guess gas limits.
5. **gnoweb URL passthrough.** Users can paste any `https://gno.land/...` URL
   directly — gnopie strips the `https://`, modifiers (`$source`, `$help`),
   and fragments (`#func-Name`) to route to the right operation.
6. **Caching.** Network discovery is cached 24h; source file and function
   signature queries are cached 1h. State queries (eval, render, storage) are
   never cached.

## Decision

Add `contribs/gnopie` as a standalone Go module with its own `go.mod`,
following the pattern of other contribs (`gnodev`, `gnofaucet`, etc.).

### Module structure

```
contribs/gnopie/
  main.go         — flag parsing, dispatch, query helpers, type cleaning
  get.go          — GET, EVAL, READ, INSPECT verbs
  call.go         — CALL verb (MsgCall) with gas estimation
  run.go          — RUN verb (MsgRun via generated wrapper package)
  paths.go        — URL/expression parser → GnoPath
  discover.go     — gnoconnect meta-tag discovery + disk cache
  querycache.go   — qfile/qfuncs query cache (1h TTL, file-based)
  config.go       — TOML config (key, gas-buffer)
  cmd_config.go   — `gnopie config {get,set,list}` subcommand
  cmd_version.go  — `gnopie version` subcommand
  cmd_completion.go — bash/zsh/fish shell completion
```

### Path parsing

`ParsePath` classifies any expression into one of:
`PathNetwork | PathNamespace | PathPackage | PathSymbol | PathCall | PathFile | PathAddress | PathUser`

This lets `dispatch` and `execGet` route without conditional string matching
spread across the codebase.

### Signing and gas estimation

CALL signs a dummy transaction, sends it to the simulation endpoint, reads
`gas_used`, then adds `gas_used * gasBufferPercent / 100` to get `gas_wanted`.
The 20% default buffer is configurable via `gnopie config set gas-buffer=30`.

### RUN implementation

RUN wraps the target function call in a generated Go package (`main` package
with a single `main()` that calls the target), uploads it as `MsgRun`, and
broadcasts the transaction. This mirrors `gnokey maketx run`.

### Crossing functions

Before sending a `qeval` query, gnopie fetches `vm/qfuncs` for the package and
checks if the first parameter of the function is a `realm` type. If so, it
auto-injects `cross` as the first argument — matching the GnoVM requirement for
cross-realm calls evaluated via qeval.

## Alternatives considered

1. **Extend `gnokey`** — `gnokey` is a key-management and transaction tool. Its
   UX is deliberately explicit. Adding auto-discovery and smart dispatch would
   add complexity without benefiting its core audience. A separate binary keeps
   concerns clean.

2. **Extend `gnodev`** — `gnodev` is a local development tool with hot-reload.
   gnopie targets any network (mainnet, testnet, local), not just dev nodes.

3. **Single-file CLI** — considered, but the logic for path parsing, discovery,
   gas estimation, caching, and signing is substantial enough that splitting
   into focused files improves readability without adding abstraction cost.

4. **Use `go-toml/v2` instead of `go-toml v1`** — the repo already uses
   `go-toml v1` in several places. gnopie follows the same convention to avoid
   adding a new major dependency.

5. **In-process caching (sync.Map)** — rejected because gnopie is a CLI
   invoked once per command. Disk caching persists across invocations and
   avoids redundant network requests for source files that change rarely.

## Consequences

- Users can explore any gno.land realm or package with a single command.
- CALL/RUN require only a key name and password — no need to know RPC,
  chain ID, or gas values manually.
- The `gnopie config set key=<name>` workflow is the only required setup for
  signing operations.
- Source code caching means `gnopie READ` on the same function is instant on
  repeated calls (within 1h). Cache can be cleared by removing
  `$GNOHOME/gnopie/cache/`.
- gnopie does not handle multi-message transactions or batch operations — those
  remain the domain of `gnokey`.
