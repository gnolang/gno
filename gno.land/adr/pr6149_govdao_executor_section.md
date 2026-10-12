# ADR: Give the govDAO executor its own section, and link its source

## Context

`r/gov/dao/impl/v0` renders a proposal page as a title, an author line, then
three blocks of prose with nothing between them: the proposer's description,
the executor's description, and the line `Executor created in: <path>`. A voter
reading it cannot tell which words come from the person asking for the vote and
which come from the code that runs if the vote passes.

The creation realm is the only thing on the page that names that code, and it
is not a link. Reading the code means copying the path out of the page and
finding its source view by hand.

`CreationRealm()` is dispatched through the public `dao.Executor` interface, so
only `SimpleExecutor` has the value captured by the machine from
`rlm.PkgPath()`; any other executor returns whatever string it likes. That is
why the value renders through `sanitize.InlineCode` today, and why turning it
into a link needs a rule about which values may become one.

## Decision

One metadata block under the title, then the executor's own words, then the
proposer's.

The block is five lines: the author, the realm that runs if the proposal passes,
the status with the denial reason where there is one, the tiers that may vote,
and the tally with a link to the vote list.

The realm line is always drawn, and reads `Executes: none reported` when there
is nothing to name. A proposal can carry no executor at all, and a
`SimpleExecutor` built outside a code realm captures an empty path, so dropping
the line would read as a page with one line fewer rather than as a disclosure
nobody made. It replaces the `### Stats` list at
the foot of the page, which is where a voter had to go to learn whether the vote
was still open. The tally is rounded to two places, since the page is a summary
and the vote list carries the detail.

`### Execution` holds the executor's own description of what it does.
`### Description` holds the proposer's prose. A proposal whose executor
describes nothing renders no Execution section. A built-in executor's
description carries text the proposer chose, a member's portfolio for one, so
both sections are clamped and quoted the same way.

Everything the page writes comes first, and nothing the proposer controls
precedes it: the title is escaped and folded to one line. A description
imitating the disclosure therefore renders below the real one, and a voter
reading from the top meets the DAO's statement first.

A horizontal rule separates the two sections, and the description renders
inside a quote the page draws: every line of it is prefixed with `> `, so any
block it opens, a fence, an HTML block, a list, ends where the quote ends and
nothing it writes renders outside it. A rule the description draws sits inside
the quote, while the page's own rule, the Description heading and the action
bar below it stay the page's. The quote is indented two spaces, so each line
starts at column 4: a tab keeps its width and tab-indented code stays code.

Two markers are escaped and nothing else, which keeps the portfolio headings
and the token lists that proposals already carry, now quoted. A footnote
definition is the one block goldmark lifts out of its container, to the foot of
the page below the action bar, and it needs `[^` with `]:` after it, so `[^` is
escaped wherever a `]:` follows on its line. A `[^` in code with no `]:` after
it keeps its bytes. A leading `[!` is escaped too, since gnoweb draws a quote
whose first line is `[!KIND]` as an alert box.

That makes the quote the boundary between the proposer's words and the page's.
It does not make the section unforgeable: a description can still write a line
that reads like the disclosure, and a heading above it, inside its quote.

The creation realm keeps the escaping it has: `sanitize.InlineCode` still
produces the code span, clamped first per `clamp.gno`. The span becomes the
text of a link to `<path>$source` when, and only when, the value parses as a
package path this chain serves:

- the chain domain, matched whole against `runtime.ChainDomain()`
- a letter directory of `r` or `p`
- one or more segments, each matching the grammar the VM enforces on a deployed
  path (`Re_name` in `gnovm/pkg/gnolang/mempackage.go`): a leading letter, an
  alphanumeric body, single `_` or `-` separators, never consecutive and never
  at either end

`StringifyProposal`, which prints a proposal as text for other realms, calls the
same section helper rather than carrying its own copy of the disclosure.

## Alternatives considered

- **Link whatever the executor returns.** Rejected: it hands a proposal a
  clickable destination on the govDAO's own page. The grammar check leaves
  nothing a URL or a markdown link can be steered with: `[a-z0-9]`, `/`, `.`,
  `-` and `_`.
