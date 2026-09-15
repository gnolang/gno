# Comparison of the various ways to communicate with Gno.land

This document is an overview of the various methods to interact with gnoland. This is mainly for developers writing applications.

## JSON-RPC

**What is it?** The RPC interface is the base-level method to interact with the gno.land blockchain over an HTTP-RPC connection. Requests and responses are JSON-RPC, and the result payloads are encoded with Amino JSON. See [JSON-RPC endpoints](./rpc-endpoints.md) for the methods, the encoding rules and the limits.

**When to use it?** The HTTP-RPC interface is used under-the-hood by `gnokey` and all the other methods. Queries are simple enough to issue over a plain HTTP connection. Submitting a transaction means building and signing an Amino-encoded `std.Tx`, which is why most applications use one of the following methods instead.

## Protobuf message definitions

**What is it?** The core Gno codebase uses the light-weight [Amino package](https://github.com/gnolang/gno/tree/master/tm2/pkg/amino) to encode messages. Amino can generate Protobuf definitions for those message types, which a non-Go application can use to build and parse them.

**When to use it?** Amino is a Go package, so an application written in another language cannot import it. Generated Protobuf definitions cover the message types — but the transport is still the JSON-RPC interface above. There is no gRPC server: the `grpc_laddr` setting present in `config.toml` is not read by anything.

## Go code with `gnoclient` and `crypto/keys`

**What is it?** The [`gnoclient`](https://github.com/gnolang/gno/blob/master/gno.land/pkg/gnoclient/docs/using-gnoclient.md) and [`crypto/keys`](https://github.com/gnolang/gno/tree/master/tm2/pkg/crypto/keys) packages are part of the core Gno codebase, written in Go. The `gnoclient` API is for querying the blockchain, calling realm functions, and other basic interactions. The `crypto/keys` API is for generating cryptographical keys and managing the local keys database.

**When to use it?** If your app is written in Go, you can use these packages (as `gnokey` and other command-line apps do).

## Gno Native Kit

**What is it?** [Gno Native Kit](https://github.com/gnolang/gnonative) is a framework that allows developers to build and port gno.land dApps written in the dApp's native language such as TypeScript, Java or C#. It is a wrapper on top of the `gnoclient` and `crypto/keys` which allows your app to use the functionality in these core Gno packages.

**When to use it?** If your app is not written in Go (for example, React Native on mobile), then Gno Native Kit can provide access to all the functionality.

## tm2-js / gno-js

**What is it?** [tm2-js](https://github.com/gnolang/tm2-js-client) is a JavaScript/TypeScript client implementation for Tendermint2-based chains (such as Gno.land). [gno-js](https://github.com/gnolang/gno-js-client/blob/main/README.md) is an extension of tm2-js, but with Gno-specific functionality.

gno-js is designed to make it easy for developers to interact with the Gno.land blockchain by providing a simplified API for account and transaction management.

**When to use it?** When developing a browser app without access to the Go runtime (which is used by all the other methods).

## `tx-indexer` with GraphQL

**What is it?** [tx-indexer](https://github.com/gnolang/tx-indexer) is a tool
which makes an indexed database of the transactions on the gno.land blockchain.
It comes with a GraphQL interface which you can use directly in a web page or from your app through WebSocket.

**When to use it?** If your app needs to do CPU-intensive processing of your realm data, then it is better to have a dedicated service running on a server instead of on-chain processing. `tx-indexer` provides efficient read-only queries of just the data that needs to be processed. Then your end-user app can fetch the result from the service.

|                                        | JSON-RPC                                                                                                                                                     | Protobuf message definitions                           | Go code with gnoclient crypto/keys       | Gno Native Kit                                                                                                          | tm2-js / gno-js   | tx-indexer with GraphQL                                                |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------- | -----------------------------------------| ----------------------------------------------------------------------------------------------------------------------- | ----------------- | ---------------------------------------------------------------------- |
| **App programming language**            | Any language with an HTTP library                                                                                                                            | Any language with a Protobuf and an HTTP library         | Go                                       | Go, JavaScript, Java, Swift, C#, via the kit's own gRPC                                                                | JavaScript        | Any language with a WebSocket library                                  |
| **Platform**                           | Any constrained platform where you can't use any of the other methods                                                                                        | Any platform that has a Protobuf library                | Desktop utilities                        | Mobile, desktop apps                                                                                                    | Browser           | Off-chain service apps                                                 |
| **Uses core Gno code?**                | No                                                                                                                                                           | No                                                      | Yes                                      | Yes                                                                                                                     | No                | Yes                                                                    |
| **Requires programming in Protobuf?** | No                                                                                                                                                           | Yes                                                     | Yes                                      | Yes                                                                                                                     | Yes               | Yes                                                                    |
| **Supports browser applications**      | Yes                                                                                                                                                          | No                                                      | No                                       | No                                                                                                                      | No                | N/A                                                                    |
| **Recommended for**                    | Situations where you can't use any of the other methods due to significant constraints. (e.g. a Raspberry Pi querying the blockchain for an account balance) | Apps which can't link to the core Go code               | Basic Go apps (like command-line utils)  | Robust mobile and desktop apps where devs use gRPC to bypass the complexity of interacting directly with the blockchain | Browser apps      | Complex read-only interactions (e.g.: compute the home feed of a dApp) |
