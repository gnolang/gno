# ADR: Accept `gno.land/...` URLs in the gnoweb search bar

## Context

The gnoweb search bar (`SearchbarController` in
`gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts`) lets users navigate
gnoweb by entering a path or URL. Today it accepts:

- relative paths (`r/foo`, `/r/foo`) — joined to `window.location.origin`
- absolute URLs with a scheme (`https://...`) — followed verbatim

Realm and package import paths in Gno are written as `gno.land/r/foo` or
`gno.land/p/foo`. Users frequently copy these from chat, docs, source code, or
the gno.land address bar and paste them into the search bar. With the previous
behavior:

- `gno.land/r/foo` (no scheme) was joined to the origin literally, producing
  `https://<gnoweb-host>/gno.land/r/foo` — a broken path.
- `https://gno.land/r/foo` was always followed externally even when the user
  was browsing a local node, gnodev, or a staging host. The user had to
  hand-edit the host to navigate.

## Decision

Normalize input in the search bar by stripping a leading `gno.land` host and
treating the remainder as a local gnoweb path. Resolution lives entirely in
the client (`SearchbarController.resolveTarget`):

1. If the input starts with `http://` or `https://`, parse it as a URL.
   - If the hostname is exactly `gno.land`, rewrite to the current origin
     while preserving `pathname`, `search`, and `hash`.
   - Otherwise, follow the URL verbatim (existing behavior for arbitrary
     external URLs, including subdomains like `staging.gno.land`).
2. Otherwise, strip a leading `gno.land` host using
   `/^gno\.land(?=\/|$|\?|#)/i`, prepend `/` if missing, and join to
   `window.location.origin`.

Examples (assuming `window.location.origin = "https://gnoweb.example"`):

| Input                              | Resolved                                              |
|------------------------------------|-------------------------------------------------------|
| `r/foo`                            | `https://gnoweb.example/r/foo`                        |
| `/r/foo`                           | `https://gnoweb.example/r/foo`                        |
| `gno.land/r/foo`                   | `https://gnoweb.example/r/foo`                        |
| `gno.land/r/foo?bar=1#x`           | `https://gnoweb.example/r/foo?bar=1#x`                |
| `https://gno.land/r/foo`           | `https://gnoweb.example/r/foo`                        |
| `https://staging.gno.land/r/foo`   | `https://staging.gno.land/r/foo` (unchanged)          |
| `https://example.com/r/foo`        | `https://example.com/r/foo` (unchanged)               |

## Alternatives Considered

### Server-side redirect for `/gno.land/...` paths

Add HTTP redirect middleware that strips a leading `/gno.land/` from the
request path. Rejected for this change: the search bar is the user's stated
entry point, and a URL pasted into the address bar already lacks the
`gno.land` prefix on the local host. A future server-side redirect can be
added independently if the need arises (e.g. for deep links typed directly
into the address bar).

### Rewrite `gno.land` subdomains too

Rejected to keep the rule simple and predictable. Subdomains
(`staging.gno.land`, `test6.gno.land`, etc.) are distinct deployments with
potentially different state; a paste from one should not silently land on
another. A user who wants cross-host navigation can edit the host explicitly
or paste a path-only form.

### Rewrite all hosts, not only `gno.land`

Rejected to preserve the existing escape hatch where a fully-qualified
external URL (e.g. a documentation link) is followed verbatim.

## Consequences

- Realm/package paths copied from anywhere — chat, docs, source, address bar
  — work in the search bar without manual editing.
- A `gno.land` URL pasted while browsing a local node or gnodev lands on the
  current origin without manual host editing.
- Subdomain URLs and arbitrary external URLs continue to behave as before.
- Logic is client-only and has no server-side surface; existing handlers,
  redirects, and tests are untouched.

## Files

- `gno.land/pkg/gnoweb/frontend/js/controller-searchbar.ts` — new
  `resolveTarget` static method; `searchUrl` delegates to it.
- `gno.land/pkg/gnoweb/public/js/controller-searchbar.js` — rebuilt bundle.
