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
gno.land answers for, from what the page displays; and tell search engines,
page by page, what they may index and follow.

### Who answers for the page

#3910 asks that search engines and link previews tell gno.land's own pages
from user content, so that nobody can use metadata to speak under gno.land's
name. `pagePolicy.kind` in `page_kind.go` sorts every page into one of four
kinds, using the `-trusted-paths` list that #6191 introduced for the realm
notice:

- **official**: a `/r/`, `/p/` or `/u/` page whose package is under a
  trusted path;
- **operator**: a markdown page an operator passed to `-aliases`; official,
  and the author of its own links;
- **community**: any other `/r/`, `/p/` or `/u/` page;
- **site**: a view that belongs to no package, such as the bare `/r/`
  listing.

An alias to a realm takes the kind of the package it serves: it publishes
whatever its target renders, so aliasing a community realm must not lend it
gno.land's card. A user page is one name: `/u/gnoland/anything` is not
found, and is classified community, rather than a page that borrows the
trust of `gnoland`. The zero value is community, so a page nobody classified
fails closed.

`pagePolicy` needs no request and no RPC, so anything that lists pages, such
as a sitemap, classifies them the same way. `Get` classifies once
(`classifyPage`) and hands the views a `pageRender` with the decided link
policy; no view classifies again. The notice, the head, the robots policy and
the link policy all read that one classification.

`DefaultTrustedPaths` ships with the gnoweb package and
`NewDefaultAppConfig` uses it. A `*` entry trusts every path; gnodev sets it,
since every package on its chain is the developer's own. Trust in a
namespace holds only on a chain that enforces who may deploy under it (see
below); an entry that can match no package (an `/r/` or domain prefix, an
uppercase letter) is logged at startup and still matches nothing.

### What each kind publishes

`setHead` in `handler_http.go` writes the whole head once the body is
rendered and its status known, against the URL the client asked for rather
than the alias target, so `/about` names `/about`. The views hand it a
`pageLead`, what the document says about itself, or an operator's front
matter, instead of writing `HeadData` themselves.

`markdown.Lead` reads the lead as plain text: the title is a leading h1, the
first block of the document, and the summary the first paragraph long enough
to summarise between that h1, or the document start, and the next heading or
rule. Text further down belongs to a section, a list or a post a trusted
realm shows on behalf of its users; a forum's "# Claim your airdrop" post
must not title the forum. `r/gov/dao` opens on `# GovDAO` then
`## Members`, so it keeps its title and gets the site sentence instead of an
`Author: g1...` line. Format characters (bidi overrides, zero-width joiners)
are dropped from lifted text.

| | title | description | card |
|---|---|---|---|
| official, no args or query | the h1, or the path title | the paragraph, or the site sentence | `og-gnoland.png` |
| official, with args or query | the path title | the site sentence | `og-gnoland.png` |
| site | "Realms", "Packages", "Users", or the path | the site sentence | `og-gnoland.png` |
| community | the path title | built from the path | `og-community-{realm,package,user}.png` |

The path title, `pathTitle`, reads what the path says and nothing a realm
rendered: `/r/nym/games/chess` is "games/chess · realm by nym", `/p/nym/lib`
"lib · package by nym", `/u/nym` "nym · user profile", `/r/nym` "realms by
nym". An address namespace is shortened as the user page shows it
(`g1jg...sqf5`), an alias names itself without its slash, and the title is
capped like an h1. The domain follows the title, unless the title already
names it, because a browser tab and a search result both truncate the tail.

The path is the only part of the URL a title repeats. Arguments and query
are typed by whoever wrote the link, on any realm, trusted ones included:
`/r/gnoland/blog:Official_GNOT_airdrop_claim_at_evil.example` used to title
itself with that sentence. Both also reach `Render`, and trusted realms echo
them: `p/gnoland/blog` writes a tag into its tag page's heading
(`# Gno.land's blog / t / <tag>`), and `r/gov/dao` prints the parse error of
a proposal ID. So an official page lends its h1 and paragraph only when its
link carries neither args nor query. A page an operator aliases keeps its h1,
since the alias URL has no args and the target's are the operator's.

A community summary is built from the path, the one part its author does not
write freely: "games/chess, a realm deployed on gno.land by nym.", "lib, a
package deployed on gno.land by nym.", "nym's profile on gno.land." It names
who published the page rather than who did not review it: a stamp of what
gno.land has not reviewed would read as a gatekeeper on a permissionless
network. The copy uses the configured domain.

