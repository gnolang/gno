# ADR-003: Genesis Streaming Cache

## Status

Implemented (branch `feat/genesis-streaming-cache`, single PR pending).

## Context

`GenesisDoc.AppState` is typed `any` and historically held the entire
decoded genesis app state in memory for the lifetime of the node process.
For gnoland1 the on-disk `genesis.json` is ~200 MB — dominated by two
arrays inside `app_state`:

- `balances`: ~3.26M short strings (`g1...=N ugnot`), max ~63 bytes each.
- `txs`: ~2.7K transaction objects, with at least one ~569 KB outlier.

This causes two distinct problems:

1. **Steady-state memory bloat.** Every node carries ~200 MB of essentially
   write-once data forever, on top of state that has already been applied
   to IAVL during InitChain.
2. **`/genesis` RPC is a DoS vector.** `WriteRPCResponseHTTP` calls
   `json.MarshalIndent` on the full `RPCResponse`, allocating a single
   ~170 MB buffer per request, and panics on broken-pipe writes. A handful
   of concurrent requests OOMs the node.

### Requirements

1. **Bounded memory.** Both `/genesis` and the genesis preprocessing pass
   must run in O(1) memory relative to balance/tx count — peak under
   ~50 MB regardless of input size.
2. **`encoding/json` only at the streaming layer.** `amino` is used only
   where polymorphic types force it (validators carry `crypto.PubKey`).
3. **No backwards compatibility shim.** No alias, re-export, or "compat"
   wrapper for symbols introduced on this branch.
4. **Atomic cache updates.** A crashed preprocessing pass must not leave
   a half-written cache that looks valid.
5. **Source-driven invalidation.** The cache regenerates iff the source
   `genesis.json` hash changes. Steady-state startups do not rerun
   preprocessing.
6. **tm2 stays gnoland-agnostic.** The `app_state` shape is gnoland-
   specific (auth/bank/vm/balances/txs). tm2 must not learn about it.

### Alternatives considered

- **Replace `AppState any` with a concrete `*GenesisAppStateRef` in
  tm2.** Rejected: the field's role is to carry an opaque per-chain
  payload, and changing it to a tm2-defined struct couples tm2 to
  gnoland's schema. `any` is exactly correct here.

- **Hand-roll the `/genesis` RPC handler outside the generic RPC
  framework.** Rejected: every RPC method that ever needs to stream
  would re-implement the same envelope/error-handling boilerplate.

- **Cache the rendered JSON-RPC response on disk, serve via `io.Copy`.**
  Solves `/genesis` memory but does nothing for in-memory bloat or
  InitChain. Also requires a render pass that itself buffers ~170 MB
  (`json.Encoder.Encode` materializes the full output before writing).

- **Stream-marshal on every request via `json.Encoder`.** Same
  materialization problem. `Encoder` "streams" multiple newline-delimited
  values, not the bytes of one large value.

- **Drop `*GenesisDoc` entirely after InitChain.** Too blunt: chain_id,
  validators, and app_hash are read from many places throughout node
  lifetime. Targeted nilling of `AppState` (see "node persistence" below)
  is the smaller, safer change.

## Decision

Three coupled changes, shipped together:

1. **Preprocess `genesis.json` on the gnoland side** into a content-
   addressed cache of small files plus newline-delimited JSONL files for
   the bulk arrays. Bind the cache to `GenesisDoc.AppState` as a
   `*GenesisStateRef` (gnoland-side type). `AppState` stays `any`;
   read-path consumers type-switch.

2. **Add a generic `StreamableResult(ctx, w)` hook in the tm2 RPC
   layer**, so any result type that knows how to write itself
   incrementally bypasses the buffer-everything code path.

3. **Strip `AppState` before persisting `GenesisDoc` to the node DB**,
   and re-invoke the genesis provider on each boot to re-attach a fresh
   ref. This recovers the steady-state memory savings and avoids a
   hidden amino-marshal landmine: the persisted-doc encoder cannot
   serialize a `*GenesisStateRef`.

