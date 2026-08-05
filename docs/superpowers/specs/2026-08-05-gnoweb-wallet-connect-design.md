# gnoweb as a dapp: wallet connect, identity, install page, QR

Design doc. Status: proposed.

## Context

This branch (`feat/gnoconnect-in-page-wallets`) already built the discovery
half of GnoConnect:

- `frontend/js/wallet-discovery.ts` — the `gno:registerWallet` /
  `gno:requestWallet` announce protocol, with the untrusted-input clamps.
- `frontend/js/controller-wallet-launch.ts` — a chooser dialog merging
  announced extensions, registry apps and legacy Adena, claiming the Execute
  submit at window-capture.
- `docs/resources/gnoconnect.md` — the standard, including `connect`,
  `getAccount`, `getNetwork`, `switchNetwork` and the `signer` pin, none of
  which gnoweb calls yet.

What is missing is the layer above: gnoweb has no notion of *who* the user is.
It asks which wallet on every Execute and never asks which identity. This
design turns a per-transaction chooser into a persisted session, and fills the
two gaps the existing ADR deferred — a result surface for both transports, and
a cross-device QR.

## Scope

One spec covering four things:

1. Connect (login) with a chosen public key from a chosen wallet.
2. A generated avatar in the header, with an identity menu.
3. A gnoweb page listing compatible wallets and where to install them.
4. A QR of the tx-link, for signing on a phone.

Out of scope: any server-side session, proof of control, per-user rendering,
or realm-side changes.

## Decisions

### 1. `connect()` becomes mandatory for in-page wallets

`docs/resources/gnoconnect.md` currently lists `connect` under "Optional
methods". A wallet may therefore announce itself, implement `sendTx` alone, and
be impossible to identify — it cannot be logged into, and cannot be pinned
against as `signer`.

Promote it. Edits:

| Section | Edit |
|---|---|
| `provider` — what the page calls (~L413) | `connect` and `GnoAccount` move into the core block beside `sendTx` |
| Optional methods (~L489) | `connect` removed; `getAccount` / `getNetwork` / `switchNetwork` / `signTx` / `sendMsgs` stay |
| Connecting, and what it gates (~L589) | Rewritten. Today it argues *"`connect` is listed as optional because a wallet that gates nothing does not need it"*. That inverts: `connect` is how a page learns who the user is, and origin-gating becomes one thing a wallet may hang off it. The "`sendTx` MAY carry its own approval / a page MUST handle `not_connected` and retry" rule survives unchanged |
| The `signer` pin (~L290) | *"A wallet that implements neither cannot be pinned against"* narrows to launch links |
| Launch links, `connect` host (~L900) | The `connect` host becomes required, symmetric with the above. Without it a mobile login flow has no defined route, and the failure is silent — indistinguishable from the user abandoning the launch |

Additionally: **`connect` on an already-approved origin MUST resolve without
prompting.** Otherwise gnoweb's session-restore throws a wallet popup on every
page navigation, which makes the feature unusable.

