# ADR: Integrated Playground, Eval, and Fork Views in gnoweb

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

Add three interactive features to gnoweb as Go-native, single-binary
extensions with no additional runtime dependencies:

### 1. `/_/play` — Playground scratch pad

A multi-file code editor backed by a plain `<textarea>` (no CodeMirror yet).
Supports:
- URL sharing via `?code=` query parameter (base64-encoded)
- Multi-file mode via `// --- filename.gno ---` separators
- Fork-from-package via `?from=` query parameter
- Tab addition/switching, keyboard shortcuts (Ctrl+Enter to run, Tab for indent)

The "Run" button currently provides useful output for packages with `Render`
by calling `/_/api/eval`. For scratch-pad code that has no on-chain package,
it prints CLI instructions instead. This is intentional: the playground is a
first step, not a full REPL.

### 2. `?eval` on realm/package pages — Expression Evaluator

Adds an "Eval" tab to the existing realm/package navigation. Renders:
- A text input for arbitrary Gno expressions
- A result pane updated via `POST /_/api/eval`
- Quick-call buttons for exported, non-crossing, non-method functions
- An expression history with re-run support

The evaluator is read-only by design: only non-crossing functions are shown in
Quick Call, and arbitrary expressions are limited to what `vm/qeval` allows
(no state mutation).

### 3. `?fork` on package/realm pages — Fork to Playground

Loads all `.gno` source files from a package via the existing `ListFiles` +
`File` client methods, concatenates them with `// --- filename.gno ---`
separators, and redirects to the playground view pre-filled with that code.

### API endpoints

Two new JSON endpoints behind `/_/api/`:

- `POST /_/api/eval` — Evaluates a `pkg_path` + `expression` pair via
  `vm/qeval` ABCI query. Returns `{result}` or `{error}`.
- `GET /_/api/funcs?path=...` — Returns exported, non-crossing functions for a
  package using the existing `Doc()` client method. Used by the eval quick-call
  buttons and (in future) playground-aware tooling.

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
ABCI endpoint to also include `'self'`, enabling the JS to call `/_/api/eval`
and `/_/api/funcs` without CSP violations.

## Alternatives Considered

- **WebSocket REPL:** More interactive but much more complex server state.
  Deferred to a later iteration.
- **Server-side gno run / gno test:** Would require sandboxing, resource
  limits, and execution isolation. Out of scope for this PR; instructions are
  printed instead.
- **CodeMirror editor:** Better UX but adds a non-trivial JS dependency.
  Deferred; the `<textarea>` approach is functional and upgradeable in place.
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
- **Positive:** The `/_/api/eval` and `/_/api/funcs` endpoints form a stable
  base for future tooling (IDE integrations, CLI helpers, etc.).
- **Trade-off:** The playground cannot execute scratch-pad code that isn't
  deployed on-chain. This is acceptable for an initial iteration; the UI
  prints CLI instructions to bridge the gap.
- **Trade-off:** No rate-limiting or sandboxing on `/_/api/eval`. Acceptable
  for a dev/exploration tool; should be addressed before exposing this to
  mainnet at scale.
- **Trade-off:** Compiled JS is committed to the repo. This is consistent with
  the existing gnoweb frontend approach.

## Not Yet Implemented

- CodeMirror syntax highlighting and editor features
- Server-side `gno run` / `gno test` execution with sandboxing
- WebSocket REPL
- Wallet integration (signing transactions from playground)
- gnodev hot-reload integration
- Rate limiting / abuse prevention on eval API
- Full test coverage for playground handler (current patch coverage ~16%)
