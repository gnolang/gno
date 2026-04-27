# SimpleAnalytics taxonomy — gnoweb

This file documents every event gnoweb emits to SimpleAnalytics, what each event carries, what it deliberately does not carry, and how to add a new one. The taxonomy is designed for a future internal indexer / DX dashboard (see #5467) to consume alongside the SA dashboard: flat event names, stable dimension keys, bounded cardinality.

## Privacy posture

gno.land is privacy-first, including at the analytics layer.

We never collect:

- Wallet addresses, public keys, or transaction signatures
- Function arguments or return values
- Form input values (address inputs, parameter inputs, search queries) as event payload. Note: the help-page URL pattern `…$help&func=…&arg1=…` reflects typed param values in its query string and is captured by SA as part of the standard pageview URL until we strip it client-side.
- Cookies, persistent visitor IDs, fingerprinting data
- IP addresses beyond what SA's defaults handle (hashed, daily-rotated, never stored raw)
- Anything that distinguishes a person across sessions

SA respects `Navigator.doNotTrack` by default: pageviews and events are not recorded for users who set DNT.

All metadata is derived from public, server-side context (URL pattern, layout mode, view type, chain id) or from low-cardinality DOM signals (mode enum, `open` boolean, scroll threshold).

## Pageview metadata

Set on every pageview via `window.sa_metadata`. A synchronous classic
`<script src="sa-bootstrap.js" data-page-type="…" data-chain-id="…">` loads
before SA's async `latest.js`, reads its own `data-*` attributes, and
assigns `window.sa_metadata`. This keeps the bootstrap CSP-safe (no inline
script) and deterministic (classic scripts block parsing, so SA's async
scripts cannot start loading until the metadata is set).

| Key | Source | Cardinality |
|---|---|---|
| `page_type` | `components.ClassifyPageType(mode, view)` | 11: home, user, pure, realm, source, help, directory, status, redirect, explorer, other |
| `chain_id` | `cfg.ChainID` (env constant) | low (one per deployment) |

## Custom events

### DISCOVER

| Event | Trigger | Payload |
|---|---|---|
| `search_action` | Header searchbar form submit | none (count only — never the query) |
| `network_popup_toggle` | `change` on the network-info popup checkbox | `{open: bool}` |
| `breadcrumb_click` | Click on any breadcrumb anchor | none |
| `back_navigation` | Browser back / forward (`popstate`) | none |
| `toc_toggle` | Native `<details>` toggle on `details.accordion` | `{open: bool}` |

### BUILD (action page)

| Event | Trigger | Payload |
|---|---|---|
| `mode_change` | Action header dispatches `mode:changed` | `{mode: 'secure' \| 'fast' \| 'url'}` |
| `send_mode_toggle` | Click on the send-mode label (Add/Remove from command) | `{active: bool}` |
| `qeval_preview` | qeval result element transitions between success and failure (placeholder text and error class read from `data-qeval-*` attrs on the element) | `{success: bool}` |
| `address_filled` | First non-empty value typed in `#action-user-address` | none (fires once per page-load) |
| `params_filled` | First non-empty value typed in any param input | none (fires once per page-load) |
| `submit_action` | Action form submission | `{func, pkgpath}` (capped at 64 / 128 chars) |

### Settings

| Event | Trigger | Payload |
|---|---|---|
| `theme_toggle` | Theme controller dispatches `theme:changed` with the user-chosen preference | `{theme: 'light' \| 'dark' \| 'system'}` |
| `devmode_toggle` | `change` on `#header-input-devmode` (3-dot menu on home) | `{enabled: bool}` |

### Package listing (user page)

| Event | Trigger | Payload |
|---|---|---|
| `list_filter_search` | Debounced (250ms) `input` on `#packages-search` | none (count only — never the query) |
| `list_sort_change` | `change` on `input[name="order-mode"]` | `{order: 'asc' \| 'desc'}` |
| `list_display_change` | `change` on `input[name="display-mode"]` | `{mode: 'display-grid' \| 'display-list'}` |

### Read engagement

| Event | Trigger | Payload |
|---|---|---|
| `scroll_depth` | Window scroll on source/help pages, fires once per threshold per page-load | `{threshold: '50' \| '75' \| '100', surface: 'source' \| 'action'}` |

### Copy actions

| Event | Trigger | Payload |
|---|---|---|
| `copy_action` | Click on any `button[data-controller~="copy"]`; kind inferred from `data-copy-*` attributes | `{kind: 'link' \| 'source' \| 'func_signature' \| 'gnokey_command' \| 'unknown'}` |

### Outbound

Two layers fire in parallel.

**Generic catch-all (SA `auto-events.js`, no code):** every outbound link, download, and `mailto:` click fires SA's built-in `outbound`, `download`, or `email` event with the host or filename as payload. Covers all outbound clicks regardless of tagging.

**Named priority outbounds (`outbound_<target>`):** anchors carrying `data-outbound="<target>"` fire an additional `outbound_<target>` event so high-traffic destinations show up under stable names in the dashboard rather than being aggregated by host.

| target | URL pattern |
|---|---|
| `docs` | docs.gno.land |
| `faucet` | faucet.gno.land |
| `status` | status.gnoteam.com |
| `github` | github.com/gnolang/* |
| `twitter` | twitter.com/_gnoland |
| `discord` | discord.gg/* |
| `youtube` | youtube.com/@_gnoland |

`data-outbound` is set in Go (`layout_footer.go`, `layout_header.go`) and rendered by `footer.html` / `ui/header_link`. To tag a new outbound: add `Outbound: "<target>"` to the `FooterLink` / `HeaderLink` in Go.

## Cardinality caps

| Dimension | Cap | Why |
|---|---|---|
| `func` (submit_action) | 64 chars | realm authors set this attribute freely |
| `pkgpath` (submit_action) | 128 chars | same |
| All enum payloads (`page_type`, `mode`, `kind`, `theme`, `surface`, `threshold`, `order`, `display`) | fixed enum set | enumerated above |
| Outbound target | fixed enum set | only `data-outbound` values defined in Go are emitted |

## Adding a new event

1. **Confirm it's not derivable.** If a dashboard query against existing pageviews + events would answer the question, do not add an event.
2. **Pick a stable, flat name.** snake_case, no namespacing colons. Match an existing prefix when applicable (`copy_*`, `list_*`, `outbound_*`, `*_toggle`, `*_change`).
3. **Define the payload schema.** Low cardinality only. Booleans or fixed enums preferred. No raw user input. Document the cap if a string field could grow.
4. **Wire the listener** in `frontend/js/analytics.ts`. Use capture phase if any controller stops propagation. Lazy-attach with `if (element)` for elements that may not exist on every page.
5. **Rebuild** `make -C gno.land/pkg/gnoweb ts` and commit both `frontend/js/analytics.ts` and the generated `public/js/analytics.js`.
6. **Document the event in this file** under the appropriate section.
7. **If the event needs a server-side data attribute** (e.g. `data-outbound`), extend the relevant struct and template in `components/`.

## Third-party posture

- `sa.gno.services` proxies to SimpleAnalytics. SA respects `Navigator.doNotTrack` by default: DNT users are not tracked.
- IPs are SHA-256 hashed and daily-rotated; never stored raw. No cookies, no fingerprinting.
- The `<noscript>` pixel uses `referrerpolicy="no-referrer"` so JS-disabled users don't leak the page URL via Referer.
- `auto-events.js` records the destination URL of every outbound click and download as event payload: SA-side behavior, not gnoweb instrumentation.

## Files

- `frontend/js/analytics.ts` — event delegation source of truth
- `public/js/analytics.js` — esbuild output, embedded via `go:embed`
- `components/analytics.go` — `AnalyticsData`, `ClassifyPageType`
- `components/layouts/analytics.html` — script wiring (sa-bootstrap + SA scripts + analytics.js)
- `frontend/js/sa-bootstrap.ts` — classic IIFE that reads `data-*` attrs and sets `window.sa_metadata`
- `components/layout_footer.go` + `layouts/footer.html` — footer outbound tagging
- `components/layout_header.go` + `layouts/header.html` — header outbound tagging
