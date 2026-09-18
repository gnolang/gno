# ADR: gnoweb Wallet Connect, Identity, Install Page and QR

## Context

Two earlier ADRs on this branch built the transports:
[`gnoconnect_in_page_wallets.md`](./gnoconnect_in_page_wallets.md) (announcement
discovery, `sendTx` on an in-page provider) and
[`pr5970_gnoconnect_external_wallets.md`](./pr5970_gnoconnect_external_wallets.md)
(the registry, the chooser, the `<scheme>://sendtx?…` launch link). Both answer
"which wallet takes this transaction". Neither gives gnoweb a notion of *who the
user is*.

The consequences showed up in the flow. Execute asked which wallet on every
single submit, even for someone with one wallet who had already used it a dozen
times. Nothing on the page ever said which key was about to sign — the address
appeared for the first time inside the wallet's own approval screen. A visitor
with no wallet at all got a chooser listing nothing useful and no route to
getting one. And signing from a phone meant retyping the URL by hand.

This change adds the identity layer on top of the transports, plus the two pages
that make it reachable: `/wallets` to install one, and a QR on `$help` to move a
transaction to a device that has one.

## Decision

### `connect()` becomes a mandatory method of the standard

`docs/resources/gnoconnect.md` listed `connect` under optional methods, on the
reasoning that a wallet gating nothing does not need it. That framing is
backwards for a page that wants to know its user: `connect` is the only way a
page learns an address, and an address is what the header renders, what the
`signer` pin carries, and what the session remembers. A wallet that cannot be
identified cannot be logged into.

Two rules come with it. `connect` on an already-approved origin MUST resolve
without prompting — gnoweb restores the session on every navigation, and a
wallet popup per page load would make the feature unusable. And the `connect`
launch-link host becomes required of an external wallet, because a mobile login
that goes nowhere is indistinguishable, to the producer, from the user
abandoning it.

The cost is real but it is cheap *now*: no wallet implements the in-page half
yet, so promoting it breaks nothing today. It would not stay cheap.

### The session is client-side only

The connected account lives in `localStorage` under `gnoweb:session:v1`, as
`{ rdns, address, chainid, name }`. There is no cookie, no signature challenge,
no server-side session, and no per-user rendering: every page stays byte-identical
for every visitor and fully cacheable.

The session is keyed on `rdns`, never on the per-page-load `uuid`, because it has
to survive a reload. It is display-and-routing state, not authorization: nothing
on the chain trusts it, and the worst a tampered entry can do is address the
wrong wallet, which then declines. Rendering happens immediately from storage —
no flash of "Connect" on every navigation — and reconciles asynchronously
afterwards via `getAccount`, falling back to `connect`.

What connecting buys: the header avatar and address, an Execute that goes
straight to the remembered wallet, and a `signer` pin on every intent.

### gnoweb ships one external wallet in the registry, by choice

`wallets.json` lists gnokey-mobile (app) and Adena (extension). Shipping a
curated list is defensible here on different ground than the old hardcoded
`window.adena` probe was: an app *cannot* announce itself in the page, so there
is no discovery protocol being overridden — the registry is the only mechanism
that exists for that transport. For extensions, the registry entry is a fallback
for one that has not adopted announcements yet, and an announced wallet always
wins over its own registry entry (see `legacyCandidate`).

An entry is server-controlled and validated at package init, so it is trusted as
markup; an announced wallet's `name` is not, and is rendered with `textContent`
and clamped, its `icon` accepted only as a `data:image/` URI.

### One Execute rule, both transports

The rule is: **Execute always ends in a navigation.** Previously an in-page
rejection left the user on an unchanged page with no feedback, and a launch link
that failed to open an app left them staring at nothing at all.