### Architecture

```
gnoland start
        │
        │  StreamingGenesisProvider(genesisFile, cacheRoot)
        │  → node.GenesisDocProvider
        ▼
genesis.json (on disk, indented JSON, ~200 MB)
        │
        │  LoadStreamingGenesisDoc — single-pass token walk
        │  encoding/json.Decoder.Token() — O(1) memory
        │
        │  Pass 1: streaming SHA-256 → srcHash
        │  Pass 2: walk top-level fields
        │            small fields → in-memory map
        │            app_state   → cache writer (cold) or skip (warm)
        ▼
<dbDir>/genesis-cache/<srcHash>/
        ├── version          ← "1\n"
        ├── manifest.json    ← {source_hash, balance_count, tx_count}
        ├── envelope.json    ← app_state siblings (auth, bank, vm, …)
        ├── balances.jsonl   ← one balance per line
        └── txs.jsonl        ← one tx per line, ~569 KB max
        │
        ▼
*GenesisStateRef (cacheDir, manifest, smallFields)
        │
        ▼
GenesisDoc{AppState: *GenesisStateRef}
        │
        │  consumed by InitChain (streaming) and /genesis RPC (streaming)
        ▼
node persistence: AppState nilled out before amino-marshal to DB
node restart:    provider re-invoked, fresh *GenesisStateRef attached
```

### Type changes

```go
// tm2/pkg/bft/types/genesis.go — UNCHANGED
type GenesisDoc struct {
    GenesisTime     time.Time
    ChainID         string
    ConsensusParams abci.ConsensusParams
    Validators      []GenesisValidator
    AppHash         []byte
    AppState        any  // gnoland-specific payload
}

// tm2/pkg/bft/rpc/lib/types/types.go — NEW
type StreamableResult interface {
    StreamJSON(ctx context.Context, w io.Writer) error
}

// tm2/pkg/bft/rpc/core/types/responses.go — method NEW on existing type
func (r *ResultGenesis) StreamJSON(ctx context.Context, w io.Writer) error

// tm2/pkg/bft/node/node.go — NEW constructor + behavior changes
func DefaultNewNodeWithGenesisProvider(
    config *cfg.Config,
    provider GenesisDocProvider,
    evsw events.EventSwitch,
    logger *slog.Logger,
    options ...Option,
) (*Node, error)

// gno.land/pkg/gnoland/genesis_state_ref.go — NEW
type GenesisStateRef struct { /* unexported fields */ }

func OpenGenesisStateRef(cacheDir string) (*GenesisStateRef, error)
func LoadStreamingGenesisDoc(genesisPath, cacheRoot string) (*types.GenesisDoc, error)

func (r *GenesisStateRef) BalanceCount() int
func (r *GenesisStateRef) TxCount() int
func (r *GenesisStateRef) SmallField(key string) (json.RawMessage, bool)
func (r *GenesisStateRef) IterBalances(ctx context.Context) iter.Seq2[[]byte, error]
func (r *GenesisStateRef) IterTxs(ctx context.Context) iter.Seq2[[]byte, error]
func (r *GenesisStateRef) StreamJSON(ctx context.Context, w io.Writer) error  // satisfies tm2's StreamableResult

// gno.land/pkg/gnoland/genesis_provider.go — NEW
func StreamingGenesisProvider(genesisFile, cacheRoot string) node.GenesisDocProvider
```

`*GenesisStateRef` lives in `gno.land/pkg/gnoland` precisely because
its on-disk schema mirrors gnoland's `app_state` shape. tm2 only sees
the `any` AppState and the generic `StreamableResult` interface.

### Why `AppState` stays `any`

A concrete tm2-defined replacement type would force tm2 to know about
`balances`, `txs`, and the rest of gnoland's `app_state` keys. Type-
switching at the read sites is one extra branch; renaming the field is a
tm2-wide invasive change for no architectural gain. The `any` field is
already the project's transport-layer convention for opaque per-chain
payloads.

