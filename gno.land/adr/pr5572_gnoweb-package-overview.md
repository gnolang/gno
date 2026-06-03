# ADR: gnoweb package overview (source landing page)

## Status

Accepted.

## Context

PR #4542 (stale 9+ months) proposed a pkg.go.dev-like overview page for
gnoweb's `$source` URL. Review feedback on that PR identified several
architectural issues: a monolithic 537-line template, global mutable
state via `SetRenderer`, a parallel markdown pipeline with ReDoS risk,
out-of-scope edits across `views/user.html`, `views/help.html` and
`views/source.html`, and metadata fields without a real data source.
The sibling PR #5562 (which supersedes the stale #4466) delivers the
proper infrastructure for doc-context markdown rendering
(`RenderDocumentation` + `ExtCodeExpand` extension) on current master.

A redesigned implementation supersedes #4542 by reimplementing the
feature on the current codebase, reusing #5562's infrastructure, and
focusing strictly on the overview view. Fixes #3665. Also addresses
part of the mainnet UX effort tracked in #5463.

## Decision

1. **New view `OverviewView`** (`components.OverviewViewType`) is rendered
   at `/r/<pkg>$source` and `/p/<pkg>$source` when no `file=` query
   parameter is set. `/r/<pkg>$source&file=X` keeps the existing
   `SourceView`. `/r/<pkg>` (no `$source`) is unchanged: `RealmView` if
   `Render()` is defined, otherwise `DirectoryView`.

2. **Overview content** is derived from existing RPCs only. `vm/qfile`
   supplies the file list plus the gnomod.toml and LICENSE bodies used
   for metadata and license detection. `vm/qdoc` supplies package doc,
   exported functions/types/constants/variables, imports, and `BUG(...)`
   comments. `vm/qpaths` supplies subpackage children. No new VM endpoint
   is introduced.

3. **`vm/qdoc` now carries source positions and imports**. `gnovm/pkg/doc`
   adds `File`/`Line` to `JSONValueDecl`, `JSONFunc`, `JSONType` and an
   `Imports []string` to `JSONDocumentation` (all `json:",omitempty"`).
   Positions come from a single new `(*pkgData).extractPosition(ast.Node)`
   helper; imports from `(*pkgData).imports()`, which reuses the
   `gnovm/pkg/packages` path-extraction idiom on the already-parsed AST
   (no second parse) and returns a deduplicated, sorted, non-test set.
   This unlocks per-symbol deep links and lets the overview render the
   import list without fetching and re-parsing source files. Additive /
   backward-compatible with existing `gno doc` CLI consumers.

4. **Metadata derivation is pure.** Helpers are split by concern across
   `components/overview_{build,symbols,imports,license,files}.go` and
   operate on `[]string`, `*doc.JSONDocumentation`, `map[string][]byte`.
   They are unit-tested without RPC mocks.

5. **Doc string rendering reuses PR #5562's pipeline** via a minimal
   `DocRenderer` interface (`RenderDocumentation(io.Writer, []byte)` +
   `RenderSource(io.Writer, name, []byte)`). No regex-based doc parsing.
   No `docparser/` package.

6. **The `DocRenderer` is injected per-request** through the handler.
   The global mutable `components.SetRenderer` pattern from #4542 is not
   reintroduced.

7. **Parallel data fetching**. `GetOverviewView` runs `ListFiles`,
   `Doc`, README rendering, and `ListPaths` concurrently via `errgroup`,
   then fetches only the two metadata files it still needs (`gnomod.toml`
   and `LICENSE`) — imports now come from `vm/qdoc`. An overview request
   is bounded to a small, constant set of RPCs instead of growing with
   the package's source-file count.

8. **CUBE CSS extension**. `06-blocks.css` adds block-level styles for
   the overview page, following the existing `b-*` naming convention
   and using semantic design tokens exclusively (`--s-color-*`,
   `--g-space-*`, `--s-border*`, `--s-rounded*`). Dark mode works
   automatically via the existing `[data-theme="dark"]` attribute.

## Alternatives considered

- **Extend `DirectoryView` in place** — rejected. Would conflate two
  semantics (explorer listing vs package landing) into one view. Leaving
  `DirectoryView` minimal preserves its purpose as a fallback path
  explorer.
- **Add an explicit `$overview` URL parameter** — rejected. `$source`
  already means "code browsing"; the overview is the landing page of
  that browsing. A new parameter would fragment the URL surface.
- **Keep `docparser/` package** — rejected. It duplicates Goldmark and
  introduces ReDoS surface through regexes on arbitrary doc content.
  PR #5562 provides the proper doc-context rendering pipeline.
