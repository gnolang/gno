# ADR: gnoweb indexer search, rows the filter did not match, and a search page laid out as a listing

## Context

On onyx, `$search&q=deploys` for `r/gnoland/wugnot` listed 20 transactions,
none of them a deploy of wugnot. Each was labelled `UnexpectedMessage`. They
were `auth` `create_session` and `revoke_session` messages, which tx-indexer
cannot decode, and tx-indexer lets an undecoded message through any message
filter. `aggregates.go` already re-checked its rows for that reason (19 of 20
rows wrong on gnoland-1); `recent`, behind `deploys`, `activity`, `account:`,
`content:` and `importers`, did not.

The provenance footer also printed the indexer's endpoint. On onyx that is
`http://tx-indexer.onyx.internal:8546`, a host no reader can reach.

The results page itself did not look like the rest of gnoweb: a full-width
article with a second search form under the omnibar, green monospace rows, and
the list of search keywords printed under the results as if it were one.

## Decision

- `recent` takes a `match func(Message) bool` and keeps only the rows it
  accepts, before they count towards the limit, so the walk goes on to the
  next band for real rows. Each query passes the check matching its filter:
  path for `RecentByPackage` and `Deploys`, the address in the field each
  message kind names it in for `RecentByAddress`, the message type for the two
  source searches, whose body is not selected.
- Every `where` also names its message's `typeUrl`. tx-indexer applies that one
  to a message it could not decode, so the rows stop being returned at all: on
  gno.land, wugnot's deploy filter gets 26 rows without it, 25 of them session
  messages, and exactly its deploy with it. `match` stays: gnoweb does not trust
  a filter it cannot see applied. `typeUrl` became filterable with `value`, so
  any indexer that accepts these queries accepts it.
- A band that hits the indexer's element cap no longer ends the walk. The rows
  the cap kept may all be dropped, and the rows below it were not searched, so
  the next band starts there, at the page's lowest block, read again whole
  since an older tx-indexer ends a page inside a block. Those rows stay in hand
  until it answers, as the walk may end first. It stops at once only when one
  block fills the cap, which heights cannot page past.
- The endpoint leaves the footer, the map, the dependencies page and the JSON
  answer (`indexer.url`). The freshness stays.
- The results page is laid out as a directory listing: the same grid
  (`SearchViewType`, registered like `DepsViewType`), `.b-pkg-section` headers
  and `.b-list` rows with an icon per kind of page. The provenance note stays
  under the rows, where a phone shows it. The page's own form goes: the
  omnibar is a form and submits here without JavaScript, and on this page it
  holds the query instead of the path. There a click leaves the caret where it
  lands, `/` only moves it to the end, and Enter on the query it holds searches:
  the URL fills the bar, so that query must never navigate, or a link could
  send the reader off the site with one key. It is read back from the bar: a
  text input drops a newline the URL kept, so comparing with the URL's query
  would let `//evil%0a.com` through. The keywords move to a native
  popover behind a "Keywords" button in `.header-info`, where the source page
  puts Copy and Download; an empty search prints them instead.
- The rail is always there, as a listing's is: the count per kind, each a
  link to its group, then the searches one click away. On a package page
  those are the package's other keywords that need no argument, on a
  whole-chain answer too, since they still work there; without an answer, a few queries any chain answers. Without an indexer, only `imports`
  is such a keyword, and the rail falls back to the examples.
- One pill per group heading: an indexer group says so inside its count
  (`20 · indexer`), the overview's `.count` pill in the indexer tag's colour.
- The title repeats the query, whole, only when the group heading does not
  already name it: a bare keyword alone (`deploys`) is not repeated.
- A pasted address is a search for that account: with an indexer, its
  activity, followed by what discovery finds for it (its packages and a
  Users row linking `/u/<address>`); without one, discovery alone. Before,
  only the paths deployed under that address showed, and nothing pointed to
  `account:`. A bad checksum gets no user link, and a qualifier beside the
  address takes the row out with the rest of what it narrows. The row needs no
  listing, so it survives a failed one, and a discovery that failed beside the
  account search still says so. The omnibar shows what its path list knows
  while the account's walk runs.
- Discovery ranks its matches before keeping ten, and the omnibar ranks its
  local matches with the same rules, so both show the same order, and the
  Users group follows the ranked paths. Throwaway deploys go last: a raw
  address namespace (unless it is the query), a namespace numbered after a test
  (`moultest123`), or a segment that names one, as a word of its own (`test`,
  `avl_test`) or a helper (`testutils`), not as letters inside `latest` or
  `attestation`. A number only counts in a namespace: a package may be
  `latest2`, and `contest2026` as a namespace is the price of `moultest123`.
  Then by relevance: the package's name (last segment
  before any `/vN`) equal to the query, a segment equal to it (a namespace
  typed whole), the name starting with it, containing it, a segment starting
  with it, anywhere; a shallower path first. Then the shortest path. The
  listing is alphabetical, which put `/p/g1…/avl` ahead of `/p/nt/avl/v0`.
  There is no list of official namespaces to keep up to date.
- A row is named and linked by its first decoded message. The resolver says
  when a detail ends in an address or a hash, which is cut in its middle; prose
  is left to the stylesheet. It also says what a row names when its link does
  not, which picks the icon: a transaction links to the package it touched, and
  a block or a standard library import links nowhere.

## Alternatives considered

- Fix tx-indexer instead. That is the root cause and gets an issue there, but
  gnoweb must not trust a filter it cannot see applied, as `aggregates.go`
  already concluded.
- Drop only `UnexpectedMessage` rows. A transaction carrying a session message
  and a call to another realm would still pass.
- Resume a capped band strictly below its lowest height, which a tx-indexer
  cutting pages to whole blocks allows. An older one ends a page inside a
  block, and the rest of that block would be lost; reading it again costs one
  block on either.
- Drop the omnibar's bech32 checksum and send an address to the server. It
  would remove a copy of what `crypto.AddressFromBech32` does, but a pasted
  address would cost a request without an indexer, and show nothing while the
  account's walk runs with one.
- Keep the keyword list on the page, restyled. It is help, not an answer.
  Suggesting keywords in the omnibar as one types would serve better still;
  it touches the omnibar on every page, so it is left to its own change.
- Rank with a list of official namespaces (`nt`, `gnoland`, `sys`…). It
  would need upkeep and would bury a new team's packages; the throwaway
  signals are in the paths themselves.
- Scope a plain word typed on a realm page to that realm. The omnibar sends
  every query from the page path, so `author:` or a word on a realm page
  would stop finding anything else. A group of the package's own matches
  beside the chain-wide ones is the better answer, and costs node queries
  per search, so it is left to its own change.
- Make the results full width when there is no rail. Every other page keeps
  that column (table of contents, listing rail, file list), and rows that
  wide set a name 1200px away from its details.

## Consequences

- A deploy history, an activity list or an account history can come back
  shorter, and is then right.
- A capped band can be followed by more bands, each bounded by the same step
  count, element cap and time budget. With `typeUrl` the cap is reached by real
  rows, which fill the limit first.
- On gno.land's public indexer the band down to genesis takes 8 s, against 15 s
  without `typeUrl`, over the page's 6 s budget. A package deployed that long
  ago answers "timed out" there rather than listing its deploy; that is not
  new, and it reads as a failure, not as "never deployed".
- `indexer.url` disappears from the JSON answer; nothing in gnoweb read it.
- The search page takes the dev chrome of listings and source pages.
