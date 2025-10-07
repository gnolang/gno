package txindexerexample

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
)

// Simplified transaction data structure
// This struct represents a cleaned-up version of the raw GraphQL data
type Transaction struct {
	Hash        string  // Transaction ID - unique identifier for each transaction
	BlockHeight float64 // Block number - tells us when this transaction happened
	Amount      float64 // Amount in ugnot (1 GNOT = 1,000,000 ugnot)
	From        string  // Sender address - who initiated the transaction
	To          string  // Receiver address - who received the funds
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
			Hash:        hash,
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
		if i >= 5 {
			break
		} // Limit to top 5

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
	transactions := parseTransactions([]byte(jsonData)) // 1. Parse JSON into structs
	sorted := sortTransactions(transactions)            // 2. Sort by amount (largest first)
	displayTransactions(sorted)                         // 3. Display results nicely

	// At this point, you have clean, sorted transaction data ready for:
	// - Saving to a database
	// - Serving via an API
	// - ...
}
