# ADR: Integrated Playground, Expression Evaluator and Run Script in gnoweb

## Context

gno.land had no interactive code evaluation in gnoweb. Users who wanted to
call realm functions or experiment with Gno code had to use external tools
(gnokey CLI, the separate React + WASM play.gno.land, or copy-paste workflows).

The key problems:

1. **Friction for exploration.** Reading a realm's `Render` output or calling
   a view function required a local environment or CLI tools.
2. **No in-browser eval.** Users could not test expressions against a
   deployed package from the web UI.
3. **No code scratch pad.** There was nowhere in gnoweb to write and share
   short Gno snippets.
4. **Separate studio dependency.** The only interactive option was a separate
   application, breaking the single-binary model.

## Decision

Add interactive features to gnoweb as Go-native, single-binary extensions,
built from gnoweb's own UI vocabulary (CUBE CSS tokens, blocks and
components) rather than a parallel design.

### 1. `/_/play`: playground

A multi-file editor backed by the shared `@gnoweb/js/code-editor` component
(CodeMirror, Go/TOML modes). Supports:

- sharing through `?code=...&z`: the base64 of raw DEFLATE. Without `&z`,
  `code` is plain text; with it, a value that does not inflate shows the
  default code. Share refuses links over 8,000 characters, so the server
  reads no longer value and inflates to 64 KiB at most;
- multi-file code, one tab per file, with add and close. A share writes
  every file, a lone one too, under a `// --- name ---` header. Reading it
  back, a header whose name is not a file name (or repeats one) is a comment
  that only looks like a header and stays in its file, and text before the
  first header is `main.gno`;
- file names from a strict allowlist (letters, digits, `_`, `.`, `-`; a
  `.gno` name or `gnomod.toml`; no leading dot or dash; 100 bytes at most,
  what a ustar header holds) for Add file, reopening a share and Download;
- opening an editable copy of a deployed package with `?from=<pkgpath>`
  (`&file=` picks the tab shown first). `<pkgpath>$play` redirects there,
  and source file views link to it.

A copy never goes through the text format, which a source line could split:
the page carries its files as JSON in a `<script type="application/json">`,
one entry per file, sources only. They are read by the fetch `$download`
uses, handed to the playground as a `PackageFetcher` (as state is handed a
`FileFetcher`), so both share one bound (100 files, 2 MiB of source), one
60 s cache and one coalesced read per package. A missing package is a 404, a
node timeout a 408, a package over the bounds a 413. `?from=` is read once,
by `playground.ParseFrom`, for the copy, the header and the breadcrumb: only
a bare realm or pure package path (with or without the domain); a file,
args or a query is a 400 and is never fetched.

The playground is its own page, titled "Playground", or
"r/demo/foo · Playground" for a copy, which share cards repeat. Its
breadcrumb ends with a `playground` pill, after the copied package (or
`_ / play` and the shared `?code=`, whose `&z` goes along as a hidden input
when the field is submitted). A copy keeps its package's tabs, none of them
current, so the reader can go back to it, and a copy of a community package
carries the realm notice that package's pages carry.

The playground does not execute scratch-pad code. Its button says what it
does. On an unmodified copy of a realm that declares `Render`, it reads
"Preview" and shows what that realm renders, `Render("")` through
`/_/api/eval` at the copy's own path. Anywhere else (edited code, a package,
a share, the default code), it reads "Run locally" and prints `gno run .`
and `gno test .`, with no file name in any command. A result that comes back
after the output moved on is dropped, and a newer Run aborts the request it
supersedes. Share prints the link when the Clipboard API is missing.
Analytics counts the page as `playground`, and `$run` as `run`.

### 2. Expression evaluator on the action page (`$help`)

Below the function list: an input for arbitrary read-only Gno expressions, a
result panel and a history with re-run (ArrowUp/ArrowDown walk it). The
Functions list ends with an entry that links to it. It and the function
cards' live results evaluate through `POST /_/api/eval` (`vm/qeval`), so
nothing mutates state.

A card's Eval button sends the card's params in order, one per input, so
unnamed and repeated (`_`) params count: the value typed, quoted when the
param is a `string` or an `address` and not already a quoted literal, or a
zero literal for the param's type (`""` for both). The template renders the
types from the package doc (`.Params`) rather than parsing the signature.
The card's live result and its printed `gnokey query vm/qeval -data` command
use the same literals, so all three agree on what a typed value means; the
`gnokey maketx call -args` command keeps the raw value, which gnokey parses
by type itself. These helpers are pure functions in `frontend/js/utils.ts`,
pinned by `make -C gno.land/pkg/gnoweb/frontend test` (node's own test
runner, no dependency), which CI runs beside the frontend lint; the
playground's file helpers (`playground-files.ts`) are pinned the same way. A
chain error reaches the reader in the node's own words, without gnoweb's
internal prefix. A newer live result
or evaluation aborts the request it supersedes. A rate limit, an unreachable
node or a network failure is shown as "unavailable", not as the function's
error, and analytics does not record it as a failed `qeval_preview`.

