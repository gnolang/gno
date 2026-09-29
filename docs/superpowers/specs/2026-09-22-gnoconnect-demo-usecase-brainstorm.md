# Brainstorm: what a connected gnoweb can do for the user

**Status:** scope decided 2026-09-29 (see Decisions); design next.
**Date:** 2026-09-22
**Branch:** `feat/gnoconnect-in-page-wallets`

## The question

`connect()` is a core GnoConnect method — every in-page wallet must implement it
(`docs/resources/gnoconnect.md`, "Every in-page wallet MUST implement `connect`").
We need a demo use case that shows *why*: what a dapp such as gnoweb can offer a
user once it knows who they are.

## Where we are today

What connect currently buys gnoweb:

- `connect()` returns `{address, chainid, pubkey}`, stored by `session.ts`.
- The header renders an identicon plus a truncated address
  (`gno.land/pkg/gnoweb/feature/connect/frontend/controller-connect.ts:67`).
- The only *use* of the session is pinning `tx.signer` on outgoing intents
  (`gno.land/pkg/gnoweb/frontend/js/controller-wallet-launch.ts:183`).

So connect is purely a **write-side** convenience right now. It makes signing
safer, but a connected gnoweb otherwise looks and behaves identically to a
disconnected one. The whole reason `connect` exists as a separate method — the
dapp knows who you are *before* anything is signed — is unused.

## The structural constraint

gnoweb renders server-side (`vm/qrender` via `gno.land/pkg/gnoweb/client.go:144`),
but the connect session lives client-side in `localStorage`. Every personalization
idea below is therefore either:

- **client-side**, done in the browser with its own RPC queries (keeps the address
  local). CORS is not the blocker: tm2 allows any origin by default
  (`tm2/pkg/bft/rpc/config/config.go:99`), every `misc/deployments/` config keeps
  `cors_allowed_origins = ["*"]`, and `https://rpc.gno.land` answers cross-origin
  POSTs with `access-control-allow-origin: *` (checked 2026-09-29). The real
  prerequisite is that `<meta name="gnoconnect:rpc">` (`RemoteHelp`) holds a
  browser-reachable HTTPS URL: without `-help-remote` it falls back to gnoweb's
  own node address (`gno.land/cmd/gnoweb/main.go:236`), often `127.0.0.1` or an
  internal hostname. Degrade gracefully when an operator disables CORS; a
  same-origin `/rpc` passthrough would fix that but hands the address back to
  the server, or
- **server-side**, which means shipping the connected address to the gnoweb server
  (simplest, reuses `gnoclient`, no CORS — but gnoweb learns the user's address,
  and it breaks shareable-URL/caching assumptions).

This choice is not yet made and cuts across every option. With CORS open by
default it is a privacy and caching trade-off, not a feasibility one.

## Findings worth keeping

Discovered while exploring; each one is load-bearing for at least one idea below.

1. **Address → username reverse lookup already exists on chain.**
   `ResolveAddress(addr) *UserData` in
   `examples/gno.land/r/sys/users/users.gno:21`. A connected gnoweb could show
   a username instead of `g1abc…xyz`.

2. **The `$help` page already has a manual address field.**
   `gno.land/pkg/gnoweb/frontend/js/controller-action-header.ts:49` restores it
   from `localStorage["actionAddressInput"]` — the developer types their own
   bech32 by hand today. Connect makes the field obsolete.

3. **The copy-paste gnokey command has three unfilled placeholders.**
   `gno.land/pkg/gnoweb/components/ui/command.html:15-17`. Secure mode's first
   line is literally "run `gnokey query auth/accounts/ADDRESS`", after which you
   hand-edit `ACCOUNTNUMBER` and `SEQUENCENUMBER` into line 3. Connect plus one
   `auth/accounts` query fills all three.

