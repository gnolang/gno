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

An indexer is a tool that transforms raw blockchain data query-optimized databases by:

1. **Processing** each transaction as blocks are created
2. **Extracting** key relationships (addresses, contracts, events)
3. **Structuring** data for efficient retrieval
4. **Maintaining** real-time synchronization with the chain

This creates a "database view" of the blockchain while preserving its decentralized nature.