Promoting an optional method to mandatory is a breaking change to the standard.
It is free now and will not be later: no wallet implements the in-page half yet
(`gno.land/adr/gnoconnect_in_page_wallets.md` — *"No such extension announces
yet"*).

This does **not** make gnokey special. `gnoconnect.md:981` states that nothing
in the standard is specific to any named implementation, and that holds.

### 2. The session is client-side only

The connected address lives in `localStorage`. No cookie, no server session, no
signature challenge, no per-user rendering. Every page stays cacheable and
identical for all visitors.

Connecting buys three things:

- the avatar and address in the header;
- Execute goes straight to the remembered wallet, with no chooser;
- the address is pinned as `signer` on every intent, so the wallet signs as the
  account the user picked rather than whichever is to hand.

Persisted: `{ rdns, address, chainid, name }`, keyed on `rdns` — the durable
identity per `gnoconnect.md`. Never the `uuid`, which is a per-page-load handle.

### 3. gnoweb ships one external wallet, by choice

The registry keeps its mechanism and stays at one app entry (gnokey-mobile).
On mobile there is no multi-wallet chooser.

This is a deliberate narrowing and should read as one. The ADR that opens this
branch argues `window.adena` was wrong *because* it picks a winner
(`gnoconnect_in_page_wallets.md:11`). Hardcoding one app is the same shape of
decision, one transport over. It is defensible on a different ground: apps
cannot announce themselves, so there is no discovery protocol being overridden,
and a chooser over a list of length one is speculative generality. Announced
in-page wallets remain honoured on every device, mobile included.

### 4. One Execute rule, both platforms

> **Execute while unconnected always ends in a navigation. If exactly one
> wallet is available, it also fires.**

The destination is the function help page in every case except a successful
signature, which lands on `?status=success&hash=…` instead. Nothing stays put.

This works because the destination already exists. Execute is a `method="GET"`
submit whose `action` *is* the function help URL, rewritten on every keystroke:

- `components/views/action.html:106` — `<form class="params" method="GET" action="{{ buildHelpURL $data . }}">`
- `frontend/js/controller-action-function.ts:191` — `executeForm?.setAttribute("action", updatedUrl)`
- `components/view_action.go:73` — `buildHelpURL` emits `$help&func=Name&p1=v1&p2=v2`
- `components/views/action.html:120` — arriving back, `getSelectedArgValue` prefills each input

So the "redirect" is the same `$help` page reloading with `func=` and the args
pinned. Three consequences fall out for free:

- **The silent-launch problem disappears.** Fire the deeplink *and* navigate. If
  the app opens, the page is behind it. If it does not, the user is already
  looking at the fallback. No timeout, no visibility heuristic. This replaces
  the "always show the chooser" escape hatch of `161d32807` with something that
  handles the failure it was actually for.
- **The QR cannot go stale.** It is rendered on the destination page, whose URL
  carries the exact submitted args. Page render and tx-link are one event.
- **Retry loses nothing.** Args round-trip through the URL.

Add `#func-<Name>` to the navigation target so scroll position survives.
Accepted loss: arguments typed into *other* functions on the same `$help` page.
Rare, and not worth a special case.

The send toggle is deliberately **not** encoded in `buildHelpURL`, so it resets
on every landing. Re-confirming a coin transfer after a failed attempt is
correct.

#### Ordering, per transport

Navigation and wallet trigger are both consequences of one click. Which happens
first is forced by the transport, and the user never sees the difference.

**In-page** — sign first, navigate on the outcome. `sendTx()` returns a Promise
in the live page; navigating first would destroy it, and some extensions cancel
a pending approval when the requesting tab navigates away. So the popup opens
immediately, and the navigation is the *result*:

| Outcome | Destination |
|---|---|
| `Approved` | `?status=success&hash=…` — the landing state both transports converge on |
| `Rejected` | the function help page |
| `not_connected` | neither: call `connect()` and retry, per `gnoconnect.md:589`. Only a declined connect becomes `Rejected` |
| any other error / throw | the function help page |

**External** — fire the launch link, then navigate. There is no promise to
protect; the result returns through the `callback` URL on a fresh load.

`Rejected` navigates like everything else. On mobile it has no choice — a
cancel returns `<callback>?status=cancelled&state=…` (`gnoconnect.md:928`), so
coming back *is* a navigation — and making desktop stay put would reintroduce
the platform split this rule removes. It is also the right destination:
rejecting usually means "not like this" (wrong account, wrong wallet, I would
rather use my phone), and the help page answers all three.

**Verify during implementation, do not assume:** whether an extension opens its
approval popup without user activation, and whether iOS Safari blocks a
gesture-less custom-scheme navigation. Both affect only the internal ordering,
not the user-visible flow.

## Components

Each is separately understandable and testable.

### `components/wallet_registry.go` — schema change

Adena joins the registry, which currently requires a launch-link scheme on
every entry (`wallet_registry.go:59`).

| Field | Change |
|---|---|
| `Scheme` | Optional. Validation still rejects malformed non-empty schemes; `seenSchemes` dedup skips empties or two extensions collide |
| `Kind` | New: `extension` \| `app`. The controller never launches a scheme-less entry; the install page groups by it |
| `RDNS` | New. Matches a registry entry to an *announced* wallet, so Adena announcing itself is not also offered as "install Adena". Replaces the hardcoded `LEGACY_PROVIDERS` table in `controller-wallet-launch.ts:32` |
| `InstallURL` | Becomes a per-platform list — Chrome Web Store and GitHub where both exist |
| `Platforms` | Was "informational for now". Now load-bearing: the install page groups by it, and the touch-only external-login rule reads it |

Validation stays at package init with a panic — a malformed registry is an
authoring error.

### `feature/connect/` — the session

New self-contained feature package, following `feature/state/`: Go component +
handler + templates + `frontend/`.

Frontend module owning `localStorage`, exposing: current identity, connect,
disconnect, switch. Sits above `wallet-discovery.ts`, which is unchanged.

Session restore on load: render the remembered address immediately — no flash
of "Connect" on every navigation — then reconcile asynchronously. Prefer
`getAccount()` (guaranteed not to prompt); fall back to `connect()`, which per
the spec change above must resolve silently for an approved origin. If
reconciliation returns a different address, update; if it fails, drop the
session. The optimistically-rendered address is the user's own and is display
only, so a brief unverified render costs nothing.

### Avatar

A ~60-line port of `ethereum-blockies-base64`: a 32-bit xorshift PRNG seeded
from the full bech32 address, an 8×8 grid mirrored horizontally, three colours
(background, primary, spot).

Rendered as **inline SVG rects**, not a base64 PNG. It scales, needs no
encoder, and costs less code than the original.

### Header identity menu

`components/layouts/header.html` has no identity control today. It gains a
"Connect" button, replaced once connected by the avatar plus truncated address,
opening a menu offering copy address, change wallet, disconnect.

"Change wallet" opens the wallet list **always**, even with one wallet
installed — the one place that differs from login, which auto-picks a lone
wallet. The list offers "install a compatible wallet", linking to the page
below.

### Install page

A gnoweb route at `/wallets`, rendering the registry: name, icon, kind, and
per-platform install links. Grouped by platform, with the current platform
first. Registered in `app.go` alongside `/status.json` and `/search.json` —
gnoweb-native, not a realm alias, so it stays accurate to the gnoweb version
and works on any chain.

Server-rendered from the same embedded registry, so it works offline and on any
chain, and cannot drift from what the chooser offers.

### QR on the help page

Adopt the approach of [PR 4602](https://github.com/gnolang/gno/pull/4602):
server-side generation via `buildHelpQR` in `registerHelpFuncs`, returning a
base64 PNG embedded in the template, using `github.com/boombuler/barcode`. No
JS QR encoder to vendor.

That PR is an open draft whose remaining blocker is *"make the QR code update
when the tx link changes"* — unsolvable for a server-rendered QR on a live
form. Our flow dissolves it: the QR renders on the destination page, whose URL
already carries the submitted args.

Coordinate with that PR rather than duplicating it. It adds a root `go.mod`
dependency that propagates to `contribs/gnobro` and `contribs/gnodev`.

The help page is pointer-aware. On desktop it leads with the QR; on touch it
leads with "Open in gnokey" plus install links and drops the QR, which is
useless on the phone holding it.

**Open detail:** what the QR shows if a user edits an argument *after* landing
on the help page. Preferred resolution — the QR is presented behind a button
whose target is the live `form.action`, so opening it navigates to the current
URL and renders fresh. This needs confirming against the layout before it is
settled.

## Error handling

- A wallet that announces without `connect` is now non-conforming, not a
  variant. gnoweb still must not dead-end: it degrades to the native submit and
  warns, as it does today for a missing `sendTx`.
- `not_connected` → `connect()` and retry. Never surfaced to the user.
- Announcements stay untrusted: the existing clamps in `wallet-discovery.ts`
  (16 entries, 64-char names, `data:image/` icons only) apply unchanged to
  every new surface that renders a wallet name.
- A registry entry is server-controlled and validated at init, so it is trusted
  markup-wise; an announced name never is.

## Testing

- **Go**: registry validation including scheme-less entries and dedup; QR
  generation; install page rendering; header rendering in both states.
- **Frontend**: gnoweb has no JS test runner, and adding one remains a larger
  decision than this change should make. Follow the existing ADR's method —
  headless Chrome against the built bundle, on a fixture reproducing the
  `$help` markup, with stub wallets announcing per the spec.

Cases to cover: connect with one wallet (auto-picked) and with several;
disconnect; switch wallet; session restore across reload; Execute connected
(straight to wallet) and unconnected (navigate + fire); `Rejected` and
`not_connected`; a wallet announcing without `connect`; args surviving a
rejection; the install page with zero wallets announced.

## Consequences

- gnoweb gains an identity without gaining a backend.
- The chooser dialog's role shrinks to "two or more wallets, unconnected", and
  the `_centerInVisualViewport` workaround may become removable — to confirm,
  not assume.
- A *connected* desktop user has no route to the QR, since Execute goes
  straight to their extension. The existing "Link" button copies the help URL,
  so it is reachable manually. Left as is.
- Signing while unconnected leaves the user unconnected: `sendTx` returns
  `{hash}`, not an address. gnoweb calls `connect()` after a successful
  unconnected sign to adopt the identity.
- The standard gains a mandatory method, which every future wallet must
  implement. That is the cost of being able to say "log in with any conforming
  wallet" at all.
