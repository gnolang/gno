# PRxxxx: Host-aware help URL on the Action page (fix #5611)

## Context

The "Link" button on the Action page (`/r/<pkg>$help&func=...`) copies a URL to the
clipboard meant to be shared. Issue #5611 reports that the copied value is a
path-only string (`/r/sys/cla$help&func=Sign&hash=`) — unusable when pasted
outside the current site context.

History of the regression:

- **#4917** (Dec 2025) removed `data.Domain` from `buildHelpURL` to fix an
  unrelated bug: the rendered string `gno.land/...` (no scheme) was treated by
  browsers as a relative path and appended to the current route. The fix
  produced a path-only URL with leading `/`, and JS in
  `controller-action-function.ts` re-absolutized it via
  `new URL(rel, window.location.origin).toString()` whenever the user typed in
  an input.
- **#4964** (Jan 2026) replaced that `new URL` parsing with a `split("&")` /
  `join("&")` to preserve the gno.land URL syntax (`$help&key=val&...` keeps
  `&` as a path separator, which `URL.searchParams.set` cannot model). The
  client-side absolutization was a side effect of `new URL` and was lost in the
  refactor.

Net effect since #4964: every `data-copy-text-value` from `buildHelpURL` is
path-only, so the clipboard never receives a shareable URL. Execute kept
working only because the form `method="GET"` + path-only `action` is resolved
natively by the browser against the current origin.

## Decision

Two layers, both correct on any deployment (gnodev, standalone gnoweb,
prod gno.land, behind reverse proxy, custom domain, ngrok tunnel).

### Layer 1 — Server-side: `buildHelpURL` returns absolute URLs

- New helper `requestOrigin(*http.Request) string` derives `scheme://host`
  from the request, honoring `X-Forwarded-{Proto,Host}` for proxied
  deployments. Returns `""` when no host can be determined; callers degrade
  gracefully to path-relative.
- New `Origin` field on `weburl.GnoURL`, set once in
  `prepareIndexBodyView` after `ParseFromURL`. Any handler that takes
  `*weburl.GnoURL` can read it without signature changes.
- `HelpData.Origin` carries the value to the template. `buildHelpURL`
  prepends it.

### Layer 2 — Client-side: defense-in-depth in `controller-copy.ts`

`_copyTextToClipboard` prefixes `window.location.origin` when the text starts
with `/`. Catches the case where Origin is empty (test fixtures, malformed
requests) and any future template that emits a path-only `data-copy-text-value`.

## Alternatives considered

- **Server-side `data.Domain`** (the pre-#4917 form): broken on gnodev where
  the configured Domain (`gno.land`) does not match the serving host
  (`127.0.0.1:8888`); the clipboard URL would point users back to prod.
- **JS-only fix in `controller-action-function.ts`**: would require either
  re-introducing `new URL` (which corrupted the gno.land URL syntax in #4964),
  or duplicating origin-prefix logic on every update path.
- **Client-side only (`controller-copy.ts` heuristic)**: sufficient for the
  functional bug, but leaves the DOM (form `action`, `data-copy-text-value`)
  showing path-only URLs at first paint. View-source / DevTools inspection is
  surprising. Adopted as Layer 2 alongside the server-side fix.
- **`Origin` as a parameter threaded through handlers**: rejected because it
  pollutes the signature of every view handler. `GnoURL.Origin` keeps the
  per-request value where the handlers already look.
- **`X-Forwarded-*` behind a `-trusted-proxies` flag**: deferred. Current
  trust model is "operators must strip these headers at the edge if exposing
  gnoweb directly"; documented in the helper's doc comment.

## Consequences

- The Action page Link button copies an absolute URL on every deployment.
- The form `action` is also absolute. Submit goes to the same origin (no
  behavioral change), but DOM inspection is now consistent with the clipboard
  value.
- `_updateArgInDOM` in `controller-action-function.ts` continues to work
  unchanged: its `split("&") / join("&")` logic preserves the `https://` /
  `http://` prefix as `parts[0]`.
- One extra `string` per `GnoURL` (negligible).
- Operators behind reverse proxies must continue to strip
  `X-Forwarded-Host/Proto` from inbound requests if gnoweb is also reachable
  directly; otherwise a client could inject an arbitrary host into the
  clipboard URL.

## Files

- `gno.land/pkg/gnoweb/handler_http.go` — `requestOrigin`,
  `gnourl.Origin = requestOrigin(r)`, `Origin` propagated into `HelpData`.
- `gno.land/pkg/gnoweb/handler_origin_test.go` — table-driven tests for
  `requestOrigin` (TLS, X-Forwarded-*, IPv6, empty host).
- `gno.land/pkg/gnoweb/handler_http_test.go` — integration test
  `TestHTTPHandler_HelpURLOrigin` covering direct http, reverse proxy, custom
  domain.
- `gno.land/pkg/gnoweb/weburl/url.go` — `Origin` field on `GnoURL`.
- `gno.land/pkg/gnoweb/components/view_action.go` — `HelpData.Origin`,
  `buildHelpURL` uses it.
- `gno.land/pkg/gnoweb/frontend/js/controller-copy.ts` — Layer 2 normalize.
- `gno.land/pkg/gnoweb/public/js/controller-copy.js` — regenerated bundle.
