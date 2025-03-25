package my_mempool

import (
	"testing"
)

// TestNewMempool tests the creation of a new mempool
func TestNewMempool(t *testing.T) {
	mp := NewMempool()
	if mp == nil {
		t.Fatal("NewMempool() returned nil")
	}
	if mp.txsBySender == nil {
		t.Error("txsBySender map not initialized")
	}
	if mp.highestFeeTxBySender == nil {
		t.Error("highestFeeTxBySender map not initialized")
	}
}

// TestCheckTx tests adding transactions to the mempool
func TestCheckTx(t *testing.T) {
	mp := NewMempool()

	// Test adding a valid transaction
	tx1 := Transaction{
		Sender: "sender1",
		Nonce:  1,
		GasFee: 10,
	}
	err := mp.CheckTx(tx1)
	if err != nil {
		t.Errorf("CheckTx() returned error for valid transaction: %v", err)
	}

	// Test adding a transaction with empty sender
	tx2 := Transaction{
		Sender: "",
		Nonce:  2,
		GasFee: 20,
	}
	err = mp.CheckTx(tx2)
	if err == nil {
		t.Error("CheckTx() did not return error for transaction with empty sender")
	}

	// Test that transaction was added to txsBySender
	txs := mp.GetTransactionsBySender("sender1")
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction for sender1, got %d", len(txs))
	}

	// Test that transaction was added to highestFeeTxBySender
	highestTx, exists := mp.GetHighestFeeTxBySender("sender1")
	if !exists {
		t.Error("No highest fee transaction found for sender1")
	}
	if highestTx.GasFee != 10 {
		t.Errorf("Expected highest fee to be 10, got %d", highestTx.GasFee)
	}

	// Test updating highest fee transaction
	tx3 := Transaction{
		Sender: "sender1",
		Nonce:  3,
		GasFee: 30,
	}
	err = mp.CheckTx(tx3)
	if err != nil {
		t.Errorf("CheckTx() returned error for valid transaction: %v", err)
	}

	highestTx, exists = mp.GetHighestFeeTxBySender("sender1")
	if !exists {
		t.Error("No highest fee transaction found for sender1")
	}
	if highestTx.GasFee != 30 {
		t.Errorf("Expected highest fee to be 30, got %d", highestTx.GasFee)
	}
}

// TestUpdate tests selecting and removing transactions from the mempool
func TestUpdate(t *testing.T) {
	mp := NewMempool()

	// Add transactions from multiple senders with different gas fees
	txs := []Transaction{
		{Sender: "sender1", Nonce: 1, GasFee: 10},
		{Sender: "sender1", Nonce: 2, GasFee: 20},
		{Sender: "sender1", Nonce: 3, GasFee: 15},
		{Sender: "sender2", Nonce: 1, GasFee: 30},
		{Sender: "sender2", Nonce: 2, GasFee: 25},
		{Sender: "sender3", Nonce: 1, GasFee: 5},
	}

	for _, tx := range txs {
		err := mp.CheckTx(tx)
		if err != nil {
			t.Errorf("CheckTx() returned error for valid transaction: %v", err)
		}
	}

	// Verify initial state
	if mp.Size() != 6 {
		t.Errorf("Expected mempool size to be 6, got %d", mp.Size())
	}

	// Test Update with maxTxs = 3
	selectedTxs := mp.Update(3)
	if len(selectedTxs) != 3 {
		t.Errorf("Expected 3 selected transactions, got %d", len(selectedTxs))
	}

	// Verify that transactions were selected in the correct order
	// First from sender2 (highest fee), then from sender1 (second highest fee)
	if selectedTxs[0].Sender != "sender2" || selectedTxs[0].Nonce != 1 {
		t.Errorf("First selected transaction should be from sender2 with nonce 1, got sender=%s, nonce=%d",
			selectedTxs[0].Sender, selectedTxs[0].Nonce)
	}
	if selectedTxs[1].Sender != "sender2" || selectedTxs[1].Nonce != 2 {
		t.Errorf("Second selected transaction should be from sender2 with nonce 2, got sender=%s, nonce=%d",
			selectedTxs[1].Sender, selectedTxs[1].Nonce)
	}
	if selectedTxs[2].Sender != "sender1" || selectedTxs[2].Nonce != 1 {
		t.Errorf("Third selected transaction should be from sender1 with nonce 1, got sender=%s, nonce=%d",
			selectedTxs[2].Sender, selectedTxs[2].Nonce)
	}

	// Verify that transactions were removed from the mempool
	if mp.Size() != 3 {
		t.Errorf("Expected mempool size to be 3 after update, got %d", mp.Size())
	}

	// Verify that sender2 has no more transactions
	txs = mp.GetTransactionsBySender("sender2")
	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions for sender2, got %d", len(txs))
	}

	// Verify that sender1 has 2 transactions left
	txs = mp.GetTransactionsBySender("sender1")
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions for sender1, got %d", len(txs))
	}

	// Verify that highestFeeTxBySender was updated correctly
	highestTx, exists := mp.GetHighestFeeTxBySender("sender1")
	if !exists {
		t.Error("No highest fee transaction found for sender1")
	}
	if highestTx.GasFee != 20 {
		t.Errorf("Expected highest fee for sender1 to be 20, got %d", highestTx.GasFee)
	}

	// Verify that sender2 was removed from highestFeeTxBySender
	_, exists = mp.GetHighestFeeTxBySender("sender2")
	if exists {
		t.Error("Highest fee transaction still exists for sender2 after all transactions were removed")
	}
}

