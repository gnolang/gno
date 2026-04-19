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
VIEW AT:    http://127.0.0.1:8888/r/demo/counter
```

The gnoweb URL is derived from the `--remote` RPC address with two simple
transformations:

1. **Strip `rpc.` prefix** from hostname (e.g. `rpc.gno.land` → `gno.land`).
2. **Replace the port** with gnoweb's default port (`8888`).

The `gno.land` prefix is stripped from the package path since gnoweb routes
use only the relative path (e.g. `/r/demo/counter`).

The logic lives in a `GnowebURLFromRemote` helper in
`gno.land/pkg/keyscli/root.go`, called from the `OnTxSuccess` callback in
`addpkg.go`.

## Alternatives Considered

| Alternative | Reason rejected |
|-------------|----------------|
| Well-known hosts map | Unnecessary complexity; simple rules cover all cases |
| New `--gnoweb-url` flag | Extra flag burden for uncommon edge cases |
| Print only pkg path, no URL | Less useful; users still need to construct URL |

## Consequences

- **Positive:** Immediate feedback after deploy; reduces onboarding friction.
- **Negative:** Heuristic may produce incorrect URLs for non-standard setups
  (e.g. gnoweb on a different port). The URL is best-effort and clearly labeled.
- **Future:** If a gnoweb discovery endpoint is added, the heuristic can be
  replaced with a query.
