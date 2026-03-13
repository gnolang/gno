# ADR-003: State Explorer for gnoweb

## Status

Accepted

## Context

Gno realms persist their entire state on-chain as an object graph (structs,
maps, slices, pointers, closures, etc.) encoded in Amino JSON. While
developers can inspect state via CLI queries (`vm/qpkg_json`, `vm/qobject_json`,
`vm/qtype_json`), there was no visual tool for browsing this state tree.

Understanding realm state is critical for debugging, auditing, and building
confidence in on-chain logic. The raw Amino JSON is verbose and requires
knowledge of encoding details (base64 primitives, `RefValue` references,
`HeapItemValue` wrappers, positional struct fields without names).

## Decision

Add a **State Explorer** tab to gnoweb that renders the persisted state of any
realm or package as an interactive, expandable tree.

### Architecture

The system has three layers:

1. **VM query endpoints** (`vm/qpkg_json`, `vm/qobject_json`, `vm/qtype_json`)
   return raw Amino JSON for package variables, individual objects by ObjectID,
   and type definitions by TypeID respectively.

2. **[`@gnojs/amino`](../../misc/gnojs/README.md)** (TypeScript library in
   `misc/gnojs/`) decodes Amino JSON into a UI-friendly `StateNode` tree. Each node has a name, type, kind,
   optional value, and optional children. The decoder handles:
   - Primitive values (base64 `N` field, `StringValue`)
   - Structs with positional fields (names resolved via `qtype_json`)
   - Collections (arrays, slices, maps) with inline or lazy-loaded children
   - Pointers and `RefValue` references for lazy object graph traversal
   - `HeapItemValue` transparent unwrapping
   - `ExportRefValue` cycle detection
   - Functions with source location extraction
   - Closures with captured variable decoding (via `Captures` field)
   - Type values and package references

3. **gnoweb controller** (`controller-state-explorer.ts`) renders `StateNode`
   trees as interactive HTML with:
   - Expandable/collapsible rows with toggle arrows
   - Color-coded types by kind (struct, map, func, closure, primitive, etc.)
   - Lazy-loading of persisted objects via `fetch()` to the state JSON API
   - Struct field name resolution via `qtype_json` round-trips
   - Inline syntax-highlighted source code for function declarations
   - Closure capture display with "Captured variables:" label
   - ObjectID display on hover with click-to-copy
   - Source file links for navigating to function definitions

### gnoweb Integration

- **State tab**: Added as a top-level navigation tab (Content, State, Source,
  Actions) for realm and package views.
- **State JSON API**: `$state&json` serves raw JSON; `$state&oid=...&json`
  serves individual objects; `$state&tid=...&json` serves type definitions;
  `$state&file=...&start=N&end=N&json` serves syntax-highlighted source
  snippets for function bodies.
- **State HTML view**: `$state` renders the full page with the state explorer
  component, which bootstraps from server-rendered initial data.

### Closure Support

Closures are detected by the presence of a non-empty `Captures` array in
`FuncValue` (the `IsClosure` boolean field is unreliable in persisted state).
Captures are `TypedValue` entries with `heapItemType` types pointing to
`RefValue` heap items. The decoder assigns kind `"closure"` (rendered in blue)
to distinguish from regular functions (purple). When expanded, closures show
both the syntax-highlighted source code and the captured variables as child
nodes.

### OID Navigation

The searchbar detects ObjectID patterns (hex/colon format) and redirects to
`$state&oid=...` on the current realm, enabling direct navigation to any
persisted object.

## Consequences

### Positive

- Visual debugging of on-chain state without CLI tools
- Lazy-loading enables browsing arbitrarily large object graphs
- Struct field names resolved from type definitions improve readability
- Closure captures made visible for understanding captured variable state
- Consistent with gnoweb's existing tab navigation pattern

### Negative

- Each object expansion requires a network round-trip to the node
- Struct field name resolution adds an additional round-trip per unique type
- PurgeCSS requires safelist entries for dynamically-constructed CSS classes
  (e.g., `b-state-kind--${kind}`)

### Files

- `gno.land/pkg/gnoweb/handler_http.go` — `GetStateView`, `ServeStateJSON`
- `gno.land/pkg/gnoweb/components/views/state.html` — state view template
- `gno.land/pkg/gnoweb/components/layout_header.go` — State tab in navigation
- `gno.land/pkg/gnoweb/frontend/js/controller-state-explorer.ts` — tree controller
- `gno.land/pkg/gnoweb/frontend/css/06-blocks.css` — state explorer styles
- `gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts` — OID detection
- `gno.land/pkg/gnoweb/frontend/postcss.config.cjs` — PurgeCSS safelist
- `misc/gnojs/src/decode.ts` — Amino JSON decoder
- `misc/gnojs/src/types.ts` — Amino type definitions
- `misc/gnojs/src/type-utils.ts` — type name/kind/signature utilities
- `gno.land/pkg/sdk/vm/keeper.go` — `QueryPkgJSON`, `QueryObjectJSON`, `QueryTypeJSON`
- `gno.land/pkg/sdk/vm/handler.go` — `qpkg_json`, `qobject_json`, `qtype_json` routes
