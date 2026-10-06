# ADR: gnoweb draws listings as a map and packages as a dependency graph

## Context

gnoweb lists what is on the chain, one path per line. A reader cannot see at
a glance where the code is, which namespaces hold most of it, which realms are
used, or what a package depends on and what depends on it. Block explorers
outside the repo answer this visually, from their own index.

Two constraints frame the answer.

- gnoweb runs with nothing but its RPC node (#2799), with no background work
  (#5584, and the "Background refresh of the chain tip" alternative rejected
  in `pr6231_gnoweb_indexer.md`).
- An indexer can be plugged in with `-indexer-url` (#6231). Whatever it adds
  must be marked as indexer data with its freshness, and must vanish, not
  degrade, when the flag is absent.

So the visual has to be complete and honest on the node alone, and the
indexer may only add to it.

## Decision

### The map is the list, drawn

Every explorer listing (`/r/`, `/p/`, `/r/<namespace>`, …) gains two tabs in
the header slot where a package shows Content, Source and Actions: **List**
and **Map**. `$map` on the same path draws the same listing.

`GetPathsListView` lists the paths once and hands the same slice to either
rendering, so the two cannot disagree about what exists. The listing now asks
the node for one more path than it shows: past `maxListedPaths` both
renderings say "only the first N paths are listed" (`ui/listing_truncated`)
instead of passing for complete. `qpaths` has no cursor, so stating the cut is
all a listing can do.

The map is a squarified treemap: one tile per package, grouped by the next
path segment below the listed path, each group sized by how many packages it
holds. Package count is the one size the node answers for a whole listing in
a single query; lines or bytes would need every file of every package. A
group with packages below its key links to that key's own map, so zooming is
navigation.

### SVG, laid out by the server, no script

The map is an `<svg>` of nested `<svg>` elements placed by percentage `x`,
`y`, `width` and `height` attributes; every tile is an `<a>` named by its
`<title>`. Attributes, not inline styles: gnoweb's Content-Security-Policy is
`style-src 'self'`, which drops `style` attributes — the first version of this
map rendered as an empty box for exactly that reason, and a test now pins it.
Percentages keep the text at the stylesheet's size at any width, and each
nested `<svg>` clips its label.

Colours live in `chainmap.css` as custom properties set by modifier classes.
The classes are written out whole in Go so the production purge, which keeps
only names it finds in the sources, keeps them; this avoids touching the
purge configuration.

### Activity, when an indexer is configured

On realm maps, tiles are coloured by successful `MsgCall`s over the last
seven days, on a log scale relative to the busiest realm on that map. The
aggregate is computed band by band (`indexer.CallsBetween`), bands split in
half when the indexer's element cap refuses them, and cached for five
minutes. A refresh is started by a request, detached from it and shared by
every concurrent reader; a stale aggregate is served while it runs. A page
waits three seconds, then says the activity is still being counted.

A band still over the cap at the narrowest width makes the aggregate partial:
counts become lower bounds ("at least N calls") and a realm with no counted
call is drawn as unknown, never as zero. A pure-package map says that nothing
calls packages rather than showing an empty scale.

### The dependency graph

The overview's Imports section becomes a graph: the package between what it
imports, read from `vm/qdoc`, and what imports it. The import row is one
partial (`ui/pkg_import`) shared by both sides.

Importers are not something the chain can list. With an indexer, a `$deps`
page proposes candidates by asking for every deploy quoting the import path
(`indexer.DeploysQuoting`), then checks each candidate's `vm/qdoc` imports on
chain. Both steps are needed: on gnoland-1 in October
2026, 11 of the 210 deploys quoting `gno.land/p/nt/avl/v0` were batched
deploys carrying 127 packages between them, and the indexer matches a
transaction as a whole; a quote can also sit in a test or a string. Only the
chain can say which package imports it, so the side is labelled "Imported by"
and contains nothing the chain did not confirm.

Candidates are deploys holding the path as a Go string literal, interpreted
or raw: an import may be written with backquotes. The scan walks the chain in
bands of 75 000 blocks, two at a time, the bottom band without a lower bound
so it reaches genesis at height 0, and the first failing band cancels the
others. One query over the whole chain took 4.9 s on gnoland-1, past the
indexer client's 4 s request timeout; bands of 150 000 took 0.6 to 1.7 s and
returned exactly the 210 rows of the whole-chain query, and the narrower band
keeps a margin, because that client's breaker counts timeouts and an open
breaker turns off search for every reader.

The lookup costs that scan plus a node read per candidate — about 38 s for
`gno.land/p/nt/avl/v0`'s 265 candidates against the public RPC, which found
183 importers — so it is a page of its own: never run by the overview, `noindex`, capped at
400 candidates, limited to six lookups a minute per address (burst three)
and to two lookups at once across all readers. The slot and the limiter are
taken inside the shared fetch, which runs once however many readers ask: only
the reader who starts a lookup pays, a cached answer or joining one in flight
is free, and a lookup finding both slots taken is refused at once ("other
lookups are running") before the limiter is asked, rather than queued on its
own budget. Neither refusal is remembered. Behind a reverse proxy this
limit, like gnoweb's others, needs `-trusted-proxies` to tell readers apart. The
package's own imports are read first, so a path with no live package answers
404 without costing the indexer anything. Answers are
cached per package for ten minutes, and each candidate's imports are cached
too, so packages sharing importers read them once. A reader waits at most 12 s;
the lookup runs detached for up to 60 s and a reload finds it. Only the
reader's own wait reads as "still looking": a lookup that failed, even on its
own deadline, reads as unavailable. An answer some candidate reads failed to
complete is kept one minute rather than ten, so a reload retries them. Past any of its
bounds — a band the indexer capped, the candidate cap, an unread candidate —
the count reads "at least".

A map exists only where the list does, that is where no package lives at the
path: on a package, its List tab would open the package instead of the
listing it drew. The package check runs beside the listing; when the node
cannot answer it, the map is an error, not a guess.

No band query starts after its lookup or refresh has given up, since it could
only fail against the shared client's breaker; the check sits inside each
band's goroutine, past any wait for a concurrency slot. An activity refresh
that runs out of time fails as a whole, and the last good week stays served.

The graph says why an answer may be incomplete: the search stopped at a limit
(a capped band or the candidate cap), which a reload cannot lift, or some
candidates could not be read, which it may.

### One cache shape

Activity, importers and candidates' imports share one small generic cache
(`flight`): answers per key, a fetch started by a reader but detached from it
and shared by every concurrent reader, a TTL, a size bound, a short memory of
failures so a failing indexer is not asked again by every reader, and, for
activity, the last good answer served while a refresh runs.

### Undecoded indexer rows

tx-indexer lets a transaction whose message it could not decode
(`__typename: UnexpectedMessage`) through any message filter: 19 of the 20
rows a deploy filter returned for one package on gnoland-1 were such rows. The
two new queries re-check every row against the message they filtered on. The
existing queries of #6231 have the same exposure; they are left as they are
here and reported to that PR.

## Alternatives considered

**Lines of code as the map's area, as block explorers draw it.** Rejected for
the node-only mode: it needs every file of every package, thousands of reads
per page or a crawler, and gnoweb runs no background work. Possible later
from the indexer, whose deploys carry the sources.

**Absolute positioning with inline styles, or CSS `attr()`.** The first is
blocked by the CSP; the second is not portable yet.

**Showing importers on the overview.** Rejected: a five-second indexer scan
on every overview view, including crawlers'.

**Treating every deploy that quotes the path as an importer.** Rejected: it
would list the whole of a batched deploy, and test-only quotes.

**A separate indexer flag for the map.** Rejected: `*indexer.Client` already
answers both features, and #6231 settled on one flag. The wiring asserts the
configured indexer to `chainmap.Indexer` and logs a warning if it cannot, so
the feature never switches off silently. Typing the config field as both
interfaces would be the deeper fix; it changes #6231's field, so it is left to
that PR.

## Consequences

### Positive

- Every listing can be seen as a whole, with no indexer and no script.
- The map and the list cannot drift apart, and a cut listing says so in both.
- With an indexer, activity and importers appear, tagged and dated; without
  one nothing mentions them.
- Importers are exact as of the chain's current state.

### Negative

- Without an indexer, area is package count, not code size.
- The first `$deps` of a heavily imported package is slow: about 40 s for
  `gno.land/p/nt/avl/v0` against the public RPC. The page says it is still
  looking, and a reload finds the cached answer.
- The map's labels are clipped, not ellipsized; small tiles carry no label,
  only their accessible name and tooltip.
- Activity spans a week estimated from the indexer's block times over the
  last 20 000 blocks.
- Importer scans and the activity refresh share the indexer client, its
  16 slots and its breaker, with search. Bands, concurrency and lookup limits
  keep them well under its timeout, but a separate client for aggregate
  queries would isolate them; that is a change to #6231's wiring.
- The provenance footer partial and its type copy omnisearch's, and reuse its
  styles; omnisearch can switch to them in a follow-up so the two cannot
  drift. Likewise omnisearch's `importers:` qualifier (unchecked mentions)
  could answer from `Handler.Importers`.

## Files

- `gno.land/pkg/gnoweb/feature/chainmap/`: layout (`squarify.go`,
  `tree.go`), activity (`activity.go`), importers (`importers.go`), the shared
  cache (`flight.go`), the map and dependencies pages (`view.go`, `deps.go`),
  their templates and stylesheet.
- `gno.land/pkg/gnoweb/indexer/aggregates.go`: `CallsBetween`,
  `DeploysQuoting`.
- `gno.land/pkg/gnoweb/components/`: List/Map tabs, `ExplorerView`, the
  graph partial `ui/pkg_graph`, and the shared partials
  `ui/listing_truncated` and `ui/indexer_status`.
- `gno.land/pkg/gnoweb/handler_http.go`: listing cap and truncation, `$map`
  and `$deps` routing, wiring.
