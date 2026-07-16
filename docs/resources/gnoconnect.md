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

Launch links hand a transaction intent off to an external wallet — a mobile app
or standalone desktop signer registered under a custom URL scheme — when an
in-page provider is not available. Gnoweb emits them from `$help` Execute; any
producer may author them.

```
<scheme>://tx?path=<pkgPath>&func=<Foo>&arg.<name>=<value>&send=<coins>&rpc=<rpc>&chainid=<chainid>&callback=<url>
```

- `<scheme>` is the wallet's registered custom scheme (e.g.
  `land.gno.gnokey`). The host is `tx`; future hosts (`run`, `sign`) may be
  added under the same scheme. Wallets should accept `call` as a silent
  back-compat alias for `tx` but emit and document only `tx`.
- Function arguments are named like TxLink arguments, but namespaced under
  `arg.` so realm parameter names cannot collide with the link's own keys
  (`path`, `func`, `send`, `rpc`, `chainid`, `callback`). Every name and value
  is percent-encoded. As with TxLinks, a link may prefill only some arguments;
  the wallet resolves parameter order and remaining fields via `vm/qdoc`.
- `send` (optional) is the coin amount to attach, in `gnokey` coin syntax
  (e.g. `1000000ugnot`).
- `rpc` and `chainid` mirror the `gnoconnect:rpc`/`gnoconnect:chainid`
  metadata of the emitting page, verbatim. `rpc` may be scheme-less
  (`127.0.0.1:26657`); the wallet assumes `http://` when the scheme is
  missing.
- `callback` (optional) is the URL to reopen after signing. The wallet
  appends `status` (`success` or `error`) and, on success, the transaction
  `hash` as query parameters.

- **Gnoweb** (provider)
- **Adena Wallet** (client)
- **Gnobro** (coming soon)
- _Add your clients here_

## Further Reading

- [Issue #2602](https://github.com/gnolang/gno/issues/2602)
- [Issue #3283](https://github.com/gnolang/gno/issues/3283)
- [PR #3609](https://github.com/gnolang/gno/pull/3609)
- [PR #3459](https://github.com/gnolang/gno/pull/3459)

