# ADR: gnoweb serves realm markdown via Accept negotiation

## Status

Implemented. Not consensus-affecting — gnoweb is a read-only presentation layer.

## Context

A realm's `Render()` returns markdown. gnoweb has always converted it to HTML
through goldmark and wrapped it in the page layout, so the markdown source was
reachable only by re-deriving it: query `vm/qrender` over RPC, or scrape the
rendered page back. Non-browser clients — agents, doc tooling, scripts — want the
source, and the chain already produced it one step earlier in the same request.

The concrete case that prompted this: Claude Code's `WebFetch` sends
`Accept: text/markdown, text/html, */*`. Before this change it received the full
HTML page and had to strip it; the home realm is roughly 4.7 KB of markdown
inside a ~58 KB page (measured against `gnodev`).

## Decision

Negotiate on `Accept`. When the client names `text/markdown`, a realm page and a
static-markdown alias return the raw bytes with
`Content-Type: text/markdown; charset=utf-8`, no page layout and no goldmark
pass. Everything else is unchanged.

Three design points carry the weight.

**Negotiation is explicit-only.** `negotiatesMarkdown` matches `text/markdown`
and the alias `text/x-markdown`, parsed with `mime.ParseMediaType` rather than
hand-rolled splitting. It never matches the `*/*` or `text/*` wildcards. This is
the load-bearing rule: every browser sends `*/*`, so wildcard matching would
serve markdown source to browsers and break the site. An explicit `q=0` is
honored as a refusal.

Relative q-ranking is deliberately not implemented. Naming markdown at any
non-zero q selects it, even where another type ranks higher — a shortcut over
RFC 9110 preference ordering. Only a client that asks for markdown *by name*
reaches the branch at all, so there is no competing preference worth ordering,
and mixing explicit-only matching with q-ordering has no coherent reading (what
should `text/html;q=0.9, */*;q=0.8` mean when the wildcard is ignored by
design?). The reasoning is recorded in the function's doc comment so it is not
mistaken for an oversight.

**The signal travels as a view type, not a second code path.** `MarkdownView`
returns a `*components.View` tagged `MarkdownViewType`; `Get` checks that tag on
the returned body view and, if set, writes the view directly instead of feeding
it to `IndexLayout`. The alternative — branching on `wantMarkdown` inside `Get`
before dispatch — would have to duplicate the routing (alias lookup, weburl
parse, realm-vs-pure-vs-user) that decides whether markdown is even available.
Routing runs once and reports back what it produced.

`wantMarkdown` still threads through `prepareIndexBodyView` → `GetPackageView`
as a dispatch selector, but it stops at the leaf: `GetRealmView` (HTML) and
`GetMarkdownRealmView` (raw) are separate functions sharing a `fetchRealm`
helper. Keeping the flag out of the leaves is what makes the fallback ordering
below hold structurally.

**Fetch failures fall back to HTML, always.** `fetchRealm` performs the
`Client.Realm` call and owns the three error cases: no `Render()` declared →
directory view, package not found → paths-list view, anything else → error view.
All three are HTML, returned as-is. The markdown short-circuit sits *after* that
switch, so a realm with no `Render()` under `Accept: text/markdown` gets the HTML
directory page rather than an empty markdown body.

Two response headers follow from serving two representations of one URL:
`Vary: Accept` on GET, so shared caches key them separately; and
`X-Content-Type-Options: nosniff` on the markdown write, because these bytes
reach the client without the goldmark sanitization the HTML path applies, and
`nosniff` stops a browser sniffing them back into an executable type. The GET
`Content-Type` default also changed from `Header().Add` to `Header().Set`, so the
markdown path overrides it instead of appending a second value.

## Consequences

- Browsers, curl without an `Accept` header, and every existing client keep
  receiving byte-identical HTML. The negotiation is unreachable without naming
  the type.
- Realm `Render()` output now leaves gnoweb unsanitized on this path. That is
  the point of it, and the reason for `nosniff`. Anything downstream treating the
  response as trusted markup is the caller's problem, as it already is for anyone
  querying `vm/qrender` directly.
- Scope is realm pages and static-markdown aliases. Source, help, directory,
  overview and user views still return HTML under any `Accept`; they are
  assembled from templates, not from a markdown source that exists to be handed
  over. Extending them means deciding what their markdown *is*, which is
  follow-up work, not a gap here.
- `GetPackageView` and `prepareIndexBodyView` gained a parameter. Both are
  exported/package-internal to gnoweb only.

## Alternatives considered

**A `?format=md` query parameter.** Explicit, trivially testable, no header
parsing. Rejected because it changes the URL: the markdown and the HTML would be
different resources, so agents would need per-site configuration to find it, and
the identity between "the page" and "the page's source" is lost. `Accept` is the
mechanism HTTP already has for one resource with two representations, and it
works with zero client configuration — which is exactly why `WebFetch` gets
markdown now without anyone opting in.

**A `.md` path suffix** (`/r/gnoland/home.md`). Same URL-identity objection,
plus it collides with the existing path grammar, where a trailing segment after a
realm path is already meaningful (`$source`, file names, render args).

**Content negotiation via a middleware wrapping the whole handler.** Would apply
uniformly to every view, but there is no generic "give me the markdown" for views
built from templates, so it would need per-view opt-in anyway — the flag, spread
wider.

## Testing

- `negotiate_test.go` — the header rules as a table: both media types, charset
  params, case-insensitivity, `q` values including `q=0` / `q=0.0` / negative,
  an unparseable `q` (ignored, so the type still counts as accepted), the `*/*`
  and `text/*` wildcards, and the literal
  `Accept: text/markdown, text/html, */*` that `WebFetch` sends, pinned as a
  regression case.
- `components/view_markdown_test.go` — `MarkdownView` renders its bytes
  verbatim and carries `MarkdownViewType`.
- `handler_http_test.go` — end-to-end through `ServeHTTP`: `Content-Type`,
  `Vary: Accept`, `nosniff`, markdown vs HTML bodies across seven `Accept`
  values; a static alias served verbatim; and the ordering invariant — a realm
  with no `Render()` falls back to the HTML directory view even when markdown is
  requested.
- Verified live against `gnodev`:
  `curl -H 'Accept: text/markdown' http://127.0.0.1:8888/r/gnoland/home` returns
  markdown, a plain request returns the HTML page.
