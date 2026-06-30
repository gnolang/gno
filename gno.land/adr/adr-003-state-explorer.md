# ADR-003: State Explorer for gnoweb

## Status

Accepted (revised — see `Revision history` at the bottom).

## Context

Gno realms persist their state on-chain as an object graph (structs, maps,
slices, pointers, closures, etc.) encoded in Amino JSON. CLI queries
(`vm/qpkg_json`, `vm/qobject_json`, `vm/qtype_json`) return the raw form, but
the encoding is verbose and only navigable by hand: positional struct fields,
`RefValue` references, `HeapItemValue` wrappers, base-64 primitives.

The State Explorer surfaces this graph as a browsable view inside gnoweb so
developers and auditors can inspect realm state without reaching for CLI tools
or writing one-off decoders.

## Decision

Add a **State** tab to gnoweb that renders the persisted state of any realm or
package as an interactive, expandable tree. The page is **server-rendered** so
URLs are shareable, the view screenshots and prints correctly, and browsers
without JS still produce a working page.

### Architecture

The system has three layers:

1. **VM query endpoints** — `vm/qpkg_json`, `vm/qobject_json`, `vm/qtype_json`
   return Amino JSON for package-level declarations, individual persisted
   objects by ObjectID, and type definitions by TypeID. The previous
   `vm/qeval_json` is reused for read-only expression evaluation. These are
   the **source of truth** for the decoded shape and remain unchanged from
   the original design.

2. **Go decoder + handler** (`gno.land/pkg/gnoweb/feature/state/`) walks the
   Amino JSON into a UI-friendly `StateNode` tree and enriches it with
   `Href`s, syntax-highlighted source snippets, doc comments, and lazy
   previews of top-level references. The handler is a self-contained
   feature module that declares its own `ClientAdapter` subset interface so
   it does not import the parent `gnoweb` package (cycle-free), and consumes
   per-request `Deps` so it stays stateless. Fan-out to the chain is bounded
   by a process-wide RPC semaphore on the client wrapper, by per-IP rate
   limiting on the HTTP entry, and by request-scoped `context.WithTimeout`
   on both the full page (10 s) and per-fragment paths (2 s).

