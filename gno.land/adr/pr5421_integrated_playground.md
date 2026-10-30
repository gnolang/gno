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

- sharing through `?code=...&z`: the base64 of raw DEFLATE of the JSON
  list of files, `[{"name","content"}]`, every file, a lone one too. The
  server decodes it into the same file list a copy opens (below), so each
  file comes back as it was written: no line of it is read as anything,
  and its blank lines, at either edge too, are kept. Without `&z`, `code`
  is plain text, always, one `main.gno`: a base64 value without `&z` is
  not decoded, and opens as the base64 text it is. One meaning per form
  keeps a snippet that happens to be valid base64 from being decoded into
  something else; no release served another form, so no link breaks. With
  `&z`, a value that does not decode to such a list, or names a file the
  allowlist below rejects (or one twice), shows the default code, under a
  notice saying why the link was not opened (too long, more than 64 KiB of
  code, not decodable, or a file name the playground does not take): the
  default code alone would read as the link's. `/_/play?z` with no code
  is the default code, with no notice. The server reads no value over
  8,000 characters and inflates to 64 KiB of JSON at most, and Share
  refuses, with a message, any link over either cap, so it never writes a
  link that opens as the default code; a Go test checks both sides hold
  the same caps. A file with CRLF line ends decodes as it is; the page
  serves it as the editor holds it, with LF ends;
- multi-file code, one tab per file, with add and close;
- file names from a strict allowlist (letters, digits, `_`, `.`, `-`; a
  `.gno` name or `gnomod.toml`; no leading dot or dash; 100 bytes at most,
  what a ustar header holds) for Add file, opening a share (checked by the
  server, which holds the same list), opening a copy and Download. The
  chain takes a few names this rejects (a leading dash, more than 100
  bytes): a copy leaves those files out and says so;
- opening an editable copy of a deployed package with `?from=<pkgpath>`
  (`&file=` picks the tab shown first). `<pkgpath>$play` redirects there,
  and source file views link to it.

