# Gno.land Go Client

The Gno.land Go client is a dedicated library for interacting seamlessly with the Gno.land RPC API.
This library simplifies the process of querying or sending transactions to the Gno.land RPC API and interpreting the responses.

Documentation may be found at [gnoclient](../../../docs/reference/gnoclient).

## Installation

Integrate this library into your Go project with the following command:

    go get github.com/gnolang/gno/gno.land/pkg/gnoclient

## Development Plan

The roadmap for the Gno.land Go client includes:

- **Initial Development:** Kickstart the development specifically for Gno.land. Subsequently, transition the generic functionalities to other modules like `tm2`, `gnovm`, `gnosdk`.
- **Integration:** Begin incorporating this library within various components such as `gno.land/cmd/*` and other external clients, including `gnoblog-client`, the Discord community faucet bot, and [GnoMobile](https://github.com/gnolang/gnomobile).
- **Enhancements:** Once the generic client establishes a robust foundation, we aim to utilize code generation for contracts. This will streamline the creation of type-safe, contract-specific clients.

