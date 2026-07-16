# ADR: GnoConnect External-Wallet Transport (Registry + Chooser + Launch Link)

## Context

`gnoconnect:*` meta tags and TxLinks already carry the network and transaction
intent on gnoweb's `$help` pages. The in-page JS draft
(`registerWallet`/`getWallets`, issue #2799) covers browser extensions — wallets
that inject code into the page. Neither reaches a wallet that cannot run code
in the page: a mobile app or a standalone desktop signer. In-page discovery is
structurally extension-only (a native app cannot announce itself on `window`),
and a same-domain Execute submit cannot hand off to an external app.

This change adds the missing transport and discovery path for external wallets,
composing with the in-page layer rather than replacing it.

## Decision

### Launch link, host `tx`

On Execute, gnoweb opens the wallet with:

```
<scheme>://tx?path=<pkgPath>&func=<fn>&args=<v1>&args=<v2>&send=<coins>&rpc=<gnoconnect:rpc>&chainid=<gnoconnect:chainid>&callback=<page-url>
```

- No new tx encoding: the exact TxLink fields, plus `rpc`/`chainid` from the
  existing `gnoconnect:*` meta so the wallet targets the right node (including
  localnet).
- Host is `tx`, not `call`: it names the intent and leaves room for future
  hosts (`run`, `sign`) under the same scheme. Wallets should accept `call` as
  a silent back-compat alias but emit/document only `tx`.
- `args` repeats once per positional parameter, in declaration order, every
  value percent-encoded; empty values are included to keep positions aligned.
- Args are read live from the form inputs at submit time — never from the
  address bar, which goes stale between edits and submit.
- `callback` is the current page URL (minus any `status`/`hash` params left by
  a previous wallet round trip, so they don't accumulate). The wallet reopens
  exactly where the user was, appending `status` and tx `hash`. Works on
  localhost.
- RPC normalization is wallet-side: gnoweb forwards `gnoconnect:rpc` verbatim
  (it may be scheme-less like `127.0.0.1:26657`); the wallet assumes `http://`
  when the scheme is missing.

### Registry: in-repo, server-embedded

`components/wallets.json` is embedded via `//go:embed`, parsed and re-marshaled
once at package init (marshal, not the raw file, because `json.Marshal`
HTML-escapes), and exposed to the frontend through a
`<script type="application/json">` tag. Entry shape:

```json
{ "name": "Gnokey", "id": "gnokey", "icon": "data:image/svg+xml;base64,…",
  "scheme": "land.gno.gnokey", "platforms": ["ios","android"],
  "install_url": "https://github.com/gnolang/gnokey-mobile" }
```

- The registry stores the bare scheme (`land.gno.gnokey`, not
  `land.gno.gnokey://tx`); gnoweb composes the launch prefix, so the registry
  stays reusable when the standard grows other hosts.
- Embedding (vs a served endpoint or on-chain catalog) is decisive for gnoweb:
  it renders offline, needs no fetch, and is cache-safe — the localnet
  requirement. A served/on-chain catalog can come later without changing the
  entry shape. Wallets register by PR.
- `platforms` and `install_url` are informational for now: no platform
  filtering and no "not installed" fallback yet.

### Selection: gnoweb, never the OS

A duplicated custom scheme on iOS routes to one arbitrary app with no system
chooser, so the OS cannot pick. gnoweb merges candidates and shows its own
chooser `<dialog>`, which also makes desktop and mobile behave identically.
One wallet → launch directly; more than one → chooser; none → today's
copy-paste TxLink, untouched.

### Additive and fail-safe by construction

`preventDefault()` fires only when actually routing to an external wallet.
The controller falls through to the native submit when: an in-page provider
(extension) is present, the device is not mobile, or the registry is
empty/malformed. Mobile detection is `(pointer: coarse)` only —
`maxTouchPoints` would also match touchscreen laptops, where a failed
custom-scheme launch would leave Execute dead.

### Custom schemes over Universal/App Links

Origin-independent, work offline/localhost, need no AASA/assetlinks/CDN.
UL/AL are at most an optional "not installed" fallback, never the primary path.

## Alternatives considered

- **OS-level chooser** — impossible: iOS resolves duplicate schemes to one app.
- **Served or on-chain registry** — breaks offline/localnet rendering; can be
  added later behind the same entry shape.
- **Universal/App Links as primary transport** — origin-dependent, needs CDN
  infrastructure, fails on localhost.
- **Reading args back from the TxLink URL** — stale between input edits and
  submit; live form values are authoritative.
- **Subscribing to `params:changed` events for arg state** — rejected during
  review: reading the DOM at submit time is always fresh and reads the same
  inputs, so the event subscription was pure duplication.

## Consequences

- Mobile users with a registered wallet get a native sign-and-broadcast flow;
  everyone else sees no behavior change.
- Wallet authors must implement the `<scheme>://tx?...` link format above.
- The registry gates wallet listing on gnoweb PRs until a served/on-chain
  catalog exists.
- Follow-ups (not gaps): QR for desktop→phone (the launch link is already the
  payload), merging in-page announced extensions into the same chooser
  (blocked on the `registerWallet`/`getWallets` draft), and normalizing
  gnodev's scheme-less `gnoconnect:rpc` meta (separate PR).

## Validation

Full round trip validated against a local `gnodev` with gnokey-mobile
(`feat/deeplink-tx` branch) on the iOS simulator and Android emulator:
Execute → chooser → wallet opens prefilled → sign & broadcast → `callback`
reopens gnoweb with `?status=success&hash=…`. Multi-arg ordering exercised via
`r/demo/profile.SetStringField(field, value)` under a scoped session
allow-path, confirmed on-chain.