Every file the editor opens, a copy's, a share's or the default code,
comes the same way: the page serves each file once, as text in a `<pre>`
named after it, which a reader without JS reads and the controller reads
back into the editor (each `<pre>` opens with a newline, which HTML drops,
so a file's own leading blank line is kept). A copy's files are read by
the fetch `$download` uses, handed to the playground as a `PackageFetcher`
(as state is handed a `FileFetcher`), so both share one file cap (100), one
60 s cache and one coalesced read per package and kind of read. A copy reads
only what it opens, the Gno sources and `gnomod.toml` (`InCopy`), so a README
or a filetest costs no node query and does not count against the bounds:
`r/gnoland/boards2/v0` lists 252 files on gno.land, 228 of them filetests,
which `$download` refuses and a copy opens as its 23 sources. Filetests stay
out of every copy, whatever its size, so what a copy holds does not depend on
a threshold; Run locally says so, a filetest's source page links to no copy,
and a link that still names one says the file is not in it. A
`gnomod.toml`'s source page links to its tab. A read skips a name vm/qfile
cannot read (`validFileName`), such as a lowercase `license`, which the chain
takes. `$download` reads 2 MiB at most, and a copy 1 MiB of sources, which
the page carries once, as HTML-escaped text. The read outlives the
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
does, from the server's first render on: the server alone decides whether a
copy previews, and the button carries the label and title the controller
switches to on the first edit. On an unmodified copy of a realm that
declares `Render`, crossing or not, it reads "Preview" and shows what that
realm renders, `Render("")` through `/_/api/eval` at the copy's own path,
where `vm/qeval` hands a crossing `Render` its `cur`.
Anywhere else (edited code, a package, a share, the default code), it reads
"Run locally" and prints the commands that would run, with no file name in
any of them (`localCommands`): `gno run .` for code that declares `main`,
`gno run -expr 'println(Render(""))' .` for one that declares
`Render(path string) string` without a `main`, and `gno test .` when the
editor holds a `gnomod.toml`, which that command needs. A crossing `Render`
gets no run command: `-expr` hands one its `cur` only as the outermost call,
where `gno run` discards the result. A Run result or a Share confirmation that comes back after the
output moved on is dropped, and a newer Run aborts the request it
supersedes. Share prints the link when the Clipboard API is missing.
Analytics counts the page as `playground`, and `$run` as `run`.

The page is served with its code as text, each file of a copy under its
name, in the source view's frame, so a reader without JS still reads it; the
controller replaces it with the editor. Share, Download, Run, the file tabs
and the Output do nothing without JS: they are served hidden and shown by
the controller. PurgeCSS drops the global `[hidden]` rule, so the page,
like the run page, keeps one of its own.

### 2. Expression evaluator on the action page (`$help`)

Below the function list: an input for arbitrary read-only Gno expressions, a
result panel and a history with re-run (ArrowUp/ArrowDown walk it). The
Functions list ends with an entry that links to it. It and the function
cards' live results evaluate through `POST /_/api/eval` (`vm/qeval`), so
nothing mutates state.

A card's Eval button sends the card's params in order, one per input, so
unnamed and repeated (`_`) params count: the value typed, as one operand,
or a zero literal for the param's type (`""` for a `string` or an
`address`). On `$help` a `string` is taken as typed and any other type
trimmed, as whitespace belongs to no number, bool or address (`paramValue`).
A form's fields have no type and are trimmed, as on master: a stray space
in a number fails on chain after spending gas, a lost one in a text does not.
A typed value is kept only when it is exactly one valid Gno
string literal, or, for a param of any other type, one number, rune or
name; anything else is quoted. An expression never gets in: a link's
prefilled value cannot make a card evaluate something else on load and
show it as the function's result. The template renders the types from the
package doc (`.Params`) rather than parsing the signature.
The card's live result and its printed `gnokey query vm/qeval -data` command
use the same literals, so all three agree on what a typed value means; the
`gnokey maketx call -args` command keeps the raw value, which gnokey parses
by type itself. On `$help` a printed arg and its input share the param's
position, so an unnamed or a repeated param keeps its own value; a form's
inputs have no position to share (a radio is several inputs of one param),
and go by name. The URL a card shares still goes by name, so repeated `_`
params collapse there. These helpers are pure functions in `frontend/js/utils.ts`,
pinned by `make -C gno.land/pkg/gnoweb/frontend test` (node's own test
runner, no dependency); the playground's file helpers (`playground-files.ts`)
are pinned the same way. A chain error reaches the reader in the node's
own words, without gnoweb's internal prefix. A newer live result
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
  node's `vm/qdoc` answers for, on its path or on one of its files that the
  package lists (the Run tab of a source file links there). A listing, a
  user page, a missing file, a path with no package or a `$source` page
  renders as if `$run` were not there, with that page's tabs; a node failure
  gets the error page `$help` gives. Error pages have no Run tab: the header
  is built knowing the page is an error. The Run tab stays a realm tab; a
  pure package's `$run` is reached by URL. It makes a realm's fifth tab, so
  on phones a realm's tabs show their labels without icons, and on tablets,
  beside the omnibar, they sit closer so the icons keep their size; pages
  with fewer tabs keep their icons and spacing. gnodev loads a package
  lazily on its first query, so its proxy counts `vm/qdoc` as one, as it
  does `vm/qrender`: a first `$run` gets the run page.
- **Starter script:** built on the server from the same qdoc, so it runs as
  generated, for a gas that does not depend on the package: `_ = pkg.Render`
  under a commented `println(pkg.Render(""))` when the package declares
  `Render(string) string`, the same with `Render(cross(cur), "")` from a
  `main(cur realm)` when that `Render` is crossing, and otherwise
  `_ = pkg.F` for the first exported function, since an unused import does
  not compile. A package with no exported function gets a blank import. The
  name is the import's own (`gno.LastPathElement`, which skips a version
  suffix).
