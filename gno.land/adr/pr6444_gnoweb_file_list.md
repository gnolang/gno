# ADR: One file list for a package, and its trailing-slash address

## Status

Proposed. Builds on #6435 (package cards in listings).

## Context

gnoweb showed a package's files in two places, with two templates:

- the overview (`$source`), in its Source Files section: one row per file with
  `readme`, `license` and `test` tags;
- the directory page (`ui/dir_list`): a bare list of names, served at the
  trailing-slash address (`/r/gnoland/blog/`), the Browse button of a listing
  card and, for a pure package or a realm without Render, the bare address
  under the Content tab.

On a realm with Render, the directory page sat under the Content tab although
it is not the rendered page, and it repeated the overview with less in it.
Neither list said what a file holds: a reader opened `admin.gno` to learn that
it is where the realm's eleven admin calls live.

The overview already fetches `vm/qdoc`, whose functions, types and values each
name the file that declares them (`file`), and whose functions say whether they
cross into the realm (`crossing`).

## Decision

### One template, `ui/file_list`, for every list of a package's files

The overview's Source Files section and the directory page render the same
template from the same data (`[]FileLink`, built by `buildFileLinks`). Each row
shows:

| Column | Source |
|---|---|
| Name, linked to its source | file list |
| What the file declares: its first two exported top-level names, then `+N` (`Post, AddPost() +1`); for a file of methods only, its first two exported methods (`Uint.Cmp(), Uint.Eq() +10`); else `no exports`. The full list, capped at 12 names, in the tooltip | `vm/qdoc` types, functions (shown with `()`), methods, values, by `file` |
| What a non-code file is for: `Documentation` (README.md), `Module path and Gno version` (gnomod.toml) | file name |
| `N actions`: exported top-level crossing functions, each a call that sends a transaction | `vm/qdoc` `crossing` |
| Tags `render` and `state` (a realm file declaring Render, or package variables; a pure package has neither), `readme`, `license`, `config`, `test` | `vm/qdoc`, file name |

`N actions` counts what `ui/pkg_cards` counts for a card, so a package's
cards and its files agree: one predicate, `components.IsAction`, decides it
for both. It is not the Actions tab's count, which lists every
exported function, crossing or not; the tooltip says what is counted.

`ui/dir_list` goes; `.b-list` stays, for the search results page. The list is
`b-pkg-list` with a `b-pkg-list--files` modifier for the description column.
The name is what a reader scans for, so it is never cut. From `--lg` the rows
share one grid (`subgrid`), whose name column is as wide as the longest name,
up to half the row, so descriptions line up; a longer name wraps, and a row
without tags lends their column to its description. Below `--lg`
the description column is hidden, and the tags wrap under a name they would
squeeze. A width computed from the names server-side would need an inline
`style`, which gnoweb's `style-src` does not allow. The Source Files heading
is no longer a link to the directory page.

### The trailing slash of a realm with Render opens the overview's files

A realm with Render already has its Content tab. Its trailing-slash address
now answers `302 Found` to `$source#files`, keeping only the height (the
realm's query is its Render's). A pure package and a realm without Render keep
the directory page, at both addresses, as their Content tab: it is the only
page they can show there, so the header's tabs are unchanged.

The redirect needs to know whether the realm declares Render. The directory
view now reads `vm/qdoc` anyway, for the rows, once the file list says a
package lives there (a listing reads none), so it decides from it with no
extra query, by signature: `components.RenderShape`
accepts `Render(string) string` and `Render(realm, string) string`, and the
`$run` starter script now uses the same check. A view asks for a redirect
through `pageRender.redirect`, still returning a `RedirectView` so no caller
sees a nil view; `Get` answers it before rendering, through `serveRedirect`,
which the redirect middleware, `/faucet` and `$play` now share. The header's
counterpart lookup is not waited for on a redirect.

`302`, not `301`: what the address shows depends on chain state (no package
yet, then a listing), so a client must not remember it. When `vm/qdoc` cannot
be read, whether the realm declares Render is unknown: the address then serves
the directory page with the doc-unavailable notice instead of redirecting.

### Landing on `#files` shows the section, and says so

The redirect makes an anchor landing the common way into the overview, which
exposed two defects the section tabs already had. The heading landed under the
sticky header and section bar. The overview's one anchor rule,
`.c-overview-view [id]`, now covers both: the header alone below `--md`, and
from `--md` the header and the bar, whose seven tabs at most wrap to three rows
up to `--xl` and to two from it (measured with four-digit counts). And
the section bar marked the section above, still in its top band: the
`observer` controller now marks the fragment's link when the page opens on
one, as a click does.

### Cards drop Browse and say Actions

A listing card's Browse led to the directory page: for a realm it is now the
overview's files, which Source already opens, and for a pure package it is the
page its name opens. Browse is removed; Action becomes Actions, as the tab,
and the buttons' group is labelled Pages so it no longer shares that name.
`folder-closed`, which #6435 moved into the chrome sprite for Browse, goes
back to the markdown icon set it came from.

### The overview's Directories are listing cards

The packages one level below a package were names in a list. They are now the
listing's cards (`ui/pkg_cards`, via `chainmap.Cards`), read beside the
overview's other queries. They sit at the bottom of the page, so the overview
waits for them at most `subpackageCardsWait` (500 ms), below a listing's 1.5 s:
a card not ready by then shows its path and links, and fills in for a later
reader. A card under a section heading takes an `h3` title (`Nested`).
Rows would be denser, but would drop what a card says about each package, its
description and activity, which is what the section gains; each card keeps
its full path, as everywhere else a package is named.

### A failed `vm/qdoc` read says so once

The overview used to replace a failed doc read with an empty doc: every row
would then read `no exports`, and the page "No documentation or exported
symbols". Both views now keep a nil doc: rows claim nothing about
declarations, and one notice (`ui/doc_unavailable`) says the documentation
could not be loaded.

## Alternatives considered

- **Delete the directory page everywhere** and send every Content tab of a
  package without Render to `$source#files`. Content and Source would then
  open the same page on a pure package: one real tab in two. Rejected for this
  change; giving a pure package's Content tab a role of its own (usage
  examples, importers) is a separate proposal.
- **Keep both templates** and enrich only the overview. Two lists of the same
  files that disagree on what they show; rejected.
- **Per-file size and line count.** The node loads every file body for
  `vm/qfile` and returns names only. Worth a separate node change; out of
  scope here.
- **Detect Render with a `Realm` call** instead of `vm/qdoc`. One more query
  per directory page, for the same answer; rejected.

## Consequences

- A directory page costs one `vm/qdoc` read more than before, after the file
  list and beside the README, which it no longer asks for when the file list
  names none. A listing costs what it did. The overview costs its
  subpackages' card reads: at most 500 ms of wait, within `chainmap`'s
  process-wide fetch cap.
- `/r/<realm>/` links keep working: a realm with Render redirects, one without
  serves its Content tab.
- The rows depend on `vm/qdoc` reporting `file` for each declaration; a
  declaration without it is not attributed to any row.
