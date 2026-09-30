# ADR-XXXX: gnoweb crawl policy

## Context

gnoweb serves every network under the gno.land name: mainnet, the testnets,
staging, gnodev and PR previews. Nothing told a crawler which one is real.
`robots.txt` and `sitemap.xml` answered 400 on every deployment, which Google
reads as "no robots.txt", and `head.html` hardcoded
`<meta name="robots" content="index, follow">` on every page.

The result is in the search results today. `site:staging.gno.land` returns
stress-test packages, `$source&file=` pages and a `$help&func=...` form URL;
`site:testnets.gno.land` returns old testnet hosts and an RPC node. Those pages
compete with mainnet for the same queries, and send people who search for a
realm to a chain that may be gone.

This is the crawl part of #5769 and §2.6 of #6121. #6103/#6229 added
`-canonical-origin`, empty by default, for canonical tags.

## Decision

**Every deployment stays indexable unless its operator passes `-noindex`.**
The one outcome this change must never produce is mainnet dropping out of
search, so the default changes nothing for any existing deployment: forgetting
a flag leaves things as they are today, it never removes a site from the index.
Testnets, staging and previews opt out explicitly.

With `-noindex`:

- every response carries `X-Robots-Tag: noindex, nofollow`, set by one
  middleware around the whole mux, so assets and JSON endpoints are covered too;
- the page meta says the same (`HeadData.NoIndex`). It is not redundant: a
  static export such as the PR previews on GitHub Pages (#6194) cannot send
  headers, so there the meta is the only signal;
- `robots.txt` still allows crawling. A `Disallow: /` would stop crawlers from
  fetching the pages that carry the noindex, and the URLs already in the index
  would stay there as bare links;
- `/sitemap.xml` is a 404, even with a canonical origin;
- no page names a canonical (see below).

Without it:

- `robots.txt` allows everything except `/search.json` and `/status.json`.
  Today it answers 400, which crawlers already treat as "allow everything", so
  nothing changes for pages;
- with `-canonical-origin` set, `robots.txt` also names `/sitemap.xml`. The
  sitemap lists the operator's alias pages only (home, about, ecosystem, ...),
  the curated entry points of the site, served with a one-hour
  `Cache-Control`. Realms and packages are left out on purpose: on mainnet 39%
  of them have under 300 characters of text and nearly a third are demos, so listing
  them adds nothing crawlers do not already reach through links, and pushes
  thin pages forward. A static alias is always listed. A path alias is listed
  only when its realm exists on this chain (checked against the same
  `RealmDirectory` as `/search.json`), so the default aliases do not publish
  404s on a chain that lacks `r/gnoland/pages`, and `/docs` (a user page, not a
  realm) is left out. Alias keys gnoweb would redirect or rewrite are skipped.
  Without a canonical origin there is no sitemap, as today.

On every deployment, an action form (`$help&func=...`) is noindex. A realm can
link to one per call, so they are an unbounded set of near-identical pages; the
function list at `$help` and the `$source` views stay indexable. This is done
with the meta, not robots.txt: `$` is an end anchor in robots patterns, and
Google compares rules against the raw path, so no single pattern works for both
the RFC and Google.

**A noindex page names no canonical.** Google may carry a noindex over to the
canonical target, so a staging deployment that copied mainnet's flags and added
`-noindex` would otherwise point every page at gno.land with a noindex on it.
Under `-noindex`, `NewRouter` drops the canonical origin itself, so no view
(including `$state`, which bypasses the page handler's head logic) can emit a
canonical, `og:url` or `og:image`. On other deployments the page handler clears
them for action forms and error pages, as #6229 already did for errors.

**Absolute URLs come from the configured origin only**, never from `Host` or
`X-Forwarded-Host`, so a request cannot put another domain into a cached
sitemap. The origin is normalized at startup (lowercase, default port and
trailing slashes dropped, surrounding spaces trimmed), because a harmless
spelling should not stop gnoweb from starting. Anything that is not a bare
`http(s)://host[:port]` (a path, a query, credentials) still stops it: the
value is copied into every canonical tag, and a wrong canonical costs more than
a failed rollout. Paths are escaped for the URL and then for XML.

gnoweb logs its crawl policy (`noindex`, `canonical_origin`) at startup, so a
deployment's policy is visible without fetching a page.

## Alternatives considered

- **Indexable only when `-canonical-origin` is set** (the first revision of
  this change). Rejected: it turns a forgotten flag in production into mainnet
  leaving Google, and recovery takes days to weeks. The opt-out fails the other
  way: a testnet that forgets `-noindex` stays indexed, as it is today.
- **Derive noindex from `-network-kind` (#6158).** Same objection: #6158
  defaults to `testnet`, so a mainnet that forgets `-network-kind=mainnet`
  would go noindex. If the two are ever linked, `-noindex` must stay the only
  thing that can turn indexing off.
- **`Disallow: /` off mainnet.** Rejected, see above: it hides the noindex.
- **`lastmod` in the sitemap.** No honest source: gnoweb does not know when a
  realm last changed, and a wrong `lastmod` teaches Bing to ignore the field.
- **`llms.txt`.** Left out. Measured adoption is close to zero, and what agents
  do use, `Accept: text/markdown`, is #5794.
- **Per-bot rules for AI crawlers.** Left out: the goal of #5769 is to be read
  and cited, so every bot is treated the same.

## Consequences

- A deployment that sets no new flag, mainnet included, keeps every page
  indexable. The only differences are `robots.txt` answering 200 instead of
  400, `/sitemap.xml` answering 404 instead of 400 (200 once a canonical origin
  is set), and action forms turning noindex without a canonical.
- Testnets, staging and previews must pass `-noindex` to leave the index; they
  drop out as crawlers revisit them. Removal requests in Google Search Console
  and Bing Webmaster Tools speed that up for the hosts already listed.
- The sitemap only exists once production sets
  `-canonical-origin=https://gno.land`, and should then be submitted to Google
  and Bing.
- Adding realms to the sitemap is a curation question, not a technical one: a
  registry of featured realms (#6191's territory) could feed it later. The
  path alias check reads the realm list, which gnoweb caps at the node default
  of 1000 per prefix, as for `/search.json`; the node accepts up to 10000, so
  forwarding the limit is a small follow-up.
- `-noindex` is the one switch that takes a site out of search. An external
  check that `https://gno.land/` carries no `X-Robots-Tag` and still says
  `index, follow` catches `-noindex` leaking into mainnet through shared
  deployment config.