- **Served whole:** the script is highlighted in the editor's place and the
  command is in the page, so the page is complete without JS. The controller
  replaces the script with the editor and fills the command's slots from the
  fields. Reset, Download, the "Include script" switch and the command
  settings do nothing without JS, so they are served hidden and shown by the
  controller, as the Eval buttons are on `$help`. Copy stays, as on every
  code block.
- **Modes and key:** the command follows `$help`: Full Security (query the
  account, `maketx run -broadcast=false > run.tx`, `sign`, `broadcast`) is
  shown and copied first, Fast is one `maketx run` under `$help`'s warning.
  The mode select and the Key / Address field are `$help`'s, under the same
  `action-header` controller and stored values, so a choice on one page
  carries to the other. A line above the command says running it signs with
  the key and broadcasts to the chain ID, which charges its fee. Empty, the
  field and all the commands read `ADDRESS`, as on `$help`: `maketx` and
  `sign` take an address as well as a key name, while the account query the
  secure mode opens with takes only an address.
- **Shell safety:** each typed value (key, gas, fee, coins) is single-quoted
  when it holds a character a shell treats specially, so it stays one
  literal argument; `escapeShellSpecialChars` leaves spaces, so it would not.
  `shellQuote` is a shared helper: `$help`'s commands quote the key the same
  way.
  The heredoc's delimiter (`GNO_EOF`, then `GNO_EOF_1`, ...) is one no line
  of the script equals, so the script cannot end the heredoc and run lines as
  commands. Both helpers are pinned by the frontend tests, which run the
  quoted values through `bash` and the heredoc through `sh`.
- **Gas:** 50,000,000 gas with a 100000ugnot fee. The starter script names
  `Render` (or the first exported function) and shows the call in a
  comment: a `Render` costs what its package makes it cost (145 million gas
  for `r/gnoland/home`, 448 million for `r/docs/markdown` on gnodev), so a
  starter that called it would run out of gas, and be charged the fee, on
  exactly the realms people open. Named only, the example realms' starters
  take 1 to 9 million gas; uncommented, `r/sys/users`'s takes 10 million and
  `r/gov/dao`'s 6 million, both broadcast under the default. The fee pays
  for 100 million gas at the chain's initial price (1ugnot per 1000 gas),
  since the whole fee is charged, even out of gas; the line above the
  command says so and names the flags to raise. `$help`'s calls keep their
  1,000,000,000 and 1000000ugnot: they call any function. Both pairs are Go
  values in `components` (`CallGas`, `RunGas`), read by `command.html`, the
  run page's inputs and, for an emptied field, the controller (the input's
  `defaultValue`).
- **"Include script"** is a `.b-switch`, whose checkbox stays focusable on
  this page; left out, the command asks for `script.gno` in a comment.

The playground and the run page are built unconditionally: `NewRouter`
always sets the help remote and the chain ID their commands print.

### API endpoint

`POST /_/api/eval` evaluates a `pkg_path` + `expression` pair through
`vm/qeval` and returns `{result}` or `{error}`.

- **Requests:** a body that is not `application/json` is refused with 415,
  and a request `Sec-Fetch-Site` names as another site's with 403, both
  before the limiter is charged. The JSON body is what stops a cross-site
  page, which needs a preflight gnoweb never approves. `Origin` is not read,
  so net/http's `CrossOriginProtection` is not used: browsers send no
  `Sec-Fetch-Site` on plain HTTP, and its fallback, comparing `Origin` with
  `Host`, refuses gnoweb's own pages behind a proxy that rewrites `Host`.
  A body over 1 MiB is a 413; another method a 405 with `Allow: POST`.
- **Rate limit:** `state.NewIPLimiter`, built with `-trusted-proxies` like
  the state and search limiters: 40 of burst and 60 per minute per client,
  so a load and a reload of a realm with a dozen zero-argument functions
  fit. Only a query that reaches the node is charged: an answer served from
  the cache below costs no budget. Behind a reverse proxy, operators must
  set `-trusted-proxies` to it; left empty (the default), every reader
  resolves to the proxy and shares one bucket, which the free cached
  answers keep a popular `$help` page from draining.
