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

### Key Capabilities
| Capability           | Use Cases                                                                 |
|----------------------|---------------------------------------------------------------------------|
| **Real-Time Indexing** | • Wallet balance updates<br>• Exchange transaction monitoring<br>• Live dashboards |
| **Historical Indexing** | • Regulatory compliance<br>• Smart contract debugging<br>• Chain analytics |
| **Event Extraction** | • DeFi liquidation alerts<br>• DAO governance tracking<br>• Custom notifications |
| **State Snapshots**  | • NFT provenance tracking<br>• DeFi performance metrics<br>• Historical queries |


## [`tx-indexer`](https://github.com/gnolang/tx-indexer): The official [TM2](https://github.com/tendermint/tendermint2) Indexer

`tx-indexer` is the reference implementation for Tendermint2 chains like Gno.land, providing:
- Dual-protocol API server: **JSON-RPC 2.0** + **GraphQL**
- **HTTP and WebSocket Support**
- **Concurrent block** processing pipeline
- **PebbleDB**: embedded storage engine

### **Query Capabilities**

#### GraphQL
```graphql
query {
    transactions(
        filter: {
            sender: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
            fromHeight: 10000
            toHeight: 20000
        }
    ) {
        hash
        timestamp
        messages {
            type
            data
        }
    }
}
```

#### JSON-RPC

### **Subscription System**
The WebSocket-based subscription system enables instant notifications for on-chain activity. This architecture eliminates the need for constant polling.

#### GraphQL
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

#### JSON-RPC
```go
// WebSocket subscription example
conn.Subscribe("address:g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", func(event Event) {
    fmt.Printf("New transaction: %v", event.TxHash)
})
```

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` README.

## Building Custom Indexers

### Step-by-Step Development Guide

1. **Define Your Data Model**
   ```go
   type NFTIndex struct {
       TokenID     string
       Owner       string
       MetadataURI string
       MintHeight  int64
   }
   ```

2. **Implement the Processor**
   ```go
   func (n *NFTIndexer) ProcessBlock(block *types.Block) error {
       for _, tx := range block.Txs {
           if isNFTMint(tx) {
               nft := extractNFT(tx)
               n.store.Put(nft.TokenID, nft)
               n.ownerIndex.Add(nft.Owner, nft.TokenID)
           }
       }
       return nil
   }
   ```

3. **Add Query Support**
   ```go
   func (n *NFTIndexer) GetNFTsByOwner(owner string) ([]NFTIndex, error) {
       tokenIDs := n.ownerIndex.Get(owner)
       var nfts []NFTIndex
       for _, id := range tokenIDs {
           nft, _ := n.store.Get(id)
           nfts = append(nfts, nft)
       }
       return nfts, nil
   }
   ```