4. **Caller-aware dry-run is possible, and `MsgCall` needs no signature for it.**
   `.app/simulate` exists (`gno.land/pkg/gnoclient/client_txs.go:501`,
   `SimulateResult`), and `gno.land/pkg/gnoland/app.go:161` sets
   `RequireSigForSimulate: txCarriesCode` — only `MsgAddPackage`, `MsgRun` and
   `MsgEnablePackage` need a real signature. A `MsgCall` simulates with a
   **pubkey-only placeholder**, and `connect()` returns `pubkey` precisely so a
   page can build one.

   This is the key discovery: **connect is the only way gnoweb can answer "what
   happens if *I* call this" before anything is signed**, because the answer
   depends on the caller.

## Audience

Asked and answered once: the demo should show that **a connected gnoweb is
meaningfully nicer to use than a disconnected one** (end users / stakeholders),
rather than proving the standard to wallet implementers or teaching realm devs a
pattern.

Then revised: **focus on developers**, who are gnoweb's actual audience. Both
framings are kept below because the user-facing ideas may still matter later.

## Ideas — user-facing framing

### U1. Identity layer over existing realms
Header resolves your username and profile link (finding 1) instead of a truncated
hash. Every occurrence of your address in rendered realm output gets badged
**you**. Execute prefills `address`-typed params.
*No realm changes; visible on pages that already exist.*

### U2. "My gno.land" — a personalized `/me` page
Profile, balance, realms you deployed, realms you're a member of.
*Needs history/indexing gnoweb does not have. Likely blocked.*

### U3. "View as me" — a path convention
Realms that already render per-address paths (`/r/x:u/g1…`) get a toggle that
substitutes your address.
*Small mechanism, but depends on realms following a convention.*

### U4. Safer Execute
Prefill, a "signing as X" banner, an "you are/aren't the owner" check.
*Smallest; most useful to devs, least visible to end users.*

Also: surface `signer_unavailable`. The `tx.signer` pin's only visible effect is
a wallet declining because it is not on the pinned identity (account switched,
or a different wallet picked in the chooser). Today gnoweb has no branch for it
— the code sits only in the type union
(`gno.land/pkg/gnoweb/frontend/js/wallet-discovery.ts:33`) — so `_signInPage`
falls through to the generic rejection and lands back on the help page
(`gno.land/pkg/gnoweb/frontend/js/controller-wallet-launch.ts:266`). Show
"your wallet is on another account than `g1…`; switch or reconnect" with a
reconnect action instead. The same applies to the launch-link callback
(`status=error&code=signer_unavailable`).
*Tiny; makes the pin legible when it fires.*

## Ideas — developer framing (current focus)

### D1. Fill the gnokey command
Address from connect, account number and sequence from `auth/accounts/<addr>`.
Turns a four-step copy-edit-query-edit ritual into one copyable block, and
deletes the manual address input (findings 2 and 3).
*Tiny. Uses connect for a read.*

### D2. Dry-run as you
A "Simulate" button beside Execute. Build the same `MsgCall` the Execute button
would send, sign it with a pubkey-only placeholder from `connect().pubkey`, run
`.app/simulate`, and report:

- would this succeed,
- the panic string if not,
- the gas actually used — which also replaces the hardcoded
  `-gas-wanted 1_000_000_000` in the command template.

*Impossible without connect. The narrative: connect turns `$help` from a
transaction form into a call console.*

Open sub-question: server-side endpoint (reuses `gnoclient`, no CORS, address
reaches gnoweb) vs. purely client-side (address stays local, needs CORS).

### D3. Permission map
D2 applied across every exported function: the function list marks the ones your
key would be rejected from ("not the owner", "not a DAO member"). Turns `$help`
into a permissions audit view for an unfamiliar realm.
*N simulations per page. Natural follow-up once D2's machinery exists.*

### D4. Account panel
Balance, account number and sequence in the header dropdown — the three things
you check when a transaction fails.
*Cheap, unglamorous, genuinely used.*

### D5. Result and events inline after broadcast
Execute currently ends in a navigation (commit `3a59097cb`). With a known signer
and the returned hash, poll the transaction and render its result and emitted
events on the page. Events are how gno devs debug realms.
*Not yet discussed in depth.*

## Recommendation on the table

**D2 with D1 alongside.** D2 is the headline because it is the one thing on this
list that cannot exist without `connect`. D1 is nearly free, removes a manual
field, and reinforces the same story. D3 follows naturally afterwards.

