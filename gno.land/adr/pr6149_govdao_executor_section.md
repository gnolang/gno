# ADR: Give the govDAO executor its own section, and link its source

## Context

`r/gov/dao/v3/impl` renders a proposal page as a title, an author line, then
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

Two sections on the proposal page. `### Execution` holds everything the
executor supplies, the creation realm first and the executor's own description
under it. `### Description` holds the proposer's prose. A proposal carrying no
executor renders no Execution section, as before.

Execution comes first, above the description. Nothing the proposer controls
precedes it: the title is escaped and folded to one line, and the author line is
the page's. A description imitating the disclosure therefore renders below the
real one, and a voter reading from the top meets the DAO's statement first.

A horizontal rule separates the two sections, and only the page may draw one:
every line of the description that would render as a thematic break is escaped,
so it prints its own dashes as text. Nothing else in the description is touched,
which is what keeps the portfolio headings and the token lists that proposals
already carry.

That makes the rule a mark of the page's own chrome rather than a decoration.
It does not make the section unforgeable: a description can still write a line
that reads like the disclosure, and a heading above it.

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
- **Verify that the named realm is the one whose code runs.** Out of scope, and
  not possible from the page: the proposal holds an `Executor` interface value,
  and the realm behind it is only recoverable for the built-in implementation.

## Consequences

- Every proposal page and every golden covering one changes shape. Filetests in
  `r/gov/dao/v3/impl` and `r/sys/namereg/v1` move with it.
- A hostile creation realm renders exactly as it did: escaped, on one line,
  inside a code span, with no link.
- A creation realm the check rejects renders with no link rather than with a
  link to a page that does not exist. `/e/` run realms fall in that set: a
  proposal can be created from one, and gnoweb serves no source for it.
- The link is reachability, not trust. It does not assert the named realm is the
  one whose code runs, which the page has never asserted either.
- A proposal can still write a second `### Execution` section of its own. It
  renders below the real one, without the rule above it, and it cannot be moved
  higher. Ending the imitation outright means the page stops rendering proposer
  markdown, which is a decision about the proposal format rather than about
  this page.
