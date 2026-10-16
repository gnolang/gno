# ADR: GnoConnect External-Wallet Transport (Registry + Chooser + Launch Link)

## Context

`gnoconnect:*` meta tags and TxLinks already carry the network and transaction
intent on gnoweb's `$help` pages. Neither reaches a wallet that cannot run code
in the page: a mobile app or a standalone desktop signer. A same-domain Execute
submit cannot hand off to an external app.

The GnoConnect standard (#6373) defines the transport for those wallets, the
launch link. Its protocol decisions — custom schemes, named `arg.<name>`
params, hosts as verbs, callback rules — are recorded in
[`pr6373_gnoconnect_standard.md`](./pr6373_gnoconnect_standard.md). This ADR
covers how gnoweb produces launch links: which wallets it offers, how it picks
one, and how it builds the link from the Execute form.

## Decision

### Building the link from the Execute form

On Execute, gnoweb opens the chosen wallet with a launch link
(`docs/resources/gnoconnect.md` § Launch Links). gnoweb's side of it:

- Args are read live from the form inputs at submit time — never from the
  address bar, which goes stale between edits and submit.
- `rpc` and `chainid` come from the page's `gnoconnect:*` metas, so the wallet
  targets the right network, localnet included. gnoweb forwards
  `gnoconnect:rpc` verbatim, even scheme-less (`127.0.0.1:26657`).
- `callback` is the current page URL, minus any `status`/`hash` params left by
  a previous wallet round trip so they don't accumulate. The wallet reopens
  exactly where the user was. Works on localhost.

This branch emits the host as `tx`, its name when this PR was opened. The
standard has since renamed it `sendtx`; #6016 moves gnoweb to it.

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
  stays reusable as the standard adds hosts.
- Embedding (vs a served endpoint or on-chain catalog) is decisive for gnoweb:
  it renders offline, needs no fetch, and is cache-safe — the localnet
  requirement. A served/on-chain catalog can come later without changing the
  entry shape. Wallets register by PR.
- `platforms` and `install_url` are informational for now: no platform
  filtering and no "not installed" fallback yet.
- Entries are validated at package init — RFC 3986 scheme grammar (which
  cannot contain `:`, so a registry entry can never smuggle a `javascript:`
  payload into `window.location`), unique ids and schemes, `data:image/`
  icons. A bad registry is an authoring error and panics.

### Selection: gnoweb, never the OS

A duplicated custom scheme on iOS routes to one arbitrary app with no system
chooser, so the OS cannot pick. gnoweb merges candidates and shows its own
chooser `<dialog>`, which also makes desktop and mobile behave identically.
The chooser appears whenever at least one wallet is registered — even for a
single wallet — because it carries a "Continue in browser" action that
performs the native submit. A failed custom-scheme launch is silent, so
without that escape hatch a mobile user without the wallet installed would
find Execute dead-ended. No registered wallet → today's copy-paste TxLink,
untouched.

### Additive and fail-safe by construction

`preventDefault()` fires only when actually routing to an external wallet.
The controller falls through to the native submit when: an in-page provider
(extension) is present, the device is not mobile, or the registry is
empty/malformed. Mobile detection is `(pointer: coarse)` only —
`maxTouchPoints` would also match touchscreen laptops, where a failed
custom-scheme launch would leave Execute dead.

## Alternatives considered

Alternatives to the link format itself (positional args, flat names, Universal
/ App Links) are in the standard's ADR.

- **OS-level chooser** — impossible: iOS resolves duplicate schemes to one app.
- **Served or on-chain registry** — breaks offline/localnet rendering; can be
  added later behind the same entry shape.
- **Reading args back from the TxLink URL** — stale between input edits and
  submit; live form values are authoritative.
- **Subscribing to `params:changed` events for arg state** — rejected during
  review: reading the DOM at submit time is always fresh and reads the same
  inputs, so the event subscription was pure duplication.

## Consequences

- Mobile users with a registered wallet get a native sign-and-broadcast flow;
  everyone else sees no behavior change.
- The registry gates wallet listing on gnoweb PRs until a served/on-chain
  catalog exists.
- Follow-ups (not gaps): QR for desktop→phone (the launch link is already the
  payload), merging in-page announced extensions into the same chooser (#6016),
  and normalizing gnodev's scheme-less `gnoconnect:rpc` meta (separate PR).

## Validation

Full round trip validated against a local `gnodev` with gnokey-mobile
(`feat/deeplink-tx` branch) on the iOS simulator and Android emulator:
Execute → chooser → wallet opens prefilled → sign & broadcast → `callback`
reopens gnoweb with `?status=success&hash=…`. Multi-arg ordering exercised via
`r/demo/profile.SetStringField(field, value)` under a scoped session
allow-path, confirmed on-chain.

That round trip used the original positional `args=` encoding; the switch to
named `arg.<name>=` params (post-review) still needs an end-to-end pass with
the updated wallet branch.
