package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Simple dashboard server that serves transaction statistics via JSON API
func RunApiExample() {
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
	data := `{"data": {"getTransactions": []}}`

	// Process data using our existing functions
	transactions, err := parseTransactions([]byte(data))
	if err != nil {
		http.Error(w, "Failed to parse transactions", http.StatusInternalServerError)
		return
	}
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
		"top_transactions":   sorted[:min(5, len(sorted))], // Top 5 transactions
	}

	// Set response headers for JSON API
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // Enable CORS for frontend access
	json.NewEncoder(w).Encode(response)
}
