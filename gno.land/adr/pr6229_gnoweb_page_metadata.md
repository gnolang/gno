# ADR: gnoweb page metadata

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

Fill every slot on every page that answers 200, from the URL and, on pages
gno.land answers for, from what the page displays.

### Who answers for the page

#3910 asks that search engines and link previews tell gno.land's own pages
from user content, so that nobody can use metadata to speak under gno.land's
name. `packageKind` in `page_kind.go` sorts every page into one of three
kinds, using the `-trusted-paths` list that #6191 introduced for the realm
notice:

- **official**: a markdown page an operator passed to `--aliases`, or a
  `/r/`, `/p/` or `/u/` page whose package is under a trusted path;
- **community**: any other `/r/`, `/p/` or `/u/` page;
- **site**: a view that belongs to no package, such as the bare `/r/`
  listing.

An alias to a realm takes the kind of the package it serves: it publishes
whatever its target renders, so aliasing a community realm must not lend it
gno.land's card. Only a markdown alias, whose bytes the operator wrote, is
official by itself. The zero value is community, so a page nobody classified
fails closed. The notice and the metadata read the same classification, and `cmd/gnoweb` passes
the trusted list even when `-no-realm-notice` hides the notice. A library
caller that sets no `TrustedPaths`, gnodev included, gets community metadata
on every package page.

### What each kind publishes

`setHeadMetadata` in `handler_http.go` runs once the body is rendered,
against the URL the client asked for rather than the alias target, so
`/about` names `/about` and not `/r/gnoland/pages:p/about`. The body hands it
a `pageLead`, what the document says about itself, or an operator's front
matter; `HeadData` is written only by `setHeadMetadata`, so no early return
can print text it has not vetted. `markdown.Lead` reads the lead as plain
text: the title is a leading h1, the first block of the document, and the
summary the first paragraph long enough to summarise between that h1, or the
document start, and the next heading or rule. Text further down belongs to a
section, a list or a post a trusted realm shows on behalf of its users; a
forum's "# Claim your airdrop" post must not title the forum.
`r/gov/dao` opens on `# GovDAO` then `## Members`, so it keeps its title and
gets the site sentence instead of an `Author: g1...` line. `setHeadMetadata`
then decides:

| | title | description | image |
|---|---|---|---|
| official, no args or query | the h1, or the path | the paragraph, or `siteDescription` | `og-gnoland.png` |
| official, with args or query | the path | `siteDescription` | `og-gnoland.png` |
| site | the path | `siteDescription` | `og-gnoland.png` |
| community | the path | a fixed sentence per kind (realm, package, user) | `og-community.png` |

The domain follows the title: `The gno.land blog - gno.land`. The page comes first
because a browser tab and a search result both truncate the tail.

The path is the only part of the URL a title repeats. Arguments and query
are typed by whoever wrote the link, on any realm, trusted ones included:
`/r/gnoland/blog:Official_GNOT_airdrop_claim_at_evil.example` used to title
itself with that sentence. Both also reach `Render`, and trusted realms echo
them: `p/gnoland/blog` writes a tag into its tag page's heading
(`# Gno.land's blog / t / <tag>`), and `r/gov/dao` prints the parse error of a
proposal ID. So an official page lends its h1 and paragraph only when its link
carries neither args nor query; otherwise it gets the path and the generic
sentence. A page an operator aliases keeps its h1, since the alias URL has no
args and the target's are the operator's.

The community wording names who published the page rather than who did not
review it: "A realm deployed on gno.land by its author." A stamp of what
gno.land has not reviewed would read as a gatekeeper on a permissionless
network. `og-community.png` marks the card as community content; its caption
is to be reworded the same way. An empty card reads as a broken site, so every 200 page carries a title, a description and,
when `-canonical-origin` is set, an image.

`canonicalURL` prefixes `Static.CanonicalOrigin`, set by `-canonical-origin`,
to `GnoURL.EncodeWebURL`, the encoder gnoweb's own links use, with the query
removed. `head.html` wraps `description`, `og:description`, `og:image`,
`og:url` and `twitter:*` in plain `{{ if }}`s for the shells that have no
input, and `twitter:card` falls back from `summary_large_image` to `summary`
when there is no image.