3. **Server-rendered HTML + htmx hydration**
   (`gno.land/pkg/gnoweb/feature/state/templates/`) emits the full page in
   one response so the SSR critical path is exactly two parallel RPCs
   (state + doc). Top-level references hydrate their children on viewport
   entry (htmx `revealed` trigger), so the common case ("what's in
   `boardTree`?") populates without a click and the initial render still
   stays at 2 RPCs. JavaScript is purely additive: a minimal Stimulus
   controller persists open/closed `<details>` state per realm and toggles
   a `Pretty / Tree` view preference via cookie + localStorage. No
   client-side decoding.

### JSON API surface

The CLI-style endpoints from the original design are preserved at the gnoweb
boundary so external tools (block explorers, IDE plugins, SDKs) keep working:

- `$state&json` — paginated package-level state for the current path; wrapped
  in `{pkg_path, total, offset, limit, names, values}` so consumers can
  iterate the realm by hopping `offset`/`limit` in cache-friendly steps
- `$state&oid=…&json` — a single persisted object by ObjectID (passthrough)
- `$state&tid=…&json` — a type definition by TypeID (passthrough)
- proper HTTP statuses — `400` on invalid `oid`/`tid`/`offset`/`limit`,
  `404` on missing package/object, `408` on request timeout, `500` on
  internal error, `502` on upstream response over `maxRPCResponseSize` —
  and a stable `{"error":"…"}` envelope on failure
- `Cache-Control: max-age=1` on every response (latest-only; see
  Historical querying)
- `Vary: HX-Request` on every response so a shared cache cannot serve a
  partial fragment to a non-htmx visitor (or vice versa)

The previous `$state&file=…&json` snippet endpoint is removed — source code is
now embedded server-side via chroma. Raw file bytes remain available via the
existing `$source&file=…` route.

`@gnojs/amino` is retained as a standalone TypeScript library in `misc/gnojs/`
for external consumers that want to decode Amino JSON in the browser. gnoweb
itself no longer depends on it — the Go decoder is the single source of truth.

### Historical querying

Out of scope. Gno object state lives on an unversioned store
(`baseKey` mounted on `dbadapter`), so an ABCI height pin only rewinds
the iavl side; combined with `prune_strategy = "syncable"` on deployed
networks, only the latest block is queryable. The State Explorer
always renders the latest block; `?height=N` is not honored.

### Closures

Closures are detected by a non-empty `Captures` slice on `FuncValue` (the
`IsClosure` flag is unreliable for persisted values). They render with the
function body **and** the captured-variable list in a single card. The
closure tag is injected at hydration via an OOB swap so it appears alongside
the captured environment.

### OID Navigation

The searchbar detects ObjectID patterns (hex `:` int) and redirects to
`<currentRealm>$state&oid=…`.

### Pagination, sidebar TOC, and search

Realms with many top-level declarations paginate at `maxTopLevelDecls = 500`
cards per page; page links carry `&offset=N` and survive search /
view-mode. The sidebar TOC lists every top-level decl (capped at
`maxSidebarTOC = 1000`) regardless of which paginated window is currently
decoded, with off-page entries carrying pre-computed `&offset=N#anchor`
URLs so a click jumps to the right slice and scrolls to the card.
`peekTopLevelKind` reads only the kind discriminator from each amino value,
so the full TOC renders without decoding values the user may never look at.

A server-side search input in the subheader filters top-level decls by name
(case-insensitive substring). It is driven by pure htmx — no JS rewrite
hook, no client-side filter — and the response swaps `#state-results` plus
OOB-swaps `#state-sidebar` and `#state-kind-tabs` so the TOC and badge
counts stay coherent with the visible cards. `hx-push-url` makes the
filtered view shareable.

### Resource bounds

The decoder and HTTP entry protect against pathological state shapes
(intentional or otherwise) and against hostile HTTP burst:

- `maxDecodeDepth = 256` — recursion stop in the walker
- `maxChildrenPerNode = 500` — bound DOM size per collection / struct /
  block / captures slice; excess collapses into a single truncated sentinel
- `maxTopLevelDecls = 500` — pagination cap on the page handler
- `maxSidebarTOC = 1000` — peek-only walk for the full sidebar TOC;
  above this the realm only stays usable through the paginated decl list
- `maxStateIDLength = 256` — bounds attacker-controlled `oid` / `tid`
  parameters before any RPC
- `MaxSearchQueryLen = 128` plus `unicode.IsControl` reject on the `search`
  parameter — keeps the substring scan cheap and the value safe to round-
  trip into `HX-Push-Url` (no CRLF header injection)
- `maxRPCResponseSize = 8 MiB` per outbound RPC — rejected before any
  decode so a misbehaving or compromised RPC node cannot amplify memory
- `maxConcurrentRPC = 64` process-wide semaphore on the `rpcClient` — caps
  gnoweb's outbound load on the chain node no matter the HTTP-side burst
- per-IP token-bucket rate limiter on the feature entry — transitional
  defense-in-depth, primary HTTP rate-limit belongs to a future nginx
  layer; `X-Real-IP` is honored only from configured `TrustedProxies`
- `context.WithTimeout` on the request context — `pageTimeout = 10 s`,
  `fragmentTimeout = 2 s` — propagated to every downstream RPC
- panic recovery at every fetch goroutine **and** at every decoder
  boundary (`parsePackage`, `decodePackageSlice`, `DecodeObject`,
  `DecodeObjectFull`) so amino's hard panics on hostile chain bytes
  return a clean 500 instead of unwinding to net/http's top-level recover

### Frontend hardening

htmx is configured to refuse the script-amplification surface:
`allowEval: false`, `allowScriptTags: false`, `selfRequestsOnly: true`,
`historyCacheSize: 0`. OOB swap targets (`#state-results`,
`#state-sidebar`, `#state-kind-tabs`) carry stable server-stamped IDs;
user-controlled decl names feed `#state-<anchor>-{pretty,tree}` card IDs
with a prefix separator so they cannot collide with the swap targets.
Outbound URLs in templates are typed `template.URL` and HTML strings
(e.g. chroma output) are typed `template.HTML` so escaping is enforced
by construction.

## Consequences

### Positive

- URLs are shareable: drop a link in chat, in an audit report, in a Linear
  ticket; the recipient sees the exact same view
- Pages screenshot and print correctly; crawlers see the full tree
- Works without JavaScript (degraded UX, full data)
- Single Go decoder eliminates TS-vs-Go drift on the decoded shape
- SSR critical path is **2 RPCs** (state + doc, fetched in parallel);
  previews, source bodies, and closure captures hydrate on demand via
  htmx fragments so one HTTP fetch never triggers N RPCs
- Search amplification factor stays at **1×** per keystroke (no extra
  fan-out per filter; combined with the 200 ms client debounce + per-IP
  rate limit, a hostile client cannot multiply chain-side load)
- Defense-in-depth bounds (payload cap + RPC semaphore + panic recovery
  + per-IP rate limit + request timeouts) close the 1-to-many
  amplification class and survive a compromised RPC node

### Negative

- First-paint cost is server-side: a cold page does the 2-RPC critical path
  before any HTML lands; the planned nginx + ETag layer (see Roadmap)
  amortizes this on repeat / cached flows
- Tree state (open / closed nodes) lives client-side; cleared if cookies /
  localStorage are wiped
- PurgeCSS requires a safelist entry for state-explorer kind classes
- Inline raw-JSON in page bodies was dropped (was a memory amplification
  surface scaling with payload size per page view); the **Copy JSON**
  button now fetches `?state&json` async on click — endpoint unchanged

### Files

Self-contained module under `gno.land/pkg/gnoweb/feature/state/`:

- `handler.go` — feature entry point; dispatches json / fragment / page
  paths; applies per-IP rate limiting and request timeouts
- `feature.go` — handler constructor; `ClientAdapter` interface; rate
  limiter config
- `page.go` — full HTML page rendering (`servePackagePage`,
  `serveObjectPage`); parallel `StatePkg + Doc` fetch via `errgroup`
- `render.go` — orchestrator (`DecodePackage`, `parsePackage`,
  `decodePackageSlice`); `RenderConfig` bounds; fragment-vs-page depth
  bucketing; decoder-boundary panic recovery
- `walker.go` — TypedValue-to-StateNode tree walker; `Kind/Shape`
  constants; `maxDecodeDepth` / `maxChildrenPerNode` truncation; closure
  detection via non-empty `Captures`; `DecodeObjectFull` with the same
  panic-recovery discipline
- `json.go` — `?state&json`, `?state&oid&json`, `?state&tid&json`
  endpoints; `pkgJSONWrapper` pagination envelope; `ValidateOID/TID`
  checks before fetch
- `validate.go` — input validators (`MaxStateIDLength`, `MaxSearchQueryLen`,
  `ValidateOffset`, `ValidateLimit`, control-char rejection)
- `ratelimit.go` — per-IP token-bucket; LRU eviction; trusted-proxy gating
  for `X-Real-IP`
- `helpers.go` — URL builders (typed `template.URL`); `recoverFetcher`,
  `recoverToErr`, `recoverDecodeToErr` panic helpers
- `sidebar.go` — `BuildPackageSidebar`, `BuildObjectSidebar`; truncation
  at `maxSidebarTOC`
- `fragments.go` — `serveFragment` dispatcher; per-fragment timeout;
  `frag=node` and `frag=source` handlers
- `errors.go` — `mapClientError` for status classification
- `component.go`, `template.go`, `view.go` — page-level glue and view types
- `templates/{page,_nodes,_pagination,frag_*}.html` — server-rendered
  templates; htmx config locked in `page.html` meta tag
- `frontend/controller-state.ts` — minimal Stimulus controller (open /
  closed tree state, view-mode toggle, doc hydration on htmx swap)
- `frontend/state.css` — feature-scoped Cube CSS

Outside the module:

- `gno.land/pkg/gnoweb/client.go` — `rpcClient` wrapper shared across the
  whole gnoweb process; enforces `maxRPCResponseSize = 8 MiB` (rejected
  before any decode) and the `maxConcurrentRPC = 64` semaphore that caps
  outbound load on the chain node
- `gno.land/pkg/gnoweb/handler_http.go` — wires the feature handler into
  the gnoweb router
- `gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts` — OID detection
- `gno.land/pkg/gnoweb/frontend/postcss.config.cjs` — PurgeCSS safelist for
  state-explorer dynamic classes
- `gno.land/pkg/gnoweb/weburl/url.go` — `?height=N` parsing
- `gno.land/pkg/sdk/vm/keeper.go` — `QueryEvalJSON`, `QueryPkg`,
  `QueryObjectJSON`, `QueryObjectBinary`, `QueryType` (unchanged from
  the original design)
- `gno.land/pkg/sdk/vm/handler.go` — `qeval_json`, `qpkg_json`,
  `qobject_json`, `qobject_binary`, `qtype_json` routes (unchanged)
- `gnovm/pkg/gnolang/values_export.go` — `ExportValues`, `ExportObject`,
  cycle-breaking via `ExportRefValue` (unchanged)
- `misc/gnojs/` — standalone TypeScript library (external consumers only)

### Module shape and ADR-4 alignment

The `feature/state/` package is pre-aligned with the upcoming feature-
framework refactor (planned ADR-4): self-contained module, one interaction
= one RPC (or bounded fan-out for typed-children + source body), no
core ↔ feature back-imports, per-request `Deps` for statelessness. When the
`featureapi.Feature` interface lands as its own PR, the state explorer
becomes the v1 feature with zero logic rewrite, only an interface wrap
(`New() featureapi.Feature`, capability methods on the handler).

## Roadmap

Tracked here so follow-up work is not lost across rewrites. The State
Explorer's full performance story spans three PRs:

| PR | Scope | Outcome |
|----|-------|---------|
| **This PR** | Server-rendered pivot, UX features (pagination, full sidebar TOC, htmx search), hardening pass (panic recover at fetch + decoder boundaries, RPC payload + concurrency caps, per-IP rate limit, search input validation), scroll-paced lazy hydration, stable JSON API | Feature-rich + defensible foundation |
| **PR #N+1** | Default nginx config + lightweight ETag (`server_starttime + path` for static paths; finer-grained for chain-state) + opt-out rate limit + transport-level body cap | Hot paths under 10 ms |
| **PR #N+2** | In-process type / package caches, streaming response (`http.Flusher`), new diff feature | Closes the cold first-paint gap |

Other follow-ups:

- **Known-type leaf views**: types that implement well-known interfaces
  (e.g. `avl.Tree`, `avl.Node`) deserve an alternative render that
  surfaces only leaf values instead of the full internal structure.
  Originally scoped as a follow-up in #5283 by @jaekwon.

## Revision history

- **Initial** (PR #5283, @jaekwon): client-side TS rendering on top of
  `@gnojs/amino`, lazy fetches via `controller-state-explorer.ts`.
- **Revised** (PR #5283, @jaekwon): pivoted to server-side rendering.
  `@gnojs/amino` retained as a standalone library for external consumers;
  the Go decoder is the single source of truth for gnoweb. Added
  `?height=N` time-travel, bounded fan-out, JSON API surface
  stabilisation, doc-comment inlining.
- **This PR**: extracted the state explorer into a self-contained
  `feature/state/` module (pre-aligned with ADR-4), added pagination,
  full sidebar TOC with cross-page navigation, htmx-native server-side
  search, doc-comment inline rendering. Hardening pass: panic recovery
  at decoder boundaries (`parsePackage`, `decodePackageSlice`,
  `DecodeObject`, `DecodeObjectFull`), per-RPC payload cap, process-wide
  RPC concurrency semaphore, per-IP rate limiter with trusted-proxy
  gating, search input validation (`MaxSearchQueryLen` + control-char
  reject for CRLF safety), htmx config locked
  (`allowEval/allowScriptTags/selfRequestsOnly`/no history cache),
  `Vary: HX-Request` on every response, typed `template.URL` /
  `template.HTML` for escape-by-construction. Scoped to latest-block
  rendering (see §Historical querying).
