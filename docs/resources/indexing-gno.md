# Blockchain Indexing in Gno

Blockchains store data in an immutable, sequential chain of blocks containing raw transactions. While optimized for security and consensus, this structure creates significant data retrieval challenges.

We want to track all transactions initiated by a specific address. To process them, we are required to re-scan **every blocks manually or on-chain**, which is a **computationally expensive tasks** (which can be very costly).

### The Indexing Solution

**Indexers** resolve this paradox by capturing newly-created blockchain data in a searchable database, enabling instant queries and unlock complex real-time use cases (e.g., "Find all 'addpkg' transaction of 'x' address").

To do so, an indexer works by:
1. **Processing** each transaction as blocks are created
2. **Extracting** key relationships (addresses, contracts, events)
3. **Structuring** data for efficient retrieval
4. **Maintaining** real-time synchronization with the chain

This creates a "database view" of the blockchain while preserving its decentralized nature.

## [`tx-indexer`](https://github.com/gnolang/tx-indexer): The official [TM2](https://github.com/tendermint/tendermint2) Indexer

`tx-indexer` is the reference implementation for Tendermint2 chains like Gno.land, providing:
- **Dual-protocol API**: JSON-RPC 2.0 + GraphQL
- **Transport Support**: HTTP + WebSocket
- **High Performance**: Concurrent block processing pipeline
- **Embedded Storage**: PebbleDB engine

### **Query Capabilities**
The query system enables complex data retrieval with multiple filter conditions, pagination, and relationship traversal.

#### GraphQL Example
```graphql
query {
  getBlocks(
    where: {
        {
          height: {
            eq: 10 
          }
        }
    }
  ) {
    hash       
    height     
    time       
    num_txs    
    total_txs  
    txs {
      content_raw  
    }
  }
}
```

#### JSON-RPC Example

```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "getBlock",
  "params": [
    "10"
  ]
}
```

### **Subscription System**
The subscription system enables instant notifications for on-chain activity. The architecture is WebSocket-based, which eliminates the need for constant polling.

#### GraphQL Example
```graphql
subscription {
  blocks(filter: {}) {
    height
    version
    chain_id
    time
    proposer_address_raw
  }
}
```

#### JSON-RPC Example
```json
{
  "id": 1,
  "jsonrpc": "2.0",
  "method": "subscribe",
  "params": [
    "newHeads"
  ]
}
```

For more example, refers to the [tx-indexer](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#examples) documentation.

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` repository.