### The summary repeats what the page shows, and nothing else

#3910 asked whether a permissionless page may author metadata, and #3797 was
reverted by #3924 for answering yes. The answer here is narrower: only an
official page lends its own text to the head, and only text a reader already
sees, so there is no slot it can fill without displaying what it filled. Image
alt text is excluded because an alt shows only when the image fails. Front
matter is read back, but only from the markdown files an operator passes to
`--aliases`; realm content never reaches that path.

A canonical URL needs a public origin, and a deployment that names the wrong
one tells a crawler its content belongs elsewhere. `-canonical-origin` is empty
by default: no origin declared, no canonical tag, because every deployment but
one would otherwise be claiming gno.land's. A value that is not a bare
`scheme://host[:port]` stops gnoweb at startup: `gno.land` alone would render a
relative href, and every page would name a 404 as its canonical.

The query is dropped from the canonical and `og:url` of every page. It does
reach `Render`, so two queries can render two pages, but anyone can append one
to any page: keeping it would publish one page under as many URLs as there are
queries, and let a link carry a sentence into the URL gno.land declares as
canonical. Arguments stay, since they address the pages of a realm
(`/r/gnoland/blog:p/hello-worlds`). A static page renders the same bytes
whatever the query says, so `/about?utm_source=x` also keeps its front matter.

An error shell drops its canonical, its share image and its indexability. The
head is assembled before the body knows the page is missing, so a mistyped path
would otherwise publish itself as a real URL. `unpublishErrorShell` holds the
rule, in Go; the template only renders what it is given.

## Alternatives considered

**Build the canonical from the request host.** `requestOrigin` already reads
`X-Forwarded-Host`, which any caller can set. A canonical link built from it
points crawlers at whatever host the caller named, so the configured origin
is the input instead. The cost is that a chain reachable under a second
hostname advertises only the configured one.

**Keep the domain first in the title.** Shorter diff, and it keeps the
existing look. A tab strip and a search result both cut the tail, so the
part that identifies the page is the part that survives only when it leads.

**Withhold the image from realm pages.** The first version of this change
served the share image only to operator pages. A link preview without an
image looks broken, and a community page posted anyway still shows gno.land
as its host; a dedicated card that marks the page as community content
answers #3910 better than no card.

**Keep the arguments in the title.** Each blog post had its own title that
way, even without a heading. Any link could also set the title of a trusted
page to a sentence of its choosing, which is the exact abuse #3910 names.

**Take the h1 of a page with args unless it repeats them.** This would give
each blog post its own heading as a title. A string match cannot tell an echo
from a heading: a post's slug is made of its title's words, so a check strict
enough to drop `/ t / <tag>` drops every post title too, and one loose enough
to keep post titles lets a reformatted echo through. A realm that wants its
pages titled can be aliased, or declare titles itself in a later change.

**Take the first h1 or paragraph anywhere in the document.** More pages would
get a title and a summary. A trusted realm that lists user posts, comments or
proposals would then be titled and summarised by whichever of them came first.

**Treat every alias as official.** An operator chose it, but an alias to a
realm renders what the realm's deployer wrote, and the notice already says
so on the page.

**Canonicalise every view to the content page.** `$source` and `$help`
render different content from the realm body, so each addresses its own
page and gets its own canonical.

## Consequences

Every page now carries a title, a description and, with an origin set, a
share image and a canonical URL.

An official page that opens on an h1 and has no args or query is titled by
it. Every other page is titled by its path: the posts of one realm, trusted or not, share
their realm's title, as do `$source` and the content page of one realm, while
their canonicals differ. Directory listings, `$source`, `$help` and profile
pages on official paths carry the generic site sentence, since they have no
paragraph to summarise, as does any official page whose prose sits under a
later heading.

Pages that differ only by query now share one canonical.

Operators must set `-canonical-origin` to get a canonical tag, and gno.land's
own deployment is one of them. Until it does, the tag is absent, which is the
safe direction.

`-canonical-origin` includes the scheme, so a deployment served over plain
HTTP declares `http://` there and gets an `http://` canonical.

`robots.txt` and `sitemap.xml` remain absent. Both answer 400 on gno.land
today, since neither is a valid gno path. Their content is a policy question
#5769 puts to maintainers.
