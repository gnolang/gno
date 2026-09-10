# ADR: gnoweb network kind

## Context

gnoweb is a single binary serving several chains under the gno.land name:
`gnoland1` (betanet, currently at gno.land), `pearl-1`, `staging`, `dev`, and
soon `gnoland-1` — the mainnet whose genesis is being built in #6154.

Nothing in the UI tells them apart. The only signal is the chain-id inside the
Network Info popup, behind a click, so a user cannot know which chain they are
acting on without opening it. With real value on mainnet and several public
deployments answering under the same branding, that is a footgun.

The footer made this concrete: it rendered `https://faucet.gno.land/`
unconditionally, on every network. The link works, but the Faucet Hub it points
to only dispenses testnet tokens (`faucet.pearl.testnets.gno.land`,
`faucet.sapphire.testnets.gno.land`), and #6154 states mainnet ships with **no
faucet**. So the deployment carrying real value advertised a testnet service
without saying so.

This is item §2 of #6121, plus §3.5 and part of §3.3.

## Decision

**One server-side source of truth.** `components.NetworkKind` is derived from
the chain-id in `NewRouter`, after the chain-id is settled (it may be read from
the node), and carried `AppConfig` → `StaticMetadata` → `IndexData` →
`data-network` on `<html>`, next to the existing `data-theme`.

Deriving from the chain-id rather than from a standalone flag means a testnet
cannot accidentally present itself as mainnet through a config mistake.
`-network-kind` overrides it, and an invalid value is a startup error rather
than a silent fallback.

`gnoland-1` is mainnet; **everything else is a testnet**, including `gnoland1`.
The hyphen is the whole difference and the two are different chains. An
unrecognised chain-id resolves to testnet on purpose: a testnet mistaken for
mainnet is the dangerous direction, a mainnet mistaken for a testnet is merely
ugly.

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
have been a second source of truth for the same fact.

Its *value* is deliberately not used as the link target. `-faucet-url` holds
the endpoint the `/faucet` route redirects to, and the deployments disagree on
what that is: staging points it at `https://faucet-api.staging.gno.land`, which
answers 405 to a browser GET, while its footer points at the hub. Pearl points
it at the hub. Feeding it into an `href` would have sent staging users to a
POST-only API. The footer keeps the hub URL as a constant; only the presence of
the link is configuration-driven.

This narrows the footer only. `-faucet-url` still reaches an `href` on the
`/faucet` interstitial (`components/views/redirect.html` renders it as the
canonical link, the meta-refresh target and a visible anchor), so on staging
that page still offers a click through to the API. Pre-existing and out of
scope here, but the two surfaces now disagree by design rather than by accident.

**Colour is secondary.** Only `--s-logo-hat` moves under
`[data-network="testnet"]` — layout, contrast and dark mode are untouched, and
the chip carries the text. §2.2 of #6121 proposed also moving
`--s-color-text-brand-default`; that token is the text *on* brand surfaces, so
tinting it would have broken button contrast. It is left alone.

Two absolute `https://gno.land/...` links (header About, footer Blog) are made
relative in passing: on any non-mainnet deployment they silently moved the user
to mainnet.

## Alternatives considered

- **A standalone `-network-kind` flag with no derivation.** Rejected: one more
  value an operator can forget or get wrong, with no cross-check. Derivation
  makes the chain-id the authority and leaves the flag as an escape hatch.
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
- `NetworkKind` is the hook §1.6 (per-network robots policy) and §2.4 (default
  off-mainnet banner) key off; neither is in this change.
- Deployments that want a footer Faucet link must set `-faucet-url`. Every
  deployment in this repo that had one already does (`misc/loop`,
  `misc/deployments/home-alias`, `test2`, `test3`), so none regress. `gnodev`
  does not set it and has no such flag, so a local dev server no longer shows
  the link — correct, it has no faucet.
- The `.network-chip--alert` rule is written flat rather than as a nested
  `&--alert`: `postcss-preset-env`'s `nesting-rules` is spec-compliant CSS
  nesting, which does not concatenate `&` with a suffix. The nested form
  compiles to a dead type selector, as `&--explorer` at `06-blocks.css:80`
  already demonstrates. Anything added to this file must follow the flat form.
- `public/main.css` is a tracked, embedded build artifact and CI verifies it is
  in sync (`.github/workflows/ci-dir-gnoland.yml`, `gnoweb_generate`). Any CSS
  change here requires `make -C gno.land/pkg/gnoweb generate`.
- `-chainid` still defaults to `"dev"` and is not validated against the node
  (#6121 §2.1). A wrong `-chainid` now also yields a wrong network kind, which
  makes that gap more visible but does not create it.

## Verification

`go test ./gno.land/pkg/gnoweb/... ./gno.land/cmd/gnoweb/...`, `go vet`,
`gofmt`, `npx biome check` on the changed CSS, and
`make -C gno.land/pkg/gnoweb fclean generate` with `public/main.css` committed.

New tests: the chain-id derivation table including the `gnoland1` /
`gnoland-1` pair (`components/network_test.go`), `NewRouter`'s derivation,
override and rejection of an invalid `-network-kind` (`app_test.go` — the only
place the "a testnet cannot present itself as mainnet" property is enforced),
a render-level assertion that `data-network` and `network-chip--alert` reach
the HTML (`components/layout_test.go`), and that the footer renders no Faucet
link without a configured faucet (`components/layout_footer_test.go`).

Also checked by hand against a running gnoweb, since a Go test cannot catch a
dead CSS selector: chip and `data-network` on mainnet and testnet chain-ids,
the chip surviving the 400 error page, the footer link appearing only with
`-faucet-url`, and `.b-header .network-chip--alert` present in the generated
`public/main.css`.