The live results stay server-side rather than going from the browser to the
node, as they did before: in production the CSP's `connect-src` named the node
gnoweb queries (`-remote`), while the page queried the help remote, so the
direct query was blocked there.

### 3. `$run`: run script

An editor pre-filled with a `main` that imports the realm, and the generated
`gnokey maketx run` command (optionally with the script inlined through a
heredoc). gnoweb never signs or broadcasts.

### API endpoint

`POST /_/api/eval` evaluates a `pkg_path` + `expression` pair through
`vm/qeval` and returns `{result}` or `{error}`.

- **Requests:** a body that is not `application/json` is refused with 415,
  and a request another site sent (`Sec-Fetch-Site`/`Origin`, net/http's
  `CrossOriginProtection`) with 403, both before the limiter is charged.
  A body over 1 MiB is a 413; another method a 405 with `Allow: POST`.
- **Rate limit:** `state.NewIPLimiter`, built with `-trusted-proxies` like
  the state and search limiters: 40 of burst and 60 per minute per client,
  so a load and a reload of a realm with a dozen zero-argument functions
  fit.
- **Node load:** answers are kept for 3 s in a bounded LRU, keyed on package
  and expression, and concurrent readers of the same expression share one
  query (singleflight), so a popular Actions page costs the node one query
  per function. Evals run on their own RPC slots, a quarter of
  `AppConfig.MaxConcurrentRPC` (8 of 32 by default), so they cannot hold
  the slots pages render with, and under the handler's request timeout.
- **Answers:** an expression the chain rejects is a 200 with `{error}`;
  an unreachable node is a 502 and a timeout a 504, both with a generic
  message. Results are cut at 64 KiB with a visible marker. A failed eval
  logs at Debug, without the expression.

### Editor

The editor keeps every file of the playground (text and undo history, one
CodeMirror state per file), so switching tabs never mixes undo histories.
Its tokens carry chroma's class names, so it is coloured by the same
stylesheet as the read-only source view in both themes, and it shares that
view's frame. Its own styles are a CodeMirror theme in TypeScript, because
PurgeCSS cannot see the `.cm-*` classes CodeMirror adds at runtime.

### Frontend approach

Vanilla TypeScript compiled to plain JS, one controller per feature through
the existing `data-controller` dispatch; shared helpers (`postEval`,
`downloadBlob`, icons) are re-exported from `controller.ts`. Compiled output
is committed to `public/js/`. Each icon keeps a single meaning across gnoweb
(the lightning is a transaction, `ico-run-code` runs the playground,
`ico-eval` evaluates, `ico-playground` opens the playground).

### CSP

`connect-src` is `'self'` only: the JS calls `/_/api/eval` and never the node.
CodeMirror's injected styles carry the per-response CSP nonce.

Assets loaded without the version query (`?v=`), such as a module's own
`./controller.js` import, are served `no-cache` and revalidated against the
ETag; only versioned URLs are `immutable`, so a deploy reaches returning
readers.

## Alternatives Considered

- **Server-side `gno run` / `gno test`:** needs sandboxing, resource limits
  and execution isolation. Deferred; the Run button prints commands instead.
- **Run previewing edited code:** the node can only evaluate deployed code,
  so an edit could only be previewed against what the chain holds, which
  would show the old code's output under the new code. Only an unmodified
  copy previews.
- **Copies in the share text format:** what the copy did at first. A source
  line shaped like a header split the file; JSON has no such line.
- **Dry Run through `.app/simulate`:** shipped on this branch, then removed.
  Since #6088 the node simulates a `MsgRun` only with a real signature, which
  gnoweb does not have. It can come back with wallet signing (sign without
  broadcasting).
- **`/_/api/funcs`:** removed, nothing called it.
- **Live results from the browser to the node (`<remote>/abci_query`):** what
  master did. Blocked by the production CSP (see the Actions page section);
  the endpoint is hardened instead.
- **WebSocket REPL:** more interactive, much more server state. Deferred.
- **Separate gnostudio service or WASM playground:** breaks the single-binary
  model.
- **Stimulus.js or another controller framework:** the existing dispatch is
  enough.

## Consequences

- **Positive:** read-only realm evaluation from the browser, without tooling.
- **Positive:** any package can be opened as an editable copy in one click.
- **Positive:** zero new runtime dependencies; gnoweb stays a single binary.
- **Trade-off:** the playground cannot execute scratch-pad code yet.
- **Trade-off:** `/_/api/eval` runs caller-chosen expressions on the node.
  The per-client limit, the cache and the separate RPC slots bound what one
  reader or one popular page costs; the edge proxy remains the first limit.
- **Trade-off:** compiled JS is committed to the repo, like the rest of the
  gnoweb frontend.

## Not Yet Implemented

- Server-side `gno run` / `gno test` execution with sandboxing
- Wallet integration (signing, dry run, publishing a package copy)
- WebSocket REPL
- gnodev hot-reload integration
