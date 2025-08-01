# Blockchain Indexing in Gno

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

### Key Capabilities

#### **Real-Time Indexing**  
Enables **live tracking** of on-chain activity as it happens. It is used for dashboards, alerts, or applications requiring instant updates.  
**Example**: Building a **search tool for decentralized social posts**. Instead of scanning every block manually or on-chain (which can be very costly), real-time indexing automatically captures new posts and updates a searchable database, enabling instant full-text search (e.g., "Find all posts mentioning 'AI' in the last hour").  

---

#### **Historical Indexing**  
Stores and organizes past transactions or events in a queryable format. Useful for compliance, analytics, or reconstructing activity over time.  
**Example**: Creating a **[username](https://gno.land/r/gnoland/users/v1) search engine**. Historical indexing scans all past profile creations and builds a searchable index, allowing queries like "Show all usernames containing 'Alice' registered before 2024."

---

#### **Event Extraction**  
Filters and extracts specific on-chain events (e.g., token mints, governance votes) for custom triggers or notifications.  
**Example**: Watching for **new comments on a decentralized blog**. Event extraction could monitor realm for interactions with specific posts (e.g., "Notify me when someone replies to my post on X board") without processing irrelevant transactions.

---

#### **State Snapshots**  
Captures the state of the indexed blockchain (e.g., realm storage, NFT ownership) at specific blocks. Critical for proving past conditions or auditing.  
**Example**: Auditing a **decentralized social graph**. A snapshot at block #X could prove "How many users followed Profile Y on January 1st? Essential for token airdrops or reputation systems relying on historical data.  


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