- **Add `vm/qmeta` endpoint to fetch Creator / Height / Draft / Private**
  — rejected for this PR. Out of scope (VM-layer work). Backlogged as
  a prerequisite for future metadata features.
- **Extract a `pkgmeta/` package** — rejected (YAGNI). The current
  gnoweb pattern is colocated helpers in `components/`. Extraction is
  a cheap refactor if/when the logic needs to be shared.
- **Ship a minimal set of CSS blocks** — rejected. The UI needs distinct
  styles for the sticky section nav, the code-block toolbar with its
  copy/view-source affordances, and the generic list wrapper shared by
  imports/files/subpackages. Extracting them as named blocks follows
  existing gnoweb patterns and keeps the template readable. Total is
  eight blocks — `b-pkg-meta`, `b-pkg-quality`, `b-pkg-toc`,
  `b-pkg-nav`, `b-pkg-section`, `b-pkg-symbol`, `b-pkg-code-block`,
  `b-pkg-list` — plus tag variants (`b-tag--kind`, `b-tag--type`,
  `b-tag--crossing`). Each block is single-purpose.

## Consequences

- **Breaking UX on `$source` landing**: users hitting `/r/<pkg>$source`
  (without `file=`) will now land on Overview rather than the first
  file's code. Direct-to-file access via `$source&file=X` is unchanged.
  Mitigation: Overview prominently lists all files with one-click
  navigation, and every symbol links directly to its source line.
- **Dependency on PR #5562**: this PR cannot merge before #5562 and is
  developed on a branch based on `fix/gnoweb-doc-render`, rebased onto
  master once #5562 merges.
- **Grid layout tweak**: on the overview page only, `.b-sidebar` spans
  all implicit content rows (`grid-row: 1 / span 99`) and `height:
  auto`, so its natural height no longer forces row heights. Scoped
  via `main.dev-mode > section:has(.c-overview-view)`. Other views are
  unaffected.
- **Existing tests updated**: two tests in `handler_http_test.go` that
  asserted the old first-file-on-`$source` behavior were removed
  (`TestHTTPHandler_GetSourceView_NoFiles`,
  `TestHTTPHandler_GetSourceView_FilePreference`). One case in
  `TestHTTPHandler_Get` updated to explicitly request `&file=gno.mod`.
- **JSON wire schema for `vm/qdoc` gains optional fields** (`file`,
  `line`, `imports`). New-client ↔ new-server and new-client ↔ old-server
  are both safe (the fields are `omitempty` and unknown-on-the-wire when
  absent). Old-client ↔ new-server may fail to decode if the consumer's
  JSON deserializer is strict about unknown fields. In this monorepo the
  two known consumers (`gno doc` CLI and `gnoweb`) are rebuilt in the
  same release cycle, so no in-tree breakage occurs; external consumers
  pinned to an older schema would need to update.
- **`gno.land/pkg/sdk/vm/handler_test.go`** fixture updated with expected
  `file` / `line` values for `TestVmHandlerQuery_Doc`.
- **Rollback**: a single `git revert` restores the previous behavior —
  no persistent state is modified, no DB migration, no impact on other
  services.
- **Forward compatibility**: packages already deployed on-chain keep
  rendering identically; direct `$source&file=X` access is unchanged.

## Scope in

- `gnovm/pkg/doc/json_doc.go` — add `File`/`Line`/`Imports` to JSON types, populate
- `gnovm/pkg/doc/pkg.go` — add `extractPosition` + `imports` helpers
- `gnovm/pkg/doc/json_doc_test.go` — fixture + imports test
- `gno.land/pkg/gnoweb/components/view_overview.go` — data types + factory
- `gno.land/pkg/gnoweb/components/overview_build.go` — orchestration, stats, quality, TOC
- `gno.land/pkg/gnoweb/components/overview_symbols.go` — funcs / types / values
- `gno.land/pkg/gnoweb/components/overview_imports.go` — import parsing
- `gno.land/pkg/gnoweb/components/overview_license.go` — license detection
- `gno.land/pkg/gnoweb/components/overview_files.go` — file classification
- `gno.land/pkg/gnoweb/components/{view_overview,overview_*}_test.go` — unit tests
- `gno.land/pkg/gnoweb/components/views/overview.html` — template
- `gno.land/pkg/gnoweb/components/ui/icons.html` — `ico-shield`, `ico-search`, `ico-kind-*` sprites
- `gno.land/pkg/gnoweb/components/layout_index.go` — dev-mode switch
- `gno.land/pkg/gnoweb/frontend/js/controller-search.ts` — symbol filter
- `gno.land/pkg/gnoweb/frontend/js/controller-observer.ts` — scroll-spy section nav
- `gno.land/pkg/gnoweb/handler_http.go` — `GetOverviewView`, routing
- `examples/gno.land/p/demo/showcase/` — demo package exercising the overview
- `gno.land/pkg/gnoweb/handler_http_test.go` — overview tests
- `gno.land/pkg/gnoweb/app_test.go` — 3 routing cases in `TestRoutes`
- `gno.land/pkg/gnoweb/frontend/css/06-blocks.css` — new blocks
- `gno.land/pkg/gnoweb/public/main.css` — compiled artifact