### Preprocessing pass (cold path)

`LoadStreamingGenesisDoc(genesisPath, cacheRoot)`:

1. **Hash pass.** Stream `genesisPath` through `tmhash.New()` + `io.Copy`
   to compute `srcHash`.
2. **Cache lookup.** `OpenGenesisStateRef(<cacheRoot>/<srcHash>/)`.
   If it succeeds, the cache is valid; jump to step 5.
3. **Walk pass.** Open `genesisPath`, wrap in `*json.Decoder`. Token-walk
   top-level keys:
   - Small fields (`genesis_time`, `chain_id`, `validators`, …):
     `amino.UnmarshalJSON` into the corresponding `*types.GenesisDoc`
     field. Amino is used here because validators carry polymorphic
     `crypto.PubKey`.
   - `app_state`: enter the object. Route children:
     - `balances` / `txs` → enter the array, loop `Decode` per element,
       `json.Compact` each `RawMessage`, write as one line to the
       appropriate `.jsonl` file.
     - Any other key → buffer into the envelope map.
   - **Unknown top-level keys are silently consumed**, matching the
     tolerance of `GenesisDocFromJSON`. Production hardfork tooling emits
     fields like `initial_height` that `GenesisDoc` does not model;
     rejecting them would break startup against real genesis files.
4. **Finalize.**
   - Write `manifest.json`, `envelope.json` to the tmp dir.
   - Write `version` LAST (its presence is the signal the cache is
     complete).
   - `os.Rename` tmp dir → `<cacheRoot>/<srcHash>/`.
   - `gcCacheRoot(cacheRoot, keepHash)` removes any other directories.
5. **Re-open** via `OpenGenesisStateRef` and set `doc.AppState = ref`.

#### Why `json.Compact` per JSONL line

`streamArrayToJSONL` originally wrote the `RawMessage` verbatim, which
preserved the source file's whitespace. Indented genesis files (the
real shape on disk) thus produced JSONL files where a "line" spanned
many physical newlines — not parseable as JSONL. `json.Compact` per
element is mandatory.

This was caught by `TestLoadStreamingGenesisDoc_IndentedSourceProducesValidJSONL`,
added as a regression guard.

### Cache layout

```
<dbDir>/genesis-cache/<srcHash>/
    version          ← contents: "1\n" (Path C — schema version separate from manifest)
    manifest.json    ← {source_hash, balance_count, tx_count}
    envelope.json    ← {auth: ..., bank: ..., vm: ..., …}
    balances.jsonl   ← one balance per line
    txs.jsonl        ← one tx per line
```

The `version` file is intentionally separate from `manifest.json` so the
manifest can evolve incompatibly without losing the version-check
shortcut. `OpenGenesisStateRef` reads `version` first; mismatch → error
without parsing the rest.

`<dbDir>/genesis-cache/` is colocated with the chain DB so it shares
the same lifecycle as other node-managed state — `gnoland
unsafe-reset-all` clears it for free.

### Read-path consumers — type-switch

`gno.land/pkg/gnoland/app.go`:

```go
func (cfg InitChainerConfig) loadAppState(ctx, appState any) ([]abci.ResponseDeliverTx, error) {
    switch state := appState.(type) {
    case GnoGenesisState:
        return cfg.applyInMemoryAppState(ctx, state)
    case *GenesisStateRef:
        return cfg.applyStreamingAppState(ctx, state)
    default:
        return nil, fmt.Errorf("invalid AppState of type %T", appState)
    }
}
```

The two paths share four extracted helpers — `applyBalance`,
`applyUnrestrictedAddrs`, `installAuthParams`, `deliverGenesisTx` — so
parity is structural, not by repeated test.

`applyStreamingAppState` orchestration: bank → balances (streamed) →
auth → unrestricted → vm → install params → txs (streamed).

