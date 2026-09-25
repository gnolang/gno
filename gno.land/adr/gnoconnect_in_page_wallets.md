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
whether the user wants it for this transaction. EIP-6963 and the Wallet
Standard both exist because the Ethereum and Solana ecosystems hit exactly this
wall; the draft in issue #2799 (`registerWallet` / `getWallets`) points at the
same answer for Gno.

The global is only half of how a wallet takes a transaction today. Adena also
claims the Execute submit outright, by scraping gnoweb's markup from a
document-capture listener and cancelling the event. That is the same problem in
a second form — first one to the event wins, the page is never asked — and it
has to be addressed here too, or discovery has nothing to discover with.

## Decision

### Announce protocol: two events on `window`

`gno:registerWallet` (wallet → page, `detail: { info, provider }`) and
`gno:requestWallet` (page → wallets, no payload). Wallets announce on load and
on every request; the page listens before it asks and keeps listening after.
Specified in `docs/resources/gnoconnect.md` § "In-Page Wallets".

- **On `window`, not `document`** — an extension announces at `document_start`,
  before any page script exists; `window` is the one surface both sides can
  rely on at that point.
- **Both directions are required.** Either alone loses to a race: announce-only
  misses pages that start listening late, request-only misses wallets that
  load late. This is the single most-repeated implementation mistake in
  EIP-6963, so the spec states it as an obligation on each side rather than as
  advice.
- **`rdns` alongside `uuid`** — `uuid` is a per-page-load handle for
  deduplication; `rdns` is what identifies a wallet across loads and versions,
  and therefore what a page persists when it remembers a choice. A display
  name cannot serve: names are neither unique nor stable.

### Provider surface: mirror the launch link, feature-detect the rest

The only method gnoweb calls is `sendTx(tx)`, whose argument
is the `sendtx` launch link field for field — named args included — returning
`UserResponse<{hash}>`. Wallets may implement more of the draft's surface
(`connect`, `signMessage`, network switching); a page must feature-detect every
method it calls.

- **One intent, two transports.** The in-page call and the launch link
  serialize the same `GnoTxRequest`, built by one function reading the live
  form. Divergence between the two paths would be invisible until a wallet
  bound an argument to the wrong parameter.
- **`Rejected` is a status, not an exception.** A user declining is an answer;
  reserving rejection of the promise for genuine failures keeps "the user said
  no" from being reported as an error.
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
  `docs/resources/gnoconnect.md` states that a wallet MUST NOT cancel or consume
  the page's events; it deliberately says nothing about deferring to one that
  does. That rule is permanent and wallet-agnostic, whereas "keep working with
  the extensions installed today" is a gnoweb rollout decision with an expiry
  date — putting it in the standard would have every future implementer carry
  our migration forever. Once wallets announce, the legacy entry and the claim
  both stop mattering and both come out — with one carve-out, measured after
  this was first written: see "With page JavaScript disabled" below.

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
does not register. What it costs is a commitment, stated here so it is not
broken by accident:

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

What the fallback does **not** cover is a visitor with no such wallet: for them
Execute reaches the native submit, which discards the typed arguments (see
[`prxxxx_gnoweb_wallet_connect.md`](./prxxxx_gnoweb_wallet_connect.md) §
Deviations — the same bug on `master`). An intercepting wallet masks it by
cancelling the submit before the browser acts on it; nothing else does.

Two further options were considered and **not** adopted, both of which remain
compatible with the decision above rather than alternatives to it:

- **Give the Execute inputs `name` attributes**, so the native submit navigates
  to a URL that fully encodes the intent instead of emptying it. This does not
  interact with the fallback — an intercepting wallet cancels the event before
  the browser ever reads the form — so it only repairs the no-wallet case, and
  additionally lets a wallet reach the intent by **observing** the navigation,
  which `gnoconnect.md` permits ("Observing is not intercepting"). Deferred as
  a separate change, since it alters a URL shape `master` also produces.
- **Sanction interception conditionally in the standard**, via a liveness marker
  gnoweb sets on boot. Unnecessary: with script disabled gnoweb registers
  nothing, so the wallet is unopposed without needing to be told. Adopting it
  would put a gnoweb-specific migration into a wallet-agnostic standard, which
  the bullet above this section rejects on its own terms.

In-page signing by a desktop extension with script disabled therefore works, and
works for exactly one wallet — the first to register. That is a property of the
page being unable to run, not a choice, and is the honest boundary of this
fallback.

### Announcements are untrusted input

Any script in the page can announce. The chooser renders `name` as text with a
64-character clamp, accepts `icon` only as a `data:image/` URI (a remote URL
would leak the page visit and render fetched content), and caps the list at 16
entries so a flood cannot push the real wallet out of view.

This does not authenticate anything, and is not meant to: picking a name from a
list is the user's trust decision, as it was when they installed the extension.
What the page owes them is a list that is legible and cannot be crowded out.

## Alternatives considered

- **Keep the `window.adena` check** — works only for one wallet, by name, and
  is what the draft standard exists to replace.
- **A `window.gno` namespace object with a wallet array** — still a shared
  mutable global: extensions overwrite each other's writes, and load order
  decides who is visible.
- **Adopting the Wallet Standard's registration callback handshake verbatim**
  (`app-ready` carrying a `register` function) — richer, but it makes the
  protocol depend on a shared library implementing it. gnoweb has no npm
  dependency for this and must work offline; announce/request is enough to
  merge into a chooser.
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
- The in-page half has no automated test: gnoweb's frontend has no JS test
  runner, and adding one is a larger decision than this change should make.
  Verified in-browser against a stub wallet that announces per the spec.
- Everything this ADR adds is page script, so a visitor with JavaScript disabled
  falls back to `master`'s flow by design: an intercepting wallet takes the
  submit as it does today. That path is single-wallet by construction and cannot
  be arbitrated, and it makes the scraped Execute markup a surface this repo has
  agreed to keep stable — see "With page JavaScript disabled".

## Validation

Headless Chrome against the built bundle, on a fixture reproducing the `$help`
Execute markup, with stub wallets announcing per the spec. Nine cases, all
passing:

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

`go test ./gno.land/pkg/gnoweb/...` green (the Go side is untouched).

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
(`gno:registerWallet`, `sendTx`, `getAccount`, and the spec's error codes), so
the earlier claim that no such extension exists is no longer true, but it is not
a shipped wallet and was disabled for these measurements.
