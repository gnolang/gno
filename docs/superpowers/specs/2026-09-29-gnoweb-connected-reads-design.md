# gnoweb connected reads: account, username, signer_unavailable

Design doc. Status: implemented (see Deviations).

## Context

`connect()` is mandatory for every in-page GnoConnect wallet, yet a connected
gnoweb behaves like a disconnected one. The session (`feature/connect/frontend/session.ts`)
feeds exactly two things: the header avatar, and the `tx.signer` pin on
outgoing intents (`frontend/js/controller-wallet-launch.ts:183`). The pin only
shows itself when a wallet declines, and gnoweb renders that decline as a plain
rejection.

The brainstorm in `2026-09-22-gnoconnect-demo-usecase-brainstorm.md` picked the
features below. This spec is **Spec 1**, which covers plain reads only. **Spec 2**
(D2 caller-aware dry-run, then D3 permission map) follows separately. It needs
a hand-written amino-binary encoder and a confirmed `MsgCall` simulation on a
real node first.

## Scope

1. Explain a `signer_unavailable` decline, on both transports.
2. Show GNOT balance, account number and sequence in the identity menu.
3. Fill `ACCOUNTNUMBER` and `SEQUENCENUMBER` in the `$help` gnokey command.
4. Show the registered username (`r/sys/users`) in the header.
5. Offer the connected address on `address`-typed `$help` parameters.
6. Badge the connected address where a realm renders it.

Out of scope: dry-run and gas estimation (Spec 2), other tx outcome codes and
`status=success`, a `/me` page, "view as me" URLs, a "signing as" banner, and
post-broadcast results (D5; needs only the hash, not `connect`).

## Decisions

### 1. Reads run in the browser

Every query about the connected address runs client-side against the RPC in
`<meta name="gnoconnect:rpc">`. The address never reaches the gnoweb server, and
pages stay identical for every visitor and cacheable. This is not new
territory: `_fetchQEval` (`frontend/js/controller-action-function.ts:236`)
already calls `abci_query` cross-origin from `$help` in production.

CORS is not a constraint. tm2 allows any origin by default
(`tm2/pkg/bft/rpc/config/config.go:99`), every `misc/deployments/` config keeps
`["*"]`, and `https://rpc.gno.land` answers cross-origin POSTs with
`access-control-allow-origin: *` (checked 2026-09-29). The real prerequisite is
operational: `-help-remote` must be a browser-reachable HTTPS URL. Without it,
gnoweb falls back to its own node address (`gno.land/cmd/gnoweb/main.go:236`).

Rejected: server-side endpoints. They would reuse `gnoclient`, but gnoweb would
learn who is connected, and per-user responses break the caching assumption.

### 2. One shared data layer (approach A)

A single RPC helper and a single per-page loader, with every feature a thin
consumer. Rejected: each controller fetching for itself, which duplicates the
same `auth/accounts` query and its error handling. Also rejected: caching chain
data in `localStorage` at connect time. The balance and sequence go stale at
once, and a stale sequence breaks the copied command.

### 3. The loader is keyed by address, not by session

The `$help` ADDRESS field stays as an editable fallback. So account number and
sequence follow whatever address is in the field, and the session only supplies
the default.

### 4. No silent prefill of parameters

