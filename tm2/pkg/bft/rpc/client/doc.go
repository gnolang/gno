// Package client provides a general purpose interface (Client) for connecting
// to a tendermint node, as well as higher-level functionality.
//
// The main implementation for production code is client.HTTP, which
// connects via http to the jsonrpc interface of the tendermint node.
//
// For connecting to a node running in the same process (eg. when
// compiling the abci app in the same process), you can use the client.Local
// implementation.
//
// For mocking out server responses during testing to see behavior for
// arbitrary return values, use the mock package.
//
// In addition to the Client interface, which should be used externally
// for maximum flexibility and testability, and two implementations,
// this package also provides helper functions that work on any Client
// implementation.
package client
