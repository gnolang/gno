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

Two sections on the proposal page, `### Description` for the proposer's prose
and `### Execution` for everything the executor supplies, the creation realm
first and the executor's own description under it. A proposal carrying neither
renders no section, as before.

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
- The proposer's description is markdown and is not sanitized, so a proposal can
  write a second `### Execution` section of its own above the real one. That
  predates the section: the same description could already forge the
  `Executor created in:` line. Sanitizing it would cost every existing proposal
  its formatting, and belongs to whoever takes that decision.
