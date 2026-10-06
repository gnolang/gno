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

An explorer listing (`/r/`, `/p/`, `/r/<namespace>/`, …) of at least ten
packages gains two tabs in the header slot where a package shows Content,
Source and Actions: **Directory** and **Map**. `$map` on the same path draws
the same listing. Below ten packages a map shows nothing the list does not:
the listing offers no tabs, and `$map` renders the list.

The two renderings share one page: the listing's header, and a side rail on
the realm grid, where a realm has its table of contents. The rail holds what
the listing contains (packages, folders), a filter that works on either
rendering (the overview's own filter controller; on the map it dims tiles
rather than removing them, so the map keeps its shape), and, on the map, its
key. Every rail section uses the overview rail's heading, meta-card and list
styles.

The listing lists the paths once and hands the same slice to either
rendering, so the two cannot disagree about what exists. It asks the node for
one more path than it shows: past `maxListedPaths` both renderings say "only
the first N paths are listed" (`ui/listing_truncated`) instead of passing for
complete. `qpaths` has no cursor, so stating the cut is all a listing can do.

The map is a squarified treemap on three levels: the first path segment
below the listed path, then the next one, then one tile per package. A
subgroup needs three packages to keep a name band of its own; smaller ones
are pooled with the group's loose packages. A tile is named below the
nearest band actually drawn, so a group too small for its band keeps its key
in its tiles' labels. Package count is the area: it is the one size the node
answers for a whole listing in a single query; lines or bytes would need
every file of every package.

A named box with packages below it links to its own map, so zooming is
navigation. The key offers the way back out, but only to a listing: the
handler checks, beside the listing query, that the parent path is not itself
a package, whose page is not a listing.

A map exists only where the list does, that is where no package lives at the
path. The package check runs beside the listing; when the node cannot answer
it, the map is an error, not a guess.

### SVG, laid out by the server, no script required

The map is one `<svg>` in a fixed coordinate space (960 × 600) that scales to
its column; every tile is an `<a>` named by its `<title>`. Because the space
is fixed, text sizes are known in the same units as the boxes, and every
label is fitted on the server to the character, shortened with an ellipsis,
or left out below four characters. The `viewBox` and each label's font size
are written from the layout's own constants, so the stylesheet repeats
neither.

Geometry and font sizes are attributes, not inline styles: gnoweb's
Content-Security-Policy is `style-src 'self'`, which drops `style` attributes.
The first version of this map rendered as an empty box for exactly that
reason, and a test now pins it.

Colors live in `chainmap.css` as custom properties set by modifier classes.
The classes are written out whole in Go so the production purge, which keeps
only names it finds in the sources, keeps them.

A small controller names the tile under the pointer or the keyboard focus in
a status line under the map; without it, each tile's tooltip says the same.
The status line is not a live region, since it changes on every pointer move.
A skip link lets keyboard readers pass the map's links at once.

### Activity, when an indexer is configured

On realm maps, tiles are colored by successful `MsgCall`s over the last seven
days, on a log scale relative to the busiest realm on that map. The key lists
the range of calls each shade stands for, and the five most called realms.

The window is read in bands of 25 000 heights aligned on multiples of that
width (`indexer.CallsBetween`); a band the indexer's element cap refuses is
split in half until it fits. Whole bands below the indexer's tip are final,
so they are kept across refreshes: a refresh reads the band the window now
starts in, the one it ends in, and any band closed since — at most four
instead of about twenty-five. Bands that leave the window are dropped.

The aggregate is cached for five minutes. A refresh is started by a request,
detached from it and shared by every concurrent reader; a stale aggregate is
served while it runs. A page waits three seconds, then says the activity is
still being counted, with a reload link.

A band still over the cap at the narrowest width makes the aggregate partial:
counts become lower bounds ("at least N calls") and a realm with no counted
call is drawn as unknown, never as zero. A pure-package map says that nothing
calls packages rather than showing an empty scale. Without an indexer the map
says nothing about activity; pale tints only tell neighboring namespaces
apart.

### The dependency graph

The overview's Imports section becomes a Dependencies graph: the package
between what it imports, read from `vm/qdoc`, and what imports it. The
section, its jump link and its table-of-contents entry share one condition and
one name.

Importers are not something the chain can list. With an indexer, a `$deps`
page proposes candidates by asking for every deploy quoting the import path
(`indexer.DeploysQuoting`), then checks each candidate's `vm/qdoc` imports on
chain. Both steps are needed: on gnoland-1 in October 2026, 11 of the 210
deploys quoting `gno.land/p/nt/avl/v0` were batched deploys carrying 127
packages between them, and the indexer matches a transaction as a whole; a
quote can also sit in a test or a string. Only the chain can say which
package imports it, so the side is labelled "Imported by" and contains nothing
the chain did not confirm. Without an indexer `$deps` is not a page: the
overview's graph already shows all there is, and the URL falls through to the
package.

