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
- The endpoint leaves the footer, the map, the dependencies page and the JSON
  answer (`indexer.url`). The freshness stays.
- The results page is laid out as a directory listing: the same grid
  (`SearchViewType`, registered like `DepsViewType`), `.b-pkg-section` headers
  and `.b-list` rows with an icon per kind of page. The provenance note stays
  under the rows, where a phone shows it. The page's own form goes: the
  omnibar is a form and submits here without JavaScript, and on this page it
  holds the query instead of the path. The keywords move to a native popover
  behind a "Keywords" button in `.header-info`, where the source page puts
  Copy and Download.
- The rail is always there, as a listing's is: the count per kind, each a
  link to its group, then the searches one click away. On a package page
  those are the package's other keywords that need no argument; without an
  answer, a few queries any chain answers. Without an indexer, only `imports`
  is such a keyword, and the rail falls back to the examples.
- One pill per group heading: an indexer group says so inside its count
  (`20 · indexer`), the overview's `.count` pill in the indexer tag's colour.
- The title repeats the query only when the group heading does not already
  name it: a bare keyword alone (`deploys`) is not repeated.
- A row is named and linked by its first decoded message. Opaque details (an
  address, a hash) are cut in their middle; prose is left to the stylesheet.

## Alternatives considered

- Fix tx-indexer instead. That is the root cause and gets an issue there, but
  gnoweb must not trust a filter it cannot see applied, as `aggregates.go`
  already concluded.
- Drop only `UnexpectedMessage` rows. A transaction carrying a session message
  and a call to another realm would still pass.
- Keep the keyword list on the page, restyled. It is help, not an answer.
  Suggesting keywords in the omnibar as one types would serve better still;
  it touches the omnibar on every page, so it is left to its own change.
- Make the results full width when there is no rail. Every other page keeps
  that column (table of contents, listing rail, file list), and rows that
  wide set a name 1200px away from its details.

## Consequences

- A deploy history, an activity list or an account history can come back
  shorter, and is then right.
- `indexer.url` disappears from the JSON answer; nothing in gnoweb read it.
- The search page takes the dev chrome of listings and source pages.
