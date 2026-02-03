# Application BlockChain Interface (ABCI)

Blockchains are systems for multi-master state machine replication.
**ABCI** is an interface that defines the boundary between the replication engine (the blockchain),
and the state machine (the application).
Using a socket protocol, a consensus engine running in one process
can manage an application state running in another.

Previously, the ABCI was referred to as TMSP.

The community has provided a number of additional implementations, see the [Tendermint Ecosystem](https://tendermint.com/ecosystem)

## Specification

A detailed description of the ABCI methods and message types is contained in:

- [A protobuf file](./types/types.proto)
- [A Go interface](./types/application.go)

## Protocol Buffers

To compile the protobuf file, first install the protobuf command-line tools, e.g. with `brew install protobuf`. Then run (from the root of the repo):

```
make -C misc/devdeps
make -C misc/genproto
```

See `protoc --help` and [the Protocol Buffers site](https://developers.google.com/protocol-buffers) for details on compiling for other languages.
