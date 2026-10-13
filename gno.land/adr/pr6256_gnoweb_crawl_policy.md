# PR6256: gnoweb crawl policy

## Context

gnoweb serves every network under the gno.land name: mainnet, testnets,
staging, gnodev. Nothing told a crawler which one is real, so testnet and
staging pages are indexed and compete with mainnet. gnoweb also had no
`robots.txt` and no `sitemap.xml` (#5769).

#6229 decides, per page, what search engines may do (`pagePolicy.robots`) and
emits the canonical from `-canonical-origin`. This PR builds on it.

## Decision

**Indexing is on unless the operator passes `-no-index`.** The default changes
nothing for an existing deployment: a forgotten flag never drops mainnet from
search. Testnets and staging opt out.

### `-no-index`

- `pagePolicy.robots` returns `noindex, nofollow` for every page, whatever its
  kind. The rest follows from #6229: the robots meta and `X-Robots-Tag` say so,
  and `setHead` names no canonical. The share card (`og:url`, `og:image`) stays,
  as on any noindex page.
- `noIndexMiddleware` sets the same header on every response, for those that
  never reach the page handler: assets, JSON routes, redirects, `/faucet`.
- `/sitemap.xml` is a 404 and `robots.txt` names no sitemap.

### `robots.txt`

Served on every deployment, the same rules everywhere:

```
User-agent: *
Disallow: /search.json
Disallow: /status.json
Disallow: /ready
```

These routes query the node and are not pages. No page is ever disallowed: a
blocked URL hides its noindex and can stay listed from links alone. Under
`-no-index` a comment line says the site is not indexed. With
`-canonical-origin` and without `-no-index`, a `Sitemap:` line follows.

### `sitemap.xml`

Mounted only with `-canonical-origin` and without `-no-index`. It lists the
operator's alias pages, the curated entry points; crawlers reach realms
through links. An alias is listed when all of these hold:

- gnoweb parses the key, no router route or `Redirects` entry owns it, and the
  router would not clean and redirect it (`/a/..`);
- its page names its own URL as canonical (`canonicalURL` of the key, as `Get`
  names it). A static key may read as a file (`/license.md`); a realm alias
  keyed `/x.md` names `//x.md` and is left out;
- `pagePolicy` indexes it, classified as `Get` classifies it: a static page by
  its key, a realm alias by its target. A `$help` target or an address
  namespace is left out;
- its target has something to show: one `vm/qfile` query for a `/r/` or `/p/`
  target, one namespace listing for a `/u/` target. One query per target, not
  the full listing, which the node caps at 1000 entries. A target that is no
  package, such as a listing or another alias, is left out.

URLs come from the configured origin only, never from `Host` or
`X-Forwarded-Host`. A failed query leaves its alias out, not the whole
sitemap, logs one warning per request, and that response is not cached
(`no-store`); otherwise it carries a one-hour `Cache-Control`. Queries share
the handler's request timeout; `HEAD` makes none.

gnodev always runs with `NoIndex`: a dev or staging chain is never the
canonical site.

gnoweb logs `noindex` and `canonical_origin` at startup.

## Alternatives Considered

- **Indexable only with `-canonical-origin`, or noindex derived from the
  network kind.** A forgotten flag on mainnet would take it out of search.
- **`Disallow: /` off mainnet.** It hides the noindex, so indexed URLs stay.
- **Middleware only, clearing the canonical by hand.** #6229's policy rewrote
  the header to `noindex, follow` on query and `$` pages, and the canonical
  had to be cleared in two places. One decision in `pagePolicy` covers every
  view.
- **Every realm in the sitemap.** Most are demos or near empty; links reach
  the rest. A curated registry could feed it later.

## Consequences

- A deployment without the new flag keeps its pages indexable; it gains
  `robots.txt`, and a sitemap once `-canonical-origin` is set.
- Every testnet and staging deployment must pass `-no-index`. Already-indexed
  URLs drop as crawlers revisit them; Search Console's removal tool speeds this
  up. Their pages must stay crawlable.
- Mainnet must run with `-canonical-origin https://gno.land` for the sitemap to
  exist, then submit it to Google and Bing. `-canonical-origin` is each
  deployment's own origin, never copied from mainnet: share cards follow it.
- After a mainnet deploy, `curl -sI https://gno.land/` must show no
  `X-Robots-Tag`.
- An operator alias keyed `/robots.txt` or `/sitemap.xml` is now shadowed.
- Two aliases of one target are two pages naming themselves, so both are
  listed. Pointing one at the other is #6229's naming.
- A realm that renders "not found" with status 200 behind an alias is still
  listed. A `/u/` alias whose packages all sit under the address half of a
  registered name is left out.

## Key Files

- `gno.land/pkg/gnoweb/crawl.go`: `noIndexMiddleware`, `robots.txt`, `sitemap.xml`
- `gno.land/pkg/gnoweb/community_index.go`: `pagePolicy.robots` under `-no-index`
- `gno.land/pkg/gnoweb/app.go`: routes, and when the sitemap is mounted
- `gno.land/cmd/gnoweb/main.go`: the `-no-index` flag
