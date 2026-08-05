# Frontend fixtures

gnoweb has no JS test runner, and adding one is a larger decision than any one
change should make. These fixtures are what the wallet work is validated
against by hand, in a real browser, so the cases are reproducible rather than
re-invented.

## Running

```bash
make -C gno.land/pkg/gnoweb generate
cd gno.land/pkg/gnoweb && python3 -m http.server 8000
```

Then open <http://localhost:8000/frontend/fixtures/wallet-connect.html>. The
bundle is loaded from `/public/js/`, so the server must run from
`gno.land/pkg/gnoweb/`. An origin (not `file://`) is required: the session uses
`localStorage`.

The stub buttons can also be armed from the URL — `?stub=one&stub=legacy` — in
which case they fire before the controllers boot. Any case that outlives a
navigation needs this: Execute always ends in one, and a wallet announced by
clicking afterwards was not there when the page asked.

## wallet-connect.html — cases to cover

| # | Case | Expected |
|---|---|---|
| 1 | Announce one wallet, click Connect | connects without a chooser; avatar and truncated address appear in the header |
| 2 | Announce two, click Connect | chooser lists both; picking one connects as that wallet |
| 3 | Connected, open the menu → Change wallet | chooser opens **even with one wallet**, unlike login |
| 4 | Connected, menu → Disconnect | header returns to "Connect"; `gnoweb:session:v1` is gone |
| 5 | Connected, reload | address renders immediately, no "Connect" flash; `getAccount` is called, not `connect` |
| 6 | Connected, Execute | goes straight to the remembered wallet, no chooser; the logged `sendTx` carries `signer` |
| 7 | Unconnected, one wallet, Execute | wallet fires **and** the page navigates |
| 8 | Unconnected, two wallets, Execute → Cancel | navigates to the help URL with the args pinned |
| 9 | `Approved` | lands on `?status=success&hash=…#func-Transfer`; an unconnected sign adopts the identity |
| 10 | `Rejected` | lands on the function help page, args intact |
| 11 | The `not_connected` stub | `connect` runs, `sendTx` retries once, never surfaced to the user |
| 12 | The no-`sendTx` stub | falls through to the native submit, with a console warning |
| 13 | The no-`connect` stub, Connect | warns that it cannot be logged into; nothing is persisted |
| 14 | Legacy interceptor alone | it still intercepts once; no chooser |
| 15 | Legacy interceptor + one announcement | chooser opens with both; the interceptor does not fire until picked |
| 16 | `/wallets` with nothing announced | the install page still lists every registry entry |

Record the results in the PR description or the ADR.
