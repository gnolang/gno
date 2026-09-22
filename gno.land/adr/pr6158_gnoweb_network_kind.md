# ADR-6158: gnoweb network kind

## Context

gnoweb is a single binary serving several chains under the gno.land name:
mainnet (`gnoland-1`, at gno.land), the current testnet (`pearl-1`), staging,
and `dev` under gnodev.

Nothing in the UI tells them apart. The only signal is the chain-id inside the
Network Info popup, behind a click, so a user cannot know which chain they are
acting on without opening it. With real value on mainnet and several public
deployments answering under the same branding, that is a footgun.

The footer made this concrete: it rendered `https://faucet.gno.land/`
unconditionally, on every network, including deployments that have no faucet at
all — gnodev has none, and the mainnet genesis funds none
(`misc/deployments/mainnet.gno.land/README.md`: **"No faucets."**) — while the
hub it points to lists testnet dispensers.

This is item §2 of #6121, plus §3.5 and part of §3.3.

## Decision

**One server-side source of truth, set by the operator.** `components.NetworkKind`
comes from `-network-kind` and is carried `AppConfig` → `StaticMetadata` →
`IndexData` → `data-network` on `<html>`, next to the existing `data-theme`.
An invalid value is a startup error rather than a silent fallback.

It is deliberately NOT derived from the chain-id. An earlier revision matched
the `gnoland-N` series, but that encodes a chain-naming assumption into gnoweb
that the naming scheme does not promise to keep — aeddi, reviewing this PR:
`gnoland-1` will not necessarily be the only mainnet, the number may be
incremented later under certain conditions. Instead the default is **testnet**
and mainnet is explicit: a mainnet that
forgets the flag shows the alert chip, the safe direction, visibly and
immediately; a testnet can only present as mainnet through explicit
misconfiguration, which no derivation prevents either (the override existed).

**The chip is rendered on every network, mainnet included.** Marking only the
testnets would make the signal an absence, and an absence is unreadable — a
user who has never seen the chip cannot tell "this is mainnet" from "this build
predates the chip". The chip names the chain on every page and escalates to a
coloured `--alert` variant off-mainnet. It sits next to the logo and does not
replace the Network Info popup, which keeps the full remote/chain-id detail.

