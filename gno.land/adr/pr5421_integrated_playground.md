# ADR: Integrated Playground, Eval, Fork, and Run Views in gnoweb

## Context

gno.land previously had no interactive code evaluation in gnoweb. Users who
wanted to call realm functions or experiment with Gno code had to use external
tools (gnokey CLI, a separate gnostudio/studio app, or copy-paste workflows).

The key problems:

1. **Friction for exploration.** Reading a realm's `Render` output or calling
   a view function required setting up a local environment or using CLI tools.
2. **No in-browser eval.** Users could not quickly test expressions against a
   deployed package from the web UI.
3. **No code scratch pad.** There was nowhere in gnoweb to write and share
   short Gno snippets.
4. **Separate studio dependency.** The only interactive option required running
   a separate application, breaking the single-binary model.

## Decision

Add four interactive features to gnoweb as Go-native, single-binary
extensions with no additional runtime dependencies:

### 1. `/_/play` — Playground scratch pad

A multi-file code editor backed by CodeMirror (Go/TOML modes, via the shared
`@gnoweb/js/code-editor` component integrated in #5674). The `<textarea>` is
retained as a hidden, progressively-enhanced fallback that also carries the
initial code into the editor.
Supports:
- URL sharing via `?code=` query parameter (base64-encoded)
- Multi-file mode via `// --- filename.gno ---` separators
- Fork-from-package via `?from=` query parameter
- Tab addition/switching, keyboard shortcuts (Ctrl+Enter to run, Tab for indent)

The "Run" button currently provides useful output for packages with `Render`
by calling `/_/api/eval`. For scratch-pad code that has no on-chain package,
it prints CLI instructions instead. This is intentional: the playground is a
first step, not a full REPL.

### 2. Expression Evaluator on the Actions page (`?help`)

Rather than a tab of its own, the evaluator is a section of the existing
Actions page, above the per-function transaction builders. Renders:
- A text input for arbitrary Gno expressions
- A result pane updated via `POST /_/api/eval`
- An expression history with re-run support

Each non-crossing function in the list below it also gets an "Eval" button that
seeds the input with a call to that function, which is the quick-call path.

The evaluator is read-only by design: only non-crossing functions offer the
Eval button, and arbitrary expressions are limited to what `vm/qeval` allows
(no state mutation).

### 3. `?fork` on package/realm pages — Fork to Playground

Loads all `.gno` source files from a package via the existing `ListFiles` +
`File` client methods, concatenates them with `// --- filename.gno ---`
separators, and redirects to the playground view pre-filled with that code.

### 4. Run scratchpad on the Actions page — Dry Run

A code editor seeded with a script that imports the realm being viewed, so the
user can do what a single function call cannot express: sequence several calls,
or pass a value that has to be constructed. It produces a copy-pasteable
`gnokey maketx run` command, a `script.gno` download, and a **Dry Run** button
that simulates the transaction against the node without committing it.

The Actions page now reads top to bottom as increasing power: package overview,
expression evaluator (read-only), run scratchpad (a whole script), then the
per-function transaction builders.

Dry Run posts to `/_/api/dryrun`, which builds a `vm.MsgRun` from the script
and runs it through the node's simulate path. It requires a bech32 address,
since a dry run has no way to resolve a local key name.

### API endpoints

Three JSON endpoints behind `/_/api/`, all served by the playground feature:

- `POST /_/api/eval` — Evaluates a `pkg_path` + `expression` pair via
  `vm/qeval` ABCI query. Returns `{result}` or `{error}`.
- `GET /_/api/funcs?path=...` — Returns exported, non-crossing functions for a
  package using the existing `Doc()` client method. Used by the eval quick-call
  buttons and (in future) playground-aware tooling.
- `POST /_/api/dryrun` — Simulates a `MsgRun` transaction for a `pkg_path` +
  `script` + `address` triple. Returns `{result}` or `{error}`.

`eval` and `dryrun` share one per-IP token-bucket limiter, since both put work
on the same node; `dryrun` additionally caps its request body.

### Frontend approach

Vanilla TypeScript compiled to plain JS (no React, no npm at runtime). Each
feature is a standalone controller function loaded dynamically by the existing
`data-controller` dispatch mechanism in `index.ts`. No new build tooling;
compiled output is committed to `public/js/`.

Cache busting is handled by extracting the `?v=` version suffix from
`index.js`'s own `<script src>` URL and forwarding it to dynamic controller
imports.

### CSP update

`connect-src` in the Content Security Policy was extended from just the remote
ABCI endpoint to also include `'self'`, enabling the JS to call `/_/api/eval`,
`/_/api/funcs` and `/_/api/dryrun` without CSP violations.

## Alternatives Considered

- **WebSocket REPL:** More interactive but much more complex server state.
  Deferred to a later iteration.
- **Server-side gno run / gno test:** Would require sandboxing, resource
  limits, and execution isolation. Out of scope for this PR; the scratchpad
  simulates instead (`/_/api/dryrun`), which reuses the node's existing
  transaction path and commits nothing, and the playground prints CLI
  instructions.
- **A separate Run page and header tab:** Rejected. It would duplicate the
  Actions page's purpose and force a tab choice before the user knows whether
  their work is one call or several, with the package path and chain context
  they already have on screen not following them across.
- **CodeMirror editor:** Adopted (#5674) via the shared `@gnoweb/js/code-editor`
  component — Go/TOML syntax modes, with the `<textarea>` kept as a hidden
  progressive-enhancement fallback.
- **Separate gnostudio service:** Breaks the single-binary model. The goal is
  to keep gnoweb self-contained.
- **Stimulus.js or other controller framework:** The codebase already has a
  minimal controller dispatch in `index.ts`. Keeping new controllers as
  standalone functions avoids framework coupling.

## Consequences

- **Positive:** Users can evaluate read-only realm expressions from the
  browser without any local tooling.
- **Positive:** Developers can fork any package's source into the playground
  with one click.
- **Positive:** Zero new runtime dependencies; gnoweb stays a single binary.
- **Positive:** The `/_/api/eval`, `/_/api/funcs` and `/_/api/dryrun` endpoints
  form a stable base for future tooling (IDE integrations, CLI helpers, etc.).
- **Positive:** Everything you can do to a realm from the browser is now on one
  page, in one order, sharing the package path and chain context already on
  screen.
- **Trade-off:** `components` now imports a feature package. This is the only
  such edge and exists solely to reach an embedded FS; `feature/run` is kept
  free of Go dependencies on the rest of gnoweb so it cannot cycle.
- **Trade-off:** Dry Run requires a bech32 address rather than a key name, so
  the user must look up the address if they're used to just using the key name.
- **Trade-off:** The playground cannot execute scratch-pad code that isn't
  deployed on-chain. This is acceptable for an initial iteration; the UI
  prints CLI instructions to bridge the gap.
- **Trade-off:** `/_/api/eval` and `/_/api/dryrun` are rate-limited per IP but
  not sandboxed. Acceptable for a dev/exploration tool; the isolation question
  should be revisited before exposing this to mainnet at scale.
- **Trade-off:** Compiled JS is committed to the repo. This is consistent with
  the existing gnoweb frontend approach.

## Not Yet Implemented

- CodeMirror syntax highlighting and editor features
- Server-side `gno run` / `gno test` execution with sandboxing (the scratchpad
  simulates via `/_/api/dryrun`; it does not execute uncommitted code)
- Dry Run against a key name rather than a bech32 address. As a minimum, can
  check if the name is registered on-chain
- Send / gas / deposit values from the Dry Run request (the handler currently
  hardcodes them; the UI collects them only for the generated CLI command)
- WebSocket REPL
- Wallet integration (signing transactions from playground)
- gnodev hot-reload integration
- Full test coverage for playground handler (current patch coverage ~16%)
