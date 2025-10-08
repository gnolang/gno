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


To demonstrate `tx-indexer` implementation workflow, we'll create a **real-time dashboard** that tracks GNOT transfers through the `send` transactions and identifies the biggest movers on the network.
We'll work during this implementation in local environment, but the same concepts apply when working over a network environment.

In order, we will:
1. [Get the indexer running](#step-1-getting-started---launch-your-own-indexer)
2. [Query transaction data from the GraphQL interface hosted by `tx-indexer`](#step-2-querying-for-send-transactions)
3. [Process and analyze transactions to structure and sort relevant informations](#step-3-processing-the-transaction-data-with-go)
4. [Add real-time monitoring using WebSocket](#step-4-real-time-monitoring-with-websockets)
5. [Build a very simple dashboard that serves processed informations](#step-5-api-implementation-example)
6. [Store data permanently with SQLite](#step-6-storing-data-permanently-with-sqlite)

This tutorial will go through each step, but you can find the full code at [Gnoverse's repository]().

**Hosted version:**
If you just want to experiment without setting up your own indexer, you can use the hosted version at [Test 8 GraphQL Playground](https://indexer.test8.testnets.gno.land/graphql).

#### Step 1: Getting Started - Launch your own indexer

Before we can query transaction data, we need to get `tx-indexer` running. This means setting up a local process and allowing it to fetch and index data from a live Gno.land chain. This usually takes some time, but for this tutorial, we will index our own, local chain, which will be extremely quick. 

**Quick Setup:**

1. **Clone `tx-indexer`**
```bash
git clone https://github.com/gnolang/tx-indexer.git
cd tx-indexer
```

2. **Build it**
```bash
make build
```

3. **Start indexing** 
```bash
# For local development
./build/tx-indexer start --remote http://127.0.0.1:26657 --db-path indexer-db

# For testnets development
./build/tx-indexer start --remote https://rpc.test8.testnets.gno.land --db-path indexer-db
```

Or if you prefer avoiding installation and running the services directly with Go:
```bash
go run cmd/main.go cmd/start.go cmd/waiter.go start --remote http://127.0.0.1:26657 --db-path indexer-db
```

**What's happening here?**
- `--remote` tells the indexer which Gno chain to watch (we're using the test network)
- `--db-path` is where your indexed data gets stored locally
- The indexer will start catching up with all the blockchain data and keep syncing new blocks

**Your endpoints will be:**
- GraphQL playground: `http://localhost:8546/graphql` 
- GraphQL query: `http://localhost:8546/graphql/query`
- GraphQL websocket: `ws://localhost:8546/graphql/query`
- WebSocket: `ws://localhost:8546/ws`
- JSON-RPC: `http://localhost:8546`

**Tip:** Run `./build/tx-indexer start --help` to see flags for rate limiting, chunk sizes, log levels, and more.

Once you see the indexer syncing blocks, you're ready to query transaction data! The indexer will keep running in the background, processing new transactions as they happen on the chain.

#### Step 2: Querying for send transactions

Now that your indexer is running, let's get some actual transaction data. We'll use GraphQL to ask for all the send transactions on the network.

**What we're looking for:**
We want to find every time someone sent GNOT to someone else. In Gno.land terms, these are called "BankMsgSend" transactions.

```graphql
query GetSendTransactions {
   getTransactions(
    where: {
      # Only show successful transactions 
      success: { eq: true }
      # Focus on send transactions specifically
      messages: {
        value: {
          BankMsgSend: {}  # This is the "send money" transaction type
        }
      }
    }
  ) {
    # What data do we want back?
    hash           # Unique transaction ID
    block_height   # Which block this happened in
    messages {
      value {
        ... on BankMsgSend {
          from_address  # Who sent it
          to_address    # Who received it
          amount        # How much (in ugnot)
        }
      }
    }
  }
}
```

**Tip:**
In GraphQL, press `Ctrl` + `Space` to autocomplete available fields through the built-in documentations.

**What you'll get back:**
The indexer will return a JSON response that looks like this:

```json
{
  "data": {
    "getTransactions": [
      {
        "hash": "cYC5jqY1JePb6UccqRj+w3GIZBHCt77d3eWVcDufH9o=",
        "block_height": 55,
        "messages": [
          {
            "value": {
              "from_address": "g148583t5x66zs6p90ehad6l4qefeyaf54s69wql",
              "to_address": "g1ker4vvggvsyatexxn3hkthp2hu80pkhrwmuczr",
              "amount": "10000000ugnot"
            }
          }
        ]
      },
      {
        "hash": "UP5WbGcTDnqFOXqDzdwNDdM1RGrjfWz/FWZiVrtTANo=",
        "block_height": 321,
        "messages": [
          {
            "value": {
              "from_address": "g148583t5x66zs6p90ehad6l4qefeyaf54s69wql",
              "to_address": "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2",
              "amount": "15000000ugnot"
            }
          }
        ]
      }
    ]
  }
}
```

It is also possible to only query transaction initiated from a specific address, where `$address` is a variable:

```graphql
# Add a variable to GetSendTransactions
query GetSendTransactions($address: String!) {
   getTransactions(
    where: {
      success: { eq: true }
      messages: {
        value: {
          BankMsgSend:{
            from_address: {eq: $address} # from_address == $address
          }
        }
      }
    }
  )  {
    hash
    block_height
    messages {
      value {
        ... on BankMsgSend {
        from_address
        to_address
        amount
        }
      }
    }
  }
}
```

**Understanding the data:**
- Each transaction has a unique `hash`
- `block_height` tells you when it happened
- `amount` is in "ugnot" - divide by 1,000,000 to get GNOT (`1_000_000ugnot = 1 GNOT`)
- Addresses starting with "g1" are Gno addresses

**Try it yourself!**
Copy that GraphQL query into the [playground](https://indexer.test8.testnets.gno.land/graphql) and execute it. You'll see real send transactions from the Gno network. 

#### Step 3: Processing the transaction data with Go

Now we have raw JSON data from our GraphQL query. Let's turn that into something useful for our dashboard. We'll write some Go code to parse the JSON and find the biggest transactions.

[embedmd]:# (../_assets/tx-indexer-example/process.go go)
```go
```

**Try it out:**
Copy the JSON response from your GraphQL query into the `jsonData` variable and run this code. 

#### Step 4: Real-time monitoring with WebSockets

Instead of just querying historical data, we'll set up real-time monitoring to catch new transactions as they happen on the blockchain. We're building a live feed that alerts us whenever someone sends GNOT on the network. It can enhance our application by adding live notification for example.

**GraphQL Subscription (instead of query):**
First, we need to change our GraphQL from a one-time `query` to a streaming `subscription`:

```graphql
subscription GetSendTransactions {
   getTransactions(
    where: {
      success: { eq: true }
      messages: {
        value: {
          BankMsgSend:{
          }
        }
      }
    }
  )  {
    hash
    block_height
    messages {
      value {
        ... on BankMsgSend {
        from_address
        to_address
        amount
        }
      }
    }
  }
}
```

**The difference:** Instead of getting all past transactions at once, this subscription will send us each new transaction as it gets added to the blockchain.

**Go WebSocket Client:**
Now we need a WebSocket client to receive these real-time updates:

[embedmd]:# (../_assets/tx-indexer-example/websocket.go go)
```go
```

**Testing our real-time monitoring:**

Now that we have our WebSocket listener running, let's generate some test transactions to see it work! You can try it by running:

```bash
# Send 1000 ugnot (0.001 GNOT) from one address to another
gnokey maketx send \
	-gas-fee 1000000ugnot \
	-gas-wanted 500000 \
	-send 1000ugnot \
	-broadcast \
	-chainid "dev" \
	-remote "tcp://127.0.0.1:26657" \
	-to g1destination_address_here \
	g1your_address_here
```

Your WebSocket listener is now detecting send transactions in real-time!

#### Step 5: API Implementation Example

Let's create a simple JSON API that serves transaction data. This is great for building applications, frontends, or integrating with other services.

**What we're building:**
A lightweight API that returns transaction statistics as JSON. 

[embedmd]:# (../_assets/tx-indexer-example/api.go go)
```go
```

#### Bonus: Storing data permanently with database

Our dashboard currently only works with data in memory - nothing is persisted. To store transactions permanently, you can use a database like SQLite.

We won't implement database persistence in this tutorial since it's unrelated to using the indexer itself, but it's an essential component for production deployments.
To see a complete implementation example, visit [Gnoverse's repository]().

### Resources for Continued Learning

- **[tx-indexer Documentation](https://github.com/gnolang/tx-indexer)** - Official reference and advanced configuration
- **[GraphQL Best Practices](https://graphql.org/learn/best-practices/)** - Advanced querying techniques and optimization
