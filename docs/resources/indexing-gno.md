# Blockchain Indexing in Gno

Blockchains store data in an **immutable**, **sequential** chain of blocks containing raw transactions. While optimized for security and consensus, this structure creates significant data retrieval challenges.

However, tracking transaction metrics requires re-scanning **every block** of the blockchain, which is a **computationally expensive task** (and can be very costly). 
Some interesting tasks that demonstrate this challenge include:
- Calculating **total trading volume** over time periods
- Identifying the **largest transfers** in network history.

Without proper indexing, each query would require scanning the entire blockchain from genesis, making real-time applications depending on this data practically impossible.

### The Indexing Solution

**Indexers** resolve this paradox by storing all created blockchain data in a searchable database, enabling instant queries and unlock complex real-time use cases (e.g., "Find all 'addpkg' transaction of 'x' address").

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

### Use case: `send` transactions dashboard 

To demonstrate `tx-indexer` implementation workflow, we'll create a **real-time dashboard** that tracks GNOT transfers through the `send` transactions and identifies the biggest movers on the network.

In order, we will:
1. [Get the indexer running](#step-1-getting-started---launch-your-own-indexer)
2. [Query transaction data from the GraphQL interface hosted by `tx-indexer`](#step-2-querying-for-send-transactions)
3. [Process and analyze transactions to structure and sort relevant informations](#step-3-processing-the-transaction-data-with-go)
4. [Add real-time monitoring using WebSocket](#step-4-real-time-monitoring-with-websockets)
5. [Build a very simple dashboard that serves processed informations](#step-5-api-implementation-example)
6. [Store data permanently with SQLite](#step-6-storing-data-permanently-with-sqlite)

#### Step 1: Getting Started - Launch your own indexer

Before we can query transaction data, we need to get `tx-indexer` running.

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
./build/tx-indexer start --remote https://rpc.test7.testnets.gno.land --db-path indexer-db
```

Or if you prefer running directly with Go:
```bash
go run cmd/main.go cmd/start.go cmd/waiter.go start --remote https://rpc.test7.testnets.gno.land --db-path indexer-db
```

**What's happening here?**
- `--remote` tells the indexer which Gno chain to watch (we're using the test network)
- `--db-path` is where your indexed data gets stored locally
- The indexer will start catching up with all the blockchain data and keep syncing new blocks

**Your endpoints will be:**
- GraphQL playground: `http://localhost:8546/graphql` 
- WebSocket: `ws://localhost:8546/ws`
- JSON-RPC: `http://localhost:8546`

**Tip:** Run `./build/tx-indexer start --help` to see flags for rate limiting, chunk sizes, log levels, and more.

Once you see the indexer syncing blocks, you're ready to query transaction data! The indexer will keep running in the background, processing new transactions as they happen on the chain.

**Alternative: Use the hosted version**
If you just want to experiment without setting up your own indexer, you can use the hosted version at [Test 7 GraphQL Playground](https://indexer.test7.testnets.gno.land/graphql).

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

It is also possible to only query transaction initiated from a specific address:

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
- Each transaction has a unique `hash` (like a fingerprint)
- `block_height` tells you when it happened (higher = more recent)
- `amount` is in "ugnot" - divide by 1,000,000 to get GNOT (so "10,000,000ugnot" = 10 GNOT)
- Addresses starting with "g1" are Gno addresses

**Try it yourself!**
Copy that GraphQL query into the [playground](https://indexer.test7.testnets.gno.land/graphql) and execute it. You'll see real send transactions from the Gno network. 

#### Step 3: Processing the transaction data with Go

Now we have raw JSON data from our GraphQL query. Let's turn that into something useful for our dashboard. We'll write some Go code to parse the JSON and find the biggest transactions.

```go
package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Simplified transaction data structure
type Transaction struct {
	Hash   		string  // Transaction ID
	BlockHeight float64	// Block number 
	Amount 		float64 // Amount in ugnot
	From   		string  // Sender address
	To     		string  // Receiver address
}

// Step 1 - Parse JSON from GraphQL into our Transaction structs
func parseTransactions(jsonData []byte) []Transaction {
	var data map[string]interface{}
	json.Unmarshal(jsonData, &data)
	
	// Navigate through the JSON structure
	// {"data": { "getTransactions": [...]}}
	// Returns list of transactions (txs)
	txs := data["data"].(map[string]interface{})["getTransactions"].([]interface{})
	var transactions []Transaction
	
	for _, tx := range txs {
		txMap := tx.(map[string]interface{})
		hash := txMap["hash"].(string)
		
		// block_height comes as a number, not string
		blockHeight := txMap["block_height"].(float64)
		
		// Each transaction has messages - we want the first one
		msg := txMap["messages"].([]interface{})[0]
		msgMap := msg.(map[string]interface{})["value"].(map[string]interface{})
		
		// Extract informations
		amount := msgMap["amount"].(string)
		from := msgMap["from_address"].(string)
		to := msgMap["to_address"].(string)
		
		amountStr := amount[:len(amount)-5] // Remove "ugnot" suffix
		amountInt, _ := strconv.ParseFloat(amountStr, 64)

		transactions = append(transactions, Transaction{
			Hash:   	 hash,
			Amount:      amountInt,
			From:        from,
			To:          to,
			BlockHeight: blockHeight,
		})
	}
	return transactions
}

// Step 2 - Sort transactions by amount (biggest first)
func sortTransactions(transactions []Transaction) []Transaction {
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Amount > transactions[j].Amount
	})
	return transactions
}

// Step 3 - Show the transactions in a nice format
func displayTransactions(transactions []Transaction) {
	fmt.Println("🏆 Top GNOT Transactions:")
	for i, tx := range transactions {
		if i >= 5 { break } // Show top 5
		fmt.Printf("%d. %.2f GNOT from %s to %s\n", 
			i+1, tx.Amount, tx.From[:10]+"...", tx.To[:10]+"...")
	}
}

func main() {
	// This would be your actual JSON from the GraphQL query
	jsonData := `{ "data": { "getTransactions": [...] } }`
	
	// Process the data in 3 simple steps:
	transactions := parseTransactions([]byte(jsonData))  // 1. Parse JSON
	sorted := sortTransactions(transactions)             // 2. Sort by amount  
	displayTransactions(sorted)                          // 3. Display results 
}
```

**Try it out:**
Copy the JSON response from your GraphQL query into the `jsonData` variable and run this code. 

**Tip:** In a real application, you'd probably store this data in a database instead of just displaying it. But this gives you the foundation to work with transaction data in Go.

#### Step 4: Real-time monitoring with WebSockets

Instead of just querying historical data, we'll set up real-time monitoring to catch new transactions as they happen on the blockchain. We're building a live feed that alerts us whenever someone sends GNOT on the network. It can enhance our application by adding live notification for example.

**GraphQL Subscription (instead of query):**
First, we need to change our GraphQL from a one-time `query` to a streaming `subscription`:

```graphql
subscription MonitorNewTransactions {
  transactions(
    filter: {
      success: { eq: true }      # Only successful transactions
      messages: {
        value: {
          BankMsgSend: {}         # Only send transactions
        }
      }
    }
  ) {
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

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	
	"github.com/gorilla/websocket"
)

func main() {
	fmt.Println("🔗 Connecting to tx-indexer WebSocket...")
	
	// Connect to the indexer's WebSocket endpoint
	u := url.URL{Scheme: "ws", Host: "localhost:8546", Path: "/ws"}
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("❌ WebSocket connection failed:", err)
	}
	defer conn.Close()
	
	fmt.Println("✅ Connected! Setting up subscription...")
	
	// Send our subscription request
	subscription := map[string]interface{}{
		"type": "start",
		"payload": map[string]interface{}{
			"query": `subscription {...}`,
		},
	}
	
	subscriptionBytes, _ := json.Marshal(subscription)
	conn.WriteMessage(websocket.TextMessage, subscriptionBytes)
	
	fmt.Println("📡 Listening for new transactions...")
	
	// Listen for new transactions forever
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			break
		}
		
		// Parse the incoming transaction
		var response map[string]interface{}
		json.Unmarshal(message, &response)
		
		if response["type"] == "data" {
			fmt.Println("🔥 NEW TRANSACTION DETECTED!")
			
			// You can process this using our parseTransactions() function from Step 3
			// or handle it directly here for real-time alerts
			
			// Example: Extract amount for quick alert
			if data, ok := response["payload"].(map[string]interface{}); ok {
				// Process the transaction data...
				fmt.Printf("📊 Transaction data: %+v\n", data)
			}
		}
	}
}
```

#### Step 5: API Implementation Example

Let's create a simple JSON API that serves transaction data. This is great for building mobile apps, frontends, or integrating with other services! 🚀

**What we're building:**
A lightweight REST API that returns transaction statistics as JSON. 

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Simple dashboard server
func main() {
	http.HandleFunc("/stats", handleStats)
	fmt.Println("📊 Dashboard running on http://localhost:8080/stats")
	http.ListenAndServe(":8080", nil)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	// Sample data (in real app, get from indexer)
	jsonData := `{ "data": { "getTransactions": [...] } }`
	
	// Process data
	transactions := parseTransactions([]byte(jsonData))
	sorted := sortTransactions(transactions)
	
	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_transactions": len(sorted),
		"biggest_amount":     sorted[0].Amount,
		"top_transactions":   sorted[:2], // Top 2
	})
}
```
#### Step 6: Storing data permanently with SQLite

Our dashboard only works with data in memory - nothing is persisted. Let's fix that by adding a database to store transactions permanently.
We'll use SQLite for example purpose.

**What we're building:**
A simple web server that provides transaction data through a REST API endpoint, making it easy to integrate with web frontends, mobile apps, or other services.

**Setting up the database:**

```go
import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"  // SQLite driver
)


// Create our database and transactions table
func setupDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./transactions.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	
	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			hash TEXT PRIMARY KEY,     -- Unique transaction ID
			amount REAL,              -- Amount in GNOT (converted from ugnot)
			from_addr TEXT,           -- Sender address
			to_addr TEXT,             -- Receiver address
			block_height INTEGER,     -- When it happened
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	
	return db
}

// Save a single transaction to our database
func insertTransaction(db *sql.DB, tx Transaction) error {

	_, err := db.Exec(`
		INSERT OR IGNORE INTO transactions 
		(hash, amount, from_addr, to_addr, block_height) 
		VALUES (?, ?, ?, ?, ?)`, 
		tx.Hash, tx.Amount, tx.From, tx.To, tx.BlockHeight,
	)
	
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %v", err)
	}
	
	return nil
}

// Get the biggest transactions from our database
func getTopTransactions(db *sql.DB, limit int) ([]Transaction, error) {
	rows, err := db.Query(`
		SELECT hash, amount, from_addr, to_addr, block_height 
		FROM transactions 
		ORDER BY amount DESC 
		LIMIT ?`, limit)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var transactions []Transaction
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.Hash, &tx.Amount, &tx.From, &tx.To, &tx.BlockHeight)
		if err != nil {
			continue
		}
		transactions = append(transactions, tx)
	}
	
	return transactions, nil
}

// Save multiple transactions to the database
func saveTransactionsToDB(db *sql.DB, transactions []Transaction) {
	for _, tx := range transactions {
		err := insertTransaction(db, tx)
		if err != nil {
			log.Printf("Failed to save transaction %s: %v", tx.Hash, err)
		} else {
			fmt.Printf("✅ Saved transaction %s (%.2f GNOT)\n", tx.Hash, tx.Amount/1000000)
		}
	}
	fmt.Printf("💾 Processed %d transactions\n", len(transactions))
}

func main() {
	// Set up our database
	db := setupDB()
	defer db.Close()
	
	// Get new transactions from the indexer
	jsonData := `{"data": {"getTransactions": [...]}}`
	transactions := parseTransactions([]byte(jsonData))
	
	// Save them to our database
	saveTransactionsToDB(db, transactions)
	
	// Get the top 10 transactions ever recorded
	topTx, _ := getTopTransactions(db, 10)
	fmt.Println("🏆 Top 10 biggest transactions in our database:")
	displayTransactions(topTx)
}
```

Your transaction data is now persistent and ready for productions!

## Examples

For more example of queries, refers to the [tx-indexer](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#examples) documentation.

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` repository.
