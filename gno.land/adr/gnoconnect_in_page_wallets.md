# ADR: GnoConnect In-Page Wallet Discovery (announce protocol + merged chooser)

## Context

PR 5970 added the external half of GnoConnect: a server-embedded registry of
native wallets and a launch link (`<scheme>://tx?…`) that opens one. It
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

The only method gnoweb calls is `signAndSubmitTransaction(tx)`, whose argument
is the `tx` launch link field for field — named args included — returning
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
  announces without `signAndSubmitTransaction` falls back to the native submit,
  so its legacy interception still works and Execute is never dead-ended.

### One chooser, both kinds

Announced wallets and registry entries are merged into the existing dialog,
each labelled with its transport (Extension / App). In-page wallets are offered
on every device; registry wallets stay mobile-only until the cross-device QR
lands. The dialog re-renders on announcements that arrive while it is open.

- **The user chooses, not the load order.** Two extensions and a mobile app are
  three entries in one list — the outcome the announce protocol exists for.
- **`window.adena` is still honoured, but only when nothing announced.** A
  wallet that speaks the protocol is picked explicitly; an extension that does
  not still owns the submit exactly as before. That ordering is what keeps this
  change additive for today's Adena users.

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
- **Rendering the result in place after an in-page signature** — the external
  transport already lands on `?status=success&hash=…` via the callback, so the
  in-page path navigates to the same URL shape. One result surface for both
  transports, to be built once (see Consequences).

## Consequences

- A user with several wallets installed can pick between them; a user with one
  sees the same one-tap flow.
- Wallet extensions must announce to appear in the chooser. Until they do, they
  keep working through their existing submit interception — no wallet breaks on
  this change.
- Rendering the outcome (`status`/`hash`) is still a follow-up. Both transports
  now converge on one URL shape, so it is one surface to build, not two.
- The in-page half has no automated test: gnoweb's frontend has no JS test
  runner, and adding one is a larger decision than this change should make.
  Verified in-browser against a stub wallet that announces per the spec.

## Validation

Headless Chrome against the built bundle, on a fixture reproducing the `$help`
Execute markup, with stub wallets announcing per the spec. Nine cases, all
passing:

| Case | Result |
|---|---|
| Wallet announcing only on `gno:requestWallet` | discovered and listed |
| Two stubs + the registry entry (coarse pointer) | 3 entries, in order, labelled Extension/Extension/App |
| Chosen wallet's `signAndSubmitTransaction` | called with the live form values — named args, `rpc`, `chainid` |
| `Approved` | lands on `?status=success&hash=…`, the external callback's URL shape |
| `Rejected` / provider throws | page untouched, nothing submitted |
| Wallet announcing without the method | falls through to the native submit |
| `window.adena`, nothing announced | native submit, today's behaviour unchanged |
| `window.adena` **and** an announcement | chooser wins; the user picks |
| 22 announcements, a 200-char name, a remote-URL icon, 3 malformed | 16 kept, name clamped to 64, icon dropped, malformed ignored |

`go test ./gno.land/pkg/gnoweb/...` green (the Go side is untouched).

Not covered: a real wallet extension. No such extension announces yet — that
is what this change is for — so what is proven is that a conforming announcer
is discovered, merged, called, and that a non-conforming page cannot dead-end
Execute.
