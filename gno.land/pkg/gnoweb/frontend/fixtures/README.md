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

## With page JavaScript disabled

None of the cases above apply: every one of them is page script. What this
section checks is that the *absence* of the feature degrades honestly, and that
a wallet which intercepts on its own still does.

Block script for the origin in Chrome under `chrome://settings/content/javascript`
→ "Not allowed to use JavaScript" → Add, using the exact `host:port` the fixture
is served from. The setting is per-origin including the port, so a second port
serving the same build stays scripted and is the control. Confirm the block took
before reading anything into a result: on `$help` the `gnokey` command must show
`-args $''`, not the typed value.

A fixture opened straight from `python3 -m http.server` has no `gnoconnect:*`
metas of its own, so run these against `gnodev` on a realm with a two-argument
function rather than against `wallet-connect.html`.

| # | Case | Expected |
|---|---|---|
| 17 | Load `$help`, no wallet installed | page renders; params show the values from the URL; the `gnokey` command's args are **empty** |
| 18 | Click the QR button | the `:target` panel reveals and the server-rendered code is visible — this is pure CSS and must survive |
| 19 | The QR anchor's `href` before submitting | frozen at what the server rendered; typing does not update it without script. Submitting is what refreshes it, via case 20 |
| 20 | Edit an argument, Execute, no wallet | the native submit posts the named inputs as `?k=v`; gnoweb answers **303** to the canonical `$help&func=…&k=v`, and the landing page shows the **typed** value. The QR, the Link button and the form action all re-render from it |
| 21 | Same, from a `$help` URL carrying no args | same fold; the args reach the canonical URL rather than being emptied |
| 21b | Clear a field, then Execute | the empty value survives — a cleared arg must not fall back to the one in the action URL |
| 21c | An argument containing a space | the redirect spells it `%20`, matching what `buildHelpURL` and the frontend emit |
| 22 | Execute with an intercepting wallet installed (Adena from the store) | **no navigation**: the content script cancels the submit and the wallet opens, exactly as on `master`. This is the intended no-script fallback, not an accident |
| 23 | The same click on the scripted control port | for contrast: with two or more candidates the chooser opens; with one, it fires directly |

Cases 20–21c depend on the Execute inputs keeping their `name` attributes and on
the hidden `func` input: without them the browser submits an empty form data set,
nothing lands in the query, and the args are discarded. The send checkbox is
deliberately *not* named, so a submit nobody could confirm never adds coins.

**Disable every wallet extension before running 20–21c.** An intercepting wallet
cancels the submit (case 22) before the browser builds the query, so the fold
never fires and the case cannot be observed. That is correct behaviour, not a
failure — but with Adena enabled these rows measure nothing.

Both were run in Chrome with all wallets disabled: with script blocked, and again
with script enabled and no wallet to route to (the "no candidate" row of the
Execute table in `prxxxx_gnoweb_wallet_connect.md`). Editing `author` to `bob`
and submitting lands on `…$help&func=Post&author=bob&body=hello` in both, with
the edit intact and no query fields in the breadcrumb.

**Case 22 is a regression test, not an observation.** gnoweb has agreed to keep
the no-script path working by leaving the markup an intercepting wallet scrapes
intact — `article.b-action-function > form.params` and the
`data-action-function-*` attributes. Re-run it after any change to
`views/action.html`, to the `b-action-function` class, or to those attributes;
breaking it silently removes the only in-page signing route these users have,
dropping them onto case 20 with their arguments discarded. See
[`gnoconnect_in_page_wallets.md`](../../../../adr/gnoconnect_in_page_wallets.md)
§ "With page JavaScript disabled".

A `Content-Security-Policy: script-src 'none'` proxy is **not** a substitute for
the browser setting. It blocks the page's own scripts but not an extension's
injected one — measured: Adena's `inject.js` still ran and threw under CSP, so a
CSP-based test would report the wallet as present when the real setting makes it
absent.

## The external transport, on iOS

The launch link can only be checked on a real iOS browser: iOS gates a custom
scheme behind a system "Open in …?" prompt, and a page that navigates after
firing the link dismisses that prompt before the user can answer — so the wallet
never opens, silently. Desktop Chrome cannot show this.

With gnokey-mobile installed on a booted simulator, and the server above still
running:

```bash
xcrun simctl openurl booted "http://localhost:8000/frontend/fixtures/wallet-connect.html"
idb ui tap --udid <udid> 20 93   # the Execute button, in points
```

Expected: the "Open in 'Gnokey'?" prompt appears — the page must **not** have
navigated — and Open brings up the wallet prefilled with the path, function,
arguments and network. `idb` (fb-idb) needs Python ≤ 3.12; taps are in points,
so divide screenshot pixel coordinates by the device scale.
