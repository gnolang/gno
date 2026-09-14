# ADR: CSP nonce for CodeMirror styles in gnoweb

## Context

gnoweb serves a strict Content-Security-Policy when started in strict mode
(`SecureHeadersMiddleware` in `gno.land/cmd/gnoweb/main.go`). The policy included
`style-src 'self'`, which forbids inline styles: both inline `style="…"`
attributes and JavaScript-injected `<style>` elements.

The code editor used on the `run` and `playground` pages
(`pkg/gnoweb/frontend/js/code-editor.ts`) is built on CodeMirror 6. CodeMirror
ships its CSS through the `style-mod` library, which at runtime creates a
`<style>` element and appends it to the document. Under `style-src 'self'` the
browser blocks that element:

```
Content-Security-Policy: The page's settings blocked an inline style
(style-src-elem) from being applied because it violates the following
directive: "style-src 'self'".
```

With the stylesheet blocked, the editor renders unstyled and its gutter/line
numbers, cursor, and selection break. The observed violation is `style-src-elem`
(a `<style>` element), not `style-src-attr`: CodeMirror positions the cursor,
selection layers, and gutter through the CSSOM (`element.style.prop = …`), which
is exempt from CSP, so the injected stylesheet is the only thing being blocked.

## Decision

Add a per-response CSP nonce and let CodeMirror carry it on its injected
`<style>` element, keeping the policy strict (no `'unsafe-inline'`).

1. **Server (`cmd/gnoweb/main.go`)** — in strict mode, generate a fresh random
   nonce per response, emit `style-src 'self' 'nonce-<value>'`, and store the
   nonce in the request context.
2. **Shared helpers (`pkg/gnoweb/csp.go`)** — `NewCSPNonce`,
   `WithCSPNonce`, and `CSPNonceFromContext` own nonce generation and
   context plumbing so the middleware and the handler agree on the value.
3. **HTML (`components/layouts/head.html`)** — echo the nonce into
   `<meta name="csp-nonce">` when present (via a new `HeadData.CSPNonce` field
   populated in `handler_http.go`).
4. **Frontend (`frontend/js/code-editor.ts`)** — read the meta tag and pass the
   value to CodeMirror's built-in `EditorView.cspNonce` facet, which forwards it
   to `style-mod` (`styleTag.setAttribute("nonce", …)`). When no nonce is
   present (non-strict mode), behaviour is unchanged.

## Alternatives considered

- **`style-src 'unsafe-inline'`** — simplest, but re-permits arbitrary inline
  styles site-wide, weakening the policy for a UI-styling need. Rejected.
- **Hashing the injected stylesheet (`'sha256-…'`)** — CSP-compliant, but the
  hash changes with CodeMirror/theme upgrades and there are multiple style
  modules (base theme + one-dark, swapped by the dark/light toggle), so the
  allowlist would need constant maintenance. Rejected in favor of a nonce.
- **Running the editor in a shadow DOM** — adopted stylesheets escape
  `style-src`, but this is a larger refactor and does not fully sidestep CSP for
  in-tree inline styles. Rejected as disproportionate.

## Consequences

- The `run` and `playground` editors work under the strict CSP without relaxing
  it; `script-src` and the rest of the policy are untouched.
- The CSP header now varies per response (fresh nonce), which is expected for
  nonce-based policies and prevents caching of the header value.
- Nonce support is centralized in `pkg/gnoweb/csp.go`; any future inline
  `<style>`/`<script>` needing a nonce can reuse the same context value.
- In non-strict mode no nonce is emitted and the editor behaves exactly as
  before.
