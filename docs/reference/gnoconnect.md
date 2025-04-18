# GnoConnect Specification

GnoConnect is a standard enabling wallets, clients, and SDKs such as Adena
Wallet, Gnoweb, and Gnobro to interact seamlessly with Gnoland blockchains. It
revolves around HTML/HTTP metadata and formatted links, providing a
straightforward user experience.

## Page Metadata

### HTML Metadata

```html
<meta name="gnoconnect:rpc" content="127.0.0.1:26657" />
<meta name="gnoconnect:chainid" content="dev" />
<meta name="gnoconnect:txdomains" content="auto,example.com" />
```

- `gnoconnect:rpc`: RPC URL.
- `gnoconnect:chainid`: Chain ID.
- `gnoconnect:txdomains`: Domains treated as transaction sources. The value
  `auto` includes the current domain in addition to any specified domains. 

### HTTP Headers

Alternative to HTML Metadata.

```
Gnoconnect-RPC: 127.0.0.1:26657
Gnoconnect-ChainID: dev
Gnoconnect-TXDomains: auto,example.com
```

## TxLink

Transaction links define blockchain calls and can include optional arguments:

### Format

Without arguments:

```
$help&func=Foo
/r/path/to/realm$help&func=Foo
https://domain/r/path/to/realm$help&func=Foo
```

With arguments:

```
$help&func=Foo&arg1=value1&arg2=value2
/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
https://domain/r/path/to/realm$help&func=Foo&arg1=value1&arg2=value2
```

Links can be relative or absolute but must match one of the domains listed in
`gnoconnect:txdomains` (including the resolved `auto` domain if set).

TxLinks only prefill specified arguments. For non-specified arguments, clients
can call `vm/qdoc` to retrieve the remaining fields
(related discussion: [PR #3459](https://github.com/gnolang/gno/pull/3459)).

TODO: Propose a standard in doc comments to define advanced rules for fields
such as limits, format, default values, etc.

### Run Calls

`run` transactions enable advanced logic.

TODO: Define `run` calls ([discussion](https://github.com/gnolang/gno/issues/3283)).

## Supported Clients

1. **Gnoweb** (provider)
2. **Adena Wallet** (coming soon)
3. **Gnobro** (coming soon)
4. _Add your clients here_

## Related Resources
- [Issue #2602](https://github.com/gnolang/gno/issues/2602)
- [Issue #3283](https://github.com/gnolang/gno/issues/3283)
- [PR #3609](https://github.com/gnolang/gno/pull/3609)
- [PR #3459](https://github.com/gnolang/gno/pull/3459)