A test (`TestInitChainer_StreamingAppState_TxParity`) feeds the same
`GnoGenesisState` through both paths and asserts identical
`len(TxResponses)`.

### `/genesis` RPC — generic streaming hook in tm2

The streaming hook is generic, not `/genesis`-specific:

1. `tm2/pkg/bft/rpc/lib/types/types.go` defines
   `StreamableResult { StreamJSON(ctx, w) error }`. Lives in `rpctypes`
   (the type-only package) so `core/types` can reference it without a
   cycle.
2. `tm2/pkg/bft/rpc/lib/server/http_server.go` adds
   `WriteStreamingRPCResponseHTTP(ctx, w, id, result)`. Writes the
   `{"jsonrpc":"2.0","id":<id>,"result":` envelope, calls
   `result.StreamJSON(ctx, w)`, closes `}`. Returns errors instead of
   panicking.
3. `tm2/pkg/bft/rpc/lib/server/handlers.go` — both delivery paths
   (HTTP GET `/<method>` and JSON-RPC POST `/`) type-assert the result
   against `StreamableResult` after `unreflectResult`. When the
   assertion succeeds, route to `WriteStreamingRPCResponseHTTP` instead
   of the buffer-everything `WriteRPCResponseHTTP`.

`processRequest` returns `(*RPCResponse, StreamableResult)`. Batch
requests buffer (an array of streamed results would interleave); single
requests stream when the result type opts in.

#### `(*ResultGenesis).StreamJSON`

```
1. Shallow-clone the doc with AppState=nil.
2. amino.MarshalJSON(clone) → encodes validators[].pub_key with @type tags.
3. Open `{"genesis":` + cloneBytes minus the trailing `}`.
4. If original AppState was non-nil:
   - separator = ',"app_state":'  (doc had other fields)
            or = '"app_state":'   (doc had no other fields → head ends in '{')
   - if AppState implements StreamableResult: appState.StreamJSON(ctx, w)
   - else: amino.MarshalJSON(appState) and write the bytes
5. Close `}}`.
```

The amino marshal is bounded — small fields only (validators, chain_id,
…). The unbounded part (balances/txs) is streamed by
`*GenesisStateRef.StreamJSON`, which does not use amino at all.

#### `(*GenesisStateRef).StreamJSON`

```
1. Open `{`.
2. Sort smallFields keys (Go map iteration is randomized; sort for byte-stability).
3. For each sorted key: write `"<k>":<rawValue>,`.
4. Stream `"balances":[`, iterate IterBalances, comma between elements, `]`.
5. Stream `,"txs":[`, iterate IterTxs, comma between elements, `]`.
6. Close `}`.
```

`ctx` is checked between elements. Determinism matters because
consumers may key on the wire bytes (signing, equality cache).

Per-request memory: one JSONL line at a time, ≤1 MB.

### Node persistence — strip AppState on save, re-attach on load

```go
// tm2/pkg/bft/node/node.go (sketch)
func saveGenesisDoc(db dbm.DB, doc *types.GenesisDoc) error {
    saved := *doc
    saved.AppState = nil  // huge and re-derivable
    bytes, err := amino.MarshalJSON(&saved)
    ...
}

func LoadStateFromDBOrGenesisDocProvider(
    db dbm.DB, provider GenesisDocProvider,
) (state.State, *types.GenesisDoc, error) {
    doc, err := loadFromDB(db)
    if err != nil { ... }
    if doc.AppState == nil {
        // re-invoke provider so streaming providers re-attach a fresh ref
        fresh, err := provider()
        if err != nil { return state.State{}, nil, err }
        doc.AppState = fresh.AppState
    }
    ...
}
```

