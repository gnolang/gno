# GnoConnect: Wallet & Client Integration Standard

GnoConnect is a standard for enabling wallets, clients, and SDKs (such as Adena
Wallet, Gnoweb, and Gnobro) to interact seamlessly with Gno blockchains. It's a
minimalistic, URL-based alternative to the gno-js-client that allows users to
define actions in their apps without JS/TS components, making integration
straightforward for both users and developers.

## How GnoConnect Works

GnoConnect uses HTML/HTTP metadata to provide connection details for clients and
wallets.

By including the following metadata/headers in your app, clients and wallets will be able to recognize your app as Gno-compatible and get the data needed to generate transactions for users.

### HTML Metadata

```html
<meta name="gnoconnect:rpc" content="127.0.0.1:26657" />
<meta name="gnoconnect:chainid" content="dev" />
<meta name="gnoconnect:txdomains" content="auto,example.com" />
```

- `gnoconnect:rpc`: RPC URL.
- `gnoconnect:chainid`: Chain ID.
- `gnoconnect:txdomains`: Domains treated as transaction sources.
  The value `auto` includes the current domain in addition to any specified
  domains.

### HTTP Headers

Alternative to HTML Metadata.

```
Gnoconnect-RPC: 127.0.0.1:26657
Gnoconnect-ChainID: dev
Gnoconnect-TXDomains: auto,example.com
```

The two sources SHOULD agree. If both are present and conflict, the HTTP header
takes precedence: it is set by the serving infrastructure and cannot be injected
through page content (a stored `<meta>` from a compromised dependency, or a
malicious realm rendering one), whereas metadata can.

## Transaction Links (TxLinks)

Transaction links define blockchain calls and can include optional arguments.

Without arguments:

```
$help&func=Foo
/r/path/to/realm$help&func=Foo
https://example.land/r/path/to/realm$help&func=Foo
```

With arguments:

```
$help&func=Foo&arg1=value1&arg2=value2
/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
https://example.land/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
```

Here `arg1`/`arg2` stand for the function's actual parameter names — a TxLink
names each argument directly. Launch links carry the same arguments namespaced
under `arg.<name>` (see Launch Links), because a launch link also has reserved
keys like `func` that a bare parameter name could otherwise collide with.

Links can be relative or absolute but must match one of the domains listed in
`gnoconnect:txdomains` (including the resolved `auto` domain if set).

