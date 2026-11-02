# ADR: Package cards in gnoweb listings

## Status

Proposed.

## Context

A listing such as `/r/gnoland` showed one line per package: its path and four
links. Nothing said what a package does, so a reader who did not already know
the path had no reason to open it. The chain already holds most of what a
reader needs to choose: a package's doc comment, the functions it exposes, and
the block it was deployed in. With an indexer, gnoweb also knows how much each
realm is used, which `feature/chainmap` computes for the map since #6283.

Measured on gnoland-1 in November 2026, 22 of 30 realms picked at random carry
a doc comment, so a description from the doc comment covers most listings
without asking authors for anything.

## Decision

### One card per package, in the listing's list rendering only

The explorer listing draws each package as a card: its path, the opening of its
doc comment, a few facts as tags, a line of details, and the four links it
already had (Open, Browse, Source, Action; a pure package keeps Browse and
Source). The map rendering, the package's own directory and the rail are
unchanged. The cards replace the explorer branch of `ui/dir_list`, which now
lists a package's files only, and `DirLinkType`, whose explorer case had no
other user, goes with it.

### Every field comes from the chain, or from the indexer when one is set

| Field | Source | Without an indexer |
|---|---|---|
| Description | `vm/qdoc` package doc, first sentences up to 240 runes, package clause dropped | shown |
| Actions, `read-only`, `token` | `vm/qdoc` exported crossing functions | shown |
| Block, or Genesis | `vm/qpkgmeta_json` height | shown |
| Date | indexer `Block(height)` | absent |
| Weekly users, Used … ago | `chainmap` activity aggregate | absent |

`token` means the realm exposes `Approve`, `Transfer` and `TransferFrom`: it
describes the API, not the token's worth. No tag is inferred from free text or
imports: an import of `r/gov/dao` does not make a realm a DAO, and a word list
would mislabel as often as it helps. Author-chosen categories wait for Explore
(#6273), whose `Category` and `Tagline` are the place for them.

Weekly users are the distinct addresses that called the realm over the last
seven days, as on the home page's "Top realms". "Used … ago" estimates the time
of the latest call from the window's average block rate, anchored on the tip
block's time rather than the reader's clock, so a cached aggregate never dates
an old call as recent. It costs no query beyond the tip block the window
already reads.

### The cards live in `feature/chainmap`

`chainmap` already owns the other rendering of the same listing and the
activity aggregate. Putting the cards there reuses its `flight` cache and its
indexer wiring, and lets the map and the list share one activity scan. The
feature gains one optional dependency, `PackageReader`, implemented by the same
adapter that reads imports (`chainReader`, formerly `importReader`).

### A listing never waits more than 1.5 s for its cards

Each card's summary is read through a `flight` keyed by path: concurrent
readers share one fetch, the doc comment is reduced to its description once,
in the fetch, and a failure is remembered a minute. An answer is fresh for a
minute, then served while it is read again: a package can change under the
same path (a private realm redeployed, any realm reloaded by gnodev), and a
listing still renders at once.

A listing waits at most `cardsWait` (1.5 s). A card not ready by then shows
its path and links, and its fetch finishes for the next reader. Because a
fetch outlives its request, a process-wide cap (`maxCardFetches`, 8) bounds
them across every reader, chain reads and indexer block lookups together: a
quarter of the RPC client's default and half of the indexer client's, so a
flood of cold listings never takes the slots the map, search and pages need.
It refuses at once rather than queue, as `importers` does for its lookups: a
card refused a slot stays blank and is not remembered.
`TestCards_FloodStaysUnderTheCap` runs 40 cold listings of 50 packages at
once and checks that no more than 8 reads ever run and every listing returns
within its budget.

What a request can make gnoweb do is bounded by data it does not choose: the
cards' cache keys are package paths from the node's listing (at most 1,000)
and deploy heights, never a value from the URL, and each cache holds at most
4,096 entries. The activity scan is one shared aggregate.
The week's activity is read beside the packages, never before them, so a cold
indexer scan cannot spend the wait; a listing of pure packages, which are
never called, does not ask for it. A cold listing of 142 packages filled 15,
29, 43 then 61 descriptions over four loads against gnoland-1; a warm one
renders as fast as before.

### Plain text only

The description is the plain text of the doc comment's leading paragraphs,
read the way `markdown.Lead` reads a page: image alt text is skipped, format
characters (bidi overrides, zero-width joiners) are dropped, raw HTML is not
kept, and the template escapes what remains. The doc comment is untrusted, so
`markdown.Summary` parses only its first 8 KiB, under the realm renderer's
emphasis guard. Tags come from a fixed set.

### Built from what gnoweb already has

The buttons are `.b-btn` with a modifier for the touch size; the title is an
`h2` so screen readers can jump card to card, as the user page's cards do; the
actions are a labelled group, not a landmark per card. The explorer listing's
`.b-list` action layout, now unused, is removed, and package files keep their
plain rows.

## Alternatives considered

- **README as a fallback description.** One more file read per package for the
  packages without a doc comment. Left out to keep a card to two RPCs.
- **`r/nt/grc20reg` for the `token` tag.** Authoritative, but the registry
  exposes no cheap query by path: it would mean parsing its `Render`.
- **A client-side fetch per visible card.** Avoids the wait entirely, but adds a
  JSON endpoint and a controller for what the server can render.
- **Deploy dates from the node.** The node has block times, but `ClientAdapter`
  has no block query; the indexer already answers `Block(height)` and caches it.
- **A visible `indexer` tag on activity.** Dropped from the design review: the
  line only exists when an indexer is set, and its tooltip names the source.

## Consequences

### Positive

- A reader sees what a package is before opening it, in both themes and on a
  phone, with the same four actions as before.
- Without an indexer, a card is complete for what the chain says; with one, it
  adds activity without another scan.

### Negative

- A cold listing waits up to 1.5 s, and a large one fills over several loads.
- A doc comment can say anything: the card shows a claim, under the package's
  real path, not a verified fact. GovDAO-curated text comes with Explore.
- Five icons move from the vendored markdown set into the chrome sprite, so
  the cards and realms share one definition each (`icons_gen.go` is unchanged).

### Files

- `feature/chainmap/cards.go`: `Cards`, `PackageReader`, caches, description
  and tag rules; `activity.go` tracks each realm's latest call.
- `components/view_directory.go`, `components/ui/pkg_cards.html`: the card
  data and markup; `views/directory.html` uses them.
- `markdown/summary.go`: `Summary`, the plain-text opening of a document.
- `handler_map.go`: `chainReader` and the wiring.
- `frontend/css/06-blocks.css`: `.b-pkg-cards`, `.b-pkg-card`.
- `components/ui/icons.html`, `markdown/icons/`: the moved icons.
