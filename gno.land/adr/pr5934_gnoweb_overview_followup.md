# ADR: Follow-up fixes for the gnoweb package overview page

## Context

The package overview page landed in #5572. Using it surfaced defects that the
review of that PR could not see from the diff, collected in
https://github.com/gnolang/gno/pull/5572#issuecomment-4934447615. They fall
into four groups.

**License detection reported the wrong license.** `deriveLicense` matched a
license name anywhere in the first 4 KB of the file. Real license texts cite
each other: MPL-2.0 enumerates the whole GNU family under "Secondary License",
and LGPL quotes the GPL. Whichever cited name the ordered signature table
reached first won, so a body-wide match reported the citation instead of the
license. Real files also hard-wrap, so a pattern spanning two words missed
whenever upstream broke the line between them.

**Clicking a symbol never revealed its declaration.** Chroma emits its line
highlight and its `scroll-margin-block-start` as class names at render time, so
they appear in no template and no Go source. The PurgeCSS extractor never saw
them and stripped both, which dropped the highlight and left a symbol permalink
landing under the sticky header.

**Controls did not explain themselves.** The top navigation tabs and the
overview sidebar's figures and status tags carried no hover text, so the
difference between Content, State, Source and Actions had to be discovered by
clicking.

**Navigation lost the reader's place.** The Source tab dropped the open file
and sent a reader back to the package overview. The overview's Files sidebar
entry was a single anchor with no file list under it. A symbol filter query
emptied a type card of its methods, because the method cards nest inside the
type card the filter hides.

## Decision

**Scope license detection to the title block, and order by precedence.** A
license names itself in its title block and cites others in its body, so every
signature that matches a name matches the first four lines that carry text
rather than the whole sample. BSD and Unlicense are recognised by the shape of
their text instead of by a name, and that wording only appears below any title,
so they stay body-wide. Multi-word patterns match their whitespace with `\s+`
and use the `s` flag so `.` crosses a line break. Markdown decoration is
dropped from the title block: a LICENSE copied off GitHub opens with a markdown
`# MIT License` heading, and a setext underline would otherwise spend one of the
four lines.

Detection runs title SPDX, then signatures, then body SPDX. A cited identifier
therefore loses to a title that names its own license, while a file whose title
block is a preamble or a file name keeps the identifier it declares further
down.

**Safelist chroma's render-time classes** with `greedy: [/chroma-/]` in
`postcss.config.cjs`, since no extractor can find a class name that exists only
at render time.

**Carry the open file onto the Source tab**, add `Tooltip` to `HeaderLink`, and
give the overview's figures and status tags title text taken from the
`gnomod.File` doc comments rather than from paraphrase.

**Give the Files sidebar entry its file list**, through a new `Link` field on
`TocItem` that points off the current page instead of anchoring to an ID. The
README entry now keys off whether the section rendered, not off whether the
file was listed, because a listed but unfetchable README would anchor at a
section the template never emitted. The source view's own back-to-overview link
is dropped, the sidebar's package name already going there.

**Fold method names into the type card's `data-name`** so the filter still
matches a method whose card is nested inside a hidden type card.

**Link stdlib imports upstream.** Stdlibs ship with the node instead of being
deployed, so they have no package page. The link goes to `gnovm/stdlibs/` on
GitHub and carries what `markdown/ext_links.go` gives an external link: `rel`,
an external icon and a tooltip. An `External bool` on `ImportLink` drives all
three, set where `buildImportLink` already decides the link is off-site.

No `data-outbound`. That attribute tags the header and footer, which are fixed
curated link tables, and `markdown/ext_links.go` emits it for no link at all.
Import rows are content rather than site chrome, so they get the treatment
content gets, and SimpleAnalytics still counts the clicks through its generic
outbound auto-event.

## Alternatives considered

- **Keep license detection body-wide and reorder the table instead.** Rejected:
  ordering can only ever protect the citations already seen. Any license whose
  body quotes another is a new false positive, and title scoping removes the
  class rather than one instance of it.
- **Scope the SPDX line to the title block too**, which is what the review of
  this PR asked for. Rejected on measurement: over 2034 LICENSE files in a
  module cache it cost two correct verdicts and gained none, because a file
  whose title block is a preamble declares its identifier below it. Precedence
  keeps the review's intent without the loss.
- **Add every chroma class to the PurgeCSS `safelist.standard` list.**
  Rejected: the set is chroma's to change, so a greedy pattern on its prefix is
  the only form that does not rot.
- **Give the overview page its own filter that walks into hidden cards.**
  Rejected: folding the method names into the parent's `data-name` keeps one
  filter implementation for every list on the page.

## Consequences

- License detection changes 664 verdicts over 2034 LICENSE files in a module
  cache. 640 were blank on master and now resolve, most of them because master's
  patterns could not cross the line break a real file wraps at. 13 move from
  BSD-2-Clause to BSD-3-Clause, master having required a literal `3.` label
  where real files bullet the clause, and 3 move from Apache-2.0 to
  BSD-3-Clause, those files opening with the Go Authors' BSD text and carrying
  Apache further down. 8 stop resolving, every one a file covering itself with
  two licenses at once, `gopkg.in/yaml.v3` and `pelletier/go-toml` among them,
  where master named whichever the table reached first. Reporting no license on
  a dual-licensed file is the intended trade: naming one of two is worse than
  naming none.
- The four license files in `examples/` go from three resolved to four, the
  fourth being `p/onbloc/json/LICENSE` and its `# MIT License` heading.
- The chroma safelist adds 559 bytes to `public/main.css`. That file is a
  checked-in build artifact, so any CSS change has to rebuild it or ship
  nothing.
- `buildOverviewTOC` takes two more parameters. It is unexported and has one
  caller.
- `TocItem.Link` is empty for every existing caller, so realm, action and
  source tables of contents anchor exactly as before.
- Removing `SourceTocData.OverviewLink` drops the source view's back link. The
  package name in the sidebar is the remaining route to the overview.