- **Node load:** answers are kept for 3 s in a bounded LRU, keyed on package
  and expression, and concurrent readers of the same expression share one
  query (singleflight), so a popular Actions page costs the node one query
  per function. Evals count against `AppConfig.MaxConcurrentRPC`, which
  still caps every query gnoweb sends the node, and may hold a quarter of
  it at most (8 of 32 by default), so pages keep the rest. That quarter is
  for readers' expressions only (`ClientAdapter.PublicEval`): gnoweb's own
  evals, like the `/u/` registry lookup, use `Eval` and query like a page,
  so a burst of reader evals cannot fail a user page. They run under the
  handler's request timeout.
- **Answers:** an expression the chain rejects is a 200 with `{error}`;
  an unreachable node is a 502 and a timeout a 504, both with a generic
  message. Results are cut at 64 KiB with a visible marker, kept inside the
  quotes of a string result, so Preview still reads it as text. A failed eval
  logs at Debug, without the expression.

### Editor

The editor keeps every file of the playground (text and undo history, one
CodeMirror state per file), so switching tabs never mixes undo histories.
Its tokens carry chroma's class names, so it is coloured by the same
stylesheet as the read-only source view in both themes, and it shares that
view's frame. That stylesheet now writes its dark rules under `.chroma-light`
for `[data-theme=dark]`, so every code block on the site changes palette in
dark mode, not only the editor's. The editor's own styles are a CodeMirror
theme in TypeScript, because PurgeCSS cannot see the `.cm-*` classes
CodeMirror adds at runtime.

### Frontend approach

Vanilla TypeScript compiled to plain JS, one controller per feature through
the existing `data-controller` dispatch; shared helpers (`postEval`,
`downloadBlob`, icons) are re-exported from `controller.ts`. Compiled output
is committed to `public/js/`. The feature controllers build in one esbuild
run with `--splitting`: CodeMirror, which the playground and `$run` both
open, ships once as a `chunk-<hash>.js` they import, and
`@gnoweb/js/controller.js` stays an import of the `controller.js` every page
loads, so neither bundle carries its own `BaseController` or helpers. The
`@gnoweb/js` alias lives in `tsconfig.json`, which esbuild reads; the Makefile
only maps `@gnoweb/js/controller.js` to that external `./controller.js`. Each
icon keeps a single meaning across gnoweb (the lightning is a transaction,
`ico-run-code` runs the playground, `ico-eval` evaluates, `ico-playground`
opens the playground).

### CSP

`connect-src` is `'self'` only: the JS calls `/_/api/eval` and never the node.
CodeMirror's injected styles carry the per-response CSP nonce.

Assets loaded without the version query (`?v=`) that change in place, such
as a module's own `./controller.js` import, are served `no-cache` and
revalidated against the ETag, so a deploy reaches returning readers. A URL
carrying this build's `?v=`, a font and esbuild's content-hashed
`chunk-<hash>.js` stay `immutable`, as on master, and only on a 2xx or a
304: an error is `no-store`. Any other `?v=` is revalidated, so a request for the next
release's version answered by an old replica mid-deploy is not kept. A page
view costs one 304, for `controller.js`.

## Alternatives Considered

- **Server-side `gno run` / `gno test`:** needs sandboxing, resource limits
  and execution isolation. Deferred; the Run button prints commands instead.
- **Run previewing edited code:** the node can only evaluate deployed code,
  so an edit could only be previewed against what the chain holds, which
  would show the old code's output under the new code. Only an unmodified
  copy previews.
- **Copies and shares in a text format:** what both did at first, every
  file under a `// --- name ---` header. A source line shaped like a
  header split the file when the link was opened, and the blank lines
  around each header were trimmed, so a file's own blank edges were lost.
  JSON has no such line.
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
