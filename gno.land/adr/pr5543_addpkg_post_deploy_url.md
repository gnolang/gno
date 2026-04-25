# ADR: Post-Deploy Success Message with Gnoweb URL

## Context

After a successful `gnokey maketx addpkg`, the CLI prints transaction details
(gas, height, events, tx hash) but gives no indication of where the deployed
package can be viewed. Users must manually construct the gnoweb URL — a common
source of friction, especially for newcomers.

## Decision

Print two additional lines after a successful `addpkg` broadcast:

```
PKG PATH:   gno.land/r/demo/counter
VIEW AT:    https://gno.land/r/demo/counter
```

The gnoweb base URL is resolved in this order:

1. **`GNO_GNOWEB_URL` environment variable.** Highest precedence. Lets
   operators of private or custom networks point at their own gnoweb
   without needing an entry in the canonical registry.
2. **Canonical registry** (`gno.land/pkg/networks`, introduced in
   [#5596](https://github.com/gnolang/gno/pull/5596)). `--chainid` is
   looked up; the entry's `gnoweb_url` is used.
3. **`dev` chain ID.** gnodev's default chain ID maps to
   `http://127.0.0.1:8888`, gnodev's default gnoweb address.

The package path is then appended to the base URL. The leading
`gno.land/` prefix is stripped (anchored: only `gno.land` exactly or
`gno.land/...` are stripped; pkg paths like `gno.landfoo/...` are
preserved).

If none of the three resolution steps produce a URL, no `VIEW AT` line is
printed.

The logic lives in a `GnowebURLForPkg` helper in
`gno.land/pkg/keyscli/root.go`, called from the `OnTxSuccess` callback in
`addpkg.go`.

## Alternatives Considered

| Alternative | Reason rejected |
|-------------|----------------|
| Hostname heuristic (strip `rpc.`, swap port to 8888) + HTTP probe | Fragile for non-standard hostnames; produces wrong URLs for testnets where the gnoweb host differs from the RPC host; the probe added a network round-trip on every successful deploy |
| Match by `--remote` against `rpc_endpoint` | Redundant with chain ID — the user already supplies `--chainid` for signing, and chain ID is the canonical key |
| Generic loopback fallback (parse `--remote`, swap port to 8888 when host is `127.0.0.1`/`localhost`/`::1`) | Adds URL parsing and an extra parameter for negligible benefit over hardcoding `dev`; gnodev's chain ID is `dev` by default and that's the only local case worth covering |
| New `--gnoweb-url` CLI flag | Per-command flag for an option that's typically constant per environment; the `GNO_GNOWEB_URL` env var lets operators set it once in their shell config and have every `gnokey` invocation pick it up automatically, with no extra flag burden on common-case calls |
| Fetch live `/api/networks` from gnoweb | Unnecessary network round-trip; the embedded registry is the same data |

## Consequences

- **Positive:** Immediate feedback after deploy; reduces onboarding friction.
- **Positive:** URLs for known networks are guaranteed correct by the
  registry — not derived heuristically.
- **Positive:** When a testnet is rotated, only `networks.json` changes;
  this command picks the new URL up automatically.
- **Positive:** Private and custom networks are first-class — operators
  set `GNO_GNOWEB_URL` once and `gnokey` deploys against their network
  print a working `VIEW AT` line immediately, without needing to fork or
  patch the registry.
- **Negative:** The `dev` mapping assumes gnodev's defaults (gnoweb at
  `http://127.0.0.1:8888`). A user who runs gnodev with a different web
  listener will see a broken URL — they can override with
  `GNO_GNOWEB_URL`. Custom local chain IDs (`mydev`, etc.) print no URL
  unless `GNO_GNOWEB_URL` is set.
- **Dependency:** Requires #5596 (`gno.land/pkg/networks`) to be merged
  first.
