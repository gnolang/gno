package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Simplified transaction data structure
// This struct represents a cleaned-up version of the raw GraphQL data
type Transaction struct {
	Hash        string
	BlockHeight float64
	Amount      float64 // in ugnot
	From        string
	To          string
}

// GraphQL response typed model (only required fields)
type gqlResponse struct {
	Data struct {
		GetTransactions []struct {
			Hash        string  `json:"hash"`
			BlockHeight float64 `json:"block_height"`
			Messages    []struct {
				Value struct {
					FromAddress string `json:"from_address"`
					ToAddress   string `json:"to_address"`
					Amount      string `json:"amount"` // e.g. 15000000ugnot
				} `json:"value"`
			} `json:"messages"`
		} `json:"getTransactions"`
	} `json:"data"`
}

// Step 1 - Parse JSON from GraphQL into our Transaction structs
// This function takes raw JSON from the indexer and converts it to Go structs
func parseTransactions(jsonData []byte) ([]Transaction, error) {
	var resp gqlResponse
	if err := json.Unmarshal(jsonData, &resp); err != nil {
		return nil, fmt.Errorf("decode graphql response: %w", err)
	}
	if len(resp.Data.GetTransactions) == 0 {
		return nil, nil
	}
	out := make([]Transaction, 0, len(resp.Data.GetTransactions))
	for _, tx := range resp.Data.GetTransactions {
		if len(tx.Messages) == 0 {
			continue
		}
		msg := tx.Messages[0].Value
		amtStr := strings.TrimSuffix(msg.Amount, "ugnot")
		amt, _ := strconv.ParseFloat(amtStr, 64) // ignore parse error -> 0
		out = append(out, Transaction{
			Hash:        tx.Hash,
			BlockHeight: tx.BlockHeight,
			Amount:      amt,
			From:        msg.FromAddress,
			To:          msg.ToAddress,
		})
	}
	return out, nil
}

// Step 2 - Sort transactions by amount (biggest first)
// This helps us identify the largest transfers on the network
func sortTransactions(txs []Transaction) []Transaction {
	sort.Slice(txs, func(i, j int) bool { return txs[i].Amount > txs[j].Amount })
	return txs
}

// Step 3 - Show the transactions in a nice format
// Convert raw data into human-readable output
func displayTransactions(txs []Transaction) {
	fmt.Println("Top GNOT Transactions:")
	for i, tx := range txs {
		if i >= 5 {
			break
		} // Limit to top 5

		gnotAmount := tx.Amount
		fmt.Printf("%d. %.2f uGNOT from %s to %s (block %.0f)\n",
			i+1, gnotAmount, tx.From, tx.To, tx.BlockHeight)
	}
}

func RunProcessExample() {
	// This would be your actual JSON from the GraphQL query
	// In a real app, you'd get this from an HTTP request to the indexer
	jsonData := []byte(`{"data": {"getTransactions": []}}`)

	// Process the data in 3 steps:
	transactions, err := parseTransactions(jsonData)
	if err != nil || len(transactions) == 0 {
		return
	}
	if len(transactions) == 0 {
		fmt.Println("no transactions decoded")
		return
	}
	sorted := sortTransactions(transactions) // 2. Sort by amount (largest first)
	displayTransactions(sorted)              // 3. Display results nicely

	// At this point, you have clean, sorted transaction data ready for:
	// - Saving to a database
	// - Serving via an API
	// - ...
}