A `Transfer(to address)` must not be filled with the caller's own address. The
address is offered through an explicit button, and only on parameters the chain
types as `address`. Name heuristics (`user`, `owner`) are rejected. Across
`examples/`, 97 exported functions take an `address` parameter and 11 take a
user-like `string` (boards2's `user string` among them). A name-based guess
would be wrong often enough to be noise.

## Components

Paths are relative to `gno.land/pkg/gnoweb/` unless absolute.

### `frontend/js/chain.ts`: RPC helper

- `rpcURL(): string | null` reads `gnoconnect:rpc` and adds `http://` when the
  value has no scheme (the fixture uses `127.0.0.1:26657`). It returns null when
  the meta tag is absent.
- `abciQuery(path, data?): Promise<Uint8Array | null>` issues a GET to
  `abci_query`. It returns the decoded `Data`, or null when `Data` is JSON
  `null`. It throws `ChainError` on a network failure, a non-2xx response, or a
  `ResponseBase.Error`.
- `qevalJSON(expr): Promise<string>` runs `vm/qeval_json` and returns
  `results[0].V.value`. It throws `ChainError` otherwise.
- `controller-action-function.ts` moves `_fetchQEval` onto `abciQuery`, with no
  behaviour change.

### `frontend/js/account.ts`: per-address loader

```ts
interface Account {
  address: string;
  accountNumber: string; // decimal strings: uint64 exceeds Number
  sequence: string;
  ugnot: bigint;
  otherDenoms: number; // count only; denom strings are never displayed
}
loadAccount(address): Promise<Account | null> // null: not on chain yet
loadUsername(address): Promise<string | null> // null: not registered
refresh(address): void                        // drop the cached entries
isAddress(s): boolean                         // bech32 g1 + 38 chars
```

- `loadAccount` parses `auth/accounts/<addr>` (`BaseAccount.coins`,
  `account_number`, `sequence`; `tm2/pkg/std/account.go:72-74`).
- `loadUsername` runs
  `qevalJSON('gno.land/r/sys/users.ResolveAddress("<addr>").Name()')`. On
  gnoland-1 this returns `"moul"` for `g1manfred47kzduec920z88wfr64ylksmdcedlf5`
  and an error for an unregistered address, which maps to `null`. An RPC
  failure still throws, so callers can tell "not registered" from "unknown".
- The address is checked with `isAddress` before it is spliced into an
  expression. Anything else throws, and no query is sent.
- Promises are cached per address for the page's lifetime, so concurrent
  consumers share one request. The cache is cleared on `onSessionChange`.
- The loaders refuse to run when `session.chainid` and the page's
  `gnoconnect:chainid` are both non-empty and differ, and behave as if the RPC
  failed. An empty value on either side (`normalize()` defaults a missing
  `chainid` to `""`) does not block them.

### `signer_unavailable` notice

- `_signInPage`: a `Rejected` response with `code === "signer_unavailable"`
  navigates to the help URL plus `status=error&code=signer_unavailable`. That is
  the same URL a launch-link wallet's callback produces, so from here on the two
  transports are identical.
- A `tx-outcome` controller on each `$help` function block reads `status` and
  `code` on load. It acts only on the block named by the URL's `func=<Name>`
  query parameter (the anchor is not usable: `_callbackURL` strips it, so a
  launch-link return has none), and only on `signer_unavailable`. It shows, under Execute: "Your wallet isn't
  on `<truncated session address>`, the account you're connected as. Switch to
  it in the wallet, or reconnect." Without a session, the message drops the
  address.
- **Reconnect** dispatches `connect:pick`. `controller-connect` handles it with
  its existing `_pickWallet("Change wallet", true)`.
- **Dismiss** removes `status` and `code` via `replaceState`, so a reload does
  not show the notice again.
- `code` comes from the URL and is untrusted: it is matched against known codes
  and never rendered. The address comes from the session and is set with
  `textContent`.

### Identity menu (header)

Toggle button:

- The avatar is followed by `@<username>` when registered, and by
  `truncate(address)` otherwise. The tooltip and `aria-label` carry the full
  address.
- `GnoSession` gains an optional `username`, persisted in `gnoweb:session:v1`.
  Each page renders the stored name at once, then calls `loadUsername` once and
  writes back any change.
  - A new address (connect, switch, reconcile) clears it.
  - A successful "not registered" answer clears it.
  - A thrown lookup keeps it.
  - `normalize()` treats a missing or non-string value as absent and clamps it
    to 64 characters.

Menu, above Copy / Change wallet / Disconnect:

```
@moul  →  /u/moul                           (only if registered)
g1manfred47kzduec920z88wfr64ylksmdcedlf5    (full, wraps, mono)
12.5 GNOT · +3 other tokens
Account 1234 · Sequence 56
```

- `loadAccount` is called each time the menu opens (after `refresh`), not on
  page load. A closed menu costs nothing.
- States:
  - loading: `…` in the value slots
  - `null` account: `0 GNOT · not on chain yet`, with no account line
  - throw: the balance and account lines are hidden
- GNOT is `ugnot` / 10⁶, with at most 6 decimals and trailing zeros trimmed. It
  is formatted from `bigint` so large balances stay exact.
- Accessibility: the account block sits in a `role="group"` labelled "Account",
  outside the `role="menu"` that holds the three actions. `role="menu"` may only
  contain menu items.

### `$help`: gnokey command (D1)

- `controller-action-header`: when connected, the ADDRESS field starts with the
  session address and follows `onSessionChange`. Edits behave as today (saved to
  `actionAddressInput`). When disconnected, nothing changes.
- `ui/command.html` wraps the two placeholders:
  `<span data-action-function-target="account-number">ACCOUNTNUMBER</span>` and
  `…="sequence">SEQUENCENUMBER</span>`.
- On each debounced `address:changed` where `isAddress(value)` holds,
  `controller-action-function` calls `loadAccount` and fills both spans. It
  restores the placeholders when:
  - the value is not an address (gnokey also accepts a key name)
  - the account is not on chain
  - the call throws
- On `visibilitychange` to visible, it calls `refresh` and reloads the current
  address. This covers copying the command, running it in a terminal, and
  coming back.
- Secure mode's `gnokey query auth/accounts/…` line and `-gas-wanted` are
  unchanged.

### `$help`: "me" button (U1)

- `components/views/action.html:127` adds
  `data-action-function-param-type-value="{{ .Type }}"` from `vm/qfuncs`.
- When connected, inputs whose type is `address` or `.uverse.address` get a
  small **me** button. It sets the input to the session address and dispatches
  `input`, so args, the command and the qeval preview update through the
  existing handlers.

### `you-badge` controller (U1)

- Added to `<md-renderer>` in `components/layouts/article.html`.
  Space-separated controllers are supported (`frontend/js/index.ts:104`). It is
  active only on realm views with a session.
- Matches:
  - text nodes containing the full session address, found with a `TreeWalker`
    capped at 5,000 nodes
  - `<a>` whose path is `/u/<session username>`
  - `<a>` whose `href` contains the address
- Nodes under `pre` and `code` are skipped.
- After each match it inserts `<span class="b-you">you</span>`. It inserts only
  its own element and never re-parses realm text.

## Error handling

The one rule: **a read failure makes the page look like today's disconnected
page.** No feature shows an RPC error. Specifically:

- `rpcURL()` null, a thrown `ChainError`, or a chain-id mismatch: the menu hides
  its balance and account lines, the command keeps its placeholders, the header
  keeps the stored or truncated identity, and the "me" button and badges still
  work, since they need only the session.
- A `null` account is not an error. It is rendered as "not on chain yet" in the
  menu, and as placeholders in the command.
- Every failure goes through `this.warn` (console) with the query path, and
  never the full response body.

## Testing

**Unit (new `test.front` target: esbuild `*.test.ts`, then `node --test`; no
new dependency):**

- `chain.ts`:
  - `abciQuery` response envelope: data, `null` data, `ResponseBase.Error`,
    non-2xx
  - `qevalJSON` extraction
- `account.ts`:
  - `auth/accounts` parsing (single denom, extra denoms, `null`)
  - `isAddress`
  - GNOT formatting (0, 1, fractional, above 2⁵³)
  - cache sharing and `refresh`
- `session.ts`: `username` lifecycle (switch clears it, a thrown lookup keeps
  it, `normalize` clamps it).
- `tx-outcome`: URL-to-notice mapping (known code, unknown code, no session,
  another function's `func=`).
- `you-badge`: text match, `/u/` link, `href` match, skip inside `code`, node
  cap.

**Go:** template tests for the `param-type` attribute and the two command spans.

**Browser (Chrome plugin, local gnodev, Adena dev build):**

- light and dark themes, phone width
- connected, disconnected, and not on chain
- `signer_unavailable` through the in-page decline and a hand-built callback URL
- boards2 `$help` with a key name versus an address in the field
- `/r/gnoland/coins:balances?address=<own>` for the badge

**Before done:** `make -C gno.land/pkg/gnoweb test lint`, the browser pass, and
`/simplify`.

## Delivery

One commit per point, each with its rebuilt `public/js` bundle and signed off:

1. `test(gnoweb): run frontend unit tests with node --test`
2. `refactor(gnoweb): extract the RPC helper from _fetchQEval`
3. `feat(gnoweb): explain a signer_unavailable decline`
4. `feat(gnoweb): show balance and account in the identity menu`
5. `feat(gnoweb): fill account number and sequence in the gnokey command`
6. `feat(gnoweb): show the connected username in the header`
7. `feat(gnoweb): offer the connected address on address parameters`
8. `feat(gnoweb): badge the connected address in realm content`
9. `docs(gnoweb): record connected reads and the -help-remote requirement`:
   - a "Connected reads" section in `gno.land/adr/prxxxx_gnoweb_wallet_connect.md`
     (decisions 1–4 and the rejected alternatives)
   - `-help-remote` usage in `gno.land/cmd/gnoweb/main.go:119` becomes "public
     RPC URL given to browsers, for gnokey commands and client-side reads; must
     be reachable by visitors over HTTPS"

## Consequences

- A connected page sends more queries to the public RPC:
  - one `qeval_json` per page load, for the username
  - one `auth/accounts` per menu open, and per address change on `$help`
- A deployment with a wrong `-help-remote` loses these features silently.
  Execute and the command still work as today.
- The username lives in `localStorage` next to the address, which is already
  there. Balance and sequence are never persisted.
- The frontend gets its first unit-test setup, which Spec 2's encoder golden
  test will reuse.

## Deviations

Found while planning and implementing; recorded in
`gno.land/adr/prxxxx_gnoweb_wallet_connect.md` (items 5–11).

1. Session changes travel as a `window` event: each controller bundle embeds its
   own `session.ts`. The account cache lives in a `globalThis` slot for the same
   reason.
2. The cache is not cleared on session change; it is keyed by address.
3. `abciQuery` returns text, not bytes.
4. `func` comes from the `$help&func=…` web query in the URL path, with a
   query-string fallback for the static fixture.
5. The "me" button is rendered server-side for `address`/`.uverse.address`
   params, and fills `"g1…"` on query functions (their live result evaluates the
   args as Gno) and a raw `g1…` on calls.
6. `you-badge` mounts on a hidden sentinel in `realm.html`, and a link counts as
   "you" only when its path carries the address.
7. The generic `[hidden]` rule is exempt from PurgeCSS, which had been stripping
   it.