TxLinks only prefill specified arguments. For non-specified arguments, clients
can call `vm/qdoc` to retrieve the remaining fields
(see [PR #3459](https://github.com/gnolang/gno/pull/3459)).

> **Note:** A future standard may define advanced rules for fields such as
> limits, format, and default values.

### Run Calls

TODO ([discussion](https://github.com/gnolang/gno/issues/3283)).

## Launch Links (external wallets)

Launch links hand an intent off to an external wallet — a mobile app or
standalone desktop signer registered under a custom URL scheme — when an
in-page provider is not available. Gnoweb emits them from `$help` Execute; any
producer may author them.

Three **hosts** are defined — the URL's host component selects the verb: `tx`
signs **and broadcasts** a transaction, `signtx` signs **without** broadcasting
(returning the signed tx to the producer), and `connect` asks for the user's
on-chain identity. Further hosts (`run`, message signing) may be added under the
same scheme.

Sign-only is a distinct host, not a flag on `tx`, on purpose: an unknown query
parameter is silently ignored — a wallet that didn't understand a
`broadcast=false` flag would broadcast anyway, exactly what a sign-only producer
must never allow — whereas an unknown **host** is declined with
`unsupported_host`. Making not-broadcasting a property of the verb means a wallet
can never broadcast a transaction the producer asked it only to sign.

**Forward compatibility.** The standard evolves additively: a new capability is
always a new query parameter, host, or (in-page) method — existing ones are
never repurposed. Receivers therefore MUST tolerate what they don't recognise:

- A wallet MUST ignore query parameters it does not understand, and MUST NOT
  reject a request for containing them.
- A wallet that receives a host it does not implement SHOULD answer
  `status=error&code=unsupported_host` when the request carries a `callback`
  (the general answer-duty below); with no callback there is nothing to answer.
  Only `callback` and `state` may be read from such a request: every host shares
  those two, but a verb the wallet does not know may define its other parameters
  however it likes.
- For that answer to be possible at all, a wallet SHOULD register for the whole
  custom scheme rather than for an enumerated list of hosts. Where the OS routes
  links per host — an Android intent filter naming each `android:host`, say — a
  host the wallet never declared does not reach it, and it cannot answer.

A producer must therefore tolerate **both** shapes of refusal: an
`unsupported_host` callback, and a purely local launch failure when no installed
app claims the link at all. Neither is guaranteed — a wallet predating this rule
may accept the launch and simply do nothing — so **a launch is not a promise of
a response**, and a producer MUST NOT treat one as pending indefinitely.

There is no version field: additivity plus ignore-the-unknown is the whole
compatibility contract.

**Payload size.** A launch link is a URL, so it is bounded by the platform's
URL-length limits (no universal figure; keep well under ~2 KB). Launch links
suit ordinary calls, not large payloads — a bulk `MsgRun` body or very large
arguments belong on the in-page transport (no such limit) or another channel.

**Wallet not installed.** A custom-scheme link requires the wallet already
installed; with none registered for the scheme the OS behaviour is
platform-specific and there is no in-protocol fallback. Producers should keep the
always-available copy-paste TxLink command as the wallet-agnostic fallback
(gnoweb renders it beside Execute); a graceful "not installed" path needs
Universal / App Links and is out of v1.

**Encoding.** Names and values are percent-encoded:

- Producers MUST percent-encode (`encodeURIComponent`). A literal plus is
  `%2B`.
- Wallets MUST accept form-encoded argument values as well: in `arg.<name>`
  and `args` values, `+` decodes to a space. Substitute **before**
  percent-decoding, so `%2B` still yields a literal `+`.
- Everywhere else — `path`, `func`, `send`, `rpc`, `chainid`, `callback`,
  `state`, `signer` — `+` is a literal plus and is not substituted.

The leniency is there because `URLSearchParams`, the obvious way to build a
link in a browser, emits `application/x-www-form-urlencoded`, where a space is
`+`. A wallet parsing strictly per RFC 3986 shows the user `testing+board` for
a board they named `testing board`, and signs that. The leniency stops at
argument values because elsewhere a `+` may be data: `state` is often base64,
and rewriting it would break the correlation check it exists for.

### `tx` — review, sign, broadcast

```
<scheme>://tx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>&state=<token>&signer=<address>
```

- `<scheme>` is the wallet's registered custom scheme (e.g.
  `land.gno.gnokey`). Wallets should accept `call` as a silent back-compat
  alias for the `tx` host but emit and document only `tx`.
- Function arguments are named like TxLink arguments, but namespaced under
  `arg.` so realm parameter names cannot collide with the link's own reserved
  keys (`path`, `func`, `send`, `rpc`, `chainid`, `callback`, `state`,
  `signer`). As with TxLinks, a link may prefill only some
  arguments; the wallet resolves parameter order and remaining fields via
  `vm/qdoc`.
- `send` (optional) is the coin amount to attach, in `gnokey` coin syntax
  (e.g. `1000000ugnot`).
- `rpc` and `chainid` mirror the `gnoconnect:rpc`/`gnoconnect:chainid`
  metadata of the emitting page, verbatim. `rpc` may be scheme-less
  (`127.0.0.1:26657`); the wallet assumes `http://` when the scheme is
  missing.
- `callback` (optional) is the URL the wallet reopens with the result.
- `state` (optional, RECOMMENDED) is an opaque producer-generated token,
  echoed verbatim in every callback. A callback scheme is public — anything
  installed can open it — so without `state` a producer cannot tell its own
  result from one an attacker synthesised. Producers that consume callbacks
  should always send one and drop responses that match no outstanding request.
  The wallet treats `state` as opaque and SHOULD bound its length (e.g. ≤256
  characters).
- `signer` (optional) pins the **identity** the producer expects to act as —
  the `address` from a prior `connect`. If present, the wallet MUST sign as that
  identity and MUST NOT sign as another; a wallet that cannot MUST decline
  (`status=error&code=signer_unavailable`) rather than substitute one. What
  counts as "that identity" — an exact address match, or an account the chain
  links to it (e.g. a delegated/session key) — is wallet-specific.
The `tx` host always signs **and broadcasts**; the callback returns `hash`. User
review before signing is mandatory. A producer that needs the signed transaction
*without* broadcasting uses the `signtx` host below.

**One message per link.** A `tx` link carries a single `MsgCall`. Multiple
messages are a planned additive extension — an indexed `msg.<i>.path` /
`msg.<i>.func` / `msg.<i>.arg.<name>` form, with today's flat fields the implicit
`msg.0`, advertised as the `multi_msg` feature (see `connect`). The `arg.`
namespace keeps that path collision-free, so it lands without migration; v1 has
no atomic multi-call bundle.

#### `tx` callback results

The wallet appends its response to `callback`:

```
<callback>?status=success&hash=<txhash>&state=<echoed>       # signed and broadcast
<callback>?status=cancelled&state=<echoed>                   # user declined
<callback>?status=error&code=<code>&state=<echoed>           # signing/broadcast failed
```

`status` is the outcome class — `success`, `cancelled`, or `error` — a closed
set. On `error`, `code` carries an enumerated, machine-readable reason (never
human text; producers MUST NOT parse it as prose):

- `invalid_request` — the link was malformed.
- `network_declined` — the user rejected the network switch.
- `signer_unavailable` — the wallet cannot sign as the pinned `signer`.
- `no_signer` — the wallet holds no account to sign with.
- `tx_failed` — the wallet could not sign or broadcast the transaction. The
  wallet has already shown the user the cause; the producer should still confirm
  on-chain, since a failure reported here does not guarantee nothing landed.

New reasons are added to `code`, never to `status`, so a producer's
`status=error` branch keeps working as the set grows.

`state` is echoed on **every** response, including failures, and is absent when
the request omitted it.

A wallet SHOULD answer every request it accepted — a producer waiting on a
callback cannot see an error surfaced on the user's device, and without a
`cancelled` or `error` response it waits indefinitely.

`hash` is a hint, not proof: the callback scheme is public, so a producer
should confirm the transaction on its own RPC before treating it as landed.

### `signtx` — review and sign, no broadcast

```
<scheme>://signtx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>&state=<token>&signer=<address>
```

Identical to `tx` field for field, but the wallet **signs and returns the signed
transaction without broadcasting** — the producer broadcasts it on its own RPC.
This suits a dapp that owns its connection to the chain and only needs a
signature. User review before signing is mandatory, exactly as for `tx`. A wallet
that does not implement sign-only answers `unsupported_host` rather than falling
through to a broadcast, so a producer's "do not broadcast" is guaranteed by the
protocol shape, not by the wallet's goodwill. The single-message limit and the
`msg.<i>` multi-message extension apply exactly as for `tx`.

Sign-only moves real obligations to the producer, and they are easy to miss:

- **It must be able to broadcast what the wallet signed.** A wallet may sign with
  a scheme the producer's client does not know — a session key, a multisig — and
  a client that cannot represent that signature will re-encode the transaction
  into an invalid one rather than refuse it. The failure surfaces at the very
  last step and looks like the wallet's fault.
- **It owns the errors.** Out-of-gas, a rejected signature, a realm that refuses
  the call: all arrive at the producer, about a transaction the wallet composed,
  once the wallet's review screen is gone.
- **`status=success` means _signed_, not _landed_.** Nothing has been broadcast
  when the callback fires; a producer that treats it as completion will report
  success for a transaction that never reached the chain.

Prefer `tx` when the producer has no specific reason to broadcast itself: the
wallet built the transaction, resolved the account sequence, and understands its
own signatures, so it is better placed to report what happened.

#### `signtx` callback results

```
<callback>?status=success&signedtx=<base64>&state=<echoed>   # signed, not broadcast
<callback>?status=cancelled&state=<echoed>                   # user declined
<callback>?status=error&code=<code>&state=<echoed>           # signing failed
```

`signedtx` is the signed transaction as amino-JSON, base64-encoded. The `status`
/ `code` envelope and code set are the same as `tx`'s, except `tx_failed` here
always means signing failed (nothing is ever broadcast). `state` echoing and the
answer-every-request duty are identical.

### `connect` — request the user's identity

```
<scheme>://connect?callback=<url>&state=<token>&rpc=<rpc>&chainid=<chainid>
```

Asks the wallet which address the user wants to act as — the sign-in step
before any `tx`. `callback` is **required**: the verb exists only to deliver an
answer, so a request without a usable one is dropped. `state` behaves as for
`tx`. `rpc`/`chainid` (optional) name the network the producer expects; the
wallet may prompt the user to switch before answering.

The wallet MUST ask the user before disclosing anything, and MUST show the
callback's host: a producer's claimed name is self-asserted and unverifiable,
so the destination is the only anti-phishing anchor the user has. The protocol
carries no producer-supplied display name — a producer's identity to the user
is its callback destination.

```
<callback>?status=success&address=<bech32>&session=<bech32>&pubkey=<gpub>&chainid=<id>&features=<tokens>&state=<echoed>
<callback>?status=cancelled&state=<echoed>
<callback>?status=error&code=<code>&state=<echoed>
```

Error codes (`code`): `no_signer`, `network_declined`, `invalid_request`. As on
`tx`, `status` is the closed outcome class and `code` the enumerated reason.

`features` (optional) is a comma-separated list of the wallet's optional
capabilities, letting a producer tailor later requests. v1 tokens: the **hosts**
the wallet supports — `tx` (sign and broadcast) and `signtx` (sign only) — plus
`multi_msg` (accepts the indexed multi-message form). The two tx hosts are
independent, so a pure signer may offer `signtx` without `tx`. `multi_msg`
extends the single-message baseline (every wallet handles at least one message).
Absent `features` is a back-compat default only: a pre-handshake wallet is
assumed to support `tx` (broadcasting) with a single message. Unknown tokens are
ignored. A producer may also simply attempt a host and treat `unsupported_host`
as the negative answer.

The returned identity is **display-level**. It carries no challenge and no
signature, so it proves nothing about control of the address: treat it as the
user stating who they are, not as authentication. Authority comes from the
on-chain `tx` the user reviews and signs. A proof-of-control extension
(challenge + signature) is left for producers with a backend able to verify one.

### Callback URL rules

A wallet opens `callback`, so it MUST constrain it:

- Accept `https:` and custom app schemes, but **reject** schemes dangerous to
  open: `javascript:`, `data:`, `file:`, `content:`, `blob:`, `about:`, and
  (Android) `intent:`.
- Require an absolute URI with a scheme, no control characters, bounded length.
- The wallet appends its response keys, preserving any parameters already in
  `callback`. If a response-key name (`status`, `code`, `state`, `hash`,
  `signedtx`, `address`, `session`, `pubkey`, `chainid`, `features`) already
  appears, the wallet's appended value is authoritative — producers MUST read
  the **last** occurrence.
- For an `https:` callback the wallet SHOULD return the result in the URL
  **fragment** (`#status=…`) rather than the query, to keep it out of server
  logs and `Referer`. A custom-scheme callback travels no network hop, so query
  parameters are fine there.
- On violation for `connect`, drop the request — there is nowhere to answer.
  For `tx` the callback is optional, so the wallet MAY still let the user sign,
  but MUST make clear that the requesting producer will not be notified.

## Supported Clients

- **Gnoweb** (provider)
- **Adena Wallet** (client)
- **Gnobro** (coming soon)
- _Add your clients here_

## Further Reading

- [Issue #2602](https://github.com/gnolang/gno/issues/2602)
- [Issue #3283](https://github.com/gnolang/gno/issues/3283)
- [PR #3609](https://github.com/gnolang/gno/pull/3609)
- [PR #3459](https://github.com/gnolang/gno/pull/3459)

