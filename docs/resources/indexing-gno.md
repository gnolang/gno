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
./build/tx-indexer start --remote https://rpc.test7.testnets.gno.land --db-path indexer-db
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
// This struct represents a cleaned-up version of the raw GraphQL data
type Transaction struct {
	Hash   		string  // Transaction ID - unique identifier for each transaction
	BlockHeight float64	// Block number - tells us when this transaction happened
	Amount 		float64 // Amount in ugnot (1 GNOT = 1,000,000 ugnot)
	From   		string  // Sender address - who initiated the transaction
	To     		string  // Receiver address - who received the funds
}

// Step 1 - Parse JSON from GraphQL into our Transaction structs
// This function takes raw JSON from the indexer and converts it to Go structs
func parseTransactions(jsonData []byte) []Transaction {
	var data map[string]interface{}
	json.Unmarshal(jsonData, &data)
	
	// Navigate through the JSON structure
	// GraphQL returns: {"data": { "getTransactions": [...]}}
	// We need to drill down to get the actual transaction array
	txs := data["data"].(map[string]interface{})["getTransactions"]
	var transactions []Transaction
	
	// Handle both single transaction and array of transactions
	// The indexer might return either format depending on the query
	var txList []interface{}
	switch v := txs.(type) {
	case []interface{}:
		// Multiple transactions - this is the typical case
		txList = v
	case map[string]interface{}:
		// Single transaction - wrap it in an array for consistent processing
		txList = []interface{}{v}
	default:
		// Unexpected format - return empty slice to avoid crashes
		return transactions
	}
		
	// Process each transaction in the response
	for _, tx := range txList {
		txMap := tx.(map[string]interface{})
		
		// Extract basic transaction info
		hash := txMap["hash"].(string)
		blockHeight := txMap["block_height"].(float64)
		
		// Navigate to the message data (transaction details)
		// Each transaction has "messages" array containing the actual operations
		msg := txMap["messages"].([]interface{})[0]
		msgMap := msg.(map[string]interface{})["value"].(map[string]interface{})
		
		// Extract the send transaction details
		amount := msgMap["amount"].(string)        
		from := msgMap["from_address"].(string)    
		to := msgMap["to_address"].(string)        
		
		// Convert amount from string to number
		// Remove "ugnot" suffix and parse as float
		amountStr := amount[:len(amount)-5] // Remove last 5 chars ("ugnot")
		amountInt, _ := strconv.ParseFloat(amountStr, 64)

		// Create our clean Transaction struct
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
// This helps us identify the largest transfers on the network
func sortTransactions(transactions []Transaction) []Transaction {
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Amount > transactions[j].Amount
	})
	return transactions
}

// Step 3 - Show the transactions in a nice format
// Convert raw data into human-readable output
func displayTransactions(transactions []Transaction) {
	fmt.Println("Top GNOT Transactions:")
	
	for i, tx := range transactions {
		if i >= 5 { break } // Limit to top 5
		
		gnotAmount := tx.Amount 
		fmt.Printf("%d. %.2f uGNOT from %s to %s\n", 
			i+1, gnotAmount, tx.From, tx.To)
	}
}

func main() {
	// This would be your actual JSON from the GraphQL query
	// In a real app, you'd get this from an HTTP request to the indexer
	jsonData := `{ "data": { "getTransactions": [...] } }`
	
	// Process the data in 3 simple steps:
	transactions := parseTransactions([]byte(jsonData))  // 1. Parse JSON into structs
	sorted := sortTransactions(transactions)             // 2. Sort by amount (largest first)
	displayTransactions(sorted)                          // 3. Display results nicely
	
	// At this point, you have clean, sorted transaction data ready for:
	// - Saving to a database
	// - Serving via an API
	// - ...
}
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