**The faucet link is gated on `FaucetURL != ""`, but does not use its value.**
The config field already decides whether `/faucet` is routed, so it is the
authority on whether this deployment *has* a faucet — and gating on it means a
deployment without one configures nothing. A network-kind conditional would
have been a second source of truth for the same fact, and the wrong fact at
that: the mainnet deployment sets `-faucet-url` deliberately, the hub being due
to serve mainnet distribution behind a form (#6121 §3.5).

Its *value* is deliberately not used as the link target. `-faucet-url` holds
the endpoint the `/faucet` route redirects to, and the deployments disagree on
what that is: the in-repo staging and home-alias compose files point it at a
`faucet-api.*` host while their footers point at the hub. Feeding it into an
`href` would have offered those users an API. The footer keeps the hub URL as a
constant; only the presence of the link is configuration-driven.

This narrows the footer only. `-faucet-url` still reaches an `href` on the
`/faucet` interstitial (`components/views/redirect.html` renders it as the
canonical link, the meta-refresh target and a visible anchor), so that page
still offers a click through to whatever the flag holds. Pre-existing and out of
scope here, but the two surfaces now disagree by design rather than by accident.

**Colour is secondary.** Only `--s-logo-hat` moves under
`[data-network="testnet"]` — layout, contrast and dark mode are untouched, and
the chip carries the text. `--s-color-text-brand-default` is left alone: it is
the text *on* brand surfaces, so tinting it would break button contrast.

Two absolute `https://gno.land/...` links (header About, footer Blog) are made
relative in passing: on any non-mainnet deployment they silently moved the user
to mainnet.

### The chip names the kind, and a banner says it in a sentence

A chip showing `pearl-1` and nothing else only helps a reader who already knows
the chain-id list. The visible text is therefore `<chain-id> <kind>` on every
network, mainnet included: a signal that exists only off mainnet makes mainnet
an absence, and an absence is what a lookalike site produces for free. The kind
word is a sibling of the chain-id rather than part of it, so the `14ch` cap that
keeps an unbounded operator value from eating the header does not truncate a
fixed word.

Off mainnet, `NewRouter` also fills the existing `.b-banner` with one sentence:
the negation first, because "not mainnet" needs no prior knowledge, then what it
costs the reader. An operator-configured banner wins, since a deployment that
set one has something more specific to say. The RPC address is left out: it is
in the Network Info popup with a label, and it means nothing to a visitor
reading a realm.

### Three kinds, one colour each

`NetworkKind` has three values rather than two: `mainnet`, `testnet` and
`local`, the last one set by gnodev itself so no operator has to. Each gets one
colour carried by `data-network`, and off mainnet that colour reaches the logo
hat, the chip, the banner, the links and the text selection, so the signal is
one thing rather than an amber banner over a green page.

Measured, foreground over background: mainnet keeps green (6.27:1 light,
**3.77:1 dark**, an existing AA failure left untouched because fixing it would
repaint mainnet); testnet purple is 6.62:1 light and 5.60:1 dark; local blue is
8.66:1 light and 7.92:1 dark. Two purple steps were added to the palette
(`purple-200`, `purple-300`) because the scale had no equivalent of
`green-400`/`green-500`, and collapsing both onto `purple-400` dropped the dark
hover from 7.14:1 to 3.73:1.

Only identity tokens move. `--s-color-bg-success-default` resolves to the same
green primitive but stays green: success is a meaning, not a brand.

### The chain-id is validated once, at the source

The chain-id reaches markdown (the banner) and `<meta name="gnoconnect:chainid">`,
which wallets read. A backtick in it would close the banner's code span and let
the rest render as markdown, which is enough for an arbitrary link at the top of
every page. `NewRouter` refuses anything outside `^[a-zA-Z0-9_.-]{1,64}$` rather
than escaping at each use, so every consumer is covered, including the ones added
later. The value is not always operator-supplied: with `-chainid` empty it comes
from the node over the wire.

## Alternatives considered

- **Deriving the kind from the chain-id** (the first revision of this PR).
  Rejected after review: it hardcodes the mainnet naming scheme into gnoweb,
  and the failure mode of the flag-only design (mainnet forgetting the flag)
  is safe and immediately visible, while a wrong naming assumption fails
  silently in the dangerous direction the day the scheme changes.
- **Chip off-mainnet only** (as originally drafted in #6121 §2.3). Rejected:
  see above — an absent marker carries no information.
- **Faucet link conditional on `NetworkKind`.** Rejected: `FaucetURL` already
  encodes "does this deployment have a faucet", and it is what routes
  `/faucet`. Two sources of truth for one fact is how they drift.
- **Rendering `FaucetURL` as the link target.** Tried, then rejected against
  the live deployments — see above.
- **Deriving the kind from the HTTP host.** Rejected: the host is a deployment
  detail (staging, previews, port-forwards) while the chain-id is the thing a
  user's transaction actually lands on.

## Consequences

- `data-network` is available to CSS for any further per-network styling.
- `NetworkKind` is the hook §2.6 (per-network robots policy) and §2.4 (default
  off-mainnet banner) key off; neither is in this change.
- Deployments that want a footer Faucet link must set `-faucet-url`. Every
  deployment in this repo that had one already does (`misc/loop`,
  `misc/deployments/home-alias`, `test2`, `test3`), so none regress. `gnodev`
  does not set it and has no such flag, so a local dev server no longer shows
  the link — correct, it has no faucet.
- The `.network-chip--alert` rule is written flat rather than as a nested
  `&--alert`: `postcss-preset-env`'s `nesting-rules` is spec-compliant CSS
  nesting, which does not concatenate `&` with a suffix. The nested form
  compiles to a dead type selector, as the pre-existing `&--explorer` (removed
  by this change, its flat duplicate did the work) demonstrated. Anything
  added to this file must follow the flat form.
- `public/main.css` is a tracked, embedded build artifact and CI verifies it is
  in sync (`.github/workflows/ci-dir-gnoland.yml`, `gnoweb_generate`). Any CSS
  change here requires `make -C gno.land/pkg/gnoweb generate`.
- The chip prints `-chainid` verbatim, and that flag still defaults to `"dev"`,
  is never validated against the node, and shares its variable with the
  deprecated `-help-chainid`, so passing both is last-one-wins in silence
  (#6121 §2.1). The kind does not derive from it, but a wrong chain-id now
  shows on every page instead of behind the popup.

## Verification

`go test ./gno.land/pkg/gnoweb/... ./gno.land/cmd/gnoweb/...`, `go vet`,
`gofmt`, `npx biome check` on the changed CSS, and
`make -C gno.land/pkg/gnoweb fclean generate` with `public/main.css` committed.

New tests: `NewRouter`'s testnet default, explicit mainnet, and rejection of
an invalid `-network-kind` (`app_test.go`), `Valid()` and the chip text
(`components/network_test.go`),
a render-level assertion that `data-network` and `network-chip--alert` reach
the HTML (`components/layout_test.go`), and that the footer renders no Faucet
link without a configured faucet (`components/layout_footer_test.go`).

Also checked by hand against a running gnoweb, since a Go test cannot catch a
dead CSS selector: chip and `data-network` under both kinds, the chip
surviving the 400 error page, the footer link appearing only with
`-faucet-url`, and `.b-header .network-chip--alert` present in the generated
`public/main.css`.
