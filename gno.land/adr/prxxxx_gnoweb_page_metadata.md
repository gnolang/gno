# ADR: gnoweb page metadata from the URL

## Context

`components/layouts/head.html` declares `description`, `canonical`, four
`og:*` tags and three `twitter:*` tags. `HeadData` in
`components/layout_index.go` carries a field for each. The HTTP handler set
one of them.

Two consequences on the live site:

- The `<title>` was `h.Static.Domain + " - " + gnourl.Path`, and `Path` stops
  at the realm. Every post under `/r/gnoland/blog` therefore rendered
  `gno.land - /r/gnoland/blog`, so one realm published one title over an
  unbounded number of pages, and the same held for `$source`, `$help` and
  every argument-addressed page.
- `Description`, `Canonical`, `Image` and `URL` were never assigned, so the
  page shipped `content=""` on five tags and no canonical link at all.

Issue #5769 lists both, alongside a sitemap, a robots.txt and a description
source. Issue #3910 is the reason the description stalled: realm content is
permissionless, so repeating it in a meta tag lets any deployer write text
that search engines attribute to gno.land.

## Decision

Fill the slots whose only input is the URL gnoweb already parsed, and stop
declaring the slots that have no input.

`setHeadMetadata` in `handler_http.go` sets `Title`, `Canonical` and `URL`
from a `weburl.GnoURL`. Two helpers back it:

- `pageTitle` encodes path, arguments, view marker and query, then appends
  the domain: `/r/gnoland/blog:p/hello-worlds - gno.land`. The page comes
  first because a browser tab and a search result both truncate the tail.
- `canonicalURL` prefixes `https://` and the configured `Static.Domain` to
  `GnoURL.EncodeWebURL`, the encoder gnoweb's own links use.

`setHeadMetadata` runs once in `Get`, against the URL the client asked for
rather than the alias target, so `/about` names `/about` and not
`/r/gnoland/pages:p/about`. `setHeaderForRealm` keeps `HeaderData` only.

`head.html` wraps `description`, `og:description`, `og:image`,
`og:url` and `twitter:*` in `{{ if }}`, so an unfilled slot emits nothing.

## Alternatives considered

**Build the canonical from the request host.** `requestOrigin` already reads
`X-Forwarded-Host`, which any caller can set. A canonical link built from it
points crawlers at whatever host the caller named, so the configured domain
is the input instead. The cost is that a chain reachable under a second
hostname advertises only the configured one.

**Derive the description from the rendered realm.** This is the #3910
question and it is left open. Leaving the tag out is reversible; publishing
attacker-chosen text under gno.land's name is not.

**Keep the domain first in the title.** Shorter diff, and it keeps the
existing look. A tab strip and a search result both cut the tail, so the
part that identifies the page is the part that survives only when it leads.

**Canonicalise every view to the content page.** `$source` and `$help`
render different content from the realm body, so each addresses its own
page and gets its own canonical.

## Consequences

Every page under one realm now has a distinct title and a canonical URL.

Nothing yet emits a description or a share image, so a link posted to a
social network still renders without a snippet or a card image. Closing
that needs #3910 answered first.

A deployment served over plain HTTP advertises an `https://` canonical.
Adding a scheme to `AppConfig` would close it and is not done here.

`robots.txt` and `sitemap.xml` remain absent. Both answer 400 on gno.land
today, since neither is a valid gno path. Their content is a policy question
#5769 puts to maintainers.