## Scope out (explicit rationale)

Each of the following is deliberately deferred to keep this change
surgical. The rationale matters for future reviewers evaluating
follow-up work.

| Deferred | Rationale |
|----------|-----------|
| Edits to `views/user.html`, `views/help.html`, `views/source.html` | Isolates the change to the new `OverviewView` and its supporting code. Avoids cascading regressions in unrelated views. |
| Global mutable state via `SetRenderer` | Pattern rejected during #4542 review. Dependency injection through the `DocRenderer` interface is used instead. |
| `docparser/` package | Would duplicate Goldmark and re-introduce a regex-based markdown pipeline with ReDoS surface. PR #5562 already delivers the doc-context rendering path. |
| `vm/qmeta` endpoint (Creator / Height / Draft / Private metadata) | VM-layer work, out of scope for a gnoweb-only change. Backlogged as prerequisite for metadata-rich badges. |
| Interactive playground | Significant UX surface; separate RFC needed. |
| Transaction history section | Non-essential for the first release; separate feature. |

## Security

- All user-controlled strings (package doc, README, file names, import
  paths, license content) pass through Goldmark safe mode or
  `html/template` and are HTML-escaped at the boundary.
- No `template.HTML` is produced from user-controlled input. Components
  returned by `renderDocString` wrap the output of Goldmark, which is
  safe by construction.
- License detection reads at most 4 KB of license content; regexes are
  anchored and use Go's RE2-based `regexp` package (linear-time, no
  backtracking). Zero ReDoS surface.
- `ActionURL` is constructed from the validated `pkgPath` and a Go
  identifier (`token.IsExported` filter upstream); no injection vector.
- `SourceURL` concatenates the validated `pkgPath`, a file name from
  `vm/qdoc`, and `strconv.Itoa(line)`. File names are passed through
  `html/template` in `href` context which rejects `javascript:` schemes.
- HTML-injection escaping is covered at the renderer boundary in
  `render_test.go` (`TestHTMLRenderer_RenderDocumentation_StripsRawHTML`),
  which is the same pipeline the overview handler uses.

## Testing

- **Unit tests** (`components/overview_*_test.go`,
  `components/view_overview_test.go`) cover every pure helper with
  table-driven cases and `t.Parallel()`: license detection (8 cases),
  imports (5), stats, quality, synopsis, file links, subpackages, TOC,
  info, symbols (5), values, BuildOverview, buildSourceURL,
  filterNonTestSources, packageTypeOf, rawHTMLComponent.
- **Handler tests** (`handler_http_test.go`): success with all sections,
  degraded on qdoc failure, 404 on package not found, routing matrix
  (overview vs source view).
- **Routing integration** (`app_test.go`): 3 cases added to `TestRoutes`
  covering Overview keyword, Source Files keyword, and source view
  deep-linking.
- **Documentation positions** (`gnovm/pkg/doc/json_doc_test.go`):
  fixture extended with 23 expected `File`/`Line` assertions across
  values, funcs, and types.
- **Coverage** on `gno.land/pkg/gnoweb/components/`: above 88 % after
  the new tests (every overview helper exercised).

### Follow-up work

Candidate refinements that build on this decision, ordered by priority:

1. Render interface methods (`JSONType.InterElems`) inside
   `b-pkg-symbol`. Current templates render only the interface
   signature; method-by-method display would match pkg.go.dev.
2. `txtar` integration tests that drive `gnoweb` over HTTP. Blocked
   on adding `gnoweb start` / `http_get` directives to the testscript
   runner. Covered by `TestRoutes` in the interim.
3. `vm/qmeta` endpoint to expose Creator, Block Height, Draft and
   Private flags; prerequisite for the richer sidebar badges #4542
   originally proposed.
4. CUBE CSS migration of `views/user.html` (large rewrite, separate
   PR).
