---
id: tm2-js-client
---

# Tendermint2 JavaScript Client

[@gnolang/tm2-js-client](https://github.com/gnolang/tm2-js-client) is a JavaScript/TypeScript client implementation 
for Tendermint2-based chains. It is designed to make it
easy for developers to interact with TM2 chains, providing a simplified API for 
account and transaction management. By doing all the heavy lifting behind the
scenes, `@gnolang/tm2-js-client` enables developers to focus on what really 
matters - building their dApps.

## Key Features

- JSON-RPC and WebSocket client support via a `Provider`
- Simple account and transaction management API with a `Wallet`
- Designed for easy extension for custom TM2 chains, such as [Gnoland](https://gno.land)

## Installation

To install `@gnolang/tm2-js-client`, use your preferred package manager:

```bash
yarn add @gnolang/tm2-js-client
```

```bash
npm install @gnolang/tm2-js-client
```

## Common Terminology

### Provider

A `Provider` is an interface that abstracts the interaction with the Tendermint2 
chain, making it easier for users to communicate with it. Rather than requiring 
users to understand which endpoints are exposed, what their return types are,
and how they are parsed, the `Provider` abstraction handles all of this behind 
the scenes. It exposes useful API methods that users can use and expects
concrete types in return.

Currently, the `@gnolang/tm2-js-client` package provides support for two
Provider implementations:

- `JSON-RPC Provider`: executes each call as a separate HTTP RPC call.
- `WS Provider`: executes each call through an active WebSocket connection,
which requires closing when not needed anymore.

### Signer

A `Signer` is an interface that abstracts the interaction with a single 
Secp256k1 key pair. It exposes methods for signing data, verifying signatures,
and getting metadata associated with the key pair, such as the address.

Currently, the `@gnolang/tm2-js-client` package provides support for two 
`Signer` implementations:

- `Key`: a signer that is based on a raw Secp256k1 key pair.
- `Ledger`: a signer that is based on a Ledger device, with all interaction
flowing through the user's device.

### Wallet

A `Wallet` is a user-facing API that is used to interact with an account. 
A `Wallet` instance is tied to a single key pair and essentially wraps the given
`Provider` for that specific account.

A wallet can be generated from a randomly generated seed, a private key, or
instantiated using a Ledger device.

Using the `Wallet`, users can easily interact with the Tendermint2 chain using
their account without having to worry about account management.
