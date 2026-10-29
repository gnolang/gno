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
  what a ustar header holds) for Add file, reopening a share, opening a
  copy and Download. The chain takes a few names this rejects (a leading
  dash, more than 100 bytes): a copy leaves those files out and says so;
- opening an editable copy of a deployed package with `?from=<pkgpath>`
  (`&file=` picks the tab shown first). `<pkgpath>$play` redirects there,
  and source file views link to it.

A copy never goes through the text format, which a source line could split:
the page carries its files as JSON in a `<script type="application/json">`,
one entry per file, sources only. They are read by the fetch `$download`
uses, handed to the playground as a `PackageFetcher` (as state is handed a
`FileFetcher`), so both share one bound (100 files, 2 MiB of source), one
60 s cache and one coalesced read per package. The read outlives the
request that started it, for the requests after, but a request waits for it
no longer than its own deadline. A missing package is a 404, a node
timeout or a deadline passed a 408 (its text in the body, as a page shows
it), a package over the bounds a 413. `?from=` is read once, by
`playground.ParseFrom`, for the copy, the header and the breadcrumb: only a
bare realm or pure package path (with or without the domain); a file, args
or a query is a 400 and is never fetched.

The playground is its own page, titled "Playground", or
"r/demo/foo · Playground" for a copy, which share cards repeat. Its
breadcrumb ends with a `playground` pill, after the copied package (or
`_ / play` and the shared `?code=`, whose `&z` goes along as a hidden input
when the field is submitted). A copy keeps its package's tabs, none of them
current, so the reader can go back to it, and a copy of a community package
carries the realm notice that package's pages carry.

The playground does not execute scratch-pad code. Its button says what it
does, from the server's first render on. On an unmodified copy of a realm
that declares `Render`, it reads "Preview" and shows what that realm
renders, `Render("")` through `/_/api/eval` at the copy's own path.
Anywhere else (edited code, a package, a share, the default code), it reads
"Run locally" and prints `gno run .` and `gno test .`, with no file name in
any command. A Run result or a Share confirmation that comes back after the
output moved on is dropped, and a newer Run aborts the request it
supersedes. Share prints the link when the Clipboard API is missing.
Analytics counts the page as `playground`, and `$run` as `run`.

The page is served with its code as text, each file of a copy under its
name, in the source view's frame, so a reader without JS still reads it; the
controller replaces it with the editor.

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

The Eval buttons and the evaluator's form and result need JS: they are served
hidden and shown by their controllers, and a `<noscript>` line says the
evaluator needs JavaScript.

The live results stay server-side rather than going from the browser to the
node, as they did before: in production the CSP's `connect-src` named the node
gnoweb queries (`-remote`), while the page queried the help remote, so the
direct query was blocked there.

### 3. `$run`: run script

An editor pre-filled with a `main` that imports the package, and the
generated `gnokey maketx run` command, optionally with the script inlined
through a heredoc. gnoweb never signs or broadcasts.

- **Where:** `$run` serves the page only for a realm or pure package the
  node's `vm/qdoc` answers for, on its path or on one of its files that
  the package lists (the Run tab of a source file links there). A listing,
  a user page, a missing file or a path with no package renders as if
  `$run` were not there, with that page's tabs; a node failure gets the
  error page `$help` gives. Error pages have no Run tab: the header is
  built knowing the page is an error. The Run tab stays a realm tab; a
  pure package's `$run` is reached by URL.
- **Starter script:** built on the server from the same qdoc, so it runs as
  generated: `println(pkg.Render(""))` when the package declares
  `Render(string) string`, `Render(cross(cur), "")` from a
  `main(cur realm)` when that `Render` is crossing, and otherwise
  `_ = pkg.F` for the first exported function, since an unused import does
  not compile. A package with no exported function gets a blank import. The
  name is the import's own (`gno.LastPathElement`, which skips a version
  suffix).
- **Served whole:** the script is highlighted in the editor's place and the
  command is in the page, so the page is complete without JS. The
  controller replaces the script with the editor and fills the command's
  slots from the fields. Reset, Download and the "Include script" switch
  do nothing without JS, so they are served hidden and shown by the
  controller, as the Eval buttons are on `$help`. Copy stays, as on every
  code block.
- **Modes and key:** the command follows `$help`: Full Security (query the
  account, `maketx run -broadcast=false > run.tx`, `sign`, `broadcast`) is
  shown and copied first, Fast is one `maketx run` under `$help`'s warning.
  The mode select and the Key / Address field are `$help`'s, under the same
  `action-header` controller and stored values, so a choice on one page
  carries to the other. A line above the command says running it signs with
  the key and broadcasts to the chain ID, which charges its fee. Empty, the
  key reads `KEYNAME` (a shell reads `<key-name>` as a redirect) and the
  account query `ADDRESS`.
- **Shell safety:** each typed value (key, gas, fee, coins) is single-quoted
  when it holds a character a shell treats specially, so it stays one
  literal argument; `escapeShellSpecialChars` leaves spaces, so it would not.
  The heredoc's delimiter (`GNO_EOF`, then `GNO_EOF_1`, ...) is one no line
  of the script equals, so the script cannot end the heredoc and run lines as
  commands. Both helpers are pinned by the frontend tests, which run the
  quoted values and the heredoc through `sh`.
- **Gas:** 50,000,000 gas with a 100000ugnot fee. The example realms'
  starter scripts take 1 to 14 million gas (one realm's `Render` alone takes
  far more and needs a raised gas), and the fee pays for 100 million gas at
  the chain's initial price (1ugnot per 1000 gas), since the whole fee is
  charged. `$help`'s calls keep their 1,000,000,000 and 1000000ugnot: they
  call any function. Both pairs are Go values in `components` (`CallGas`,
  `RunGas`), read by `command.html`, the run page's inputs and, for an
  emptied field, the controller (the input's `defaultValue`).
- **"Include script"** is a `.b-switch`, whose checkbox stays focusable on
  this page; left out, the command asks for `script.gno` in a comment.

The playground and the run page are built unconditionally: `NewRouter`
always sets the help remote and the chain ID their commands print.

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
