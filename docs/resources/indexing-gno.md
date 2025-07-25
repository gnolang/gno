# Comprehensive Guide to Blockchain Indexing on Gno

Blockchains store data in an immutable, sequential chain of blocks containing raw transactions. While optimized for security and consensus, this structure creates significant data retrieval challenges:

 **Determining an address balance requires computationally expensive historical reprocessing** - every transaction ever associated with the address must be scanned, verified, and summed to calculate the current state.

### The Indexing Solution

Indexers resolve this paradox by transforming sequential blockchain data into query-optimized structures while preserving decentralization benefits:
1. **Eliminate reprocessing** - Balances become O(1) lookups
2. **Enable complex queries** - Temporal, relational, and semantic searches
3. **Unlock real-time use cases** - Wallets, explorers, analytics

## What is a Transaction Indexer?

### Core Definition

An indexer is a tool that transforms raw blockchain data to query-optimized databases by:

1. **Processing** each transaction as blocks are created
2. **Extracting** key relationships (addresses, contracts, events)
3. **Structuring** data for efficient retrieval
4. **Maintaining** real-time synchronization with the chain

This creates a "database view" of the blockchain while preserving its decentralized nature.

### Use case example
- Real-Time Wallet Balance Tracking
- DeFi Dashboard Analytics
- DAO Governance Monitoring
- Smart Contract Debugging

## [`tx-indexer`](https://github.com/gnolang/tx-indexer) - Official implementation of [Tendermint2 (TM2)](https://github.com/tendermint/tendermint2) Indexer

`tx-indexer` is a tool designed to index TM2 chain data (As GnoLand) and serve it over RPC, facilitating efficient data retrieval and management in TM2 networks.

### Key Features

- Support of GraphQL 
- **JSON-RPC 2.0 Specification Server**: Utilizes the JSON-RPC 2.0 standard for request / response handling.
- **HTTP and WebSocket Support**: Handles both HTTP POST requests and WS connections.
- **2-Way WS Communication**: Subscribe and receive data updates asynchronously over WebSocket connections.
- **Concurrent Chain Indexing**: Utilizes asynchronous workers for fast and efficient indexing. Data is available for serving as soon as it is fetched from the remote chain.
- **Embedded Database**: Features PebbleDB for quick on-disk data access and migration.

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` README.

## Implement Custom TM2 Indexer


