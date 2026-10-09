# ADR: a realm's overview states its storage and its recent calls

## Context

A realm's overview page shows its code: symbols, files, imports. It never
shows what the realm weighs on chain, nor whether anyone uses it.

- Every realm keeps state between transactions and pays for it up front:
  the chain locks a storage deposit, at its storage price per byte, from
  whoever writes the data, and refunds it when the data is freed. That is
  the realm's real weight and part of the cost its users carry (on
  gnoland-1 in October 2026, `r/gnoland/blog` held 1.29 MB for 129.27
  GNOT). The node answers it in one query, `vm/qstorage`.
- Who calls a realm, and what, is not something the node can list. An
  indexer can (#6231), and gnoweb's listing map (#6283) already scans the
  last 7 days of calls, every call of the chain, to color its tiles.

gnoweb runs on its node alone, with no background work; an indexer may only
add data, marked as such and absent without `-indexer-url`.

## Decision

### Storage, from the node

The Package Info card of a realm gains two rows: Storage (the stored size,
in decimal units, the ones the chain prices) and Storage deposit (in GNOT,
with cents; a deposit under a cent says so rather than "0.00"). The query
runs beside the overview's others, for realms only: a pure package keeps no
state, and the node answers "realm not found" for one. A node that cannot
answer leaves the rows out, never a zero that would read as an empty realm.
The label is "Storage deposit", not "Deposit", which a realm's own
`Deposit` function (wugnot) would make ambiguous; the tooltip names the
chain's storage price without hard-coding it, since it is a parameter.

### Recent calls, from the activity scan

With an indexer, the overview of a realm lists its last calls in a Recent
calls section, marked as indexer data with the provenance footer. A row is
one transaction into the realm: the functions it called there ("Approve,
Deposit"), the full caller address (never shortened, linked to its user
page), the transaction's gas and how many calls it held, a failed tag, when,
and its block.

The rows come from the 7-day scan the map already makes. Each band of that
scan keeps every realm's newest calls (the scan now also selects the
function name), and the aggregate keeps the newest eight overall. The
overview reads the shared aggregate, cached for five minutes and served
stale while it refreshes. There is one scan per five minutes whatever the
traffic, started by whichever reader first finds the aggregate stale, map
or overview; no page adds a query of its own, and the path read changes
nothing, so a path nobody calls, or that does not exist, costs nothing.
Before an aggregate first exists a reader does not wait for it (the scan
takes seconds): the section says calls are still being counted, and the
scan carries on for the next reader. This also applies to the map.

A failed call costs its sender little and may name any realm path, so it is
kept only beside a successful call into the same realm in the same band,
and at most two per realm: failed calls can neither grow the aggregate with
invented paths nor push a realm's real calls out. A function name, a Go
identifier of any length, is kept to 64 characters.

The cost of that choice is the window: a realm with no call in the last 7
days says "No calls in the last 7 days.", rather than listing older calls.
A window the indexer could not read in full says calls may be missing, and
never claims none.

The indexer gives no time per transaction, so a call's time is placed from
its height between the times of the window's first and last block, which
the scan already reads; blocks come at a steady rate. The row shows how
long ago in a `<time>` element whose title says it is estimated, and the
block itself on the line below.

The section exists only for a realm (a pure package cannot be called) and
only at the latest height: a page pinned to a past height would otherwise
show calls made since. Without an indexer it does not exist at all; while
the aggregate is first computed it says calls are still being counted; when
the indexer cannot be read it says so.

### Gas per row

A row shows the whole transaction's gas, as measured (154 M, 120 k), with
"tx of N calls" when the transaction held more calls than this realm's:
the indexer reports gas per transaction, and splitting it on a single row
would invent a precision the data does not have. One row per transaction
also keeps a transaction calling the realm twice from showing its gas
twice. The map, which aggregates, shares a transaction's gas between its
working messages instead.

## Alternatives considered

**A lookup per realm** (`RecentByPackage`, then the blocks' times), cached
per package. The first version. It cost up to four widening indexer scans
per uncached overview, for any path a client could invent, detached from
the request, and could fill the indexer client's slots or trip its breaker,
shared with search. Gating it on the realm existing and rate-limiting it
would have reduced, not removed, that; reading the aggregate removes it.

**Rows without an indexer**, from the node. The node keeps no call history.

**Exact times** with one block query per page. Precise to the second, at a
query per page; the estimate is within the block-time variance and free.

**Splitting gas per call on the rows**, as the map does. Rejected for a
single row, where an even split reads as a measurement it is not.

## Consequences

### Positive

- Every realm states its weight on chain, on any gnoweb, from consensus
  data.
- With an indexer, a realm shows that it is used, by whom and at what cost,
  at no indexer cost per page.

### Negative

- Calls older than the 7-day window are not listed.
- Times are estimates, within the block rate's variance.
- Recent calls refresh with the activity aggregate (five minutes), not live.
- Overview traffic alone now keeps the 7-day scan refreshed, where only the
  map did before: still one scan per five minutes.
- A realm whose only recent calls failed shows none.
- One more node query per realm overview (`vm/qstorage`, a single record
  read).

## Files

- `gno.land/pkg/gnoweb/client.go`, `client_mock.go`: `Storage`,
  `RealmStorage`, `parseStorage`.
- `gno.land/pkg/gnoweb/components/overview_storage.go`,
  `overview_calls.go`, `views/overview.html` (`ui/pkg_calls`, sidebar
  rows, jump link), `ui/icons.html` (`ico-pulse`).
- `gno.land/pkg/gnoweb/feature/chainmap/calls.go`, `activity.go`: the
  recent calls kept by the scan, their times, the section.
- `gno.land/pkg/gnoweb/indexer/aggregates.go`: the scan selects `func`.
- `gno.land/pkg/gnoweb/handler_http.go`: realm and height gating.
- `gno.land/pkg/gnoweb/frontend/css/06-blocks.css`: the section's rows.