| Situation | What happens |
|---|---|
| Connected, wallet reachable | straight to it, no chooser |
| Unconnected, exactly one candidate | it fires, then the navigation follows |
| Unconnected, two or more | chooser; picking routes, cancel still navigates |
| No candidate at all | native submit (legacy interception intact); its named inputs carry the typed args, which gnoweb folds back into `$help&func=…` with a 303 |
| `Approved` | `?status=success&hash=…#func-<Name>`, adopting the identity if none |
| `Rejected`, or a throw | the function help page, args pinned |
| `not_connected` | `connect`, retry once, never surfaced to the user |

For the in-page transport the order is sign-then-navigate: `sendTx` returns a
promise, and navigating first would destroy it.

For the launch link there is **no navigation at all** — the args are pinned with
`history.replaceState` and the link is fired last. This was originally written
as fire-then-navigate; measuring it on iOS showed that shape never reaches the
wallet (see Validation). Nothing is lost: the destination was this same page
with the args pinned, which `replaceState` reaches without tearing the page
down, and the result comes back through the callback URL.

### The QR is server-rendered on the destination page

The `$help` QR encodes the same URL the Execute form points at, rendered by
`buildHelpQR` into a `data:image/png;base64,…` URI via
`github.com/boombuler/barcode`. It sits in a `:target` panel, and the QR button
is an anchor to that URL's `#qr-<Name>` — so revealing the code *is* a
navigation to the exact args being encoded. A stale code is structurally
impossible, which is what [PR 4602](https://github.com/gnolang/gno/pull/4602)
was blocked on.

## Alternatives considered

- **A server session with a signature challenge.** Rejected: it makes every
  page per-user and uncacheable, needs a backend gnoweb does not have, and buys
  nothing here — gnoweb never acts on the user's behalf, the wallet does.
- **A JS test runner for the frontend.** Rejected: gnoweb has none, and adding
  one is a larger decision than this change should make. A checked-in fixture
  plus a documented checklist covers the same cases reproducibly.
- **A client-side QR encoder.** Rejected: vendoring a JS encoder against a
  server-side one already drafted in PR 4602, and it would need the page to
  re-encode on every keystroke rather than resolving to one URL.
- **User-Agent sniffing to order the install page.** Rejected in favour of a
  fixed server-rendered group order plus a `(pointer: coarse)` CSS `order`
  rule.

## Consequences

- gnoweb gains an identity without gaining a backend.
- The chooser's role shrinks to "two or more wallets, unconnected". It is no
  longer on the common path.
- `centerInVisualViewport` did **not** become removable. The chooser still opens
  as a modal `<dialog>` for the unconnected multi-wallet case, on mobile
  included, so the zoomed-viewport centering still applies. It is untouched by
  this change.
- A *connected* desktop user has no route to the QR, since Execute goes straight
  to their extension. The "Link" button copies the same URL, so it is reachable
  manually. Left as is.
- Signing while unconnected leaves the user unconnected — `sendTx` returns a
  hash, not an address — so gnoweb calls `connect()` after a successful
  unconnected sign to adopt the identity.
- The standard gains a mandatory method every future wallet must implement.
  That is the price of being able to say "log in with any conforming wallet".
- `github.com/boombuler/barcode` is a new dependency of the root module, and
  therefore an indirect one of `contribs/gnodev` and `contribs/gnobro`.

## Deviations from the design spec

1. **The registry gains a `global` field**, which the spec's schema table does
   not list. `rdns` replaces half of the old `LEGACY_PROVIDERS` table — matching
   an entry to an announcement — but legacy detection still needs a `window` key
   to probe. Rather than keep a second hardcoded table, the key moved into the
   registry, validated as a bare JS identifier and permitted only on
   `kind: "extension"`.
2. **"Current platform first" is CSS, not User-Agent sniffing.** The server
   emits a fixed, deterministic group order; a `(pointer: coarse)` media query
   floats the mobile groups to the front with `order`.
3. **The chooser dialog and the registry `<script>` moved into
   `layouts/header.html`.** The header renders on every page and both
   controllers need them; leaving them on `$help` would mean either duplicating
   the registry JSON or a header with no chooser.
4. **Frontend tests are a checked-in fixture plus a checklist**, not a test
   runner — see Alternatives.

## Validation

`gno.land/pkg/gnoweb/frontend/fixtures/wallet-connect.html` plus its README
checklist, run in Chrome against the built bundle. All 16 cases pass. Notable
results:

- Connected Execute goes straight to the remembered wallet and the logged
  `sendTx` carries `signer`, `rpc`, `chainid` and the live form args.
- The `not_connected` stub is refused once, `connect` runs, `sendTx` retries and
  succeeds — none of it surfaced to the user.
- Cancelling the chooser still navigates to the function help page with the
  edited args pinned; a rejected signature lands there too.
- A wallet announcing no `sendTx` degrades to the native submit with a console
  warning; one announcing no `connect` cannot be logged into and persists
  nothing.
- A legacy interceptor alone still intercepts; alongside an announced wallet it
  appears in the chooser and does not fire until picked.

Two bugs were found by running the checklist and fixed in the same branch:

- `/wallets` rendered one word per line: `main > section` is a ten-column grid
  and `.b-wallets` declared no `grid-column`, so the page rendered in one
  column. Fixed in `feature/connect/frontend/connect.css`.
- The fixture as originally drafted could not cover any case that outlives a
  navigation, since a stub clicked after load was not announced when the
  controllers booted. Stubs can now be armed from the URL (`?stub=one`).

A third bug the checklist surfaced, also fixed here: the two paths that fell
through to the *native* submit — "Continue in browser", and a wallet announcing
no `sendTx` — landed on `$help` with the args emptied, because the form's inputs
carry no `name` attribute and a native GET submit rebuilds the query from them.
Both now navigate explicitly to the help URL instead. "Continue in browser"
therefore reaches the same place dismissing the dialog does; the button stays
because it says so out loud.

The one native submit that remains is the no-candidate case, where gnoweb has
nothing to route to and a legacy extension may still want to intercept the
event.

**Superseded later on this branch.** That path no longer keeps `master`'s
behaviour — it is now better than it. The Execute inputs carry `name` attributes,
so the native submit rebuilds the query from the typed values instead of an
empty form data set, and gnoweb folds that query back into its own
`$help&func=…` shape with a 303. Measured in Chrome with every wallet disabled:
editing an argument and submitting lands on the canonical URL with the edit
intact, both with script enabled and with it blocked for the origin. The
sentence above described the bug, which was inherited from `master`, not a
property worth preserving.

The spec flagged two open questions to measure rather than assume:

1. **Does an extension open its approval popup without user activation?**
   Measured `navigator.userActivation.isActive` inside `sendTx`, under a real
   mouse click: `true` on the direct connected path, and `true` again on the
   post-chooser path (picking an entry is itself a gesture). No reordering was
   needed.
2. **Does iOS Safari block a gesture-less custom-scheme navigation?**
   Measured on an iPhone 17 simulator (iOS 26.5) with gnokey-mobile installed,
   driving real taps through `idb`. The answer is more specific than the
   question: iOS does not block the navigation, it gates it behind a system
   **"Open in 'Gnokey'?"** prompt — and *any* navigation issued after firing the
   link dismisses that prompt before the user can answer, so the wallet never
   opens. Four shapes were compared:

   | Shape | Result |
   |---|---|
   | fire, then `location.assign` (as written) | no prompt; page goes to the fallback; app never opens |
   | fire, then `setTimeout(assign, 0)` (the suggested fix) | same — the deferral does not help |
   | fire alone | prompt appears; Open launches the app |
   | `history.replaceState`, then fire | prompt appears; Open launches the app; args pinned |

   The last shape is what shipped. Verified end to end from the fixture's own
   Execute button: the prompt appears, Open brings up gnokey-mobile prefilled
   with the path, function, both arguments and the network from gnoweb's link.
   The plan's proposed remedy would not have worked, which is why it was
   measured rather than assumed.

`go test ./gno.land/pkg/gnoweb/...` passes, as do `make -C gno.land/pkg/gnoweb
lint.go` and `make -C gno.land/pkg/gnoweb/frontend lint`.
