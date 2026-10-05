# ADR-004: A discovery store for gnoweb

> The subject is discovery: how a reader who does not already know a path
> finds the apps built on gno.land, and how an app that is not this week's
> favourite keeps being found. It must make people want to explore, need
> almost no human moderation, and assume every listing is hostile.

## Status

Proposed. Not consensus-affecting for gnoweb, which stays a read-only
presentation layer. The registry realm it reads,
`gno.land/r/gnoland/store/v0`, is ordinary on-chain code governed by GovDAO
through proposals it builds. The UI calls it **Explore**; code identifiers
keep "store".

Builds on ADR-003 (state explorer: server-rendered first, feature modules own
their URLs, resource bounds), #6191 (community-realm notice and
`TrustedPaths`) and #6231 (optional indexer backend). The indexer-backed parts
depend on #6231 landing; nothing else does.

## Context

gnoweb answers one question well: "show me this path". It does not answer
"what is there?". A reader who arrives on `/` sees the home realm, the
omnibar and the footer. Finding an app means already knowing its path,
following a link someone curated by hand, or leaving gnoweb for a third-party
explorer.

The ecosystem is now large enough that this is the main gap. Third-party
realms deploy on mainnet under `g1…` and `nym-…` namespaces (#6191), and
nothing on gno.land itself helps them get found. gnoscope shows both the
demand and one way to meet it:

| gnoscope page | What it lists | Source |
|---|---|---|
| `/realms` | every deployed `/r/` realm, ranked by calls, callers, failures | tx-indexer, raw |
| `/packages` | distinct `/p/` libraries, with importers | tx-indexer, raw |
| `/apps` | "what is built here": a picture grid, one card per app, with category, blurb, screenshot | ranked by usage, then corrected by hand (`apps.json`, `moderation.toml`), plus `awesome-gno` for off-chain apps |

gnoscope's model is "the chain proposes, curation corrects". It works, but
its curation is a person editing a JSON file in a GitHub repo, and that does
not scale to the chain's front door.

There is prior art on gno.land itself. The **Hall of Realms**
(`r/leon/hor`, first `r/demo/hof`; #2842, #3674, #3709, #4087) let a realm
register itself from `init()` with a title and description, collected up-
and downvotes, and sorted by votes or recency, with an owner who could pause
the realm and delete entries. Issue #3310 ("Brainstorming features for Hall
of Fame") asked for most of what this ADR specifies: categories and tags,
sorting and filters, featured realms, usage stats, logos, and a gnoweb
explore page. The realm now sits in `examples/quarantined/` with the other
unaudited examples (#5726). Its reviews left three rules this ADR adopts:
`Register` must never panic (moul, #2842; Leon, #3674), someone must be able
to pause and clean up (an admin then, "later … DAO-managed"; GovDAO here), and sort orders should be kept in avl indexes rather
than recomputed. Its gaps are the threats below: votes from any caller,
upvote and downvote cumulable, no defence against sybil voting, and free text
shown as is.

#### Maintainer constraints

Earlier reviews and issues set constraints this design must meet:

| Constraint | Source | How the design meets it |
|---|---|---|
| "It shouldn't be "gnoland"; it's not "official" or "useful" enough, and it allows too much permissionless activity." | moul, [#2842](https://github.com/gnolang/gno/pull/2842#discussion_r1784897476) | A conscious decision to keep `gnoland`: gnoweb, an official surface, points at the realm (`-store-realm`), and every permissionless write is bounded (deposits, 5 listings per namespace, confirmation, ranked stars). See §1. |
| "The admin should be able to pause the contract completely and also have the ability to delete some entries. […] Later, we can revisit the hof contract to make it DAO-managed." | moul, [#2842](https://github.com/gnolang/gno/pull/2842#discussion_r1784900141) | DAO-managed from day one: GovDAO pauses and hides (never deletes) through proposals the realm builds (§1). |
| "please, try to make minimalist contracts; that's what we need at the moment" | moul, [#2842](https://github.com/gnolang/gno/pull/2842#discussion_r1784961429) | Not met in size; the full scope is kept on purpose, and §1 "Why it is not minimal" says what each part buys and what could be cut. |
| "`Register` should not panic. […] It's similar to a "Set" or "Upsert" […] or sending a "wakeup" UDP packet" | moul, [#2842](https://github.com/gnolang/gno/pull/2842#discussion_r1801271863) | `Register` is an upsert that never panics; a refusal is an event (§1). |
| "if you panic, you will fully block execution of the realm that called Register" | leohhhn, [#3674](https://github.com/gnolang/gno/pull/3674#discussion_r1940991269) | Same. |
| Registration "handled by the contract itself", plus `AdminRemoveEntry` and `AdminPauseSubmissions` | moul, [#2772](https://github.com/gnolang/gno/issues/2772) | The app realm lists itself; GovDAO can pause and hide. |
| "let anyone propose content, but a curation team will have the ability to approve, refine, or reject entries" | moul, [#3928](https://github.com/gnolang/gno/issues/3928) | Anyone lists; curation is automatic (tiers, confirmation, ranked stars) with GovDAO hiding after the fact. GovDAO is the curation team for the one editorial act there is: a voted, labelled, time-bounded pick (§1). No approval queue: it would reintroduce oblivion through backlog (Alternatives). |
| "An example of a section that should be dynamic. It can remain static until we have a realm to support this." | moul, [#3888](https://github.com/gnolang/gno/pull/3888#discussion_r2046450156) | The store is that realm: its markdown render can feed a home-realm block (§5). |
| "ensure search engines can distinguish between Gno.land's main pages and user-generated realms, preventing harmful content from affecting the platform's reputation" | moul, [#3910](https://github.com/gnolang/gno/issues/3910) | No owner free text or links; community apps say so after their path and carry #6191's notice, while only provenance badges (Core, Established) mark the others; owner images only for trusted or earned tiers (§2). |
| "a new helper […] that remains resource-free when not utilized" | moul, [#997](https://github.com/gnolang/gno/issues/997) | The store is absent, not broken, without `-store-realm`: the handler is nil and costs nothing (§6). |
| Metadata is auto-derived only: no developer-specified fields the page shows invisibly (the rationale for reverting #3924) | moul, [#3924](https://github.com/gnolang/gno/pull/3924) | The store's page metadata comes only from fixed gnoweb strings and the closed taxonomy, never from listing taglines or titles (§6). |
| An exhibition widget other realms can render, an importable helper to embed the list, a `RenderBlock` | moul, [#2842](https://github.com/gnolang/gno/pull/2842#discussion_r1784964302), [#3928](https://github.com/gnolang/gno/issues/3928), [#3249](https://github.com/gnolang/gno/issues/3249) | `Block(list, n)` and `Badge(path)`: read-only, non-crossing helpers that return sanitized markdown, never a pointer (Read API). A `Badge` replaces any star count an app would paste: the store computes it. |
| "A scraper to index realms is a good approach for a third party service, but it should not be part of our core offering at least for now" | thehowl, [#3518](https://github.com/gnolang/gno/issues/3518#issuecomment-2595880518) | No crawler: apps list themselves; the indexer stays optional (#6231). |

Four forces shape this ADR:

1. **Popularity lock-in.** A store ranked only by usage shows the same ten
   apps forever. New and niche apps never get the first visit that would let
   them rank.
2. **Oblivion.** An app listed once and never featured again is invisible
   after its first week, even if it is good, alive and maintained.
3. **Moderation cost.** Every free-form field (image, link, rich text) is
   something a person has to watch. A store that needs daily human review
   will either rot or become a full-time job.
4. **Hostility.** Deployment is permissionless. Listings will include
   impersonators, phishing links and bait-and-switch apps, and a store
   surface makes them look endorsed.

Forces 3 and 4 push towards a bare directory, and forces 1 and 2, together
with the goal of making people *want* to explore, push the other way. The
decision below resolves this by being strict about **where content comes
from** rather than by having less of it.

Constraints inherited from gnoweb:

- **No mandatory external dependency.** gnoweb must run against one RPC node
  and nothing else (#2799). An indexer is an optional enhancement (#6231).
- **Stateless and server-rendered.** No background jobs, no database, no
  sessions. Pages work without JavaScript (ADR-003).
- **CSP.** `img-src` is `'self' data:` plus a fixed host allowlist
  (`cspImgHost`, `gno.land/cmd/gnoweb/main.go`). Other image URLs will not
  load.
- **gnoweb serves many chains.** Whatever this adds must be absent, not
  broken, on a chain without a store realm.
- **Trust is the operator's decision** (#6191): the `-trusted-paths` list,
  decided by namespace, maintained by reviewed PR, removable by flag.

## Decision

### Design rules

Every mechanism below follows four rules. They are what keeps human
moderation near zero without making the store poor.

1. **Nothing the store realm holds is trusted.** It is permissionless input.
   gnoweb validates and escapes every field, and decides what to show.
2. **Trust is gnoweb's decision, not the realm's.** The realm stores
   everything; gnoweb displays rich fields only for listings its own
   `TrustedPaths` (or the realm's earned-reputation signal, see §2) allows.
   One trust list, one owner, same as #6191.
3. **Richness comes from safe sources first.** Live chain data, the app's own
   realm page, and visuals generated by gnoweb carry no new moderation
   burden, so they are available to everyone. Free-form owner content (images,
   external links) is the only risky source, so it is gated.
4. **Every rich right is revocable in a minute without a transaction.** An
   operator flag removes a listing or a namespace from every rich surface
   immediately.

### Overview

```
          ┌──────────────────────────── gno.land chain ──────────────────────────┐
          │  r/gnoland/store/v0  listings · taxonomy · stars · shelves          │
          │     ▲ Register(cross, …)   app realms list themselves               │
          │     ▲ GovDAO proposals (pause, hide)                                 │
          │  r/sys/names · r/sys/users   namespace ownership, registered names  │
          └─────┼────────────────────────────────────────────────────────────────┘
                │ vm/qrender "api/v1/…"   one call per page, cached
                │ vm/qpkgmeta_json, vm/qfile gnomod.toml   liveness, private flag
          ┌─────▼──────────────────────────────────┐ GraphQL ┌─────────────┐
          │ gnoweb feature/store                   │◄──────►│ tx-indexer  │
          │  tier = f(TrustedPaths, namespace,     │optional└─────────────┘
          │           reputation, private, deny)   │
          │  Content tab of the store realm,       │
          │  header icon, /explore alias           │
          └────────────────────────────────────────┘
```

Four parts:

1. **A registry realm**, `r/gnoland/store/v0`: listings, a closed taxonomy,
   stars, GovDAO pause and hide. It renders a markdown store on
   its own and exposes a versioned JSON API.
2. **Trust tiers computed by gnoweb**: what each listing may show is a pure
   function of the operator's trust list, the namespace kind and automatic
   reputation.
3. **A gnoweb feature module, `feature/store`**: a richer view of the store
   realm's own Content tab, a header entry ("Explore") and an `/explore`
   alias. The store
   is a realm; gnoweb renders it better. It has no route, no view mode and
   no app page of its own.
4. **Indexer-backed signals**, registered only with `-indexer-url` (#6231).

### 1. The registry realm

#### Where it lives: `r/gnoland/store/v0`, not `r/sys/store`

`r/sys/*` is "contracts designed for chain use … minimal, future-proof, and
gas-efficient" (`examples/gno.land/r/sys/README.md`): names, params,
validators, users, the things the chain itself relies on. A store is
editorial data. Its schema will evolve, its content changes daily, and
nothing in consensus reads it. Putting it in `sys` would weaken what that
namespace promises.

Being read by gnoweb does not make it a system realm. gnoweb already depends
on `r/gnoland/home` and `r/gnoland/pages` through `DefaultAliases`
(`gno.land/pkg/gnoweb/app.go`), and `gnoland` is in `defaultTrustedPaths`.

It borrows governance from `sys`: GovDAO pauses and hides through
proposals the realm builds, like `r/sys/users` and valopers.

**Versioned, because it cannot be upgraded.** A public realm cannot be
redeployed. Every constant (thresholds, delays, quotas, sizes) and the
taxonomy are therefore permanent for a given path, so the path carries a
version, as `boards2/v0` and `namereg/v0` do. A future `v1` is a new path
with a documented migration: apps call `Register` on it again, and stars do
not carry over.

**Why `gnoland`, despite the Hall of Fame review.** When the Hall of Fame
was proposed as `r/gnoland/hof` (#2842), moul refused the namespace: "It
shouldn't be "gnoland"; it's not "official" or "useful" enough, and it
allows too much permissionless activity"
([review](https://github.com/gnolang/gno/pull/2842#discussion_r1784897476)).
It moved to `r/demo/hof`, then `r/leon/hor`. Keeping `gnoland` here is a
conscious decision, on two grounds. It is official enough: gnoweb, an
official surface, points at it through `-store-realm` and the header. And
its permissionless writes are bounded: storage deposits, at most 5 listings
per namespace, claims that buy no exposure until confirmed, and rankings
that count only ranked stars. The other half of the objection, what
`gnoland` vouches for, is answered by construction rather than by
promise:

| HoF's problem | What the store does instead |
|---|---|
| Anything listed shows with the same weight as official content | gnoweb assigns a tier from `-trusted-paths` (§2). Trusted and earned apps carry a "Core" or "Established" badge, which states where the app comes from, never that anyone reviewed it; every other app says "community" after its path, and #6191's notice on its own page. Only trusted or earned apps reach the Spotlight. Being listed is not being endorsed. |
| Free-form title and description, shown as is | No free text but a 40-rune name and an 80-rune tagline, both spoof-checked; the description is the realm's own doc comment (§1 Listings). |
| Moderation = an owner deleting entries by hand | Automatic tiers, GovDAO proposals that pause and hide (never delete), and later flags that demote and gnoweb's `-store-deny` (§3). |
| Votes from any caller, upvote and downvote together, sybil-free | Stars only, direct user calls only, a per-address daily budget; only ranked stars (from names the store has known for 3 days) rank, and only aged ranked stars earn (T8). |
| Ranking by raw vote totals | Guaranteed exposure for new apps, rotation for the long tail, and ranked stars only (§5). |

If reviewers still judge `gnoland` too strong a signal, the fallback is a
neutral namespace governed the same way (for example `r/gnops/store`). Nothing
in gnoweb changes: `-store-realm` names the path.

gnoweb reads it through a flag:

```
gnoweb -store-realm /r/gnoland/store/v0   # gno.land deployments
gnoweb -store-realm ""                    # no store: no view, no icon, no alias
```

#### Listings

A listing has a **core** that every app sets, and **rich fields** (owner
images) that are stored for everyone but displayed only to tiers that allow
them (§2).

```go
type listing struct {
    // Core: shown for every tier.
    Slug      string // [a-z0-9-]{3,40}, unique store-wide; key for stars and the API
    Namespace string // derived from Path: builders and the per-namespace cap
    Kind      string // app | service | package
    Path      string // the realm or package, e.g. gno.land/r/acme/swap
    Title     string // ≤ 40 runes, unique store-wide on a folded key, spoof-checked (T6)
    Tagline   string // ≤ 80 runes, plain text, spoof-checked
    Category  string // key from the closed taxonomy
    Palette   int    // index into gnoweb's fixed palette (0–11), never a free colour
    Created   int64  // block height
    Updated   int64

    // Rich: displayed per tier, subject to the publication delay.
    live   rich  // Icon, Cover: ipfs://CID only
    next   *rich // the last submitted change, effective at nextAt + richDelay
    nextAt int64

    Confirmed  bool      // listed by its own realm or the seed, or given a ranked star
    stars      *avl.Tree // every star: the total, in the card's tooltip
    registered *avl.Tree // ranked stars only: rankings, reputation
    hidden     bool      // hidden by GovDAO: kept under its path, shown nowhere
}
```

**A claim alone buys no exposure.** A realm that calls `Register` proves it
exists; a `Claim`ed path is checked only for its grammar. So a listing is
**confirmed** when its own realm lists it, when the seed lists it, or once
it receives `confirmStars` (3) ranked stars (§5). The namespace owner's own
star never ranks, so an owner cannot confirm their own claim. The realm's
later `Register` confirms a claimed listing and keeps its kind (a claimed
service stays a service). Only confirmed listings reach the shelves (New,
Top, Trending, Recently updated, Rediscover, the Build lens), the activity
feed, the pulse and the builders: one function, `track`, gates the feed and
the pulse, and hidden or unconfirmed listings announce nothing, re-claims
included. Until it is confirmed, a claimed package or service shows
nowhere, and a claimed app only on its category page. Flooding
"New" or the feed with fake paths therefore costs three ranked stars per
path.

**Slugs are global in this version.** A global slug is first come, first
served, so a famous name can be squatted. The title is spoof-checked and the
path, with its namespace, is always shown, which limits the damage; the real
fix (namespaced ids, see Follow-ups) is out of scope here. gnoweb derives
the namespace from the path itself and never keys on it. Activity events
point at the listing itself, not its slug, so a slug released by `Hide` and
reused never inherits the old listing's events.

**"Updated" cannot be farmed.** A re-registration, a claim or a rich-field
change moves a listing in "Recently updated" and the activity feed only
when a stored field actually changes, and at most once per 7 days
per listing. Spamming `Register` changes nothing a reader sees, and a
hidden listing announces nothing.

Four fields are deliberately absent:

- **No description.** The description is the package's own doc comment,
  already public, already tied to the code, and needing no second
  moderation surface.
- **No owner news posts.** "Updated" is derived from the chain.
- **No tags.** At this size the closed taxonomy is enough, and a tag list
  is one more governed vocabulary to maintain. Tags and tag pages can come
  back later (Follow-ups).
- **No website.** An external link is the main phishing vector and the one
  field a stolen key would use first. Off-chain presence can come back
  later, gated like images (Follow-ups).

**Owner images are `ipfs://CID` only**, served through `ipfs.io`, already in
`cspImgHost`. A content address cannot change after anyone has looked at it
(T3). A mutable `https` URL can; inline `data:` images would bloat every
response and every page and cannot be checked by the realm; SVG from owners
is never accepted. gnoweb's own generated visuals are SVG, because gnoweb
writes them.

#### Who may write what

These are crossing functions. Caller identity always comes from
`cur.Previous()`, never from an argument, and never from
`unsafe.PreviousRealm()` in a non-crossing function (see
`docs/resources/gno-interrealm.md` § Captured Realm Values).

| Action | Caller check | Notes |
|---|---|---|
| `Register(cur, slug, title, tagline, category, palette)` | the listed path is `cur.Previous().PkgPath()` | **The app realm lists itself.** Deploying the realm proves control of it. No off-chain verification. Never panics, see below. |
| `Claim(cur, path, slug, title, tagline, category, palette, kind)` | `cur.Previous().IsUserCall()` and `ownsNamespace(caller, ns(path))`: `names.IsAuthorizedAddressForNamespace` when `r/sys/names` is enabled; otherwise (dev chains, tests, where it lets everyone through) only the caller's own address namespace | Packages, services, and realms deployed before the store. Panics with the reason. The path is checked for its grammar only (a realm cannot query whether a package exists), so the listing starts unconfirmed. |
| `SubmitRich(cur, icon, cover)` | `cur.Previous().PkgPath()` is a listed path | **A listing's owner is its realm.** There is no owner address: an app manages its listing through its own code. A claimed package has no owner images. |
| `Star` / `Unstar` | `cur.Previous().IsUserCall()` | One per address per listing, at most 20 new stars per address per day, not refunded by `Unstar`. Whether the star *ranks* is decided once, at star time (§5); the namespace owner's own star never ranks. One per-address record holds the star budget and the name clock. No downvote. |
| `ProposePause`, `ProposeUnpause`, `ProposeHide(slug, reason)`, `ProposeUnhide(path)`, `ProposeSpotlight(slug, days)` | anyone builds the proposal; it takes effect only when GovDAO votes and executes it | The safety net the Hall of Fame review asked for (#2842), DAO-managed from the start; see below. |
| later: `Flag`, `Retire`, editor shelves | see Follow-ups | |
| taxonomy, quotas, delays, thresholds | none: constants, permanent for this path | a new version (`v1`) changes them |

**`Register` never panics.** It runs from the app's `init()`. A panic there
would make the app's deployment fail, and a redeployable (private) realm's
`init` can run more than once. This was the main review point on the Hall of Fame
(#2842, #3674): registration is an upsert, like a wake-up packet that is
harmless to repeat.

- A second call from the same realm updates the listing (title, tagline,
  category, palette) and keeps its slug and creation height; the slug
  argument is ignored.
- An invalid call (bad field, slug or title taken, namespace full, store
  paused) returns without effect and emits `StoreRegisterRejected` with a
  reason, so the author can see why in the transaction events.

`Claim`, `SubmitRich` and `Star` run in ordinary transactions, not in
`init`, so they panic on bad input: there, an error is the right answer.

**Governance: GovDAO, through proposals the realm builds.** There is no
admin address and no bootstrap key. Following the `r/sys/users` and
valopers pattern, `ProposePause`, `ProposeUnpause`, `ProposeHide(slug,
reason)`, `ProposeUnhide(path)` and `ProposeSpotlight(slug, days)` return a
`dao.ProposalRequest` whose
`dao.NewSimpleExecutor` callback runs only once GovDAO has voted and
executed it. Inputs are checked when the proposal is built (a known reason,
a listing that exists, a path that is hidden, a confirmed app and a
duration), and again when it is executed. A hide or a pick is bound to the
listing's path when it is proposed: since a hide releases a slug, it fails
if the slug names another listing by execution. A pick is also bound to
the app's title (its folded key): it fails if the app was renamed by
execution. The callbacks return errors instead of
panicking, because a panicking executor would leave the proposal stuck.
GovDAO can pause, hide and pick, never edit:

- `Pause` turns `Register` into a no-op and makes `Claim` panic; stars,
  `SubmitRich` and hiding keep working.
- `Hide` takes an enum reason (impersonation, phishing, broken, offensive,
  other), emits an event, and removes the listing from every shelf, category
  and response. It also releases the listing's slug, title and namespace
  slot, so a squatter blocks no one.
- The record stays under its path, so `Unhide` takes the path and refuses
  with `slug_taken`, `title_taken` or `namespace_full` if any was taken
  meanwhile.
- `ProposeSpotlight` features one app as the front page's hero, labelled
  "Picked by GovDAO", for 1 to 30 days (`maxPickDays`); the proposal reads
  "Feature <slug> as the front page's hero for N days". One pick at a time:
  a later one replaces it. Confirmed apps only, as the front shows no
  other. The pick is for the app as it was voted on, so it ends for good
  when the app is hidden, stops being a shown app (a claim that makes it a
  service) or is renamed (the pick keeps the title's folded key): an
  unhide, a change back or the old name again does not bring it back.
  It emits `StorePicked`; `api/v1/home` carries `"pick": "<slug>"` while it
  lasts, and the markdown home shows it under "Picked by GovDAO".

**The pick is the store's only editorial act**, and it is shaped by the
minimal-curation goal: it is voted (a GovDAO proposal, not a key), labelled
(the reader sees who chose it), and bounded in time (30 days at most, so
no pick outlives the vote that made it). That is the "curation team" moul
described for awesome-gno (#3928), reduced to one slot that cannot
accumulate.

The integration test `gno.land/pkg/integration/testdata/store_govdao.txtar`
covers the path end to end: initialise GovDAO, propose a hide, vote,
execute, and the listing is gone; then propose a pick, which GovDAO shows
as "Feature boards as the front page's hero for 7 days", vote, execute,
and `api/v1/home` carries `"pick":"boards"`. The trade-off is speed: hiding a phishing
listing takes a GovDAO vote, with no fast unilateral admin. That is
consistent with moul's "Later … DAO-managed", and gnoweb's `-store-deny`
(Next) is the operator's immediate switch.

**Why it is not minimal.** moul asked for minimalist contracts (#2842). The
store keeps its full scope on purpose: most of its size is security bounds
(title and asset validation, confirmation of claims, namespace quota),
anti-sybil rules (ranked stars, the name clock, the star budget) and gas
bounds (maintained indexes, capped reads, epoch rings), each answering a
threat below. What could be cut later without weakening those: the Build
lens and builder rankings, the pulse counters and activity ring, Trending's
epoch-bucketed index (a plain "most starred" would do on a small chain),
the publication delay if owner images are dropped, and the markdown store
beyond the home page.

**Rich fields have a publication delay.** `SubmitRich` stores the change as
pending. It becomes effective `richDelay` blocks later (≈ 24 h); until then
the API keeps serving the previous values, and a new submission restarts the
delay. gnoweb never sees a pending value, so a compromised realm cannot put
a phishing image in front of readers instantly (T1).

**Spam has a price; capacity has no global cap.** Storage deposits price
every write; a namespace holds at most 5 visible listings (hidden ones
release their slot); stars have a per-address daily budget; a claim reaches
no shelf until confirmed. There is deliberately no global listing cap: a cap
is a capacity that one actor can fill to lock everyone else out. Responses
and the work behind them are bounded instead (T10, T11).

**Sort orders are maintained, not recomputed.** As in the Hall of Realms,
listings are indexed in avl trees: by creation height, update height,
ranked-star count and, for apps, an epoch-bucketed trend index and per category (paged with
`p/nt/avl/pager`). Builders are kept ranked by stars and by first listing.
Writes move only what changed: an update re-indexes a listing only when its
kind or category changes, an "updated" announcement moves only its
update-height entry, and a builder is re-keyed only when its key changes.
The activity feed is a fixed ring; the pulse and 7-day stars are per-epoch
rings; the earned check reads one entry of the ranked-star tree.

Every read is bounded whatever the store's size. A front or Build shelf
sends a pool of 16 (`gridPool`), so it reads at most 16 entries of an
index; Top reads the ranked-star index downwards and stops at the first app
below its floor, skipping only the few core apps; the builder rankings read
10 each. Trending reads a trend index of 7 slots,
one per epoch of the window: an app is filed in the slot of the epoch of its
last ranked-star change, by its 7-day score at that change (an upper bound
since: scores only decay), and only if that score reaches the floor of 2. A
slot whose epoch has left the window is dropped (replaced by a fresh tree on
the next write), so stale scores never pile up in front of live ones.
Trending reads only live slots, each downwards, stopping once a stored score
falls below the floor or the n-th best real score found so far, n being
the list's length (24, `pageSize`, for `:trending` and its shelf; at most 8
for `Block`), and after 200 apps in all (`trendingScan`). It is exact unless that cap cuts it short. The Build lens
has no Trending shelf and keeps no trend index. Rediscover finds where the new apps end in the
creation index by binary search, then pages it.

#### Taxonomy

The taxonomy is a closed list of keys and labels, permanent for this path
(a new version changes it):

`defi` · `social` · `games` · `dao-governance` · `identity` · `nft-art` ·
`dev-tools` · `infrastructure` · `education` · `content-media` · `utilities`

No free text means nothing to moderate.

#### Read API

`Render(path)` serves two audiences:

- `Render("")`, `Render("build")`, `Render("c/defi?page=N")`,
  and the lists `Render("top|latest|updated|trending[?page=N]")`: a
  complete **markdown store** (24 per page; each home section links
  `[See all](…)`), readable in any gnoweb, in
  gnobro, or by an agent through `Accept: text/markdown` (#5794). Each item
  links to the listing's own realm or package page; there is no per-app
  page. Its own links use the realm's path, read in `init`, so it works
  wherever it is deployed. It is what gnoweb falls back to when its richer
  view cannot be built (§6). One paged renderer serves categories and lists.
- `Render("api/v1/home")`, `api/v1/category/<key>/<page>`,
  `api/v1/list/<key>/<page>`, `api/v1/build`: **JSON**, versioned by path
  and trimmed to what gnoweb reads. `home` returns the pulse, the taxonomy
  with counts, every shelf, the activity feed and the listings they
  reference **in one call**, plus `pick`, GovDAO's picked app while the
  pick lasts. A shelf carries its `note`, whether it is
  `pinned` and its `empty` text, `stars_7d` for Trending, and `more`, the
  lists behind
  it (`{"key", "title", "count"}` each). A listing carries its total
  `stars` and its `ranked` stars. `api/v1/list/<key>/<page>` serves `top`,
  `latest` and `updated` (paged by 24) and `trending` (computed, one page
  of up to 24). `list` answers
  `{"version":1,"list":{…},"page","pages","listings"}`, paged like a
  category. `build` does the same as `home` for the Build lens.

- `Block(list, n)` and `Badge(path)`: **embed helpers**, exported
  non-crossing functions any realm calls from its own `Render`, which read
  state only and return markdown strings, never pointers. `Block` lists the
  first `n` (clamped to 1–8) apps of `latest` (the front's New rule),
  `top` (the front's Top rule) or `trending`, through the same functions
  as the front's shelves (`newApps`, `topApps`), so the two cannot drift, one sanitized link per line
  (the markdown store's item writer), then "See all on Explore"; any other
  list gives `""`. `Badge` gives "★ N on Explore · [Star](…$help&func=Star&slug=…)"
  for a shown app at that realm path, N its ranked stars as the store counts
  them, or `""`: an app cannot show a count of its own making.

The payloads are **bounded by construction**: fixed shelf sizes, a fixed
activity ring, and pagination (24 per page). A store full of listings must
never push `api/v1/home` past `qrender`'s gas limit (T10).

`vm/qrender` is preferred over `vm/qeval` because gnoweb's client already
speaks it (`Client.Realm`), and its output is a string the realm fully
controls. `qeval` would need a new client method and a parser for Gno's value
formatting.

The realm computes the deterministic rotation (§5) and the reputation flag
(§2), so the JSON is ready to render.

### 2. Trust tiers

gnoweb assigns each listing one tier, recomputed per request from cached
inputs. The realm never decides a tier.

| | **Trusted** | **Registered** | **Anonymous** |
|---|---|---|---|
| Rule | path under the operator's `TrustedPaths` (#6191) | namespace is a registered name (`nym-…` or other), not trusted | `g1…` address namespace |
| Icon, cover | ✅ | 🔓 when **earned** | ❌ generated visuals |
| Spotlight | ✅ | 🔓 when earned | ❌ |
| New arrivals, rotation, categories, search | ✅ | ✅ | ✅ |
| Badge on the card | **Core** | **Established** when earned, else "community" after the path | "community" after the path |
| #6191 notice on its realm page | no | yes | yes |

**Earned reputation** (registered tier only) unlocks the same rights as the
trusted tier. It is the only strict signal in the store (see §5, "Count
generously, trust carefully"), and it can be reached by either of two paths:

- **Community path**, computed by the realm from chain data alone:
  - listed for ≥ `earnAge` (14 days at launch);
  - ≥ `earnStars` (5 at launch) ranked stars (from names the store has
    known for 3 days, §5), each ≥ `starMaturity` (3 days) old;
  - no rich-field change pending (T8).
- **Usage path**, computed by gnoweb when an indexer is configured:
  - ≥ 25 distinct callers over the last 30 days.

  Calls cost gas, so usage is harder to fake than stars. It is gnoweb's
  decision, like every tier, and only for the deployment that runs that
  indexer.

Once flags and the private-realm check land (Plan), either path also
requires no open flags above the threshold and not a private realm (T2). The
thresholds are constants sized for a young chain, permanent for this path;
a later version can raise them as the chain grows.

The realm's own reputation never reads indexer data: it is off-chain and can be
wrong or manipulated (T8). Anonymous namespaces cannot earn rich rights at
all, because a `g1…` namespace costs nothing to create.

**Overrides that always win**, in this order. None is implemented yet
(Plan, Next); until then a hide goes through GovDAO.

1. `-store-deny <path|namespace>` (operator, repeatable): the listing is not
   rendered anywhere. Effective at once, no transaction.
2. **Private realm** (`private = true` in `gnomod.toml`, read through
   `vm/qfile` and cached): never above the core tier, with a "Code can be
   replaced by its author" chip (T2).
3. **Names not enforced** (startup check, T5): the trusted tier is disabled
   and everything is treated as anonymous.
4. Non-`live` packages (from `vm/qpkgmeta_json`): dimmed and kept off the
   front page, never removed.

**Badges claim provenance, never a review.** Consistent with #6191, which
claims no review or audit, there is no "Verified" badge. A card carries a
badge, with a check icon, only for a fact the store can state: **Core** for
a path on the operator's `-trusted-paths` list ("One of gno.land's core
apps"), **Established** for a registered namespace whose app earned
reputation on chain ("A named builder whose app earned its place"). Each
says where the app comes from, the operator's list or a track record of age
and aged ranked stars, not that anyone read its code. Every other app shows
a quiet "community" after its path, a word rather than a chip ("Deployed by
its author. Read the code before you interact."), so the default state of a
permissionless store is said without reading as an alarm. The limits are
real: Established is earned with stars and an app can change after earning
it (T8), and Core is only as good as the operator's list (T9). The badges
decide who reaches the most visible places; they certify no code, and the
store makes no stronger claim than the rest of gnoweb.

### 3. Zero-ops curation

The store must run with **no human in the loop**. Editors are a bonus.

| Concern | Automatic mechanism | Human role (optional) |
|---|---|---|
| What gets shown | new arrivals, rotation, stars, indexer shelves (§5) | a GovDAO pick: one app, labelled, 30 days at most |
| What looks official | tiers from `TrustedPaths` + earned reputation | maintain `-trusted-paths` by PR (already the process from #6191) |
| Spam | deposits, a 5-listing namespace cap, a daily star budget, rankings on ranked stars, claims unconfirmed until starred | none |
| Abuse | later: registered-user flags **demote** automatically (out of Spotlight, shelves and rotation) but never hide | a GovDAO `ProposeHide` with an enum reason, reversible and logged |
| Dead apps | later: liveness check, dimmed, off the front page | none |
| Emergencies | — | GovDAO `ProposeHide` (takes a vote); later, operator `-store-deny`, effective at once |

Flags (later) demote; only GovDAO hides. A flag group therefore cannot make a
competitor disappear, only drop it out of the shelves, and a hidden listing
is never deleted, so every decision can be audited and reversed.

Expected human time: **≈ 0 h/week** to operate, 1–2 h/week if the team
wants editorial (a GovDAO pick, and stories once they exist, see Follow-ups).

### 4. Rich without risk: content from safe sources

What makes the store exciting is mostly content that needs **no**
moderation, because it comes from the chain, from the app's own code, or
from gnoweb itself.

| Source | What it gives | Why it is safe | Who gets it |
|---|---|---|---|
| **The app's own page** | every card links to the listing's canonical `/r/…` or `/p/…` page: the app *running* (Content), its overview and docs (Source), Actions, State | it is the page gnoweb already serves, with its own notice rules; the store embeds and copies nothing from it | everyone |
| **Generative covers** | a unique, deterministic artwork per app: the motif (waves, network, grid, rings, blobs, bars, dots) and the position of a soft light disc both drawn from a seed of the slug, not from the category, so a category page, whose apps share nothing else, still varies; colours from the palette index only, with no category colours | written by gnoweb, inline SVG, no input but a slug and a palette index | everyone, as the fallback and default |
| **Pulse and shelves** | block height, listings and stars per window; New, Top, Trending, Recently updated, Rediscover | counters and indexes kept by the realm | everyone |
| **Auto collections** (later) | "New this week in DeFi", "Built with avl", "Games under 500 lines" | queries over listings and chain facts | everyone |
| **Moments** | one-line facts from the realm's activity feed, the latest joined to the pulse line: "Pixel Wall was listed", "… crossed 10 stars"; later "1 year on chain", "Crossed 100 users" (indexer) | facts, phrased by fixed templates in gnoweb | everyone; indexer moments labelled |
| **Activity** (indexer) | "used 3 minutes ago", 7-day sparkline, weekly users | labelled `indexer` with freshness, as #6231 | everyone |
| **Owner assets** | icon and cover, `ipfs://CID` only | the only free-form source, so it is gated by tier, delay and content addressing: an image cannot change after review, and no external link exists to swap (§1) | trusted, earned |

### 5. Discovery mechanics

#### Chain-only (no indexer)

| Mechanism | What it does | Against |
|---|---|---|
| **A fixed front** | Every section is a grid of 8, always in the same order (see below). | a store that reads differently every visit |
| **New** (pinned) | Confirmed apps of named namespaces, newest first, from the same index as Rediscover: address (`g1…`) namespaces earn no front-page slot, and the seed's core apps are left to gno.land essentials. Its list, `:latest`, holds every confirmed app. Always drawn; empty: "No new app yet. List yours from a registered namespace: it opens here." | oblivion at birth |
| **Top** (pinned) | Community apps with ≥ 3 ranked stars, by ranked stars (the core apps are in gno.land essentials), so a young chain never crowns a one-star app. Its list, `:top`, goes on with every starred app. Note: "Ranked by stars from accounts with a registered name."; always drawn; empty: "No app has 3 stars from named accounts yet. Found one you like? Star it: that's how apps get here." | |
| **gno.land essentials** (pinned) | The seed catalogue's apps, "Built by the core team. Good first stops.", the core of gno.land, frozen in code (no admin setter); no page; last. Always drawn, less the hero's and the Spotlight's apps; left out only when all of them are featured above. The community sections (New, Top, Rediscover) leave the core apps out, so they never crowd it out. | a newcomer with no idea where to begin |
| **Rediscover** | Confirmed apps past their first 14 days, in creation order; a window of 16 (like every unpinned section, room for what is shown above; gnoweb draws 8) slides by 8 every epoch (≈ 1 day) and wraps, so every app falls in the first half of a window, the part gnoweb draws, within `ceil(m/8)` epochs, with no hashing, no state and no pass over all apps. It rotates over a dedicated index of named namespaces' apps, so address (`g1…`) namespaces never take its slots: a free namespace earns no guaranteed front-page slot (its apps stay in categories and lists). The seed's core apps stay out of it too: they live in gno.land essentials, which would otherwise lose them to Rediscover and fall under 4 apps. No page. | long-tail oblivion |
| **Trending this week** | Ranked stars over the last 7 epochs, floor of 2; `:trending`, computed, one page of up to 24; read from the live epoch slots of a trend index with an early stop, at most 200 apps; exact unless the cap cuts it short. | what is picking up |
| **Recently updated** | By `Updated`: real changes only, at most once per 7 days. | rewards maintenance |
| **Lists** | `:top`, `:latest`, `:updated` (24 per page) and `:trending` (computed, one page), behind each section's "See all". The New section's list keeps the key `latest` (URL `:latest`) while its title is "New". | |
| **Exploration slots** (later) | Every ranked section reserves 2 of its slots for rotation picks from the same category. | popularity lock-in |
| **Category pages** | `:c/<key>`, newest first, 24 per page, apps only; the only place an unconfirmed claimed app shows. Tag pages and other sorts (*Stars*, *A–Z*, *Random*) are follow-ups. | "popular first" browsing |
| **Surprise me** | A plain link to a random app's page, picked per request from the listings of the cached home payload. No redirect endpoint. | serendipity |
| **Auto collections and moments** | See §4. | a store that feels static |
| **Shown once, and a minimum** | The pinned sections (New, Top, gno.land essentials) set aside only the hero and the Spotlight's apps, so Top shows the true top even when those apps are also new. Trending, Recently updated and Rediscover show no app already shown above them (the realm sends a pool of 16 so 8 remain) and are hidden under 4 cards. | the same 10 apps everywhere |

#### A fixed front

The front is the same sequence on every chain and every day, with no phase
or threshold to tune:

1. **The hero**, always (gnoweb, below): GovDAO's pick while one lasts,
   else an automatic choice among eligible apps.
2. **Spotlight** (gnoweb, below), which never repeats the hero.
3. **New** and **Top**, pinned: always drawn, each from its own head, with
   an empty text that says what to do when it holds nothing. A young chain
   therefore shows its new apps and an honest empty Top, rather than
   crowning an app with one star; a mature one shows both.
4. **Trending this week**, **Recently updated** and **Rediscover**, each
   drawn only when it has at least 4 apps not already shown above.
5. **gno.land essentials**, pinned like New and Top: always drawn, less the
   hero's and the Spotlight's apps, and left out only when all its apps are
   featured above.

Sections carry a short note saying how they are ranked ("Ranked by stars
from accounts with a registered name."), a "See all" link to their list, and weekly
star counts for Trending.

#### Count generously, trust carefully

A store that counts too strictly looks dead; a store that trusts too easily
looks endorsed. So the two are separate, and each signal says which side it
is on.

| Signal | Counted from | Used for | Strict? |
|---|---|---|---|
| Total stars | every star from a direct user call | the card's tooltip | no |
| Stars in the pulse | every star | the pulse | no |
| Recent registrations and updates | listing heights | "New", "Recently updated"; registrations also in the pulse | no |
| Ranked stars | an address with a registered name the store has known for ≥ 3 days, not the listing's namespace owner | the number on the card, "Top" (≥ 3), "Trending", milestones, builder ranks, confirming a claim (3) | yes |
| Ranked, aged stars | the same, the star itself ≥ 3 days old | earned reputation | stricter |
| Distinct callers (indexer) | transactions | "Most used", usage path to earned | partly: gnoweb's call |

**Why rankings use ranked stars.** An address is free: a script can create
a hundred and push any app into "Trending" within each address's daily
budget. A name costs at least a transaction, and the store makes it wait: a
star ranks only if its address has a registered name that the store has
known for `nameMaturity` (3 days). `r/sys/users` keeps no registration
height, so the store records the first time it sees each named address; that
first star only starts the clock (a follow-up moves this clock into
`r/sys/users`). The namespace owner's own star never ranks: it only adds to
the total. A name made for one campaign therefore
ranks nothing for three days. Names cost 0 on mainnet and onyx today, so
this is a delay, not a price, as is Rediscover's named-only rule: the full
defence also depends on GovDAO setting `registerPrice` above zero. The pulse and the card's tooltip count every star at once;
the number on the card and everything that *ranks* an app, confirms a claim
or earns reputation count only ranked stars, so the number a reader sees is
the one that orders the store. No star is seeded. Bombing is further
contained by the daily budget and the shown-once rule (a ranked section can
never be filled by one app or one push).

**Trending without an indexer.** The realm keeps a 7-epoch ring of each
listing's ranked stars, so "stars this week" is a cheap count. "Trending"
ranks by it with a floor of 2; "Rising" (later) would restrict it to
listings younger than 60 days. Trending cards show "+N this week" only from
5: smaller
counts read as a quiet chain. The chain alone can show what is picking up this
week.

#### The chain is alive: the pulse

A front page must show that things are happening, even with no indexer and
few listings. A one-line pulse sits under the title. It is built from
counters the realm keeps per epoch, plus the node's latest block:

> Block 1 234 567 · 4 s ago · **3** apps & packages listed this week · **41** stars given · Pixel Wall just joined

With an indexer, a fourth figure appears, labelled as indexer data: **2 381
calls** to listed apps today.

The realm counts listings of every kind and stars given, net of those taken
back; it keeps no "updated" figure, which a reader could not act on.

- **Activity feed.** The realm keeps a fixed ring of the last 20 events of
  confirmed, visible listings: listed (or confirmed), updated, crossed a
  ranked-star milestone (10, 50, 100). The API serves the newest 5 of
  listings still shown as apps; gnoweb joins the latest one to the
  pulse line ("Pixel Wall crossed 10 stars").
- **Honest windows.** Every figure names its window, and the pulse picks the
  shortest window that is not zero (this week, then this month). A quiet
  chain reads as "this month", never as "0 this week". Nothing is invented:
  a figure with no activity is left out, not padded.

#### For builders: packages and service realms

Builders look for code to reuse as much as for apps to use. The store keeps
apps in front for users and adds a **Build** lens
(`/r/gnoland/store/v0:build`) for builders:

- **Listing kinds**:
  - `app`: a realm with a `Render()`, for end users;
  - `service`: a realm other realms call, such as a registry or an oracle;
  - `package`: a `/p/` library.

  `Register` lists apps; `Claim` takes the kind.
- **Registering packages.** A `/p/` package cannot call `Register`: pure
  packages do not cross into realms. So packages, and realms deployed before
  the store, are listed through `Claim`, a direct user call by the address
  that owns the namespace (`ownsNamespace`, §1). It has the same validation
  and upsert rules as `Register`, but panics with the reason, since it is a
  user transaction. Until three ranked stars or the realm's own
  `Register` confirm it, a claimed package or service shows nowhere, and a
  claimed app only on its category page.
- **Shelves on the Build lens**:
  - New to build with, Recently updated, and Most starred (by ranked stars,
    from accounts with a registered name): chain-only, 16 apps sent each like the
    front's sections, same rules as apps;
  - later, "Used by listed apps" (cached `qdoc` imports of listed realms)
    and "Most imported" (indexer `importers`, labelled).
- **Package pages** are gnoweb's own `/p/…` pages: the overview already
  shows the API, the docs and the import path. The store adds no copy.
- **Builders as a shelf.** "Top builders" (by ranked stars across a
  namespace's confirmed listings, then by number of listings, so it is
  never empty nor a copy of the other list) and "New builders" (namespaces
  by first confirmed listing) on the Build lens. The core team's
  namespaces, an explicit list in the realm (`coreNamespaces`: gnoland, gno,
  demo, gov, sys, nt, gnops), have no builder entry, so the core team never
  holds the community's rankings. The list is not derived from the seed,
  which also lists a package of a personal namespace (`p/moul/md`): its
  owner is a community builder like any other. The realm keeps both
  rankings on every write and reads at most 10 of each.
- **Builder profiles.** The user page (`/u/…`) gains a "Listed in the store"
  shelf, with the builder's apps and packages and their stars. That gives
  builders a portfolio and readers a way to follow a builder.

**The hero and the Spotlight.** gnoweb chooses both with no editor, from
the *eligible* apps: apps only, trusted or earned, the first 12 in the
realm's shelf order (which puts quality first). The hero is, in order:

1. GovDAO's pick, labelled "Picked by GovDAO" (§1). The vote is the
   endorsement, so a pick can be any app.
2. Else the eligible app with the most ranked stars this week, if at least
   5, labelled "Trending now" only when it leads the realm's trending shelf
   (ties included, every app on it counted, eligible or not), else "Popular
   this week". Both labels link to `:trending`.
3. Else a daily rotation over the eligible apps, labelled "In the spotlight
   today".

Only eligible apps can be the automatic hero, so a community app cannot
take the most visible place on the page by stars alone (T6). If no app is
eligible at all there is no hero, the secure default; on a real chain the
operator's seed apps always are.

The Spotlight then rotates daily through the remaining eligible apps: 3,
left out below 2, never repeating the hero. Once 3 apps (`spotlightSize`)
have earned their place, the operator's trusted apps take one place at most
across hero and Spotlight together: the store must not look like it
promotes its host. Until then, at launch, the operator's apps fill them, so
a new chain is not left without either. Everything is derived from the
chain height in the cached home payload, so every gnoweb shows the same and
caches it. The Spotlight's caption says what it is: "A new set every day, from
gno.land's core apps and community apps with a track record."

#### Indexer-backed (only with `-indexer-url`)

These are registered only when `Deps.Indexer != nil`. Without an indexer,
the shelves do not exist: no empty rows, no placeholder.

| Shelf / signal | Query | Notes |
|---|---|---|
| **Trending** | calls in the last 7 days vs the previous 7, growth-ranked, with a volume floor | growth, not totals |
| **Most used** | distinct callers over 30 days | |
| **Rising newcomers** | Trending, restricted to listings < 60 days old | |
| **Back in action** | quiet for ≥ 60 days, active in the last 7 | resurfaces revived apps |
| **Built on** | `importers` of a listed package | |
| **Per-app stats** | calls, callers, last activity, sparkline | on cards, and in the app band once it exists (Follow-ups) |
| **Candidates queue** | active realms *not* listed, ranked by distinct callers; shown in the realm's editor view only | the chain proposes, curation corrects |

**Wire-in.** The same shape as `feature/omnisearch` in #6231:
`store.Deps` gains an optional `Signals` field, a narrow interface declared in
`feature/store` and satisfied by `*indexer.Client`:

```go
// Signals is the subset of *indexer.Client the store consumes. Nil is the
// feature switch: indexer shelves are not registered without one.
type Signals interface {
    // Activity returns calls and distinct callers per package path over
    // [from, to) block heights, for the given paths only.
    Activity(ctx context.Context, paths []string, from, to int64) (map[string]indexer.Activity, error)
    LatestBlockHeight(ctx context.Context) (int, error)
    URL() string
}
```

The wire-in assigns the field inside the `if cfg.IndexerURL != ""` branch, so
a typed-nil pointer can never advertise shelves it cannot serve. Ranking
(growth, rising, back in action) is one tested function in `feature/store`
over that map; `Activity` is one GraphQL document for every listed path on
the page, not one query per app. The realm is never involved: indexer data
stays off-chain and never feeds `earned` (T8).

All of a page's indexer queries are merged into one GraphQL document (#6231's
transport allows it) and cached 5 minutes per shelf. Results carry #6231's
provenance: an `indexer` tag and a footer with the endpoint and the last
indexed block. On breaker-open or timeout, the shelf is omitted.

#### Discovery outside the store realm

- **Header**: the store icon on every realm page (§6). Shipped.
- **Realm pages**: when the realm is listed, a slim "In the store: ★ n ·
  <Category>" band on its own page, linking back to the store. Later: it
  touches existing pages (Follow-ups).
- **User pages**: an "Apps by @user" shelf, and a "List your app" prompt for
  owners of unlisted realms.
- **Omnibar**: `is:app` and `category:<key>` qualifiers, registered with
  `feature/omnisearch` as chain provenance.
- **Home realm and any other realm**: `r/gnoland/home` can embed a "From
  the store" block with `store.Block` (Read API), and an app its own
  `store.Badge`. This is each realm owner's choice.
- **Not-found pages**: "Looking for an app?" with 3 rotation picks.

### 6. gnoweb: the store realm, rendered better

The store is a realm. gnoweb shows it at its own path like any other, and
`feature/store` only replaces the markdown of its **Content** tab with a
richer view built from the realm's `api/v1` JSON.

#### Pages

```
/r/gnoland/store/v0                  front page ("Explore"), "Apps" lens
/r/gnoland/store/v0:c/<key>?page=N   category page
/r/gnoland/store/v0:top?page=N       every starred app, by ranked stars
/r/gnoland/store/v0:latest?page=N    every app, newest first
/r/gnoland/store/v0:updated?page=N   by last real change
/r/gnoland/store/v0:trending         this week, one page
/r/gnoland/store/v0:build            Build lens: packages, services, builders
/explore                             alias of the store realm
```

Everything else is the realm as it is: other render paths (`:api/…`), the
other tabs (`$source`, `$help`,
`$state`), `Accept: text/markdown`, and every other realm.

Because these are realm pages, gnoweb's own breadcrumb, omnibar value,
Content/State/Source/Actions tabs, realm-notice rules and page title apply
unchanged. The store duplicates none of them and competes with none of them.

#### One hook

`feature/store` follows `feature/state`: a local `ClientAdapter` subset
(`Realm` only), `Deps` (client, realm path, domain, the `Trusted` predicate
that drives #6191's notice, logger) and `New(deps)`.

`Handler.View(ctx, gnourl)` returns a view for the render paths it owns on the
configured realm (`""`, `c/<key>`, `top`, `latest`, `trending`,
`updated`, `build`) and nil for anything else. `GetPackageView` calls it on the
Content tab only, after every other tab dispatch and before the realm
render, unless markdown was asked for. `HTTPHandler.Store` is nil when
`-store-realm` is empty: the nil-field switch of #6231, and nothing else
checks a flag. `-store-realm` is validated at startup (`main.go`): a realm
path, or gnoweb refuses to start.

**Nothing from the realm is trusted.** Every field is validated once per
cache fill, before a template sees it (T7): listings, category keys, labels
and counts (including the category page's own), shelf titles and notes,
weekly counts (dropped unless one per card and non-negative), list
references (a bad one only loses its "See all"), builders, and the page
numbers of a category or list answer. Empty builder lists are left out. A
list key that is not a key at all is refused before the front page is even
loaded. The Spotlight's daily window comes from the chain height in the
payload (a negative height is clamped), so it is the same on every gnoweb and caches with the page.
Text checks are about rendering safety only (length, controls, format and
bidi characters, zero-width and blank fillers, non-graphic runes, doubled
spaces); the look-alike policy lives in the realm, which refuses a superset,
so gnoweb shows whatever the realm accepts and drops nothing silently.
gnoweb checks paths for their grammar and kind; it does not fetch them.
The tier, the short path and the generated art are computed at the same
time, so a request only reads them. A category key and page are checked
against the cached home taxonomy, and a list key and page against the lists
the cached front page names (24 per page), **before** any query: an
arbitrary `:c/<anything>` or `:<anything>?page=N` never reaches the node or
fills the cache. A bad list reference only loses its "See all". The `Register` and `Claim` snippets use the configured realm and
the chain's domain.

**Not found.** An unknown category, a list the front page does not name,
or a page past the last of a category or list is answered by the store's
own not-found page with status 404, without querying the node: `View`
returns it with a `Meta{Status, Title, Description}`, here "Not found ·
Explore". Every store view sets its head title and description through
`Meta`, from fixed strings and the closed taxonomy's labels only
("DeFi apps · Explore"), never from a listing.

**Graceful fallback.** If the store API fails, or answers an unknown version
or an oversized payload, `View` returns nil and gnoweb renders the realm's
markdown. A failure is cached for
5 s and logged once, at load (`Warn`), never once per reader, and never
shown: it can carry the node's address.

#### Links, not pages

The store owns no page beyond its realm's. Everything else is a link into
pages gnoweb already serves:

- **Cards** link to the listing's canonical page (`/r/…` or `/p/…`). That
  page is the detail page: Content is the live `Render()`, Source the
  overview and docs, then Actions and State. The code is one tab away
  (`$source`).
- **Stars**: the ★ count on a card links to the store realm's Actions page,
  prefilled: `/r/gnoland/store/v0$help&func=Star&slug=<slug>`. Signing goes
  through gnoweb's normal Actions flow.
- **Surprise me** links to a random app's path, picked from the home
  payload. No redirect endpoint.
- **`/explore`** is a plain `GnowebPath` alias of the configured realm,
  added in `main.go` (`withStoreAlias`) with `-store-realm`, not with the
  default aliases. It is added before `-aliases`, so an operator's own
  `/explore` wins, and `DefaultAliases` is never mutated.

#### Header

In `components/layouts/header.html`, `.main-nav` holds the gnome logo
(`a.user-picture`) and then `.b-main-navigation` (the omnibar). The store
entry goes to the left of the logo:

```
[ explore ] [ gnome ] [ omnibar ……………………………… 🌐 ] [ menu ]
```

- `HeaderData.StoreURL`, set in `setHeaderForRealm` when the store is on,
  so every page with a realm header shows it. Empty means not rendered.
- The existing `ico-apps` symbol, the label "Explore",
  `aria-current="page"` on the store realm's Content tab only (not on its
  `$source` or `$help`), and a target of at least 44px on mobile.

#### Templates

`feature/store/templates/pages.html` (front, Build, and one paged template
for categories and lists) and
`parts.html` (card, shelf, cover, meta, lens, category strip, "List your
app"), generated covers in `cover.go`, and
`feature/store/frontend/store.css`, imported from `main.css`. Owner assets
reach a template only after validation (§2, T7).

The front page, top to bottom:

1. **Lens**: Apps · Build.
2. **Intro** with **Surprise me**, then the **pulse** and its latest event.
3. **Category strip**: label and count per category.
4. **Hero** (§5), always when an app is eligible: one card, its cover at
   2/5 beside the text on wide screens, with its label ("Picked by GovDAO",
   "Trending now", "Popular this week" or "In the spotlight today"; the two
   trending labels link to `:trending`), **Open** and **Read the code**.
5. **Spotlight** (§5): up to 3 cards, scrolling sideways on small screens,
   never the hero.
6. **Sections**, in the fixed order (§5), none repeating the hero or the
   Spotlight, every one a grid of up to 8 (a flat list, one column, on
   phones; 2 columns, then 4 on wide screens), cut to full rows of 4 once
   it has one. New, Top and gno.land essentials are always drawn, New and
   Top with their empty text when they hold nothing; the others show no app
   already shown above them and are hidden under 4 cards. Trending cards
   show "+N this week" from 5. Each section shows its note and a "See all"
   link (its list's title for screen readers), and is labelled by its
   heading. The Build lens uses the same grids.
7. **List your app**: what happens next ("A realm lists itself: deploying
   it is the proof that you control it. New apps open New, then stay in
   their category; those of registered names come back through
   Rediscover."), where to look when it does not ("Not showing up? Your
   deploy transaction carries a `StoreRegisterRejected` event with the
   reason."), "Know Go? You already know most of Gno." with a link to the
   docs, a copyable `store.Register(cross(cur), "YOUR-SLUG", "Your App
   Name", …)` snippet, and under "Add a star button to your app" a
   `store.Badge("<domain>/r/YOUR-NAMESPACE/your-app")` snippet for the
   app's own `Render`.

The pulse line carries the latest activity event; there is no separate
activity list. Cards show ranked stars, with the total in the tooltip. A
list page header says "N apps".

The Build lens shows its shelves, Top and New builders (linking to `/u/…`),
and a copyable `Claim` snippet. A category or list page shows the strip, a
grid of cards and previous/next links.

#### Footprint in gnoweb's core

About +100/−8 lines over 8 files: the flag, its validation and the alias
(`main.go`), the config field (`app.go`),
the hook and wiring (`handler_http.go`), `HeaderData.StoreURL` and the
header entry with its CSS, and `components.IsVisibleRune`, exported so the
store reuses the banner's invisible-character rule. `weburl`'s grammar,
dispatch and the omnibar are untouched, and there is no new view mode.

### 7. UX, UI and positioning

The store must make people want to explore and still read as gnoweb. "App
store" here means Apple's *editorial confidence*, delivered in gnoweb's own
visual language, not Apple's chrome.

#### Principles, taken from what gnoweb already is

1. **The code is the product's proof.** Every realm is one click from its
   Source, State and Actions. No other app store can say that, so the store
   leads with it: every card opens the app's own realm page, where Source
   is one tab away, and a community app's quiet "community" says "Deployed
   by its author. Read the code before you interact."
   Positioning: *"Apps you can read."*
2. **Calm chrome, colourful content.** gnoweb's chrome is grayscale surfaces
   (`--s-color-bg-surface-*`), one green brand (`--s-color-bg-brand-default`),
   Inter, mono for chain facts, a 6px radius. The store keeps all of it.
   Colour lives *inside* content: covers and generative art.
3. **Honest by construction.** Provenance is labelled (`indexer`), guesses
   are drawn as guesses, community apps say so after their path, the only
   badges (Core, Established) state provenance, not a review, and dead apps
   will be dimmed rather than hidden (once the liveness check lands). No fake urgency, no pay-to-rank, no sponsored slot.
4. **Server-rendered, no JavaScript needed, accessible.** Shelves are lists of
   real links, drawn as grids; only the Spotlight scrolls sideways on small
   screens, under the user's control, and nothing
   moves on its own. Focus is visible and contrast meets AA in both themes.
5. **Fast and stable.** Inline generated SVG, lazy images with explicit
   dimensions, and one store RPC call per page.

#### Who it serves

| Reader | Job | What answers it |
|---|---|---|
| Newcomer | "Show me something worth trying, now" | gno.land essentials, Spotlight, Top, **Surprise me**, one click to the app running on its own page |
| Returning user | "What's new since I was here?" | New, Recently updated, the pulse's latest event; a "new since your last visit" dot (localStorage, per-viewer, progressive) |
| User with a need | "Is there an app for X?" | Categories; later tags and `is:app category:` |
| Sceptic | "Is this alive? Can I trust it?" | trust badges, stars, the pulse, the Source tab, the realm page's own notice |
| Builder | "How do I get seen?" | **List your app** with a copyable `store.Register(cross(cur), …)` snippet and the exposure guarantees stated plainly; **List your package** with a `Claim` command on the Build lens; Top and New builders |

#### Card anatomy (one component, two sizes)

```
┌──────────────────────────────┐
│ [cover | generative]         │   a thin banner; none on phones
├──────────────────────────────┤
│ [icon] Name                  │   icon 40px tile
│        nym-acmex123/swap     │   path in mono, under the name
│ Tagline, at most two         │   2 lines, then ellipsis
│ lines…                       │
│ DeFi · ✓ Established · ★ 128 │   category chip, trust badge, ★
└──────────────────────────────┘
```

- The name links to the app's own page. The category chip and the ★ count
  are separate links with their own focus targets, never nested anchors;
  the ★ count links to the store's Actions and is hidden while the app has
  no star. A category page drops the category chip: it would repeat the
  page.
- The trust badge is "Core" or "Established" (§2). A community app has
  none: "community" follows its path instead, as a quiet word.
- Paths are shortened **in the middle**, never at the end, and the full
  path is in the element's `title`. A named namespace is always shown
  whole, and a deep path keeps it and the package's name, with its version
  (`acme/…/swap/v2`). An address namespace is cut in its own middle, both
  ends kept (`g17zyd…9cxg/poll`), so a look-alike is never hidden (T6).

#### Visual system

- **No new palette.** `store.css` uses gnoweb's existing tokens directly;
  there is no store token layer. The 12-entry app palette is a fixed set,
  validated for AA in both themes; categories have no colours.
- **Typography.** Store title 32, shelf titles 18 semibold,
  body 14, chain facts in `--g-font-family-mono`.
- **Layout.** The realm page's own content area; sections are grids, and
  only the category strip reuses `c-reel`.
- **CSS architecture.** ITCSS layering, `b-store-*` blocks.
- **Header icon.** Same size, stroke and treatment as the other `ico-*`
  symbols (`ico-apps`); green active state on the store realm's path only.
- **Motion.** Opacity and border changes only, at most 150ms, disabled under
  `prefers-reduced-motion`.

#### Small catalogues

- A section other than the pinned New, Top and gno.land essentials with
  fewer than 4 cards is not rendered; its apps stay in their category and
  lists. New and Top are always drawn, with an empty text that says what
  to do, so a young chain's front page is never blank.
- An empty front page or category says so plainly and shows **List your
  app**.
- Later: picks from neighbouring categories on an empty one.

#### Voice and microcopy

The voice is gnoweb's: direct, precise, a little warm, never hype.

| Use | Avoid |
|---|---|
| New | 🔥 Hot right now |
| Rediscover | You might have missed… |
| Back in action | Trending!!! |
| Open · Read the code | Get · Install |
| Core · Established · community | Verified · Official |
| Usage from indexer, as of block 1 234 567 | unlabelled numbers |

#### Sharing

The URLs are the chain's own: `/r/gnoland/store/v0`, `:c/<key>`, `:top`,
`:latest`, `:trending`, `:updated`, `:build`, and
each app's canonical path. Titles and summaries come from gnoweb's realm
pages, as for any realm (#6229), except that the store's own views set
their head title and description from fixed strings and the taxonomy only
(§6), so what a link preview shows is never a listing's to choose. Later, `og:image` only when a raster cover
is displayable for the tier. Social platforms do not render SVG, and gnoweb
will not run a rasteriser.

#### Launch and measurement

- **Seed before launch**: the realm's seed catalogue lists 12 existing apps,
  services and packages, on paths that exist on mainnet and onyx (paths
  found on neither were removed). Seeds show generated art: they cannot
  call `SubmitRich`, so they have no owner images. A store that opens
  empty does not get a second look.
- **Surfaces**: header icon, footer link, a home-realm block, and a "List your
  app" guide under `docs/builders/`.
- **Measure** with the SimpleAnalytics already wired in (no new tracker):
  store → app realm page, → its Source tab, ★ → Actions, "See all" →
  list pages, listing growth. Long-tail
  exposure is guaranteed by the rotation and checkable from the realm alone.

## Threat model

Assume every listing, every `Render()` and every indexer answer is hostile.
Ordered by severity.

**T1. A trusted realm is compromised, or governance is abused.** A trusted
realm's listing has the richest surface: the Spotlight and images. Whoever
controls governance can pause the store and hide listings.
*Mitigations*: rich fields are images only, content-addressed, and wait
`richDelay` (≈ 24 h) before gnoweb sees them; there is no external link to
swap; a card links only to the realm path; GovDAO can `ProposeHide` the
listing, and `-store-deny` (Next) will remove it from every surface at once.
There is no admin key to steal: only an executed GovDAO proposal pauses or
hides, it cannot edit or delete a listing, every action emits an event, and
a hide is reversible. The automatic hero and the Spotlight draw only from
the first 12 eligible apps, rotate daily, and once 3 apps have earned their
place give the operator's trusted apps one place at most across both, so one compromised trusted listing is not pinned to the top.
GovDAO's pick, the one editorial act, is visible (a proposal anyone can
read, a `StorePicked` event), labelled "Picked by GovDAO" on the page, and
bounded in time (30 days at most; hiding the app, its ceasing to be a
shown app or a rename ends it for good), so an abused vote cannot keep an
app on top, nor follow it into another title. "gno.land essentials" is frozen in code: no call
changes it.
*Residual*: an attacker who controls a realm's code controls its `Render()`,
as they would on its realm page today. The store does not make this worse.
Hiding a phishing listing takes a GovDAO vote; until `-store-deny` lands,
there is no faster switch.

**T2. A private realm swaps its code.** Only private realms can be deployed
over (#6164). An app can look clean when listed, then become a drainer.
*Mitigations* (Next: the private-realm check is not implemented yet):
private realms will never get above the core tier, whatever their
namespace; they will carry a "Code can be replaced by its author" chip;
reputation will exclude them.
*Residual*: public realms cannot be redeployed, but they can hold upgradable
logic (proxies to other realms). Every card opens the realm page, where
Source is one tab away. The store states this and does not claim more.

**T3. Mutable image hosts.** A `https` image reviewed today can be replaced
tomorrow.
*Mitigations*: owner images are `ipfs://CID` only, checked by the realm and
again by gnoweb, and served through the one gateway already in
`cspImgHost`. No `https` images, no inline `data:`, no SVG from owners; CSP
still bounds hosts.

**T4. A listed realm's `Render()` impersonates the store or gnoweb, or
costs too much.** A realm's markdown can mimic a gnoweb alert ("Verified by
gno.land") or a button, and a slow `Render()` would multiply the cost of
every store page that embedded it.
*Mitigations*: the store embeds no realm render. Its pages show only
validated listing fields through `html/template`, and the one `Render()` it
calls is its own realm's `api/v1`, cached. A listed app's `Render()` shows on
its own realm page, as it always did, under gnoweb's existing sanitising
pipeline, CSP and realm-notice rules. Embedded previews are a follow-up and
would need framing, a capped height and a timeout first.

**T5. The trust list on the wrong chain.** `defaultTrustedPaths` is compiled
in. On a chain without namespace enforcement, anyone can deploy under
`gnoswap` and would get the trusted tier.
*Mitigations* (Next: the startup check is not implemented yet): at startup,
gnoweb will check `r/sys/names.IsEnabled()`. If names are not enforced and
the operator did not set `-trusted-paths` explicitly, the trusted tier will
be disabled and logged. #6191 documents this hazard; the store will enforce
it. Until then, operators of such chains must set `-trusted-paths`.

**T6. Impersonation by name.** Squatting a famous slug or title, while
both are global; homoglyphs ("GnoSwаp" with a Cyrillic а); bidi overrides,
zero-width characters, odd spaces; look-alike namespaces
(`nym-gnoswapx123`).
*Mitigations*: the realm refuses titles and taglines with no letter or
digit, or with anything not graphic (controls, format, bidi and zero-width
characters, private-use and unassigned code points), blank fillers, spaces
other than U+0020 or doubled spaces, enclosing marks, fullwidth ASCII,
mathematical alphanumerics, and any combining mark that does not follow a
character of its own script among those allowed to mix with Latin (below):
no accent added to a Latin letter, no variation selector, no invisible
Khmer vowel, no foreign mark such as a Devanagari nukta on a Latin letter.
It also refuses Latin letters outside the alphabets in common use (IPA,
small capitals, Roman numerals, click letters), and Latin mixed with
letters of any script outside an allowlist that does not mimic Latin (Han,
Hiragana, Katakana, Hangul, Arabic, Hebrew, Devanagari, Thai): Greek,
Cyrillic, Armenian, Cherokee, Lisu, Coptic and the rest are refused next to
Latin. Titles are unique store-wide on a folded key: case; spaces, `-` `_` `.` and the
dashes and dots drawn like them (‐ ‑ ‒ – — ― ⁃ − ﹘ ﹣ · ․ ‧ ∙ ⋅ ・) dropped;
`i`, `1`, `|`, `ı`, Cyrillic `і`, Greek `ι` → `l`; `0` → `o`; Cyrillic and
Greek letters that look Latin → that letter (an all-Cyrillic "АСМЕ"
collides with "ACME"); `rn` → `m`; `vv` → `w`. The i/l and separator folds
match `r/sys/users`' canonical names, so "Mail" and "Mall" collide, an
accepted false positive. `Hide` releases a squatted slug and title for
their real owner. The path is always displayed and shortened in the
middle only: a named namespace is shown whole, an address namespace keeps
both its ends (`g17zyd…9cxg`), and the full path is in the `title`. The page's most visible places, the hero and
the Spotlight, go automatically only to eligible apps (trusted or earned),
so an impersonating community app cannot reach them by stars alone; only a
GovDAO vote can put any other app there.
*Residual*: confusables outside these folds remain; uniqueness on the full
UTS #39 skeleton and namespaced ids close them (Follow-ups).

**T7. Injection.** Free colours in `style=`, `javascript:` links, template
escaping slips.
*Mitigations*: colours are palette indexes; only gnoweb's own SVG reaches
`template.HTML`; there are no owner links, and image URLs are built by
gnoweb from a validated CID; paths are validated against `weburl`'s grammar
and must match the kind (`/r/` or `/p/`). The JSON is capped at 2 MiB, and
every field is re-validated in gnoweb even though the realm validated it
first: listings (slug, path grammar and kind, title, tagline, category,
palette, stars), category keys, labels and counts, shelf titles, builders,
and a category answer's page numbers. Text is checked for rendering safety
(T6's look-alike policy is the realm's). A new version of the realm can be
written badly; gnoweb must not care.

**T8. Gaming rankings and reputation.** Addresses are free, so stars
from fresh addresses are a sybil army; an owner may star their own app; an
app can also earn reputation while benign and then switch.
*Mitigations*: stars come from direct user calls only (a `maketx run`
script or another realm cannot star), each address has a budget of 20 new
stars per day, and there is no downvote to bomb with. Rankings ("Top",
"Trending", milestones, builder
totals), the number on the card and confirmation (3 stars)
count only ranked stars: from an address whose registered name the store
has known for 3 days, and never the namespace owner's own. Earned
reputation also needs the stars ≥ 3 days old. Other stars only move the
tooltip's total and the pulse. No star is seeded. Anonymous namespaces
cannot earn; any rich change restarts the delay and blocks earning while
pending; indexer data never feeds reputation. Top needs 3 ranked
stars, so one star cannot crown an app on a young chain, and the front's
order is fixed, so no star count reshapes the page. Later, flags from registered users demote quickly.
*Residual*: names cost 0 on mainnet and onyx today, so a patient attacker
can register names, wait 3 days and then rank; the delay and the daily
budget slow this down, and GovDAO's `registerPrice` is what prices it.

**T9. A trusted name changes hands.** GovDAO can rename or delete a user in
`r/sys/users` (`ProposeUpdateName`, `ProposeDeleteUser`), and trust follows
the name.
*Mitigations*: the same rule as #6191: a GovDAO allocation of a listed name
must be mirrored in `-trusted-paths` by PR. The store adds nothing to it: a
listing is controlled by its realm's code, not by an address, and `Claim`
re-checks namespace ownership on every call.

**T10. Store realm failure.** A caller-check bug, an unbounded payload, a
bad new version. The deployed path is permanent, so a bug stays until a new
version replaces it.
*Mitigations*: crossing functions with `cur.Previous()` only; tests that
try `maketx run` on star functions, indirect calls through a third-party
realm, and registering a foreign path; a GovDAO integration test for the
governance proposals; bounded payloads (§1); a new version is a new path
that `-store-realm` switches to; gnoweb drops any listing or shelf that fails
validation, and if the API itself fails it hands the page back to the
realm's markdown (§6), caching the failure for 5 s so a broken realm costs
one query per key, not one per reader.

**T11. Capacity exhaustion.** An actor fills the store so that legitimate
apps cannot list, or so that pages grow until they fail.
*Mitigations*: no global listing cap to exhaust; at most 5 visible listings
per namespace, and a hidden listing releases its slot, slug and title;
storage deposits price every write. A claim reaches no shelf, feed, pulse
or builder list until three ranked stars from others (or its realm's own
`Register`) confirm it: until then a fake package or service path shows
nowhere, and a fake app path only on its category page.
Free `g1…` namespaces can still list their own realms, but those never rank
without ranked stars, never get owner images, and are left out of
New and Rediscover, which draw from named namespaces only. Every read is
bounded: front sections read at most 16 index entries (Top stops at its
floor, skipping only the few core apps), builders 10,
Trending at most 200 apps from live epoch slots only (stale slots
are dropped, not scanned), cut at the n-th best score (n = 24, at most 8
for `Block`), Rediscover a binary search and a page, category
and list pages 24, so the
catalogue's size never reaches a page;
gnoweb only queries category pages that exist in the cached taxonomy.

**Lower severity.**
- *Privacy*: owner images on IPFS gateways reveal reader IPs to the gateway,
  as images on realm pages do today. Stars are public and reveal what an
  address likes; starring goes through the Actions page, where the call is
  shown before it is signed.
- *Deep links*: owners cannot supply links into Actions with prefilled
  arguments or coins. **Open** always goes to the realm root. The only
  prefilled Actions link is gnoweb's own: the store realm's `Star`, with a
  validated slug and no coins.
- *Flag brigading* (once flags exist): flags demote, never hide, and only
  count from registered names.

## Resource bounds

- Front page and Build lens: **1** `vm/qrender` each (`api/v1/home`,
  `api/v1/build`), cached 30 s, singleflight, 4 s timeout detached from the
  request. A cache hit costs neither an RPC nor a decode.
- Category and list pages: 1 `qrender` (`api/v1/category/<key>/<page>` or
  `api/v1/list/<key>/<page>`, 24 per page) plus the cached home payload,
  which also decides whether the key and page exist: an unknown one costs no
  query.
- Failures are cached 5 s and logged once per load. Validation, tiers, short
  paths and generated art are computed once per cache fill, not per request.
- The cache holds at most 512 entries; its keys are bounded by the
  taxonomy, the front page's lists and their real page counts.
- No app page and no previews: an app's page is its realm page, at the cost
  it already has.
- Later: liveness and private checks, lazily for listings on the page
  (≤ 60), cached 10 min; the realm-page band, 0 calls on a cache hit and
  omitted on timeout.
- Indexer: 1 merged GraphQL request per page on a cold cache; #6231's
  16-slot gate, 8 MiB cap and breaker unchanged.
- Realm JSON capped at 2 MiB before decoding.
- Storage, measured: a demo with about 40 listings and 70 ranked stars
  needed about 1.27 MB of realm storage (about 127 GNOT of deposit at 100
  ugnot per byte). Each listing carries several index entries (creation,
  update, stars, trend, category, named) and its own trees (stars, ranked
  stars), so storage grows with listings and stars, not with readers. Each
  writer pays the deposit for what it adds; the seed's deploy pays for the
  seed.
- In the realm, every read is bounded whatever the catalogue's size: front
  sections read at most 16 index entries (Top stops at its floor and skips
  only the few core apps), builders 10, Trending
  at most 200 apps
  (`trendingScan`) over the live epoch slots with an early stop at the
  n-th best score (n = 24, at most 8 for `Block`), Rediscover
  a binary search and a window of 16; the feed is a ring of 20 served 5 at a
  time, and category
  and list pages hold 24.

**Later: "Apps you may like".** Recommendations need to know who is reading.
gnoweb is stateless and has no accounts, by design. When gnoconnect (#5970)
gives the browser an address, recommendations can be computed from that
address's own stars. Until then, the Rediscover rotation fills that role.

## Alternatives considered

**A top-level `/a/<namespace>/<name>` path**, next to `/r/`, `/p/` and
`/u/`. Rejected. Those prefixes are the chain's own trees, and gnoweb's URL
grammar (`weburl`), dispatch, breadcrumb, omnibar and realm notice all rely on
that. An `a/` tree would be a gnoweb-only fiction that exists on no chain:
typed in the omnibar today it is a 400, and making it work would mean
teaching every one of those layers about a namespace the chain does not
have. Apps keep their real path (`/r/acme/swap`) as their address, and
that page is their store page. The omnibar can find apps through an
`is:app` qualifier in `feature/omnisearch` (#6231): an addition, not a new
grammar.

**A store-owned route and app page** (`/store`, `/store/app/<slug>`).
Rejected. The app page duplicates the realm page and its `$source`
overview; the route lives outside gnoweb's navigation (empty omnibar, no
breadcrumb, no tabs) and needs a view mode of its own. Rendering the store
realm's own Content tab reuses all of it, and leaves `/explore` as a plain
alias.

**Reviving the Hall of Realms instead.** It already has self-registration
and votes, and its reviews are a useful record. Rejected as the base: its
data model (free title and description, up- and downvotes, owner deletion)
is the part this ADR replaces, and it lives in a personal namespace. Its
rules are kept (see Context); its code is not.

**Downvotes.** Rejected: a downvote is the cheapest way to bury a competitor,
and it adds nothing to discovery that flags (demote, never hide) do not
already cover.

**Pure markdown in the realm, no gnoweb template.** Still available through
`Render("")`, and it is the fallback when the richer view cannot be built.
Rejected as the main surface: no Spotlight, no generative covers, no pulse. That
would be a directory, not a store.

**`r/sys/store`.** Rejected: see §1. Governance can be as strict as `sys`'s
without putting editorial data in a system namespace.

**A curated list in gnoweb's repo (like gnoscope's `apps.json`).** Rejected:
every change needs a release, other chains fork the file, and listing is
gated by GitHub access instead of on-chain identity.

**Trust decided by the realm** (a `Verified` flag set by editors).
Rejected: a second trust list next to `TrustedPaths` that would drift from
it, a new target for key compromise, and a stronger claim than #6191 makes.

**Owner images and links for everyone.** Rejected: this is the main
moderation cost and the main phishing vector. Gating by tier, delay and
content addressing keeps the visual richness where it is safe.

**A human review queue for every listing.** Rejected: it does not scale,
and it reintroduces oblivion through backlog. Tiers and automatic demotion do
the same job without a queue.

**Free-text descriptions, news and tags.** Rejected: the doc comment, chain
events and the closed taxonomy carry the same information with nothing to
moderate. Governed tags may come back later (Follow-ups); free-text ones
will not.

**App metadata in each app's `gnomod.toml`, crawled.** Deferred: it needs a
gnomod schema change, a crawl gnoweb cannot afford without an indexer, and a
curation layer on top. It can feed `Register` later.

**Server-side screenshots (gnoscope's `/api/shot`).** Rejected: an external
headless browser breaks the no-dependency rule. The app's own page is one
click away anyway: it is the app itself.

**Ranking computed in gnoweb in the chain-only mode.** Rejected: the rotation
must be identical across instances and readers, so the realm computes it. In
indexer mode gnoweb must rank, in one tested function.

**One global ranking for everyone.** A conscious non-goal, as jefft0
argued for boards ([#3226](https://github.com/gnolang/gno/issues/3226#issuecomment-2505573679)):
"I don't think there should be a 'one size fits all' single ranking
system." The store ranks with one signal it can defend (ranked stars) and
balances it with New, Rediscover and gno.land essentials; other rankings (indexer
usage, curated lists, other frontends reading the same API) can sit beside
it rather than replace it.

**Impression tracking to balance exposure.** Rejected: gnoweb is stateless.
Epoch rotation gives a provable exposure guarantee with no state.

## Consequences

### Positive

- A discovery surface on every deployment with a store realm, with or
  without an indexer.
- Guaranteed exposure for every listed app: new arrivals first, then
  rotation for as long as it is active.
- Near-zero moderation: no free text, no owner links, no ungated or mutable
  images, automatic tiers, rankings a script cannot buy, GovDAO hiding
  without deleting, no admin key.
- A visually rich store from day one through generative covers, an
  always-present hero, the Spotlight, the pulse, shelves and moments, none
  of which needs review; the hero and Spotlight choose themselves among
  eligible apps.
- The store is a realm page: breadcrumb, omnibar, tabs, notice and title come
  from gnoweb, and the app's page is its own realm page. Nothing is
  duplicated, and gnoweb's core grows by about 100 lines with no change to
  existing behaviour.
- One trust list for all of gnoweb (#6191), not two.
- Permissionless listing: an app realm registers itself.
- One cached `qrender` per store page view; off with a flag. Markdown
  clients and failures see the realm's own render.

### Negative

- Registered and anonymous apps look plainer until they earn reputation. This
  is deliberate, and the generative covers soften it. They also never reach
  the hero or the Spotlight on their own, however many stars they have,
  until they earn reputation or GovDAO picks them.
- `richDelay` slows down legitimate image updates by about a day.
- Owner images must be on IPFS and pinned by their owner; no website link,
  no tags, and a claimed package has no owner images.
- A new app's first stars from fresh addresses or fresh names show on its
  card but do not rank it: ranking waits for names the store has known for
  3 days, and while names are free this is a delay rather than a price.
- A claimed listing shows nowhere until three ranked stars or its realm's
  `Register` confirm it, except a claimed app on its category page.
- Hiding a listing takes a GovDAO vote: no fast unilateral moderation until
  gnoweb's `-store-deny` lands.
- The path is permanent: a wrong constant or a bug stays in `v0`, and a `v1`
  means apps re-`Register` and stars start over.
- Storage: a demo with about 40 listings and 70 ranked stars needed about
  1.27 MB of realm storage, about 127 GNOT of deposit at 100 ugnot per byte,
  since each listing carries several index entries and its own trees. Each
  writer pays its own deposit; the seed's deploy pays for the seed.
- Names are free (0 ugnot for a `nym-` name on mainnet today), so ranked
  stars and Rediscover's named-only rule are a delay, not a price, until
  GovDAO sets `registerPrice` above zero.
- `/explore` becomes a default alias when the store is on (an operator's own
  `/explore` alias wins).
- The store has no detail band on app pages yet: a reader on `/r/acme/swap`
  does not see its stars or category until that follow-up lands.
- The JSON-over-`Render` contract is versioned by path; gnoweb must tolerate
  unknown and missing optional fields.
- Indexer shelves inherit #6231's caveats (freshness, windowing).
- The trusted tier inherits #6191's trust model, including its reliance on
  namespace enforcement and on the list being maintained.
- IPFS assets need an allowlisted gateway in `cspImgHost`.

## Plan

Each phase ships on its own.

1. **Realm** (this PR): `r/gnoland/store/v0` with listings keyed by global
   slug, unique titles, the taxonomy, a 5-per-namespace cap, stars with the
   daily budget and ranked-star rankings (name maturity, no owner
   self-stars), confirmation of claims (3 ranked stars), one `track` gate
   for the feed and pulse, IPFS-only owner images with the publication delay, earned
   reputation, pulse, activity ring, shelves and the sliding rotation,
   exact bounded Trending, binary-searched Rediscover, bounded builder
   reads, writes that move only what changed, an explicit list of core
   namespaces kept out of the builder rankings, the updated-farming fix, a `Register`
   that never panics, `Claim` (`ownsNamespace`), GovDAO proposal builders
   (pause, unpause, hide that releases slug, title and slot, unhide, and a
   front-page hero pick of 1 to 30 days, ended by a hide, a kind change or
   a rename) with an integration test, the seed catalogue (12 listings on paths that exist on
   mainnet and onyx), bounded `api/v1` (`home`,
   `category`, `build`) and the paged markdown store, linked from the
   realm's own path so it works at any `-store-realm`. Next: flags, `Retire`
   and deposits. Adversarial tests for T1, T6, T8, T10 and
   T11.
2. **gnoweb, chain-only** (this PR): `feature/store`, `-store-realm`, the
   Content-tab hook with markdown fallback, the `/explore` alias, tiers, the
   header icon, the fixed front page (the always-present hero: GovDAO's
   pick, else trending, else a daily rotation; the Spotlight
   with its launch fallback, pinned New and Top
   with empty texts, then Trending, Recently updated, Rediscover over named
   namespaces and gno.land essentials, all grids, no app twice, weekly momentum) and
   their list pages, category pages, the Build lens, generative
   covers, the Star and Surprise links, field validation and failure
   caching. Next: `-store-deny`, the
   names-enforcement startup check, liveness and private checks, flags, the
   small-catalogue fallbacks. Design review against §7. Launch is gated
   on the seed content and the "List your app" guide.
3. **Indexer signals** (after #6231): trending, most used, rising, back in
   action, built on, per-app stats, indexer moments, candidates queue.
4. **Spread**: the app band on realm pages, omnibar qualifiers, user-page
   shelf, not-found suggestions, home-realm block.

## Follow-ups (out of this PR)

- **Namespaced ids.** `<namespace>/<name>` instead of a global slug, so only
  the owner of a namespace can use a name and squatting disappears. Apps
  keep their chain path as their address (no `/a/` tree, see
  Alternatives).
- **App band on canonical realm pages.** A slim "In the store: ★ n ·
  Category" band on a listed realm's own page, with the star link. Later,
  because it touches existing pages and adds a store call to them.
- **Richer surfaces, later**: live previews of a realm's render on the front
  page (framed, capped, with a timeout, T4), an author showcase render path,
  Code DNA, the ecosystem map, editorial stories, similar apps and more from
  this author, other category sorts.
- **Exploration slots in ranked sections**: 2 of a section's slots for rotation
  picks from the same category, still later (§5).
- **Tags and tag pages.** A governed tag list, a `tags` field, and
  `:t/<tag>` pages, once the catalogue outgrows its categories.
- **Website link.** Removed from this version (§1). If it returns, it is
  gated like images: trusted or earned only, behind the publication delay,
  with the domain shown and IDNs in punycode.
- **Confusable-safe titles.** Uniqueness on the full UTS #39 skeleton,
  beyond today's script-mixing refusal and ASCII folds (T6).
- **Registration height in `r/sys/users`**, replacing the store's own name
  clock (§5): a name's age would then be known from its first day, not from
  its first star here.
- **A chain-side "package exists" check** usable from a realm, so `Claim`
  can refuse a path nothing is deployed at.
- **A shared confusable-fold package** (`p/…`), so the store, `r/sys/users`
  and other registries fold names the same way.
- **An enforcement-independent owner query in `r/sys/names`**, so
  `ownsNamespace` need not special-case chains where names are not
  enforced.
- **Priced names.** Ranked stars delay fresh names but do not price them
  while registration is free; a non-zero `registerPrice` in GovDAO, or a
  store-side cost, would (T8).
- **Health score, on data that cannot be farmed.** A score built on
  docs, liveness and recency is trivial to game (a one-line comment, a
  re-registration). It comes back only on signals that cost an attacker
  real money or time, which need the indexer (#6231):
  - distinct callers and gas spent by them over 30 days;
  - longevity of that usage (weeks with activity, not a one-day burst);
  - imports by realms in *other* namespaces (a deploy fee each);
  - failure rate of calls.
  Shown with its breakdown, labelled as indexer data, never as "verified".
- **Indexer-backed discovery**, through a narrow `Signals` interface wired
  like #6231 (nil means absent): "Most used", usage-based "Trending",
  "Rising", "Back in action", "Most imported", per-app activity
  sparklines, the usage path to earned reputation, and the editors'
  candidates queue (realms with traction that are not listed).
- **Social proof.** Recent starrers as faces next to the star count, once
  profile avatars land on chain; until then the star count alone.
- **Personal recommendations** once gnoconnect (#5970) gives the browser an
  address.

## Open questions

- Editor set: a dedicated GovDAO-appointed group (`p/nt/groups`, as in
  #6009), or devrels to start with? With zero-ops curation this matters less.
- Off-chain apps (wallets, explorers) as gnoscope lists them: allow
  website-only listings for trusted namespaces?
- IPFS gateway: which one goes into `cspImgHost` (`ipfs.io` is already
  there), and do we pin seed assets ourselves?
- `richDelay`, reputation thresholds and the namespace cap are constants,
  permanent for `v0`: should `v1` read them from GovDAO parameters
  (`r/sys/params`)?
- Should `vm/qpkgmeta_json` expose `private` directly, so gnoweb does not
  need to fetch `gnomod.toml`?
- Default `-store-realm`: on for gno.land deployments only, off for gnodev?