```go
import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// Import the previously created function
func parseTransactions(jsonData []byte) []Transaction {
	...
}

func displayTransactions(transactions []Transaction) {
	...
}

func main() {
	fmt.Println("🔗 Connecting to tx-indexer WebSocket...")
	
	// Build WebSocket URL - replace localhost:8546 with your indexer's address
	u := url.URL{Scheme: "ws", Host: "localhost:8546", Path: "/graphql/query"}
	
	// Establish WebSocket connection with GraphQL-WS protocol
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("❌ WebSocket connection failed:", err)
	}
	defer conn.Close() // Clean up connection when function exits
	
	fmt.Println("✅ Connected! Initializing GraphQL-WS connection...")
	
	// Step 1: Send connection_init message
	// GraphQL-WS protocol requires this handshake before subscriptions
	initMsg := map[string]interface{}{
		"type": "connection_init",
	}
	initBytes, _ := json.Marshal(initMsg)
	conn.WriteMessage(websocket.TextMessage, initBytes)
	
	// Step 2: Wait for connection_ack from server
	// Server must acknowledge our connection before we can subscribe
	_, ackMessage, err := conn.ReadMessage()
	if err != nil {
		log.Fatal("❌ Failed to receive connection ack:", err)
	}
	
	var ackResponse map[string]interface{}
	json.Unmarshal(ackMessage, &ackResponse)
	
	// Verify server sent the correct acknowledgment
	if ackResponse["type"] != "connection_ack" {
		log.Fatalf("❌ Expected connection_ack, got: %+v", ackResponse)
	}
	
	fmt.Println("✅ Connection acknowledged! Setting up subscription...")
	
	// Step 3: Send subscription message
	// This give the query to the server
	subscription := map[string]interface{}{
		"id":   "1",                    // Unique ID for this subscription
		"type": "start",                // GraphQL-WS message type for subscriptions
		"payload": map[string]interface{}{
			"query": `subscription { ... }`, // Your GraphQL subscription here
		},
	}
	
	subscriptionBytes, _ := json.Marshal(subscription)
	conn.WriteMessage(websocket.TextMessage, subscriptionBytes)
	
	fmt.Println("📡 Listening for new send transactions...")
	
	// Step 4: Listen for incoming messages in an infinite loop
	for {
		// Read next message from WebSocket
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			continue
		}
		
		// Debug: Show raw message (remove in production)
		fmt.Printf("📨 Raw message: %s", string(message))
		
		// Parse JSON message from server
		var response map[string]interface{}
		err = json.Unmarshal(message, &response)
		if err != nil {
			log.Printf("❌ JSON parse error: %v\n", err)
			continue
		}
		
		// Handle different message types from GraphQL-WS protocol
		switch response["type"] {
		case "data":
			// New transaction data received!
			fmt.Println("🔥 NEW SEND TRANSACTION DETECTED!")
			
			// Extract payload and process transaction data
			data := response["payload"]
			dataBytes, _ := json.Marshal(data)
			parsedData := parseTransactions(dataBytes)
			displayTransactions(parsedData) // Show formatted transaction details
		case "error":
			// GraphQL query/subscription error
			fmt.Printf("❌ GraphQL error: %+v\n", response["payload"])
			
		case "complete":
			// Subscription finished (shouldn't happen for infinite subscriptions)
			fmt.Println("✅ Subscription completed")
			
		case "connection_error":
			// WebSocket connection issue
			fmt.Printf("❌ Connection error: %+v\n", response["payload"])
			
		case "ka":
			// Keep-alive message from server - ignore silently
			// Server sends these periodically to prevent connection timeouts
			continue
			
		default:
			// Unknown message type - log for debugging
			fmt.Printf("📋 Unknown message type: %s\n", response["type"])
		}
	}
}
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

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)


func parseTransactions(jsonData []byte) []Transaction {
	...
}

func sortTransactions(transactions []Transaction) []Transaction {
	...
}

// Simple dashboard server that serves transaction statistics via JSON API
func main() {
	http.HandleFunc("/stats", handleStats)
	fmt.Println("Dashboard running on http://localhost:8080/stats")
	http.ListenAndServe(":8080", nil)
}

// HTTP handler function for /stats endpoint
func handleStats(w http.ResponseWriter, r *http.Request) {
	// Sample data (in real app, get from indexer or database)
	// This would typically come from:
	// 1. Direct GraphQL query to indexer
	// 2. Your local database
	// 3. Cached data in memory
	data := ...
	
	// Process data using our existing functions
	transactions := parseTransactions([]byte(data)) 
	sorted := sortTransactions(transactions)            
	
	// Calculate statistics from our transaction data
	var totalVolume float64
	for _, tx := range transactions {
		totalVolume += tx.Amount
	}
	
	// Prepare response data structure
	response := map[string]interface{}{
		"status":             "success",
		"total_transactions": len(sorted),         
		"total_volume":       totalVolume,         
		"biggest_amount":     sorted[0].Amount,    
		"average_amount":     totalVolume / float64(len(sorted)),
		"top_transactions":   sorted[:min(5, len(sorted))],  // Top 5 transactions
	}
	
	// Set response headers for JSON API
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for frontend access
	json.NewEncoder(w).Encode(response)
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
	"fmt"
	"log"
	_ "github.com/mattn/go-sqlite3"  // SQLite driver - underscore import for side effects
)

// Create our database and transactions table
// This function sets up the SQLite database and creates the schema
func setupDB() *sql.DB {
	// Open SQLite database file (creates if doesn't exist)
	db, err := sql.Open("sqlite3", "./transactions.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			hash TEXT PRIMARY KEY,     -- Unique transaction ID
			amount REAL,              -- Amount in ugnot
			from_addr TEXT,           -- Sender address
			to_addr TEXT,             -- Receiver address
			block_height INTEGER,     -- Block number 
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP  -- When we stored this record
		)
	`)
	
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}
	
	fmt.Println("✅ Database initialized successfully")
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
	
	// Iterate through result set and scan into struct fields
	for rows.Next() {
		var tx Transaction
		err := rows.Scan(&tx.Hash, &tx.Amount, &tx.From, &tx.To, &tx.BlockHeight)
		if err != nil {
			log.Printf("Error scanning transaction: %v", err)
			continue 
		}
		transactions = append(transactions, tx)
	}
	
	return transactions, nil
}