This is what makes `*GenesisStateRef` durable across restarts. The
amino marshaler in `saveGenesisDoc` cannot serialize a
`*GenesisStateRef` (it's not a registered type, and even if it were,
re-loading from the DB would point at the wrong cacheDir). Stripping
the field on save is correct and re-deriving on load is cheap (the
hash-pass shortcut keeps the cold path off the boot path on warm
caches).

Side benefit: ALL gnoland nodes — not just streaming-enabled ones —
stop bloating their state DB with hundreds of MB of redundant
`AppState`.

### Iterator semantics

- `IterBalances` and `IterTxs` open the JSONL file, wrap in
  `bufio.NewReaderSize(f, 1<<20)` (1 MB buffer absorbs the 569 KB
  outlier tx), loop on `ReadBytes('\n')`.
- Yielded `[]byte` is the line contents without the trailing `\n`.
- **The slice MUST NOT be retained across iterations.** It is reused by
  the underlying reader.
- ctx is checked at start AND per-line.
- On error, the iterator yields `(nil, err)` once and stops.

### DoS guardrails

The `WriteRPCResponseHTTP` panic-on-write was the most visible part of
the DoS vector. Three response writers — `WriteRPCResponseHTTP`,
`WriteRPCResponseHTTPError`, `WriteRPCResponseArrayHTTP` — were
converted to return errors instead of panicking. Twelve call sites
updated. This is foundational for the streaming path: a streaming
response cannot recover from a mid-write panic.

Bounded concurrency and write deadlines are NOT introduced in this
change. They remain a separate hardening pass — out of scope for the
streaming work. The OOM-via-double-buffering vector is closed by the
streaming path itself.

### Testing

1. **Unit tests for `*GenesisStateRef`.** Open / iterate / context
   cancellation / version-mismatch / corrupt-cache rebuild.
2. **Unit tests for `LoadStreamingGenesisDoc`.** Cache hit / cache miss
   / GC of stale caches / different-source-different-cache /
   indented-source-produces-valid-JSONL (regression guard).
3. **Type-switch parity.** `TestInitChainer_StreamingAppState_TxParity`
   feeds the same `GnoGenesisState` through both in-memory and streaming
   paths, asserts identical `len(TxResponses)`.
4. **`*ResultGenesis.StreamJSON` unit tests.** Streamable AppState /
   non-streamable AppState (amino fallback) / nil AppState (omitempty
   preserved) / validators preserve `@type` polymorphism markers.
5. **`*GenesisStateRef.StreamJSON` unit tests.** Shape rehydration /
   empty arrays render `[]` / deterministic small-field order across 10
   iterations / context cancellation propagates.
6. **Generic streaming hook tests in tm2 RPC.** Both delivery paths
   (HTTP GET, JSON-RPC POST) exercised with a synthetic
   `StreamableResult`. Write-error propagation tested with a failing
   `ResponseWriter`.
7. **Node-side persistence tests.** `TestSaveGenesisDoc_OmitsAppState`,
   `TestLoadStateFromDBOrGenesisDocProvider_ReinvokesProviderForAppState`.
8. **Cross-package end-to-end.**
   `TestResultGenesis_StreamJSON_SlimFixtureEndToEnd` — load a real-
   shape slim fixture (47 KB, 5 balances + 2 txs, validator polymorphism)
   through `LoadStreamingGenesisDoc`, wrap in `*ResultGenesis`, stream
   to a buffer, parse the wire body, assert all fields rehydrated.
9. **Memory bound assertion (PENDING).** Real 200 MB fixture, hit
   `/genesis`, assert peak `runtime.MemStats.HeapInuse < 50 MB`. Gated
   on fixture path / not run in CI.

A 200 MB fixture is too large for the repo; the slim fixture covers
real-shape correctness and the memory-bound test is opt-in.

### Out of scope

- `mmap` of JSONL files. Kernel page cache is sufficient.
- JSONL compression. Disk usage is not the limiting factor.
- Multi-cache retention for hardfork rollback. Single-cache policy is
  simpler; `gnoland unsafe-reset-all` clears it.
