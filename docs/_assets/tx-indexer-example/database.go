package txindexerexample

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver - underscore import for side effects
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
