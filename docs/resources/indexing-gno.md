# Blockchain Indexing in Gno

Blockchains store data in an **immutable**, **sequential** chain of blocks containing raw transactions. While optimized for security and consensus, this structure creates significant data retrieval challenges.

However, tracking transaction metrics requires re-scanning **every block** of the blockchain, which is a **computationally expensive task** (and can be very costly). 
Some interesting tasks that demonstrate this challenge include:
- Calculating **total trading volume** over time periods
- Identifying the **largest transfers** in network history.

Without proper indexing, each query would require scanning the entire blockchain from genesis, making real-time applications depending on this data practically impossible.

### The Indexing Solution

**Indexers** resolve this paradox by storing all created blockchain data in a searchable database, enabling instant queries and unlock complex real-time use cases (e.g., "Find all 'addpkg' transaction of 'x' address").

It works by **processing** transactions as blocks are created and **extracting** key relationships (addresses, contracts, events) to store in a structured database, while **maintaining** real-time synchronization with the chain. Using them, it enables the creation of **structured datasets** for analytics and application development.

This creates a "database view" of the blockchain while preserving its decentralized nature.

## [`tx-indexer`](https://github.com/gnolang/tx-indexer): The official [TM2](https://github.com/tendermint/tendermint2) Indexer

`tx-indexer` is the reference indexer implementation for Tendermint2 chains like Gno.land, providing:
- **Dual-protocol API**: JSON-RPC 2.0 + GraphQL
- **Transport Support**: HTTP + WebSocket
- **High Performance**: Concurrent block processing pipeline
- **Embedded Storage**: PebbleDB engine

## Examples
To find query examples, refer to the [tx-indexer](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#examples) documentation.

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` repository.

### Use case: `send` transactions dashboard 

To demonstrate the usage of the indexer, a tutorial is available in the official `tx-indexer` repository: [How to Create an Indexer](https://github.com/gnolang/tx-indexer/blob/main/docs/how-to-create-an-indexer.md).