Candidates are deploys holding the path as a Go string literal, interpreted
or raw. The scan walks the chain in bands of 75 000 blocks, two at a time, the
bottom band without a lower bound so it reaches genesis at height 0; the first
failing band cancels the others. One query over the whole chain took 4.9 s on
gnoland-1, past the indexer client's 4 s request timeout.

The lookup costs that scan plus a node read per candidate — about 38 s for
`gno.land/p/nt/avl/v0`'s 265 candidates against the public RPC, which found
183 importers — so it is a page of its own: never run by the overview,
`noindex`, capped at 400 candidates, limited to six lookups a minute per
address (burst three) and to two lookups at once across all readers. The slot
and the limiter are taken inside the shared fetch: only the reader who starts
a lookup pays, a cached answer or joining one in flight is free, and a lookup
finding both slots taken is refused at once rather than queued. Neither
refusal is remembered. Behind a reverse proxy this limit, like gnoweb's
others, needs `-trusted-proxies` to tell readers apart.

The package's own imports are read first, through the same cache as the
candidates', so a path with no live package answers 404 without costing the
indexer anything. Answers are cached per package for ten minutes. A reader
waits at most 12 s; the lookup runs detached for up to 60 s and a reload finds
it. Only the reader's own wait reads as "still looking": a lookup that failed
reads as unavailable. An answer some candidate reads failed to complete is
kept one minute rather than ten. Past any of its bounds — a band the indexer
capped, the candidate cap, an unread candidate — the count reads "at least",
and the page says whether a reload can help. On a wide screen the package node
stays in view while a long side scrolls.

### One cache shape

Activity, importers and imports share one small generic cache (`flight`):
answers per key, a fetch started by a reader but detached from it and shared
by every concurrent reader, a TTL, a size bound evicting in store order, a
short memory of failures so a failing indexer is not asked again by every
reader, and, for activity, the last good answer served while a refresh runs.

A fetch past its budget is cancelled, not given a deadline. The indexer
client's breaker, shared with search, counts a deadline as the indexer failing
and ignores a cancellation; a slow aggregate must not turn search off. No band
query starts after its fetch has given up, for the same reason.

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

**Percentage-placed nested `<svg>`s, clipping labels.** The first version.
Clipped names read as a rendering fault, and the text size could not be known
against the boxes. Replaced by the fixed space and server-fitted labels.

**Absolute positioning with inline styles, or CSS `attr()`.** The first is
blocked by the CSP; the second is not portable yet.

**Shortening addresses in the middle (`g1abc…xyz`) in labels and lists.**
Rejected: a start…end short form is what a lookalike vanity address imitates
(the rule #6206 applies to `/u/`). Labels end with an ellipsis, and lists
wrap.

**Showing importers on the overview.** Rejected: a five-second indexer scan
on every overview view, including crawlers'.

**Treating every deploy that quotes the path as an importer.** Rejected: it
would list the whole of a batched deploy, and test-only quotes.

**A separate indexer flag for the map.** Rejected: `*indexer.Client` already
answers both features, and #6231 settled on one flag. The wiring asserts the
configured indexer to `chainmap.Indexer` and logs a warning if it cannot, so
the feature never switches off silently.

## Consequences

### Positive

- Every listing of ten packages or more can be seen as a whole, with no
  indexer and no script.
- The map and the list cannot drift apart, and a cut listing says so in both.
- With an indexer, activity and importers appear, tagged and dated; without
  one nothing mentions them.
- Importers are exact as of the chain's current state.
- An activity refresh reads a handful of bands, not the whole week.

### Negative

- Without an indexer, area is package count, not code size.
- The first `$deps` of a heavily imported package is slow: about 40 s for
  `gno.land/p/nt/avl/v0` against the public RPC. The page says it is still
  looking, and a reload finds the cached answer.
- Small tiles carry no label, only their accessible name and tooltip.
- Activity spans a week estimated from the indexer's block times over the
  last 20 000 blocks.
- `$map` costs one more node read than the list, to check the parent before
  offering a way back out.
- Importer scans and the activity refresh share the indexer client, its slots
  and its breaker, with search. A separate client for aggregate queries would
  isolate them; that is a change to #6231's wiring.
- The provenance partial copies omnisearch's; omnisearch can switch to it in a
  follow-up.

## Files

- `gno.land/pkg/gnoweb/feature/chainmap/`: layout (`squarify.go`,
  `tree.go`), activity (`activity.go`), importers (`importers.go`), the shared
  cache (`flight.go`), the map and dependencies pages (`view.go`, `deps.go`),
  their templates and stylesheet.
- `gno.land/pkg/gnoweb/indexer/aggregates.go`: `CallsBetween`,
  `DeploysQuoting`.
- `gno.land/pkg/gnoweb/handler_map.go`: wiring, listing, `$map` and `$deps`
  handlers.
- `gno.land/pkg/gnoweb/components/`: Directory/Map tabs, `ExplorerView` and
  the listing rail, the graph partial `ui/pkg_graph`, and the shared partials
  `ui/listing_truncated` and `ui/indexer_status`.
- `gno.land/pkg/gnoweb/frontend/js/controller-map.ts`, and an optional class
  on `controller-filter.ts`.
