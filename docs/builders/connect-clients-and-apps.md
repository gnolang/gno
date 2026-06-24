# Connecting Clients and Applications to Gno.land

This guide explains how to connect external applications to Gno.land networks
using clients in different languages. You'll learn how to use the RPC endpoints
to query the blockchain and submit transactions.

## Available Clients

Gno.land provides several client libraries to interact with the blockchain:

- **[gnoclient](https://gnolang.github.io/gno/github.com/gnolang/gno/gno.land/pkg/gnoclient.html)** - The official Go client for connecting to Gno.land networks
- **[gno-js-client](https://github.com/gnolang/gno-js-client)** - A JavaScript client for building web applications
- **[tm2-js-client](https://github.com/gnolang/tm2-js-client)** - A lower-level JavaScript client for direct RPC access

## Understanding Gno.land's RPC Interface

Gno.land networks expose several RPC endpoints that allow you to:

1. **Query blockchain state** - Retrieve account information, package data, and more
2. **Submit transactions** - Send GNOT tokens, call realm functions, and deploy code
3. **Subscribe to events** - Get real-time updates about blockchain activity

All RPC endpoints for each network can be found in the [Networks documentation](../resources/gnoland-networks.md).

### Query Endpoints

There are two families of query endpoints:

- **Text-oriented** (`vm/qrender`, `vm/qeval`, `vm/qfile`, etc.) — return human-readable strings, suitable for CLI tools like `gnokey`. See [Interacting with gnokey](../users/interact-with-gnokey.md#querying-a-gnoland-network).
- **JSON/structured** (`vm/qeval_json`, `vm/qpkg_json`, `vm/qobject_json`, `vm/qtype_json`) — return Amino JSON, suitable for programmatic access by frontends and tools. See [Querying On-Chain State (JSON APIs)](query-state-api.md).