- Bounded concurrency / write deadlines on `/genesis`. Separate
  hardening pass.
- `gnogenesis` CLI rewrite. Separate follow-up PR; the existing
  in-memory path still works through the `GnoGenesisState` branch of
  the type-switch.

## Implementation

Single PR. The streaming hook in tm2 RPC, the gnoland-side ref + cache,
the type-switch, and the node-persistence changes are coupled — none of
them ship alone, because the type-switch routes to a streaming path
that depends on the streaming hook, and the streaming path is
unobservable to a node that doesn't strip AppState on save.

Files modified:

- `tm2/pkg/bft/rpc/lib/server/http_server.go` — panic→error, plus
  `WriteStreamingRPCResponseHTTP`.
- `tm2/pkg/bft/rpc/lib/server/handlers.go` — `StreamableResult` routing
  in both delivery paths.
- `tm2/pkg/bft/rpc/lib/types/types.go` — `StreamableResult(ctx, w)`
  interface.
- `tm2/pkg/bft/rpc/core/types/responses.go` — `*ResultGenesis.StreamJSON`.
- `tm2/pkg/bft/node/node.go` — `DefaultNewNodeWithGenesisProvider`,
  `saveGenesisDoc` strips `AppState`,
  `LoadStateFromDBOrGenesisDocProvider` re-invokes provider on nil
  `AppState`.
- `gno.land/pkg/gnoland/app.go` — type-switch in `loadAppState`,
  extracted helpers.
- `gno.land/cmd/gnoland/start.go` — wire `StreamingGenesisProvider`.

Files added:

- `gno.land/pkg/gnoland/genesis_state_ref.go` — `*GenesisStateRef`,
  `LoadStreamingGenesisDoc`, `cacheWriter`, `StreamJSON`.
- `gno.land/pkg/gnoland/genesis_provider.go` — `StreamingGenesisProvider`.
- `gno.land/pkg/gnoland/testdata/slim_genesis.json` — 47 KB real-shape
  fixture.
- Tests for everything above.

## Consequences

### Positive

- ~200 MB memory reclaimed per gnoland1 node, permanently.
- `/genesis` becomes O(1) memory and DoS-resistant.
- Preprocessing runs once per `genesis.json` change, not per startup.
- Node DB stops bloating with redundant `AppState` for ALL nodes (not
  just streaming-aware ones).
- Streaming hook in tm2 RPC is generic — future result types that need
  to stream (block downloads, large exports) get the same plumbing for
  free.
- tm2 stays gnoland-agnostic. The `app_state` schema lives entirely on
  the gnoland side.

### Negative

- New on-disk artifact (`<dbDir>/genesis-cache/`) that operators must
  reason about. Mitigated by: colocated with node data, single-cache
  policy, regenerated automatically on hash mismatch, cleaned by
  `gnoland unsafe-reset-all` for free.
- Slight cold-start I/O cost on first `/genesis` request after binary
  upgrade (page cache miss). Negligible thereafter.
- `app_state` JSON output from `/genesis` is no longer byte-identical
  to the on-disk `genesis.json` — same semantic content, but compact
  rather than indented, and small-field key order is sorted rather than
  source-order. Acceptable per design.
- Iterator misuse risk: caller retains `[]byte` across iterations.
  Documented; not enforced.
- One amino dependency at the consumer level: `*ResultGenesis.StreamJSON`
  amino-marshals the doc-with-AppState-nil to preserve `Validators[].PubKey`
  polymorphism. Cannot use plain `encoding/json` here without re-implementing
  amino's interface tagging. Acceptable per the avoid-amino-where-possible
  rule.

### Migration

No data migration. The cache is regenerated on any node that starts
with a new binary and sees a hash mismatch (which it always will on
first run after upgrade). The DB-side `AppState` strip is forward-only:
old DBs with persisted `AppState` are still readable; new boots simply
overwrite with the slim form.

Operators do not need to take action.