// TestGetAllTransactions tests retrieving all transactions from the mempool
func TestGetAllTransactions(t *testing.T) {
	mp := NewMempool()

	// Add transactions from multiple senders
	txs := []Transaction{
		{Sender: "sender1", Nonce: 1, GasFee: 10},
		{Sender: "sender2", Nonce: 1, GasFee: 20},
		{Sender: "sender3", Nonce: 1, GasFee: 30},
	}

	for _, tx := range txs {
		err := mp.CheckTx(tx)
		if err != nil {
			t.Errorf("CheckTx() returned error for valid transaction: %v", err)
		}
	}

	// Test GetAllTransactions
	allTxs := mp.GetAllTransactions()
	if len(allTxs) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(allTxs))
	}
}

// TestSize tests the Size method of the mempool
func TestSize(t *testing.T) {
	mp := NewMempool()

	// Test empty mempool
	if mp.Size() != 0 {
		t.Errorf("Expected size 0 for empty mempool, got %d", mp.Size())
	}

	// Add transactions
	txs := []Transaction{
		{Sender: "sender1", Nonce: 1, GasFee: 10},
		{Sender: "sender1", Nonce: 2, GasFee: 20},
		{Sender: "sender2", Nonce: 1, GasFee: 30},
	}

	for _, tx := range txs {
		err := mp.CheckTx(tx)
		if err != nil {
			t.Errorf("CheckTx() returned error for valid transaction: %v", err)
		}
	}

	// Test size after adding transactions
	if mp.Size() != 3 {
		t.Errorf("Expected size 3 after adding transactions, got %d", mp.Size())
	}

	// Remove transactions
	mp.Update(2)

	// Test size after removing transactions
	if mp.Size() != 1 {
		t.Errorf("Expected size 1 after removing transactions, got %d", mp.Size())
	}
}

// TestConcurrency tests concurrent access to the mempool
func TestConcurrency(t *testing.T) {
	mp := NewMempool()
	done := make(chan bool)

	// Add transactions concurrently
	for i := 0; i < 10; i++ {
		go func(i int) {
			tx := Transaction{
				Sender: "sender1",
				Nonce:  uint64(i),
				GasFee: uint64(i * 10),
			}
			err := mp.CheckTx(tx)
			if err != nil {
				t.Errorf("CheckTx() returned error for valid transaction: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify that all transactions were added
	if mp.Size() != 10 {
		t.Errorf("Expected 10 transactions after concurrent additions, got %d", mp.Size())
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			mp.GetAllTransactions()
			mp.GetTransactionsBySender("sender1")
			mp.GetHighestFeeTxBySender("sender1")
			mp.Size()
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}
