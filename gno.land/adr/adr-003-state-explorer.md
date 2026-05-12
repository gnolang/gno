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
URLs are shareable, time-travel is bookmarkable, the view screenshots and
prints correctly, and browsers without JS still produce a working page.

### Architecture

The system has three layers:

1. **VM query endpoints** — `vm/qpkg_json`, `vm/qobject_json`, `vm/qtype_json`
   return Amino JSON for package-level declarations, individual persisted
   objects by ObjectID, and type definitions by TypeID. The previous
   `vm/qeval_json` is reused for read-only expression evaluation. These are
   the **source of truth** for the decoded shape.

2. **Go decoder + orchestrator** (`gno.land/pkg/gnoweb/components/`) walks the
   Amino JSON into a UI-friendly `StateNode` tree and enriches it with
   `Href`s, syntax-highlighted source snippets, and inline previews of
   top-level references. Fan-out to the chain is bounded by per-pool
   semaphores so a single page render cannot flood the RPC layer.

3. **Server-rendered HTML** (Go templates in
   `gno.land/pkg/gnoweb/components/views/state.html`) emits the full page in
   one response. JavaScript is purely additive: a minimal controller
   persists open/closed `<details>` state per realm and toggles a
   `Pretty / Tree` view preference via cookie + localStorage. No client-side
   decoding, no lazy fetches.

### JSON API surface

The CLI-style endpoints from the original design are preserved at the gnoweb
boundary so external tools (block explorers, IDE plugins, SDKs) keep working:

- `$state&json` — package-level state for the current path
- `$state&oid=…&json` — a single persisted object by ObjectID
- `$state&tid=…&json` — a type definition by TypeID
- proper HTTP statuses (`400` on invalid `oid`/`tid` length, `404` on missing
  package/object, `500` on internal error) and a stable
  `{"error":"…"}` envelope on failure
- `Cache-Control: max-age=1` for latest, `max-age=86400, immutable` for
  pinned `?height=N`

The previous `$state&file=…&json` snippet endpoint is removed — source code is
now embedded server-side via chroma. Raw file bytes remain available via the
existing `$source&file=…` route.

`@gnojs/amino` is retained as a standalone TypeScript library in `misc/gnojs/`
for external consumers that want to decode Amino JSON in the browser. gnoweb
itself no longer depends on it — the Go decoder is the single source of truth.

### Time-travel

Appending `?height=N` to any `$state` URL pins the view to a historical block.
A `↺ Latest` link rolls back to live. The pinned height propagates into every
object/type/file fetch in the page, so the entire view is consistent with the
block the user pinned. Object and type links carry the height across hops.

### Closures

Closures are detected by a non-empty `Captures` slice on `FuncValue` (the
`IsClosure` flag is unreliable for persisted values). They render with the
function body **and** the captured-variable list in a single card.

### OID Navigation

The searchbar detects ObjectID patterns (hex `:` int) and redirects to
`<currentRealm>$state&oid=…`. Time-travel preserved if a `?height=N` is in
scope.

### Resource bounds

The decoder protects against pathological state shapes (intentional or
otherwise):

- `maxDecodeDepth = 256` — recursion stop
- `maxChildrenPerNode = 500` — bound DOM size per collection / struct /
  block; excess is collapsed into a single truncated sentinel
- `maxStateIDLength = 256` — bounds attacker-controlled `oid` / `tid`
  parameters before any RPC
- per-pool semaphore caps (`maxConcurrentFileFetches = 8`,
  `maxConcurrentObjectFetches = 8`) — back-pressure on the chain
- total caps on inline previews (`maxInlinePreviewFetches = 30 ×
  maxInlinePreviewRounds = 2`) — bound total RPCs per render
- `context.WithTimeout` on the request context — propagated to every RPC

## Consequences

### Positive

- URLs are shareable: drop a link in chat, in an audit report, in a Linear
  ticket; the recipient sees the exact same view
- Pages screenshot and print correctly; crawlers see the full tree
- Works without JavaScript (degraded UX, full data)
- Time-travel makes audit narratives reproducible across blocks
- Single Go decoder eliminates TS-vs-Go drift on the decoded shape

### Negative

- First-paint cost is server-side: a cold page does one `qpkg_json` plus N
  preview fetches before any HTML lands; the planned nginx + ETag layer
  (see Roadmap) amortizes this on repeat / cached flows
- Tree state (open/closed nodes) lives client-side; cleared if cookies /
  localStorage are wiped
- PurgeCSS requires a safelist entry for state-explorer kind classes

### Files

- `gno.land/pkg/gnoweb/handler_http.go` — `GetStateView`, `ServeStateJSON`,
  status mapping
- `gno.land/pkg/gnoweb/components/state_walker.go` — Amino JSON → `StateNode`
- `gno.land/pkg/gnoweb/components/state_orchestrator.go` — bounded fan-out
- `gno.land/pkg/gnoweb/components/state_sidebar.go` — TOC + Identity /
  Lineage / Storage panels
- `gno.land/pkg/gnoweb/components/view_state.go` — page-level glue
- `gno.land/pkg/gnoweb/components/views/state.html` — server-rendered template
- `gno.land/pkg/gnoweb/weburl/url.go` — `?height=N` parsing
- `gno.land/pkg/gnoweb/frontend/js/controller-state.ts` — minimal toggle /
  view-mode controller
- `gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts` — OID detection
- `gno.land/pkg/gnoweb/frontend/css/06-blocks.css` — state explorer styles
  (Cube CSS)
- `gno.land/pkg/gnoweb/frontend/postcss.config.cjs` — PurgeCSS safelist for
  state-explorer classes
- `gno.land/pkg/sdk/vm/keeper.go` — `QueryEvalJSON`, `QueryPkg`,
  `QueryObjectJSON`, `QueryObjectBinary`, `QueryType`
- `gno.land/pkg/sdk/vm/handler.go` — `qeval_json`, `qpkg_json`,
  `qobject_json`, `qobject_binary`, `qtype_json` routes
- `gnovm/pkg/gnolang/values_export.go` — `ExportValues`, `ExportObject`,
  cycle-breaking via `ExportRefValue`
- `misc/gnojs/` — standalone TypeScript library (external consumers only)

## Roadmap

Tracked here so follow-up work is not lost across rewrites:

- **Known-type leaf views**: types that implement well-known interfaces
  (e.g. `avl.Tree`, `avl.Node`) deserve an alternative render that
  surfaces only leaf values instead of the full internal structure.
  Originally scoped as a follow-up in #5283 by @jaekwon.

## Revision history

- Initial: client-side TS rendering on top of `@gnojs/amino`, lazy fetches
  via `controller-state-explorer.ts`.
- Revised in this PR: pivoted to server-side rendering. `@gnojs/amino` is
  retained as a standalone library for external consumers; the Go decoder
  is the single source of truth for gnoweb. Added `?height=N` time-travel,
  bounded fan-out, JSON API surface stabilisation, doc-comment inlining.
