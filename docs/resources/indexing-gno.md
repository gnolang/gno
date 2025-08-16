# Blockchain Indexing in Gno

Blockchains store data in an immutable, sequential chain of blocks containing raw transactions. While optimized for security and consensus, this structure creates significant data retrieval challenges.

To track all transactions initiated by a specific address, we are required to re-scan **every blocks manually or on-chain**, which is a **computationally expensive tasks** (which can be very costly).

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

### Use case: `send` transactions dashboard 

To demonstrate `tx-indexer`, we will track and sort transactions containing token's exchange from a certain user through the `send` transactions.
We will:
- 1. Get data from GraphQL queries
- 2. Interpret them to expose relevant informations
- 3. Serve them through an interface

#### Step 1: Setting up the Transaction Filter using GraphQL

We need to filter transactions to identify all `send` transactions from an address. It is possible using the GraphQL exposed service:

```graphql
query GetTransactions { # Define a new query named "GetTransactions"
   getTransactions(     # Retreive all transactions
    where: {            # Apply filters
      # Only include transactions that succeeded.
      success: { eq: true }
      # Filter transactions containing specific MsgSend messages.
      messages: {
        value: {
          # Focus on messages of type "MsgSend" (maketx send).
          BankMsgSend: {}
      }
    }
  ) {
    # Expose filtered results
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

Which returns a JSON formated output, that can easily be processed:
```
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
      },
    ]
  }
}
```

#### Step 2: Processing Transaction Data

Here's a simple way to process transactions data.
You'll need to parse the GraphQL response:

```go
package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Simplified transaction structure
type Transaction struct {
	Hash   string
	Amount float64
	From   string
	To     string
}

// Step 1: JSON data -> Transaction structure
// This result can be indexed in a database
func parseTransactions(jsonData []byte) []Transaction {
	var data map[string]interface{}
	json.Unmarshal(jsonData, &data)
	
	txs := data["data"].(map[string]interface{})["getTransactions"].([]interface{})
	var transactions []Transaction
	
	for _, tx := range txs {
		txMap := tx.(map[string]interface{})
		hash := txMap["hash"].(string)
		
		// Get first message (1-to-1 relationship)
		msg := txMap["messages"].([]interface{})[0]
		msgMap := msg.(map[string]interface{})
		
		var msgData map[string]interface{}
		json.Unmarshal([]byte(msgMap["value"].(string)), &msgData)
		
		// Extract data
		amount := msgData["amount"].(string)
		from := msgData["from_address"].(string)
		to := msgData["to_address"].(string)
		
		amountInt, _ := strconv.ParseInt(amount, 10, 64)
		gnot := float64(amountInt) / 1000000
		
		transactions = append(transactions, Transaction{
			Hash:   hash,
			Amount: gnot,
			From:   from,
			To:     to,
		})
	}
	return transactions
}

// Step 2: Sort transactions by amount
// It can be done using database's capabilities
func sortTransactions(transactions []Transaction) []Transaction {
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Amount > transactions[j].Amount
	})
	return transactions
}

// Step 3: Output ready to be served
func displayTransactions(transactions []Transaction) {
  ...
}

func main() {
	// Sample JSON data from GraphQL
	jsonData := `{...}` // Retreive from GraphQL
	
	// Simple 3-step process:
	transactions := parseTransactions([]byte(jsonData))  // Step 1 - Process
	sorted := sortTransactions(transactions)             // Step 2 - Interpret
	displayTransactions(sorted)                          // Step 3 - Serve
}
```

#### Step 3: Real-time Monitoring with Subscriptions

Set up a WebSocket subscription to monitor new transactions in real-time:

```graphql
# We change the keyword `query` to `subscription`
subscription MonitorNewTransactions {
  transactions(
    filter: {
      success: { eq: true }
      messages: {
        value: {
          BankMsgSend: {}
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

This time, you will need a WebSocket client to handle real-time updates:

```go
package main

import (
	"fmt"
	"log"
	"net/url"
	
	"github.com/gorilla/websocket"
)

func main() {
	// Connect to tx-indexer WebSocket
	u := url.URL{Scheme: "ws", Host: "localhost:8546", Path: "/graphql"}
	
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("WebSocket connection failed:", err)
	}
	defer conn.Close()
	
	// Send subscription
	subscription := `{
		"type": "start",
		"payload": {
			"query": "subscription { ... }"
		}
	}`
	
	conn.WriteMessage(websocket.TextMessage, []byte(subscription))
	
	// Listen for new transactions
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		
		fmt.Printf("🔥 New transaction received: %s\n", string(message))
		// Process the transaction using parseTransactions() from Step 2
	}
}
``` 

#### Step 4: Dashboard Implementation Example

Here's an example of a statistics service:

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
	jsonData := `{...}`
	
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

For more example of queries, refers to the [tx-indexer](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#examples) documentation.

## Installation
Follow official [installation guide](https://github.com/gnolang/tx-indexer?tab=readme-ov-file#getting-started) on `tx-indexer` repository.