## Open questions

1. Which direction: D2+D1, D2 alone, D3, or D1+D4?
2. Client-side or server-side simulation (the structural constraint above)?
3. Does the demo need anything in `examples/`, or is it gnoweb-only?
4. Is a realistic `-gas-wanted` from simulation in scope, or a follow-up?

## Decisions (2026-09-29)

Review of the findings and ideas above:

- **Finding 1** — `r/sys/users` is live on gnoland-1 (94 registered names) and
  realms already resolve through it (`ResolveAddress` in `r/nt/commondao/v0`,
  `ResolveAny` in `r/gnoland/coins`).
  gnoweb does not yet (`gno.land/pkg/gnoweb/handler_http.go:568,623` are TODOs).
  **In scope** via U1.
- **Finding 4 / D2** — not the same as PR #6122 (base `playground2`). That PR and
  #6035 dry-run a user-written **`MsgRun` script** through a server endpoint
  (`/_/api/dryrun`), gated behind `-with-msg-run` because `MsgRun` is restricted
  and needs a real signature to simulate (`RequireSigForSimulate`). D2 simulates
  the **`MsgCall`** the Execute button already builds, which master still
  simulates with a pubkey-only placeholder. Caveat to D2's framing: #6035 fetches
  the pubkey from the on-chain account, so a `MsgCall` dry-run as any address that
  has signed before is possible without `connect`. What connect adds is the
  address without typing it, and the pubkey for an account that has never signed.
  Live `MsgCall` simulation on gnoland-1 is still unverified.
- **U1** — in scope: username in the header, "you" badge, prefill address params.
- **U2 + D4** — merged: one `auth/accounts/<addr>` fetch (coins, account number,
  sequence) shared by the header dropdown and D1. GNOT balance, account number
  and sequence in the dropdown, not next to the avatar. GNOT only, plus
  "+N other tokens" (third parties can send junk denoms). A `null` account shows
  "0 GNOT, not on chain yet". No `/me` page.
- **U3** — dropped. `Render` is a read-only query with no caller identity, so
  "view as me" only fills in your own address; realms put it in different places
  (`?address=` vs path segment), so it needs per-realm knowledge.
- **U4** — banner dropped (the sticky header already shows the identity next to
  Execute). `signer_unavailable` handling kept. Owner check moves to D2/D3.
- **D1** — in scope, fed by the U2+D4 fetch.
- **D2** — in scope, subject to the gnoland-1 check above.
- **D3** — depends on D2; there is no generic way to know a call would reject you
  without running it.
- **D5** — dropped: it needs only the tx hash, not `connect`.

**Delivery rule:** each implemented point lands in its own commit.

**Architecture (decided 2026-09-29):**

- All connected-address queries run **client-side**, against the RPC in
  `<meta name="gnoconnect:rpc">`. gnoweb already does this in production:
  `_fetchQEval` (`gno.land/pkg/gnoweb/frontend/js/controller-action-function.ts:236`).
- **Approach A:** one shared module generalized from `_fetchQEval` (`chain.ts`,
  JSON-RPC `abci_query`) plus a per-page account loader on top; features are thin
  consumers. `_fetchQEval` moves onto it in its own refactor commit.
- **Two specs.** Spec 1: `signer_unavailable`, the shared module, U2+D4, D1, U1
  (plain reads). Spec 2: D2 then D3, after a `MsgCall` simulation is confirmed on
  a real node.
- **Spec 2 encoder: hand-written** amino-binary encoding for the one fixed tx
  shape, checked byte-for-byte against Go's `amino.Marshal` in a golden test.
  Rejected `@gnolang/gno-js-client` + `tm2-js-client`: they import Node `crypto`
  (no browser bundle as-is), and a minimal encode entry bundles to 326 KB
  minified / 103 KB gzip (all gnoweb JS is 129 KB today), mostly curve, mnemonic
  and cipher code a simulation never uses.

## Next step

Approaches → sectioned design → spec → implementation plan for the in-scope set:
`signer_unavailable`, U2+D4, D1, U1, D2, D3.