- **Build the link with `md.Link`.** Rejected: it escapes its text with
  `sanitize.InlineText`, which turns the code span's backticks into visible
  ones. The link is written out instead, over a value the grammar check has
  already accepted.
- **Point at the realm's rendered page.** Rejected: the page exists to show what
  runs, and the render is the realm's own UI. `$source` with no file is the
  overview, which lists the files.
- **Put the grammar check in a shared package.** `r/sys/params` carries a looser
  copy (`assertDelegatePath`). Left where it is: one caller, and a package would
  fix the shape for callers that do not exist yet.
- **Escape the whole description with `sanitize.Block`.** Rejected, and
  measured before rejecting: it escapes the `####` portfolio heading a member
  proposal renders and the `-` token list a treasury proposal renders, printing
  both as literal text, and it still lets a forged `Executor created in:` line
  through, because inline links survive by design. The cost is real and it does
  not close the imitation.
- **Quote the description with `sanitize.BlockquoteRich`.** Rejected, and
  measured before rejecting: it keeps headings and lists, and it escapes every
  bracket outside a link, so `a[0]` in a code span renders as `a\[0\]`. It
  prefixes `> `, which leaves a tab-indented line two columns short of code, and
  it deletes a link reference definition along with the link that used it.
- **Escape every line of the description that would draw a rule.** Rejected,
  and measured before rejecting: an escaping pass has to reimplement
  CommonMark's block grammar to know where a rule renders, and each review
  round found shapes that got past the last one, among them a backtick in an
  info string, a fence of the other character, a carriage return inside a dash
  run, an HTML block left open and a rule indented under a list item. Each
  either drew the page's rule or swallowed its action bar. The quote holds on
  every one of them, through gnoweb's own goldmark options.
- **Verify that the named realm is the one whose code runs.** Out of scope, and
  not possible from the page: the proposal holds an `Executor` interface value,
  and the realm behind it is only recoverable for the built-in implementation.

## Consequences

- Every proposal page and every golden covering one changes shape. Filetests in
  `r/gov/dao/impl/v0` and `r/sys/namereg/v0` move with it, and so do the
  render assertions in `govdao_execute_reject_proposal.txtar` and
  `govdao_treasury_payment_reason_cannot_forge_label.txtar`, under
  `gno.land/pkg/integration/testdata/`.
- `proposalStatus.String` is gone, replaced by `statusLine` and `tally`. The
  page assembles its own block rather than taking a rendered one from the
  status type.
- Both descriptions render as a quote: indented, under the quote's left
  border, with their markdown otherwise as written. A footnote prints as the
  text the proposer typed rather than as a note at the foot of the page, and a
  code span carrying `]:` after a `[^` on its line shows a backslash.
- `StringifyProposal` escapes the title and clamps and quotes the description
  the way the page does.
- Rendering a description now rewrites every line into the quote, and the
  rewrite is what the page pays for: a 100KB description measured
  3,085,843,608 gas per render, against the 3,000,000,000 query cap. So both
  descriptions are clamped to `maxRenderedDescription` before the quote, like
  every other slot on the page, and the worst 8KB shape in both measured
  969,999,528; the stored proposal keeps every byte.
- Adding declarations to a genesis realm shifts the balances asserted in
  `gno.land/pkg/integration/testdata/storage_deposit_price_change.txtar`, which
  its own note says to re-derive.
- A hostile creation realm renders exactly as it did: escaped, on one line,
  inside a code span, with no link.
- A creation realm the check rejects renders with no link rather than with a
  link to a page that does not exist. `/e/` run realms fall in that set: a
  proposal can be created from one, and gnoweb serves no source for it.
- The link is reachability, not trust. It does not assert the named realm is the
  one whose code runs, which the page has never asserted either.
- A proposal can still write a second `### Execution` section of its own. It
  renders below the real one and inside the description's quote, and it cannot
  be moved higher or out of the quote. Ending the imitation outright means the
  page stops rendering proposer markdown, which is a decision about the
  proposal format rather than about this page.