The community cards lead with "Community realm", "Community package" or
"Community profile", a line on what the page is, gno.land's mark small in a
corner as the platform rather than the author, and a footer such as
"Deployed by its author · check before you sign". They are drawn by
`frontend/og-cards/generate.sh`, which holds their wording (`make og-cards`).
Cards are per kind for now, not per package: a card naming the package ("app
· realm by nym") is tracked in #6266. The head also carries `og:site_name`,
the image size and alt text, and a versioned `og:image`, so a redrawn card
reaches preview caches. An empty card reads as a broken site, so every 200
page carries a title, a description and, when `-canonical-origin` is set, an
image.

### Canonical URLs

A canonical URL needs a public origin, and a deployment that names the wrong
one tells a crawler its content belongs elsewhere. `-canonical-origin` is
empty by default: no origin declared, no canonical, og:url or image. A value
that is not a bare `scheme://host[:port]` stops gnoweb at startup.

`canonicalURL` prefixes that origin to the page's path and arguments, the
encoder gnoweb's own links use, and ignores `X-Forwarded-Host`. The query is
never kept: anyone can append one to any page. Of the `$` keys, an indexed
page carries only `source` and `file`, written in the order gnoweb's own
links use (`$source&file=<name>`). An alias's target asked for by its own
path, with no view or query, names the alias: `/r/gnoland/home` is `/`. Of
two aliases to one target, the shorter wins. Args reach only a realm's
`Render`, so a source view, a file and a pure package drop them, and a pure
package drops its trailing slash: `/r/gnoland/blog:p/a$source` names
`/r/gnoland/blog$source`, and `/p/gnoland/lib/` and `/p/gnoland/lib:x` name
`/p/gnoland/lib`. A static page names its alias key even when the key reads
as a file, as `/license.md` or `/Terms` do.

### What search engines may index

`pagePolicy.robots` in `community_index.go` decides once, in `classifyPage`,
before any branch writes. A page that may not be indexed gets the robots meta
and an `X-Robots-Tag` header (the only signal a text/markdown response can
carry), and no canonical, since a canonical on a noindex page contradicts
it. It keeps its card: link previews ignore robots.

| page | robots |
|---|---|
| official: bare, with args, `$source`, `$source&file=` | index, follow |
| official: any query, `$help`, `$state`, any other `$` key | noindex, follow |
| community, by `-index-community` | see below |
| error status, user page with nothing to show | noindex, nofollow |

A query or a `$` view lets any link multiply one page into as many URLs as
it cares to write, and a query reaches `Render`; an official page keeps its
links followed there. Args stay indexed, since they address a realm's pages,
which leaves a soft 404 (a 200 page saying "404") indexable: see
Consequences. `finalRobots` lowers the decision once the status is known: an
error, an empty user page, and a state explorer page that sent its own
noindex all become "noindex, nofollow", so the header and the meta agree.

For community pages, `-index-community` chooses:

| `-index-community` | community pages indexed |
|---|---|
| `registered` (default) | the bare page of a package or user page whose namespace is a registered name: no args, no query, no `$` view, no file, no directory listing |
| `none` | none |
| `all` | all of them, under the rule of official pages |

The `registered` line rests on the chain. `checkNamespacePermission`
(`gno.land/pkg/sdk/vm/keeper.go`) asks
`r/sys/names.IsAuthorizedAddressForNamespace` before every deploy, and once
that realm is enabled it authorizes `r/<ns>/` only for the address `<ns>`
itself or for the current holder of the name `<ns>` in `r/sys/users`
(`examples/gno.land/r/sys/names/verifier.gno`;
`gno.land/pkg/integration/testdata/addpkg_namespace.txtar`). Mainnet enables
it at genesis (`misc/deployments/mainnet.gno.land/README.md`), betanet
through its first GovDAO proposal. An address namespace therefore costs
nothing to create and throw away, while a name is held by one account.
gnoweb tells the two apart by the path alone, with the bech32 check the user
page already uses, and makes no RPC. A chain that never enables
`r/sys/names` lets anyone deploy under any name: it should run
`-index-community=none`, and its `-trusted-paths` mean nothing either.

### Links pass on authority only where gno.land answers for them

A followed link tells a search engine the linking site vouches for its
target. Each document renders with a `markdown.LinkPolicy`, carried by the
render context:

| document | internal link | external link |
|---|---|---|
| community realm, README, user page | `nofollow ugc` | `noopener nofollow ugc` |
| trusted realm or README | followed | `noopener nofollow ugc` |
| operator markdown (`-aliases` file) | followed | `noopener` |
| doc comment (`$help`, source overview) | `noopener nofollow ugc` | `noopener nofollow ugc` |

A trusted realm keeps its external links unfollowed because it may show what
its users wrote: a board, a proposal, a profile. The policy's zero value
follows nothing. Links from a `<gno-foreign>` sandbox stay unfollowed
whatever the policy. Doc comments render through a separate renderer shared
by pages of every kind, so their links are always treated as user content.
gnoweb's own header, breadcrumb and tab links are not documents and keep no
`rel`. The link transformer collects links before replacing them: replacing
during the walk skipped every link after an autolink in the same paragraph,
which then rendered with no `rel` at all.

### Error pages

An error status gets "noindex, nofollow" in the meta and the header, a fixed
title by status ("Page not found", "Invalid path") and a fixed description,
and no canonical, og:url or card. The head is built from the URL, and a
mistyped or crafted one must not publish itself or its words.

### Answering #3910

#3910 and its thread name the risks of user content under gno.land's name:
SEO bombing, dilution of the site's authority, scams and phishing, and
harmful content affecting gno.land's reputation. It asks that search engines
can tell gno.land's pages from user realms, and that risk come before SEO.

| concern | this change | what remains, and where |
|---|---|---|
| SEO bombing under gno.land's name | community titles and summaries come from the path, cards say the page is community content; community args, query and `$` views are noindex; every link in a community document is `nofollow ugc` | the bare page of a registered-name realm stays indexable, with that fixed head, under `registered` |
| Unbounded URLs from one page | official queries and `$` views other than source are noindex, canonicals drop the query and other keys, alias targets name their alias | args on official pages stay indexed, soft 404s included |
| Throwaway scam realms | address namespaces, free to create and discard, are noindex; empty user pages too | the page itself still renders |
| Scams under a registered name | indexable, but with a path title, a path summary, a community card and the #6191 notice on the page | content filtering in #5185; `-index-community=none` if that is not enough |
| Dilution of the site's authority through links | every link in a community document or a doc comment, and every external link in a trusted one, is `nofollow ugc` | none for rendered documents |
| Scams and phishing through metadata | the head repeats a document only on official pages, only its lead, never with args or a query; user pages are one name; format characters are dropped | a trusted realm whose own lead shows user text would lend it; the trusted list is the control |
| Phishing through page content | out of scope for metadata | the community notice of #6191; content filtering in #5185 |
| Crawlers telling main pages from user realms | robots meta and `X-Robots-Tag`, `ugc` links, distinct cards and summaries, all from one classification usable outside a request | the sitemap of #6256 should use `pagePolicy` and never list a noindex page; no structured data |
| Front matter as authored metadata (#3797, reverted by #3924) | front matter is read only from files an operator passes to `-aliases` | none |
| Who counts as official | `-trusted-paths`, default `DefaultTrustedPaths` | governance of that list |

### Reversibility

`-index-community` changes the community policy without a code change.
Search engines apply a noindex, or its removal, when they next crawl the
page: days to weeks, with no penalty for the change itself. A URL meant to
carry noindex must stay crawlable: blocking it in robots.txt hides the
noindex and can leave the URL listed from links alone, which matters for the
robots.txt of #6256.

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
as its host; a card that marks the page as community content answers #3910
better than no card.

**Keep the arguments in the title.** Each blog post had its own title that
way, even without a heading. Any link could also set the title of a trusted
page to a sentence of its choosing, which is the exact abuse #3910 names.

**Take the h1 of a page with args unless it repeats them.** A string match
cannot tell an echo from a heading: a post's slug is made of its title's
words, so a check strict enough to drop `/ t / <tag>` drops every post title
too, and one loose enough to keep post titles lets a reformatted echo
through. A realm that wants its pages titled can be aliased, or declare
titles itself in a later change.

**Take the first h1 or paragraph anywhere in the document.** More pages would
get a title and a summary. A trusted realm that lists user posts, comments
or proposals would then be titled and summarised by whichever came first.

**Treat every alias as official.** An operator chose it, but an alias to a
realm renders what the realm's deployer wrote, and the notice already says
so on the page.

**Index every `$` view of an official page.** `$help` and `$state` are views
of the same package, and an unknown key renders the page anyway; indexing
them would publish one package under unbounded URLs. `$source` and its files
are linked by gnoweb itself, so they stay.

**Give doc comments the page's link policy.** It would keep internal links
of trusted doc comments followed, but the doc renderer is shared through a
components interface with no page context; marking every doc link as user
content is the contained choice.

## Consequences

Every page now carries a title, a description and, with an origin set, a
share image. Every indexed page also names a canonical URL; no other page
does.

Community realms beyond the bare pages under registered names leave search
results under gno.land, and with `none` all of them do. That costs the
ecosystem the discoverability #5769 asks for; the trusted list, and the
bare-page rule, are the way back in.

An official page that opens on an h1 and has no args or query is titled by
it. Every other page is titled by its path title: the posts of one realm,
trusted or not, share their realm's title, as do `$source` and the content
page of one realm, while their canonicals differ. A realm that answers 200
for a missing post ("404" as content) stays indexable under its args.

`Renderer.RenderRealm` takes a link policy in its render context, and
`HeadData.NoIndex` became `Robots`, the meta's content as decided.

Operators must set `-canonical-origin` to get a canonical, og:url and share
image, and gno.land's own deployment is one of them. `-canonical-origin`
includes the scheme, so a deployment served over plain HTTP declares
`http://` there.

`robots.txt` and `sitemap.xml` remain absent here; #6256 adds them.
