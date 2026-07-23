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

Two hosts are defined: `tx` signs a transaction, `connect` asks for the user's
on-chain identity. Further hosts (`run`, `sign`) may be added under the same
scheme.

**Encoding.** Names and values are percent-encoded:

- Producers MUST percent-encode (`encodeURIComponent`). A literal plus is
  `%2B`.
- Wallets MUST accept form-encoded argument values as well: in `arg.<name>`
  and `args` values, `+` decodes to a space. Substitute **before**
  percent-decoding, so `%2B` still yields a literal `+`.
- Everywhere else — `path`, `func`, `send`, `rpc`, `chainid`, `callback`,
  `state`, `signer`, `broadcast` — `+` is a literal plus and is not
  substituted.

The leniency is there because `URLSearchParams`, the obvious way to build a
link in a browser, emits `application/x-www-form-urlencoded`, where a space is
`+`. A wallet parsing strictly per RFC 3986 shows the user `testing+board` for
a board they named `testing board`, and signs that. The leniency stops at
argument values because elsewhere a `+` may be data: `state` is often base64,
and rewriting it would break the correlation check it exists for.

### `tx` — review, sign, broadcast

```
<scheme>://tx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>&state=<token>&signer=<address>&broadcast=<bool>
```

- `<scheme>` is the wallet's registered custom scheme (e.g.
  `land.gno.gnokey`). Wallets should accept `call` as a silent back-compat
  alias for the `tx` host but emit and document only `tx`.
- Function arguments are named like TxLink arguments, but namespaced under
  `arg.` so realm parameter names cannot collide with the link's own reserved
  keys (`path`, `func`, `send`, `rpc`, `chainid`, `callback`, `state`,
  `signer`, `broadcast`). As with TxLinks, a link may prefill only some
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
- `signer` (optional) is the address the producer expects to sign, typically
  carried over from a prior `connect`. The wallet MUST sign with that account
  and MUST NOT silently sign as another identity; if it is unavailable, refuse
  and say so rather than substituting a different one.
- `broadcast` (optional, default `true`) selects the mode:
  - `true` — the wallet signs **and broadcasts**, returning `hash`.
  - `false` — **sign-only**: after the same review, the wallet returns the
    signed transaction and does not broadcast. The producer broadcasts on its
    own RPC. This suits a dapp that owns its connection to the chain and only
    needs a signature.

  User review before signing is mandatory in both modes.

  Sign-only moves real obligations to the producer, and they are easy to miss:

  - **It must be able to broadcast what the wallet signed.** A wallet may sign
    with a scheme the producer's client does not know — a session key, a
    multisig — and a client that cannot represent that signature will re-encode
    the transaction into an invalid one rather than refuse it. The failure
    surfaces at the very last step and looks like the wallet's fault.
  - **It owns the errors.** Out-of-gas, a rejected signature, a realm that
    refuses the call: all arrive at the producer, about a transaction the wallet
    composed, once the wallet's review screen is gone.
  - **`status=success` here means _signed_, not _landed_.** Nothing has been
    broadcast when the callback fires. A producer that treats it as completion —
    the reasonable reading in the default mode — will report success for a
    transaction that never reached the chain.

  Prefer the default when the producer has no specific reason to broadcast
  itself: the wallet built the transaction, resolved the account sequence, and
  understands its own signatures, so it is better placed to report what happened.

#### `tx` callback results

The wallet appends to `callback`, preserving any parameters already there:

```
<callback>?status=success&hash=<txhash>&state=<echoed>       # broadcast=true
<callback>?status=success&signedtx=<base64>&state=<echoed>   # broadcast=false
<callback>?status=cancelled&state=<echoed>                   # user declined
<callback>?status=error&message=<text>&state=<echoed>        # signing/broadcast failed
```

`signedtx` is the signed transaction as amino-JSON, base64-encoded.

`state` is echoed on **every** response, including failures, and is absent when
the request omitted it.

A wallet SHOULD answer every request it accepted — a producer waiting on a
callback cannot see an error surfaced on the user's device, and without a
`cancelled` or `error` response it waits indefinitely.

`hash` is a hint, not proof: the callback scheme is public, so a producer
should confirm the transaction on its own RPC before treating it as landed.

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
so the destination is the only anti-phishing anchor the user has.

```
<callback>?status=success&address=<bech32>&session=<bech32>&pubkey=<gpub>&chainid=<id>&state=<echoed>
<callback>?status=cancelled&state=<echoed>
<callback>?status=error&message=<code>&state=<echoed>
```

Error codes: `no_active_session`, `network_declined`, `invalid_request`.

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

