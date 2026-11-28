# ADR: The GnoConnect Standard (in-page wallets, launch links, identity)

## Context

`docs/resources/gnoconnect.md` on `master` tells a wallet what a page offers —
`gnoconnect:*` metas, HTTP headers, TxLinks — but gives a page no way to hand a
transaction to a wallet, and no way to learn who the user is. Issue #2799
drafted an in-page JS surface (`registerWallet` / `getWallets`) for browser
extensions, which leaves out every wallet that cannot run code in the page: a
mobile app, a standalone desktop signer.

This PR turns the document into a standard a wallet can implement without
reading any producer's source. It was written alongside its implementations:
gnoweb (#5970 external wallets, #6016 in-page wallets), Gnokey Mobile on iOS and
Android, an Adena development build, and two independent producers
(boards2-mobile, Bubble Rumble mobile). Most decisions below are a place where
two of them diverged on the same input while both conformed to the text.

This ADR records the protocol's decisions. What gnoweb chose in implementing it
(registry, chooser, submit handling, session storage) is recorded in the ADRs of
#5970 (`pr5970_gnoconnect_external_wallets.md`) and #6016
(`pr6016_gnoconnect_in_page_wallets.md`, `pr6016_gnoweb_wallet_connect.md`).

## Decision

### Two transports, one intent

A wallet that runs in the page announces itself on `window` (In-Page Wallets).
A wallet that cannot — an app — is opened with a custom-scheme launch link
(Launch Links). Neither is a fallback for the other: they serve different
wallets, and a page offers both.

They carry the same transaction **intent** (path, function, arguments, coins,
chain, signer) but not the same **envelope**. A launch link needs out-of-band
delivery (`callback`, `state`) because its answer comes back as a new
navigation; a Promise is already correlated and private. A producer should
build both from one function, so the two cannot drift. A direct call is also not
URL-bounded, so the transports differ in capacity too.

### Shared rules

**Network resolution: `chainid` selects, `rpc` advises.** A request names its
chain; the wallet resolves it among the networks the user configured and never
queries or broadcasts through the producer's `rpc`. A producer that could name
the endpoint would choose the node answering the `vm/qdoc` lookup that decides
argument order, and the labels the review screen shows. Adding a chain the
wallet lacks takes an approval of its own, never one that also signs. A declared
`rpc` that differs from the endpoint in use SHOULD be surfaced, because a chain
id is not globally unique (`dev` names every local devnet) and the divergence is
the only sign of a mismatch. A network is identified by chain id and endpoint;
a name is a SHOULD, since a shipping wallet has none and a chain added from a
request has none by construction.

**Arguments: named or positional, one form per call.** Named arguments are
resolved through `vm/qdoc`, and a failed lookup is an error, not a fallback to
arrival order. Positional arguments assert their order and need no lookup, so
they are the form that works offline. A name matching no declared parameter is
refused: dropping it and binding it positionally both sign something other than
what was asked. A call supplying no arguments is neither form, so it MUST NOT be
refused for a failed lookup — the shipping wallets read it opposite ways, and one
refused to sign `Increment()` offline. On a launch link, names sit under `arg.`
so a parameter named `func` or `callback` cannot collide with the link's keys.

**The `signer` pin names an identity, not a key.** The wallet MUST produce a
transaction authorised by that identity or decline with `signer_unavailable`,
and MAY sign with a delegated session key of it. Key equality cannot be
implemented: both shipping wallets sign with a session key under a master pin,
and a pin captured at connect time goes stale as session keys rotate. Pinning
the identity pins what the chain credits, which is what a producer relies on. So
a producer pins the `address` from `connect`, never the `session`, and the
wallet checks the pin where it signs, not on arrival — the user may switch
accounts in between.

### In-page wallets

- **Two events on `window`, both directions required.** `gno:registerWallet`
  (wallet → page) and `gno:requestWallet` (page → wallets), as in EIP-6963 and
  the Wallet Standard. `window` is the one surface an extension announcing at
  `document_start` and a page script can both rely on. Announce-only misses late
  pages and request-only misses late wallets, so each side's half is a MUST.
- **`rdns` beside `uuid`.** `uuid` deduplicates within a page load; `rdns`
  identifies a wallet across loads and versions, so it is what a page persists.
- **`connect` and `sendTx` are core, the rest is feature-detected.** `connect`
  is how a page learns who the user is and what it pins as `signer`; a wallet
  that cannot be identified cannot be logged into. A user declining is the
  `Rejected` status, not a thrown error.
- **`sendMsgs` puts the chain and the signer on the transaction.** It takes one
  batch intent: messages plus one `chainid`, one `rpc`, one `signer`. A list of
  per-message intents could name two chains for one broadcast, and four wallets
  could answer it four ways while conforming.
- **Connecting is per origin.** A wallet MAY gate any method, `getAccount` and
  `getNetwork` included (one discloses the identity, the other a chosen
  endpoint). A gated call answers `not_connected`; a signing request MAY carry
  its own approval, and a page MUST handle both answers.
- **Events go through the provider.** `provider.on` (`accountChanged`,
  `networkChanged`, `disconnect`) reaches approved origins only, and a page
  never answers `disconnect` with a connect prompt — the user just revoked it.
- **Announcements are untrusted.** Any script can announce. A chooser renders
  names as text and accepts only `data:` icons; picking one is the user's trust
  decision.
- **A wallet MUST NOT cancel or consume the page's events.** Scraping the markup
  and cancelling the submit makes the choice for the user, invisibly to the
  page — the race the announce protocol exists to remove. Observing is allowed.
  The one exception is a page that never asks: a wallet MAY intercept a TxLink
  navigation only on a page that has not dispatched `gno:requestWallet`, which a
  page implementing the protocol MUST do on load. That keeps visitors with
  JavaScript off reachable, and the two paths never compete. It is in the
  standard because no-JS is permanent, not a migration with an end date.

### Launch links

- **Custom schemes**, not Universal / App Links: origin-independent, work on
  localhost and offline, need no hosted association file. A wallet registers
  the whole scheme, because per-host routing (an Android intent filter per
  host) never delivers a request for an unknown verb and leaves the wallet
  unable to answer it.
- **A verb is a host when ignoring it would be unsafe.** An unknown parameter is
  ignored and an unknown host is declined. So sign-only is the `signtx` host,
  not `broadcast=false` on `sendtx` — an older wallet would have broadcast what
  it was asked only to sign. Revocation is the `disconnect` host for the same
  reason, and `run` will be a host too.
- **Forward compatibility is additive.** Receivers ignore what they do not know;
  an unknown host with a `callback` is answered `unsupported_host`. There is no
  version field. A launch is not a promise of an answer: a producer must
  tolerate both `unsupported_host` and a launch that fails silently.
- **Every answer is correlated and closed.** `state` is echoed verbatim, since
  the callback scheme is public and a producer must tell its own result from a
  forged one. `cancelled` and `error` (with an enumerated `code`) are specified,
  so a producer never waits forever on a failure raised on the user's device.
- **Callback rules.** The wallet shows the callback host: the protocol carries
  no producer name, so the destination is the user's only anti-phishing anchor.
  An `https` callback answers in the fragment, out of server logs. An unsafe
  callback drops a `connect`, which exists only to deliver an answer, but need
  not block a `sendtx`, where the transaction is the point.
- **Encoding.** Producers percent-encode, a plus being `%2B`. Wallets read `+` as
  a space in argument values only, and before percent-decoding, so `%2B`
  survives. `URLSearchParams` writes spaces as `+`, and a wallet reading RFC
  3986 strictly showed the user a value they never typed.
- **Sign-only moves obligations to the producer.** It must be able to broadcast
  what the wallet signed, it inherits every chain error after the review screen
  is gone, and `success` means signed, not landed. A producer that attributes
  the transaction to someone must pin `signer`, since `signedtx` is opaque.

### Identity: `connect`, challenge, sessions

- **`connect` is required on both transports.** On a launch link, a login that
  goes nowhere is indistinguishable from a user abandoning it. In the page,
  `connect` on an already-approved origin MUST resolve without a prompt, or a
  producer restoring a session on every navigation would prompt on every page;
  a challenge always prompts, since what the user approves is a signature.
  `chainid` is optional on `connect`: nothing is signed, and the answer says
  which chain it is about.
- **Without a challenge the identity is display-level.** It is the user saying
  who they are. Authority comes from the transaction they review and sign.
- **A challenge is signed as a sign-data document, and checked as one.** The
  identity, never a session key, signs a `sign/MsgSignData` document with the
  chain id kept (ADR-036's off-chain message), so a proof on one network is none
  on another and no node can broadcast it. The producer MUST parse the signed
  bytes as that document and compare challenge, signer and chain. Checking the
  signature alone is not enough: any signature the key ever made — a
  transaction on chain — would pass. `std.SignDataPayload` and
  `std.VerifySignData` build and check it, and `gnokey sign-data` (#6322) signs
  one. A `connect` cannot carry both a challenge and a key, so the identity never
  signs twice in one exchange.
- **Sessions are the wallet's business.** A producer gets two verbs: `connect`
  (with a `pubkey` to have a key it holds authorised as a session, or with scope
  hints for the wallet to pick one of its own) and `disconnect`. The private key
  never travels. Every request naming a key carries a proof of possession over a
  canonical payload covering every parameter and the host, so a request cannot
  be redirected, replayed on another chain or turned into another verb. The
  in-page provider carries the same verbs and rules. It differs only where the
  transport does: there is no callback, and the review shows the page's origin.

## Alternatives considered

- **A `window.<wallet>` global, or a shared `window.gno` array** — the last
  writer wins, so a user with two wallets reaches one.
- **The Wallet Standard's registration handshake verbatim** — richer, but it
  makes the protocol depend on a shared library; announce/request is enough to
  build a chooser.
- **Universal / App Links as the primary transport** — origin-dependent, needs
  hosting, fails on localhost. At most a future "not installed" fallback.
- **Positional arguments only** — a producer whose order diverges from the
  declaration binds values to the wrong parameters, silently.
- **Flat TxLink-style names on launch links** — collide with the link's keys.
- **`broadcast=false` on `sendtx`** — ignored by an older wallet, which then
  broadcasts.
- **A version field** — additivity plus ignore-the-unknown already covers it.
- **The `signer` pin as key equality** — implemented by no wallet and
  unimplementable against rotating session keys.
- **`sendTxs(GnoTxIntent[])`** — lets a page name two chains for one
  transaction.
- **Verifying the challenge signature alone** — accepts any signature the key
  ever made.
- **A gnoweb liveness marker to sanction interception** — puts one producer's
  signal into a wallet-agnostic standard. The `gno:requestWallet` test reuses a
  signal the standard already defines.

## Consequences

- A wallet can implement GnoConnect from `gnoconnect.md` alone, and any producer
  can use it; nothing in the standard is specific to gnoweb.
- `connect` is a method every wallet must implement. It is cheap now, while the
  in-page half has no shipped implementation, and it is the price of "log in
  with any conforming wallet".
- Proving an identity is a sign-data proof, so a producer verifying it depends on
  `std.VerifySignData` or an equivalent of it, not on a bare signature check.
- The protocol grows by adding hosts, parameters and `features` tokens. A
  multi-message launch link (`msg.<i>`) is named but not yet specified. A wallet
  that is not installed cannot be detected from a custom-scheme link.
- The no-script interception rule binds wallets on every producer, not only on
  gnoweb.

## Validation

The text was implemented end to end by the producers and wallets in Known
Implementations: gnoweb against Gnokey Mobile (iOS simulator, Android emulator)
and against stub wallets announcing per the spec, with an Adena development
build implementing the in-page half; boards2-mobile and Bubble Rumble mobile
against Gnokey Mobile, including sign-only and producer-held sessions. The
implementation PRs' ADRs record gnoweb's runs.
