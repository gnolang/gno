# ADR: An optional indexer backend for gnoweb

> The subject is the indexer integration and its removability. The search
> feature is what consumes it, and what demonstrates that removing the flag
> removes the capability rather than breaking the page.

## Status

Proposed. Implements the design discussion opened in
[#5762](https://github.com/gnolang/gno/issues/5762).

## Context

gnoweb answers from a single RPC node, one query at a time. That is the right
default and it is deliberate: gnoweb must run with no external dependency and
no internet beyond its own node (#2799), and it must stay self-describing and
stateless (#5584). Nothing in this ADR changes that default.

What the RPC node cannot answer is anything needing an index: a transaction by
hash, an account's recent activity, which packages mention a given path, a
text search across deployed source. Those are the most-requested missing
pieces, and the search bar today is a URL parser with a client-side path
filter rather than a search.

The complication is trust. An indexer is off-chain and non-consensus; the same
gnoweb binary may point at a community indexer, a self-hosted one, or none.
So the integration has to satisfy three constraints at once:

1. With no indexer, gnoweb behaves exactly as before — no broken UI, no empty
   panels, no "indexer not configured" errors reaching a reader.
2. With an indexer, whatever it contributes is visibly marked as not
   consensus data, with its freshness stated.
3. The indexer's footprint inside gnoweb's core is small enough that removing
   it is a flag, not a refactor.

## Decision

### An indexer is one flag, and its absence is a nil interface

`gnoweb -indexer-url <graphql endpoint>` builds an `*indexer.Client`. Empty —
the default — builds nothing.

The switch is a nil interface, not a config system. `feature/omnisearch`
registers indexer-backed qualifiers **only when `Deps.Indexer` is non-nil**,
so with no indexer they do not exist: they are absent from the hint list, and
typing one in its `name:value` form reports an unknown qualifier rather than
an empty result list that would imply the transaction does not exist.

The no-argument ones (`activity`, `deploys`, `importers`) are typed as bare
words, which are indistinguishable from free text, so on a deployment with no
indexer they fall through to the ordinary path search instead of being named
as unavailable. That is the honest limit of the mechanism rather than a
promise it keeps.

The wire-in assigns the interface field *inside* the `if cfg.IndexerURL != ""`
branch. An interface holding a typed nil pointer is not nil, and that is
exactly how a deployment with no indexer would end up advertising
indexer-backed search and then failing every call.

### gnoweb speaks tx-indexer's GraphQL, not a new API

#5762 left open whether to define a neutral API in this repo or adopt an
existing one. We adopt tx-indexer's GraphQL, because it is the only API with
running implementations: `gnolang/tx-indexer` serves it and
`gnoverse/mygnoscan` consumes it. Defining a gnoweb-specific contract would
produce an interface with one implementation and one consumer, and every
indexer operator would have to write an adapter to be usable.

`gno.land/pkg/gnoweb/indexer/` is a transport, not an abstraction: `Query`
takes a raw GraphQL document, because GraphQL lets several fields ride in one
request. A future caller needing three panels' worth of data merges them into
one round trip without the transport changing.

### Search is a `$search` URL, owned by one feature module

The feature owns every `$search` URL, the way `feature/state` owns every
`?state*` URL:

```
/r/demo/boards$search&q=func:Vote        server-rendered results page
/r/demo/boards$search&q=func:Vote&json   what the omnibar consumes
```

No new top-level route is mounted. The scope comes from the path the reader is
already on, which is also what makes package-scoped qualifiers meaningful.

### Server-rendered first; the omnibar stays additive

The results page is server-rendered. With JavaScript off, the header form
submits and the reader gets every answer. The omnibar dropdown is an
enhancement over that page, not a replacement for it — the discipline
ADR-003 set for the state explorer.

For that claim to be true the header form had to become a real form: it now
carries `method`, `action` and a `name` on its input, and the handler reads
`q` from either grammar — gnoweb's own links build `$search&q=`, while a plain
GET form can only produce `?q=`. Without both halves the bar was
JavaScript-only and the omnibar was a replacement for the page rather than an
enhancement over it.

The core controller learns nothing about any qualifier. It probes
`$search&json` once, renders whatever groups come back, and marks rows the
server labelled `indexer`. A 404 means the endpoint is absent and it stops
asking. Which qualifiers exist is a property of the deployment, and only the
server knows it. The client's rule for *what looks like* a qualifier mirrors
the server's exactly — a whitespace token carrying a colon, unless it looks
like a URL so a pasted link still navigates — rather than being re-derived
into a second grammar that drifts.

### GitHub-style qualifiers, with a hard fetch budget

A query is any number of `key:value` qualifiers plus free text. Qualifiers
split into **selectors** (choose what is searched) and **narrowing** keywords
(`author:`, `in:`, `is:`).

| Selector | Source | Answers |
|---|---|---|
| `func:` `type:` `file:` `imports` | chain | qdoc / file listing for the scoped package |
| `state:` | chain | hands off to `feature/state`'s own `?state&search=` |
| `render:` | chain | text in what realms actually **display** — see below |
| `tx:` `account:` `block:` | indexer | transaction, account activity, block header |
| `activity` `deploys` `importers` | indexer | scoped package's recent calls, deploy history, mentions |
| `content:` | indexer | text across deployed **source** |

Provenance is stamped at registration, from which table a selector is in,
rather than declared by each literal — so a chain selector cannot come to
claim indexer provenance, or the reverse, through a copy-paste.

**Exactly one selector runs per query.** Selectors name different subjects, so
running every registered one would turn a keystroke into a dozen chain and
indexer queries — the amplification ADR-003 §Resource bounds exists to
prevent. Without a selector the query falls to the discovery search, which
costs at most two `ListPaths` calls, or one when `is:` picks a kind. The
users group is derived from paths already fetched, so it is free.

### Provenance is part of the answer

Every result carries the `Source` that produced it, all the way into the
markup. An indexer-backed group renders an `indexer` tag; the page renders a
footer naming the endpoint and the last indexed block. A chain-only answer
never queries the indexer just to stamp a footer nothing will show.

`importers` and `content:` are labelled as **mentions** and **source**, not
imports and content: the indexer answers with a substring match over deployed
file bodies, so a path inside a comment matches too. Overstating them would be
exactly the quiet inaccuracy an indexer-backed feature must not introduce.

## Alternatives considered

**A neutral REST/Connect contract defined in this repo.** Rejected: an
interface with one implementation, and an adapter burden on every operator.
Revisit if a second indexer API gains real adoption.

**Per-feature flags (`-indexer-search`, `-indexer-recent`).** Rejected
(#5762 Q6): one flag plus a registry that reflects what is configured gives
the same control with nothing to keep in sync.

**Capability negotiation against the indexer.** Deferred. mygnoscan had to
cache per-type support because indexers differ, but the hook already exists —
a failed query degrades to a visible group error. If it becomes necessary it
is a cache inside the client, not a redesign.

**Replacing `/search.json`.** Rejected for this PR. It works over RPC today;
rewriting it would make the change substitutive rather than additive. The
qualifiers are layered on top, and plain-text search keeps its existing
client-side path ranking with no request per keystroke.

**Searching rendered `Render()` output across the whole chain.** Not possible,
and `render:` does not attempt it. Nothing indexes rendered output, and
producing it would mean executing `Render` on every realm.

What `render:` does instead is search a field somebody else has already
narrowed: it refuses to run without `author:` or `in:`, caps the result at
`maxRenderCandidates = 8` packages, fetches them 4 at a time, and — because
one execution costs what serving a page view costs — never runs on the
omnibar's per-keystroke JSON path at all. That gate is a `Selector.PageOnly`
field checked by the dispatcher, not a special case inside the one resolver
that happens to be expensive today. This is the bounded fan-out ADR-003
permits, and it is chain data: the node executes `Render`, so a hit is exactly
what a reader would see on the page.

**Background refresh of the chain tip.** Rejected: gnoweb runs no background
work. The tip is fetched per request, coalesced through a singleflight group
and cached for 5 s on the client struct.

**A stricter in-process rate limit.** Rejected once the edge configuration was
read. See the section above: the layer that can tell clients apart already
caps them, and a second, stricter, address-blind bucket underneath it only
produces outages that look like attacks.

## Consequences

### Positive

- Default behaviour is unchanged for a deployment with no `-indexer-url`: no
  outbound request beyond the RPC node, and no UI change beyond the header
  form gaining `method`/`action`/`name`. One pre-existing test was updated —
  `handler_search_test.go`, because `RealmDirectory.Paths` now returns a
  result struct carrying truncation.
- The indexer's own footprint in core is ~39 lines across three files, most
  of them comments: one `AppConfig` field, one guarded constructor call, one
  `HTTPHandlerConfig` field, one flag. Deleting them leaves a compiling
  binary. The feature's total footprint is larger — a second config field
  (`Directory`), a `HTTPHandler.Search` field, the `$search` dispatch block,
  the `isStateJSONRequest` → `isFeatureJSONRequest` widening,
  `HeaderData.SearchAction` and the header form's `method`/`action`/`name`,
  one CSS import and one purge entry.
- Chain-backed qualifiers work with no indexer, so the omnibar gets better
  even on a deployment that never configures one.
- The results page works without JavaScript.
- Adding a qualifier is a resolver, a registration and a template that already
  renders groups.

### Negative

- `content:` and `importers` are substring matches over source, with the
  false positives that implies. Labelled, not hidden.
- Result freshness is an indexer's, not the chain's. Stated per page, but a
  reader who ignores the footer can still act on stale data.
- `tx:` results link to the package a transaction touched, because gnoweb has
  no transaction page. A dead link would be worse; a real one would be a
  different PR.
- Recent listings are windowed by block height because tx-indexer exposes no
  limit or offset. A quiet package can return nothing even though history
  exists further back.
- Path listings are capped by the node at 10000 per kind. Above that the
  answer says it is truncated, but it cannot page past it: `qpaths` has no
  cursor.
- The `$search` webquery key sits next to `feature/state`'s `?state&search=`
  parameter. Different positions, similar names.
- Without JavaScript, a path typed into the bar submits as a search rather
  than navigating. The results page answers that with its path groups, but it
  is a behaviour change for a reader with scripts disabled.
- `render:` is honest about its field but not about the chain: it reports what
  it found in the 8 packages it was allowed to execute, which is not the same
  as what exists.

### Files

New, self-contained:

- `gno.land/pkg/gnoweb/indexer/` — `client.go` (GraphQL transport, per-client
  breaker, request timeout, element-cap and not-found classification),
  `queries.go` (typed queries, widening height window, GraphQL string
  escaping), `client_test.go`
- `gno.land/pkg/gnoweb/feature/omnisearch/` — `feature.go` (local
  `ClientAdapter` / `Indexer` / `Limiter` interfaces, `Deps`, registration),
  `handler.go` (dispatch, timeouts, rate limit, scope resolution),
  `query.go` (tokenising and input bounds), `selector.go` (registry types),
  `resolve_chain.go`, `resolve_indexer.go`, `resolve_render.go`,
  `discover.go`, `json.go`, `view.go`, `component.go`, `template.go`,
  `templates/`, `frontend/omnisearch.css`, tests

Touched in the core:

- `gno.land/pkg/gnoweb/app.go` — `AppConfig.IndexerURL`, client construction
- `gno.land/pkg/gnoweb/handler_http.go` — `HTTPHandlerConfig.Indexer`, handler
  construction, `$search` dispatch
- `gno.land/cmd/gnoweb/main.go` — `-indexer-url`, `-trusted-proxies`
- `gno.land/pkg/gnoweb/components/layout_header.go`,
  `components/layouts/header.html` — `SearchAction`, and the `method`/`action`
  /`name` that make the bar work without JavaScript
- `gno.land/pkg/gnoweb/feature/state/ratelimit.go` — `IPLimiter.AllowRequest`,
  which also makes the previously inert `RateLimitConfig.TrustedProxies` do
  something
- `gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts` — qualifier-aware
  querying, group rendering, submit to the results page
- `gno.land/pkg/gnoweb/frontend/css/main.css`,
  `frontend/postcss.config.cjs` — feature CSS import and purge safelist

## Where a bound belongs: the edge bounds rate, gnoweb bounds cost

Deployments put gnoweb behind Traefik. `misc/loop/traefik/gno.yml` defines a
`web-ratelimit` middleware — 20 req/s average, burst 30 — and
`misc/loop/docker-compose.yml` applies it to the gnoweb router alongside
`secure-headers`. The edge is the only layer that sees the real client
address, and it already caps per-IP rate there.

gnoweb's own per-IP bucket defaulted to 100/min, twelve times stricter than
the edge, while being blind to who was asking: behind a proxy every visitor
resolves to the proxy, so one global bucket throttled everyone at a rate the
edge would have allowed. That is useful security in the wrong place — it adds
no protection the edge does not already give, and it adds a self-inflicted
outage. ADR-003 said as much when it introduced the limiter, calling it
"transitional defense-in-depth" whose primary form "belongs to a future nginx
layer"; that layer exists, and it is Traefik.

So the division this change settles on:

> **The edge bounds the rate.** It sees the real client and already does.
> **gnoweb bounds the cost of one request.** Only gnoweb knows that a single
> request means eight `Render` executions or a scan across a million blocks.

Concretely: the default moves to 1200/min to match the edge, so it stays a
genuine ceiling for a directly exposed gnoweb without fighting the layer that
can tell clients apart, and `-trusted-proxies` lets an operator give it real
addresses where that is wanted. Everything else this change adds — `PageOnly`
selectors, deadline-bounded widening, the indexer concurrency cap, the
response cap, the `MinTerm` floors — is a cost bound, and those belong here
because nothing upstream can compute them.

The duplicated security headers are deliberately kept: Traefik overwrites its
own on the way out, so there is no double header, and gnoweb keeps them for
the case Traefik does not cover — a gnoweb exposed without a reverse proxy.

A `tx-indexer` already runs in that same stack (`indexer.staging.gno.land`,
`-http-rate-limit=500`, `-max-slots=2000`), so `-indexer-url` has a real
endpoint to point at from day one, and the indexer applies its own rate limit
independently of anything gnoweb does.

## Resource bounds

- `MaxQueryLen = 256`, `MaxFilters = 8`, `MinTermLen = 2`, control characters
  rejected before any fetch
- per-selector floors where the term's selectivity sets the cost:
  `content:` 4 characters, `render:` 3
- `MaxResults = 20` per group; `maxDiscoverResults = 10`
- one selector per query; a discovery search costs **one** directory listing
  (itself two `qpaths` calls, fanned out in parallel), shared with
  `/search.json` through the same singleflight group. `is:` narrows what is
  rendered, not what is fetched.
- `PageOnly` keeps fan-out selectors off the per-keystroke path;
  `render:` fans out to at most 8 packages, 4 at a time
- `pageTimeout = 6 s`, `jsonTimeout = 3 s`, indexer client timeout 4 s
- per-IP token bucket on the feature entry as a backstop for direct
  exposure, defaulting to 1200/min to match the edge rather than contradict
  it (see the section above); its own bucket rather than one shared with the
  state explorer. The trusted-proxy rule moved onto the limiter that was
  configured with it (`IPLimiter.AllowRequest`), and `feature/state` was
  migrated to it in the same change, so the rule has one home rather than two
- outbound concurrency to the indexer capped at 16, mirroring the RPC
  client's semaphore; responses capped at 8 MiB before decoding; redirects
  refused, and upstream response bodies never quoted into an error a reader
  will see
- the widening loop stops early when the caller's deadline is nearly spent:
  a window whose answer arrives after the reader gave up costs the indexer
  exactly as much as one they will read
- selectors that fan out or scan the whole chain (`render:`, `content:`,
  `importers`) are `PageOnly` — they answer a committed navigation, never a
  keystroke
- indexer breaker: 3 consecutive failures, 30 s cooldown; a miss or a capped
  result set never counts against it
- chain tip cached 5 s per client and fetched through a singleflight group, so
  a cold cache costs one round trip for all concurrent callers rather than one
  each; fetched on request, never by a ticker
- package paths are validated against `weburl`'s own grammar before becoming
  an href, never escaped into one: deployment is permissionless, so a path is
  attacker-supplied data

## Fixed along the way

Two defects this work depended on, both pre-existing and both corrected here
rather than worked around:

- `rpcClient.ListPaths` never forwarded its `limit` argument, so the node's
  1000 default governed silently and every caller saw the lexicographically
  first 1000 paths. It is forwarded now, raised to the node's own 10000
  ceiling, and a listing that comes back at the cap says so — through
  `PathsResult.Truncated`, into `/search.json` and onto the results page.
  Without that flag a search reports "no such realm" about a realm that
  exists, and always the same ones.
- `weburl.GnoURL.Namespace()` read `idx > 1` where it meant `idx > 0`,
  returning `"a/b"` for `/r/a/b` — wrong for any single-character namespace,
  and wrong for its two existing callers.

## Follow-ups

- Transaction and block pages in gnoweb, which would give `tx:` and `block:`
  a real destination.
- `updated:` needs a deploy timestamp; the height is known, the time is not.
- Transitive dependency depth (`deps:avl` showing `level 2 via ufmt`) needs a
  dependency graph nobody builds yet.
- Cursor pagination on `qpaths`, so a chain past 10000 packages can be listed
  rather than merely reported as truncated.
- `GNOWEB_INDEXER_URL` as an env alternative to the flag, if operators ask.
