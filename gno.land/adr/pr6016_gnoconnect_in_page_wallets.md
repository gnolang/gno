# ADR: GnoConnect In-Page Wallet Discovery (announce protocol + merged chooser)

## Context

PR 5970 added the external half of GnoConnect: a server-embedded registry of
native wallets and a launch link (`<scheme>://sendtx?…`) that opens one. It
deliberately left the in-page half alone — on Execute it checks for
`window.adena` and, if present, does nothing, letting the extension intercept
the submit as it does today.

That check is the problem this change removes. A single well-known global is a
namespace race: whichever extension writes it last wins, and a user with two
wallets installed cannot reach one of them. It is also unusable as discovery —
gnoweb learns that *an* extension exists, not which, nor what it can do, nor
whether the user wants it for this transaction.

The global is only half of how a wallet takes a transaction today. Adena also
claims the Execute submit outright, by scraping gnoweb's markup from a
document-capture listener and cancelling the event. That is the same problem in
a second form — first one to the event wins, the page is never asked — and it
has to be addressed here too, or discovery has nothing to discover with.

The GnoConnect standard (#6373) specifies the in-page protocol: the announce
handshake, the provider, the rule against consuming the page's events, and the
no-script exception to it. Its decisions are recorded in
[`pr6373_gnoconnect_standard.md`](./pr6373_gnoconnect_standard.md). This ADR
covers how gnoweb implements the page side, and how it stays compatible with
the wallets installed today.

## Decision

### The page side of the handshake

gnoweb implements the page side of `docs/resources/gnoconnect.md` § "In-Page
Wallets": it listens for `gno:registerWallet` before it dispatches
`gno:requestWallet`, keeps listening afterwards, and dispatches the request on
load rather than when the chooser opens (§ Pages that never ask makes that a
MUST). Announcements are deduplicated by `rdns`, falling back to `uuid`, and
`rdns` is what gnoweb persists when it remembers a wallet.

The only provider methods gnoweb calls are `sendTx` and `connect`.

- **One intent, two transports.** The in-page call and the launch link are
  built from the same `GnoTxRequest`, produced by one function reading the live
  form. Divergence between the two paths would be invisible until a wallet
  bound an argument to the wrong parameter.
- **Announcing is not a claim to implement everything.** A wallet that
  announces without `sendTx` falls back to the native submit,
  so its legacy interception still works and Execute is never dead-ended.

### One chooser, both kinds

Announced wallets and registry entries are merged into the existing dialog,
each labelled with its transport (Extension / App). In-page wallets are offered
on every device; registry wallets stay mobile-only until the cross-device QR
lands. The dialog re-renders on announcements that arrive while it is open.

- **The user chooses, not the load order.** Two extensions and a mobile app are
  three entries in one list — the outcome the announce protocol exists for.
- **A legacy extension is still honoured, but only when nothing announced.** A
  wallet that speaks the protocol is picked explicitly; an extension that does
  not owns the submit exactly as before, one tap fewer. That ordering is what
  keeps this change additive for today's Adena users.

### Claiming the submit at window-capture

The chooser listens for `submit` on `window`, capture phase, and calls
`preventDefault()` + `stopPropagation()` once it has a candidate.

This is not a stylistic choice. Adena 1.20.1 claims the Execute submit by
scraping the page: a `document.addEventListener("submit", h, true)` matching
`article.b-action-function > form.params`, which cancels the event and reads
`data-action-function-*` off the markup (it registers whenever the page carries
the `gnoconnect:*` metas, which gnoweb emits on every page). A listener on the
`<form>` — where this controller previously sat — is downstream of that: with
Adena installed, the chooser never opened and the announce protocol was
unreachable. Capture reaches `window` before `document` whatever order the
listeners were added in, so it is the one place the page is certain to be asked.

- **`stopPropagation`, not `stopImmediatePropagation`.** Other window-capture
  listeners are not competing for the event; only the downstream interception
  is. Analytics' `submit_action` moves to window-capture alongside, upstream of
  the claim, so the metric survives the chooser path.
- **Only with a candidate.** With nothing announced and no registry entry the
  handler returns before cancelling anything, so a legacy extension keeps
  intercepting exactly as it does today.
- **The legacy extension is listed, and reached by giving the event back.** It
  exposes nothing to call and acts only on the submit the page has just taken,
  so picking it re-dispatches through `requestSubmit()` behind a flag the
  handler honours once. Without that, a user with one announcing wallet and
  Adena installed could no longer reach Adena at all — "Continue in browser"
  calls `form.submit()`, which fires no event either.
- **This is a transition mechanism, and it lives here, not in the standard.**
  The standard forbids a wallet to cancel or consume the page's events and says
  nothing about deferring to one that does. "Keep working with the extensions
  installed today" is a gnoweb rollout decision with an expiry date; in the
  standard, every future implementer would carry our migration forever. Once
  wallets announce, the legacy entry and the claim both come out — with one
  carve-out: see "With page JavaScript disabled" below.

### With page JavaScript disabled

Adena's interception is registered by its **content script**, in the isolated
world, and gated on the `gnoconnect:*` metas — which are server-rendered markup.
Neither depends on the page running any script of its own. Measured against
Adena 1.21.1 with JavaScript blocked for the origin: Execute is still cancelled
and the wallet still opens.

Everything on gnoweb's side of this ADR, by contrast, is page script. With page
JS off there is no window-capture claim, no announcement handling, and no
chooser; gnoweb is reduced to the HTML and CSS it served. So:

- **The legacy interception outlives the protocol meant to replace it.** For a
  no-JS visitor it is the only in-page path that works at all.
- **A no-JS page cannot arbitrate.** Choosing between wallets is page code, so
  with two intercepting wallets installed the first to register wins — exactly
  the race the announce protocol exists to remove. This is not fixable by any
  protocol; it is a property of the page being unable to run.

**The decision is to keep that.** Without page script, gnoweb falls back to
`master`'s flow: it claims nothing, and a wallet that intercepts on its own
takes the submit exactly as it does today. This costs no code — it is what
happens already, because every mechanism this ADR adds is page script and simply
does not register. The standard sanctions it (§ Pages that never ask: a wallet
may intercept a TxLink navigation only on a page that has not dispatched
`gno:requestWallet`), and gnoweb dispatches that request on load, so the two
paths never compete. What it costs gnoweb is a commitment, stated here so it is
not broken by accident:

- The markup an intercepting wallet scrapes — `article.b-action-function >
  form.params` and the `data-action-function-*` attributes — is a **supported
  surface for the no-script path**, not merely an implementation detail. It may
  not be restructured without checking case 22 in
  [`frontend/fixtures/README.md`](../pkg/gnoweb/frontend/fixtures/README.md).
- The Execute form's `name` attributes are a second such surface, for the visitor
  with no wallet at all: the param inputs plus the hidden `func` input are what a
  native submit sends, and `canonicalHelpURL` in `handler_http.go` keys off `func`
  to fold that query back into `$help&k=v`. The send checkbox is deliberately
  *not* named, so a submit nobody could confirm never adds coins. Removing a
  `name=` silently returns those users to a discarded-argument Execute; cases
  20–21c cover it. The `+`→`%20` spelling is agreed in two places —
  `escapeWebArg` in `components/view_action.go` and the raw-query fold in
  `canonicalHelpURL` — and the two must not diverge.
- The registry's legacy entry and its `global` probe outlive the chooser's need
  for them. The retirement above applies to the scripted path only.

This change already satisfies the commitment: `views/action.html` adds an inert
`data-controller` to the form and a sibling QR `<div>`, so the selector and the
scraped attributes are untouched — verified against Adena 1.21.1 with script
blocked, where Execute behaves as it does on `master`.

In-page signing by a desktop extension with script disabled therefore works, and
works for exactly one wallet — the first to register. That is a property of the
page being unable to run, not a choice, and is the honest boundary of this
fallback.

### Announcements are untrusted input

Any script in the page can announce (§ Announcements are untrusted). The
chooser renders `name` as text with a 64-character clamp, accepts `icon` only as
a `data:image/` URI (a remote URL would leak the page visit and render fetched
content), and caps the list at 16 entries so a flood cannot push the real
wallet out of view. What the page owes the user is a list that is legible and
cannot be crowded out; picking from it is their trust decision.

### Following the wallet's events

The session is remembered in `localStorage`, so without being told, the header
keeps showing an account the user has since switched away from, or a site they
revoked in the wallet. `followWallet` (`feature/connect/frontend/session.ts`)
subscribes to the remembered wallet's `provider.on` (`gnoconnect.md` § Events)
and keeps the session in step:

- `accountChanged` rewrites the session with the new address and drops the
  stored username, which belonged to the old one.
- `networkChanged` records the new `chainid`, which decides whether the header's
  chain reads can run against this page's RPC.
- `disconnect` clears the session, without calling `connect`.

It only listens to the wallet the session came from, subscribes once per wallet
(announcements repeat), resubscribes when the session changes, and waits for the
wallet to announce when the session was restored before it did. An event about
another wallet, or one that changes nothing, is ignored.

These events only change what the header shows. What gets signed is still
protected by the `signer` pin on every intent: a page that missed an event has its
request declined, not signed as the wrong identity.

## Alternatives considered

Alternatives to the protocol itself (`window.<wallet>` globals, a `window.gno`
array, the Wallet Standard handshake verbatim, a liveness marker sanctioning
interception) are in the standard's ADR.

- **Rejecting a wallet whose icon is not a `data:` URI** — makes an otherwise
  working wallet unreachable over a presentational flaw; the icon is dropped
  and the entry kept instead.
- **Leaving the listener on the form and documenting the limitation** — honest,
  but it means the chooser never appears for anyone with Adena installed, which
  is most of today's users. The feature would ship unreachable.
- **Not listing the legacy extension, accepting that it becomes unreachable
  once another wallet announces** — self-resolving as wallets adopt the
  protocol, but it takes a working wallet away from a user mid-transition, for
  the benefit of a wallet they may not have chosen.
- **Rendering the result in place after an in-page signature** — the external
  transport already lands on `?status=success&hash=…` via the callback, so the
  in-page path navigates to the same URL shape. One result surface for both
  transports, to be built once (see Consequences).

## Consequences

- A user with several wallets installed can pick between them; a user with one
  sees the same one-tap flow.
- Wallet extensions must announce to appear in the chooser on their own terms.
  Until they do, they keep working: alone, through their existing interception;
  alongside an announcing wallet, as a listed entry that is handed the submit.
  No wallet becomes unreachable.
- A wallet that intercepts now shares the event with the page instead of owning
  it. For a user with one wallet installed nothing changes — same one tap, same
  extension. For a user with two, the page asks.
- Rendering the outcome (`status`/`hash`) is still a follow-up. Both transports
  now converge on one URL shape, so it is one surface to build, not two.
- Everything this ADR adds is page script, so a visitor with JavaScript disabled
  falls back to `master`'s flow by design: an intercepting wallet takes the
  submit as it does today. That path is single-wallet by construction and cannot
  be arbitrated, and it makes the scraped Execute markup a surface this repo has
  agreed to keep stable — see "With page JavaScript disabled".

## Validation

Headless Chrome against the built bundle, on a fixture reproducing the `$help`
Execute markup, with stub wallets announcing per the spec. All cases pass:

| Case | Result |
|---|---|
| Wallet announcing only on `gno:requestWallet` | discovered and listed |
| Two stubs + the registry entry (coarse pointer) | 3 entries, in order, labelled Extension/Extension/App |
| Chosen wallet's `sendTx` | called with the live form values — named args, `rpc`, `chainid` |
| `Approved` | lands on `?status=success&hash=…`, the external callback's URL shape |
| `Rejected` / provider throws | page untouched, nothing submitted |
| Wallet announcing without the method | falls through to the native submit |
| `window.adena`, nothing announced | native submit, today's behaviour unchanged |
| `window.adena` **and** an announcement | chooser wins; the user picks |
| 22 announcements, a 200-char name, a remote-URL icon, 3 malformed | 16 kept, name clamped to 64, icon dropped, malformed ignored |
| A document-capture interceptor, nothing announced | it still intercepts, once; no chooser |
| Interceptor **and** an announcement | chooser opens with both entries; the interceptor does not fire |
| Picking the announced wallet | wallet called; the interceptor never fires |
| Picking the legacy entry | the interceptor fires once; no navigation, no wallet call |

Also run against Adena 1.20.1 itself, on the same fixture: with a stub wallet
announcing, the chooser opens (before this change Adena consumed the submit),
and picking the "Adena" entry hands the event back — Adena cancels it, so the
page stays put and the stub is not called.

`go test ./gno.land/pkg/gnoweb/...` green.

### Re-run against Adena 1.21.1 from the Chrome Web Store

The rows above were taken against Adena 1.20.1. Re-measured against the store
build 1.21.1, on `gnodev` serving a two-argument realm, with every other wallet
disabled:

| Case | Result |
|---|---|
| The build's own code | `content.js` carries the document-capture submit listener matching `b-action-function` / `data-action-function`; no `gno:registerWallet` anywhere. The ADR's description holds for the shipped wallet. |
| `window.adena` | legacy surface only — `DoContract`, `SignTx`, `GetAccount`. **No `sendTx`**, so handing back the submit is the only way to reach it, as designed. |
| Adena alone, script enabled | no chooser, no navigation: one legacy candidate, fired directly. |
| Adena + one announcing stub | **chooser opens with both entries**, labelled Extension. The window-capture claim beats Adena's document-capture listener, which is the whole point of this ADR — verified against the real extension rather than a stub interceptor. |
| Picking the stub | `sendTx` receives the **live** form values (an argument edited after load, not the one in the URL), with `rpc` and `chainid`, and no `signer` while unconnected. |
| The stub returning `Rejected` | lands on the function help page with the edited arguments pinned. |
| Adena, script **disabled** for the origin | Execute still cancelled, wallet still opens — see "With page JavaScript disabled". |

Now covered: a real wallet extension, on both paths. Not covered: a real wallet
implementing the announce protocol. A development build that does exists
(`gno:registerWallet`, `sendTx`, `getAccount`, and the spec's error codes), but
it is not a shipped wallet and was disabled for these measurements.
