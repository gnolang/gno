# ADR: a standalone script for the home page's "What's new" block

## Status

Proposed. Replaces the `gnoalias.sh` part of #6327.

## Context

The static home of mainnet and the testnets never changes between two
deployments, so nothing on it shows that the project is alive. The goal is one
section that does: recent posts from X, the blog, videos and releases, next to
the most used realms, the most imported packages and the latest merged pull
requests.

#6327 did this for the blog from `gnoalias.sh`, a sidecar that rewrites the page.
Those networks run no such sidecar: their page is a static file read at startup.
Whatever updates the block has to fit that, and has to be a small thing for the
operator to adopt.

## Decision

`misc/deployments/home-alias/whatsnew.sh`, one file needing only `bash`, `curl`
and `jq`. It takes the path of the page and updates it between two marker lines,
so the pinned file stays the source of the rest of the page and can be deployed
without the script.

- **One section**, not one per source: a post on X every day keeps its top
  recent, where a section for the blog alone would show a post weeks old. X, the
  news and the rankings each get a column, on the page's three-column grid, so that X cannot
  push the others out and the block lines up with the sections around it.
- **Each source fails alone.** Its last good output is kept in a state folder
  and its last attempt sets when it is read again, so the timer can be frequent
  while the expensive sources stay rare. A ranking older than a week goes: it
  would no longer read as "this week".
- **Text from X comes from X.** FxTwitter only supplies post IDs; oEmbed supplies
  text, date and author, and the author is checked. An X Article only has a link
  there, so its title comes from X's syndication CDN. A third party can then
  choose which of the account's posts appear, but not what they say.
- **All outside data is validated or escaped**, because the page is treated as
  operator content. Text stays on one line and every ASCII punctuation mark is
  backslashed, so no post can form a link, image, tag, mention or icon, nor close
  a block. Links are built by the script, from identifiers and paths checked
  against a strict pattern. Numbers from the indexer are checked too, since bash
  arithmetic would run what a string holds. t.co links are dropped, except those
  X made out of a bare "gno.land".
- **Realms are ranked by distinct addresses**, since raw calls are dominated by
  bots (one realm's `Heartbeat` made 4,592 of its 4,664 calls in a week), and
  only registered namespaces chosen in `USED_NAMESPACES` can appear, since other
  paths are permissionless and addresses are free. Namespaces rather than single
  realms, so that a new realm of a known project needs no change.
- **The scan copies gnoweb's chain map** (`feature/chainmap/activity.go`): same
  rate probe, band width and window, in a language a lone script can run. Tuning one
  does not reach the other.
- **No relative dates, and no post counters.** They would change the page on
  every run, and each change costs a reload.
- **Two audiences.** Mainnet greets visitors. A testnet is where people come to
  build, so `home.testnet.md` becomes a page for them, with onboarding by case
  and a deploy command, and `--builders` fills it with chain activity. It
  runs with `USED_NAMESPACES` empty: real projects sit under `g1` addresses, and a
  gamed ranking costs little on a testnet.
- **The latest candidate, not the node's version.** The RPC answers `v1.0.0-rc.0`
  on a chain running v1.7, so the line names GitHub's latest candidate and
  claims no more than that.

## Alternatives considered

- **A gnoweb feature filling the block per request.** It needs no timer, but
  gnoweb would gain calls to third parties in its request path.
- **`r/gnoland/home` on mainnet.** A realm cannot read X or an indexer, so a bot
  would have to push them in transactions with a permanent key, and any template
  change would need a new version of the realm.
- **Republishing the file.** It would need a deployment for every post.

## Consequences

- The page needs something to run the script. Until then the markers render as
  nothing and the rest of the page shows.
- The served file now differs from the committed one, by the block alone. A
  deployment restores the committed page, and the next run fills it again.
- The block depends on gnoweb re-reading a changed page, from #6327. A restart
  per change only suits a single host: a pool would drop requests each time, and
  the totals of `--builders` change once a day on their own.
- FxTwitter and the YouTube feed are outside the project's control, and the feed
  answered 404 and 500 at random while this was written. Each only costs its
  part of the block.
- A new partner shows in "Top realms" once its namespace joins `USED_NAMESPACES`.
