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

- **official**: a `/r/`, `/p/` or `/u/` page whose package is under a
  trusted path, or a markdown page an operator passed to `--aliases`
  (`pageOperator`, official and the author of its own links);
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
| official, no args or query | the h1, or the path title | the paragraph, or `siteDescription` | `og-gnoland.png` |
| official, with args or query | the path title | `siteDescription` | `og-gnoland.png` |
| site | the path | `siteDescription` | `og-gnoland.png` |
| community | the path title | a fixed sentence per realm, package or user page | `og-community-{realm,package,user}.png` |

The path title, `pathTitle`, reads what the path says and nothing a realm
rendered: `/r/nym/games/chess` is "games/chess · realm by nym", `/p/nym/lib`
"lib · package by nym", `/u/nym` "nym · user profile". A namespace root such
as `/r/nym` has no realm of its own and reads "realms by nym". An address
namespace is shortened as the user page shows it (`g1jg...sqf5`). Other paths,
`/r/` or an alias without an h1, name themselves. The domain follows the
title: `The gno.land blog - gno.land`. The page comes first because a browser
tab and a search result both truncate the tail.

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
network. `og-community.png` follows: it leads with "Community realm" and
"Open source, running on-chain. Read its code, call its functions.", with
gno.land's mark small in a corner, as the platform the realm runs on rather
than its author, and "Deployed by its author · check before you sign" level
with it. Packages and profiles get the same card as "Community package" and
"Community profile". Cards are per kind for now, not per package: a card
naming the package ("app · realm by nym") is tracked in #6266. An empty card
reads as a broken site, so every 200 page carries a title, a description
and, when `-canonical-origin` is set, an image.

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

### Which community pages search engines may index

`indexable` in `robots.go` decides, once per request in `Get` and before any
branch writes, whether a page may be indexed. One that may not gets
`<meta name="robots" content="noindex, nofollow">`, an
`X-Robots-Tag: noindex, nofollow` header (the only signal a text/markdown
response can carry) and no canonical, since a canonical on a noindex page
contradicts it. It keeps its card: link previews ignore robots. Official
pages are always indexed. For community pages, `-index-community` chooses:

| `-index-community` | community pages indexed |
|---|---|
| `registered` (default) | the bare page of a package or user page whose namespace is a registered name: no args, no query, no `$` view |
| `none` | none |
| `all` | all of them, like official pages |

The `registered` line rests on the chain. `checkNamespacePermission`
(`gno.land/pkg/sdk/vm/keeper.go`) asks `r/sys/names.IsAuthorizedAddressForNamespace`
before every deploy, and once that realm is enabled it authorizes
`r/<ns>/` only for the address `<ns>` itself or for the current holder of the
name `<ns>` in `r/sys/users` (`examples/gno.land/r/sys/names/verifier.gno`;
`gno.land/pkg/integration/testdata/addpkg_namespace.txtar`). Mainnet enables
it at genesis (`misc/deployments/mainnet.gno.land/README.md`), betanet through
its first GovDAO proposal. An address namespace therefore costs nothing to
create and throw away, while a name is held by one account. gnoweb tells the
two apart by the path alone, with the bech32 check the user page already
uses, and makes no RPC. A chain that never enables `r/sys/names` lets anyone
deploy under any name; it should run `-index-community=none`.

Args, a query or a `$` view (`$source`, `$help`, `$state`, a download) let
any link multiply one package into unbounded URLs, so none is indexed.

### Links pass on authority only where gno.land answers for them

A followed link tells a search engine the linking site vouches for its
target. `pageKind.links` gives every rendered document a `markdown.LinkPolicy`,
through the one `renderContext` helper all four document views use (realm,
user page, README, operator markdown):

| document | internal link | external link |
|---|---|---|
| community realm, package README, user page | `nofollow ugc` | `noopener nofollow ugc` |
| trusted realm or README | followed | `noopener nofollow ugc` |
| operator markdown (`--aliases` file) | followed | `noopener` |

A trusted realm keeps its external links unfollowed because it may show
what its users wrote: a board, a proposal, a profile. The policy's zero value
follows nothing, so a document rendered without one is treated as user
content. Links from a `<gno-foreign>` sandbox stay unfollowed whatever the
policy. gnoweb's own header, breadcrumb and tab links are not documents and
keep no `rel`.

### Answering #3910

#3910 and its thread name the risks of user content under gno.land's name:
SEO bombing, dilution of the site's authority, scams and phishing, and harmful
content affecting gno.land's reputation. It asks that search engines can tell
gno.land's pages from user realms, and that risk come before SEO.

| concern | this change | what remains, and where |
|---|---|---|
| SEO bombing: user pages ranking on gno.land's name | community titles come from the path, descriptions are fixed, cards say the page is community content; args, query and `$` views of community pages are noindex; every link in a community document is `nofollow ugc` | the bare page of a registered-name realm stays indexable, with that fixed head, under `-index-community=registered` |
| Throwaway scam realms | an address namespace, free to create and discard, is noindex | none for search; the page itself still renders |
| Scams under a registered name | indexable, but with a path title, a fixed description, a community card and the #6191 notice on the page | content filtering in #5185; `-index-community=none` if that is not enough |
| Dilution of the site's authority through links | every link in a community document, and every external link in a trusted one, is `nofollow ugc` | links in `$help` doc comments go through the documentation renderer, which has no link policy and adds no `rel` |
| Scams and phishing through metadata | the head repeats a document only on official pages, only its leading h1 and paragraph, and never when the link carries args or a query | a trusted realm whose own lead shows user text would lend it; the trusted list is the control |
| Phishing through page content | out of scope for metadata | the community notice of #6191 on the page; content filtering in #5185 |
| Crawlers telling main pages from user realms | `noindex` and `X-Robots-Tag` on the community pages above, `ugc` links, distinct cards and descriptions, all from one classification | the sitemap of #6256 should list official pages only (a follow-up there); no structured data |
| Front matter as authored metadata (#3797, reverted by #3924) | front matter is read only from markdown files an operator passes to `--aliases` | none |
| Who counts as official | `-trusted-paths`, with a default list in `cmd/gnoweb` | governance of that list |

### Reversibility

`-index-community` changes the policy without a code change. Search engines
apply a noindex, or its removal, when they next crawl the page: days to weeks,
with no penalty for the change itself. A URL meant to carry noindex must stay
crawlable: blocking it in robots.txt hides the noindex and can leave the URL
listed from links alone, which matters for the robots.txt of #6256.

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
share image. Every indexed page also names a canonical URL.

Community realms beyond the bare pages under registered names leave search
results under gno.land, and with `none` all of them do. That costs the
ecosystem the discoverability #5769 asks for; the trusted list, and the
bare-page rule, are the way back in.

An official page that opens on an h1 and has no args or query is titled by
it. Every other page is titled by its path title: the posts of one realm, trusted or not, share
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