// Save multiple transactions to the database
func saveTransactionsToDB(db *sql.DB, transactions []Transaction) {
	savedCount := 0
	
	for _, tx := range transactions {
		err := insertTransaction(db, tx)
		if err != nil {
			log.Printf("Failed to save transaction %s: %v", tx.Hash, err)
		} else {
			savedCount++
			fmt.Printf("✅ Saved transaction %s (%.2f uGNOT)\n", 
				tx.Hash[:8]+"...", tx.Amount)
		}
	}
	
	fmt.Printf("💾 Successfully saved %d/%d transactions to database\n", 
		savedCount, len(transactions))
}

// Get transaction statistics from database
func getTransactionStats(db *sql.DB) {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM transactions").Scan(&count)
	
	var totalVolume float64
	db.QueryRow("SELECT SUM(amount) FROM transactions").Scan(&totalVolume)
	
	var maxAmount float64
	db.QueryRow("SELECT MAX(amount) FROM transactions").Scan(&maxAmount)
	
	fmt.Printf("Database Statistics:\n")
	fmt.Printf("Total transactions: %d\n", count)
	fmt.Printf("Total volume: %.2f uGNOT\n", totalVolume)
	fmt.Printf("Largest transaction: %.2f uGNOT\n", maxAmount)
}

func main() {
	db := setupDB()
	defer db.Close()
	
	// Example: Get new transactions from the indexer
	// In a real application, this would be:
	// 1. HTTP request to GraphQL endpoint, or
	// 2. WebSocket subscription data, or
	// 3. Periodic polling of the indexer
	jsonData := `{"data": {"getTransactions": [...]}}`
	transactions := parseTransactions([]byte(jsonData))
	
	// Save new transactions to our database
	if len(transactions) > 0 {
		saveTransactionsToDB(db, transactions)
	}
	
	// Display current database statistics
	getTransactionStats(db)
	
	// Get and display the top 10 transactions ever recorded
	topTx, err := getTopTransactions(db, 10)
	if err != nil {
		log.Printf("Error getting top transactions: %v", err)
		return
	}
	
	fmt.Println("\nTop 10 biggest transactions in our database:")
	displayTransactions(topTx)
	
	// Your transaction data is now persistent and ready for:
	// - Web dashboards
	// - Mobile app APIs  
	// - Analytics and reporting
	// - Real-time monitoring alerts
}
```

Your transaction data is now persistent and ready for productions!

### Resources for Continued Learning

- **[tx-indexer Documentation](https://github.com/gnolang/tx-indexer)** - Official reference and advanced configuration
- **[GraphQL Best Practices](https://graphql.org/learn/best-practices/)** - Advanced querying techniques and optimization
