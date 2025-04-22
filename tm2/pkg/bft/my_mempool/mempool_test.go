package my_mempool

import (
	"fmt"
	"sync"
	"testing"
)

func TestInsertAndSelectionOrder(t *testing.T) {
	mp := NewMempool()

	mp.CheckTx(Transaction{Sender: "a", Nonce: 2, GasFee: 50})
	mp.CheckTx(Transaction{Sender: "a", Nonce: 1, GasFee: 100})
	mp.CheckTx(Transaction{Sender: "b", Nonce: 1, GasFee: 90})

	selected := mp.CollectTxsForBlock(3)

	expectedOrder := []struct {
		Sender string
		Nonce  uint64
	}{
		{"a", 1},
		{"b", 1},
		{"a", 2},
	}

	for i, tx := range selected {
		if tx.Sender != expectedOrder[i].Sender || tx.Nonce != expectedOrder[i].Nonce {
			t.Errorf("tx %d: expected %s:%d, got %s:%d", i, expectedOrder[i].Sender, expectedOrder[i].Nonce, tx.Sender, tx.Nonce)
		}
	}
}

func TestCacheFillsAndRollbackRestores(t *testing.T) {
	mp := NewMempool()

	mp.CheckTx(Transaction{Sender: "x", Nonce: 1, GasFee: 100})
	mp.CheckTx(Transaction{Sender: "y", Nonce: 1, GasFee: 200})

	selected := mp.CollectTxsForBlock(2)

	if len(selected) != 2 {
		t.Errorf("expected 2 selected txs, got %d", len(selected))
	}

	if len(mp.cachedTxs) != 2 {
		t.Errorf("expected 2 cached txs, got %d", len(mp.cachedTxs))
	}

	mp.Rollback()

	if mp.Size() != 2 {
		t.Errorf("expected 2 txs after rollback, got %d", mp.Size())
	}
}

func TestUpdateOnlyKeepsUncommitted(t *testing.T) {
	mp1 := NewMempool()
	mp2 := NewMempool()

	txs := []Transaction{
		{Sender: "alice", Nonce: 1, GasFee: 100},
		{Sender: "bob", Nonce: 1, GasFee: 200},
		{Sender: "alice", Nonce: 2, GasFee: 150},
	}

	for _, tx := range txs {
		mp1.CheckTx(tx)
		mp2.CheckTx(tx)
	}

	selected := mp1.CollectTxsForBlock(3)

	committed := []Transaction{selected[0], selected[2]} // commit bob:1 and alice:2

	mp2.Update(committed)

	remaining := mp2.GetAllTransactions()
	if len(remaining) != 1 {
		t.Errorf("expected 1 tx in mp2 after Update, got %d", len(remaining))
	}
	if remaining[0].Sender != "alice" || remaining[0].Nonce != 1 {
		t.Errorf("expected remaining tx to be alice:1, got %+v", remaining[0])
	}
}

func TestHeapOrdering(t *testing.T) {
	mp := NewMempool()

	mp.CheckTx(Transaction{Sender: "z", Nonce: 1, GasFee: 100})
	mp.CheckTx(Transaction{Sender: "x", Nonce: 1, GasFee: 300})
	mp.CheckTx(Transaction{Sender: "y", Nonce: 1, GasFee: 200})

	selected := mp.CollectTxsForBlock(3)

	expectedFees := []uint64{300, 200, 100}
	for i, tx := range selected {
		if tx.GasFee != expectedFees[i] {
			t.Errorf("expected fee %d at index %d, got %d", expectedFees[i], i, tx.GasFee)
		}
	}
}

func TestConcurrentInsertion(t *testing.T) {
	mp := NewMempool()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		sender := string(rune('a' + i))
		wg.Add(1)
		go func(sender string) {
			defer wg.Done()
			for j := 1; j <= 5; j++ {
				mp.CheckTx(Transaction{Sender: sender, Nonce: uint64(j), GasFee: uint64(j * 10)})
			}
		}(sender)
	}

	wg.Wait()

	if mp.Size() != 50 {
		t.Errorf("expected 50 transactions after concurrent insert, got %d", mp.Size())
	}
}

func TestBasicInsertAndRetrieve(t *testing.T) {
	mp := NewMempool()

	tx1 := Transaction{Sender: "alice", Nonce: 1, GasFee: 100}
	tx2 := Transaction{Sender: "bob", Nonce: 1, GasFee: 200}

	_ = mp.CheckTx(tx1)
	_ = mp.CheckTx(tx2)

	if mp.Size() != 2 {
		t.Errorf("Expected mempool size 2, got %d", mp.Size())
	}

	allTxs := mp.GetAllTransactions()
	if len(allTxs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(allTxs))
	}

	// Check that we can retrieve by sender
	aliceTxs := mp.GetTransactionsBySender("alice")
	if len(aliceTxs) != 1 || aliceTxs[0].Nonce != 1 {
		t.Errorf("Expected 1 transaction for Alice with nonce 1, got %+v", aliceTxs)
	}
}

func TestPriorityBasedSelection(t *testing.T) {
	mp := NewMempool()

	// Add transactions with different fees
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 50})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 150})
	_ = mp.CheckTx(Transaction{Sender: "charlie", Nonce: 1, GasFee: 100})

	// Expected order: bob (150), charlie (100), alice (50)
	selected := mp.CollectTxsForBlock(3)

	if len(selected) != 3 {
		t.Fatalf("Expected 3 transactions, got %d", len(selected))
	}

	expectedOrder := []struct {
		sender string
		fee    uint64
	}{
		{"bob", 150},
		{"charlie", 100},
		{"alice", 50},
	}

	for i, tx := range selected {
		if tx.Sender != expectedOrder[i].sender || tx.GasFee != expectedOrder[i].fee {
			t.Errorf("Position %d: expected %s with fee %d, got %s with fee %d",
				i, expectedOrder[i].sender, expectedOrder[i].fee, tx.Sender, tx.GasFee)
		}
	}
}

func TestMultipleTransactionsPerSender(t *testing.T) {
	mp := NewMempool()

	// Add multiple transactions for the same sender
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 180})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 150})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 200})

	// Add one transaction from another sender
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 160})

	// Collect all transactions
	selected := mp.CollectTxsForBlock(4)

	if len(selected) != 4 {
		t.Fatalf("Expected 4 transactions, got %d", len(selected))
	}

	// First should be Alice's first tx, then Bob's, then Alice's remaining txs
	expectedOrder := []struct {
		sender string
		nonce  uint64
	}{
		{"alice", 1},
		{"bob", 1},
		{"alice", 2},
		{"alice", 3},
	}

	for i, tx := range selected {
		if tx.Sender != expectedOrder[i].sender || tx.Nonce != expectedOrder[i].nonce {
			t.Errorf("Position %d: expected %s with nonce %d, got %s with nonce %d",
				i, expectedOrder[i].sender, expectedOrder[i].nonce, tx.Sender, tx.Nonce)
		}
	}
}

func TestPartialCollection(t *testing.T) {
	mp := NewMempool()

	// Add 10 transactions
	for i := 1; i <= 5; i++ {
		_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: uint64(i), GasFee: 100})
		_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: uint64(i), GasFee: 100})
	}

	// Collect only 3 transactions
	selected := mp.CollectTxsForBlock(3)

	if len(selected) != 3 {
		t.Fatalf("Expected 3 transactions, got %d", len(selected))
	}

	// Verify remaining count
	if mp.Size() != 7 {
		t.Errorf("Expected 7 transactions remaining, got %d", mp.Size())
	}

	// Collect 5 more
	selected = mp.CollectTxsForBlock(5)

	if len(selected) != 5 {
		t.Fatalf("Expected 5 transactions, got %d", len(selected))
	}

	// Verify remaining count
	if mp.Size() != 2 {
		t.Errorf("Expected 2 transactions remaining, got %d", mp.Size())
	}
}

func TestEmptyMempool(t *testing.T) {
	mp := NewMempool()

	// Try to collect from empty mempool
	selected := mp.CollectTxsForBlock(10)

	if len(selected) != 0 {
		t.Errorf("Expected 0 transactions from empty mempool, got %d", len(selected))
	}

	// Add and remove all transactions
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	selected = mp.CollectTxsForBlock(1)

	if len(selected) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(selected))
	}

	// Try to collect again
	selected = mp.CollectTxsForBlock(1)

	if len(selected) != 0 {
		t.Errorf("Expected 0 transactions after emptying mempool, got %d", len(selected))
	}
}

func TestRollbackFunctionality(t *testing.T) {
	mp := NewMempool()

	// Add transactions
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 200})

	// Collect all
	selected := mp.CollectTxsForBlock(2)

	if len(selected) != 2 {
		t.Fatalf("Expected 2 transactions, got %d", len(selected))
	}

	// Verify mempool is empty
	if mp.Size() != 0 {
		t.Errorf("Expected empty mempool after collection, got size %d", mp.Size())
	}

	// Rollback
	mp.Rollback()

	// Verify transactions are back
	if mp.Size() != 2 {
		t.Errorf("Expected 2 transactions after rollback, got %d", mp.Size())
	}

	// Verify we can collect them again
	selected = mp.CollectTxsForBlock(2)

	if len(selected) != 2 {
		t.Errorf("Expected to collect 2 transactions after rollback, got %d", len(selected))
	}
}

func TestUpdateRemovesCommitted(t *testing.T) {
	mp := NewMempool()

	// Add transactions
	tx1 := Transaction{Sender: "alice", Nonce: 1, GasFee: 100}
	tx2 := Transaction{Sender: "alice", Nonce: 2, GasFee: 150}
	tx3 := Transaction{Sender: "bob", Nonce: 1, GasFee: 200}

	_ = mp.CheckTx(tx1)
	_ = mp.CheckTx(tx2)
	_ = mp.CheckTx(tx3)

	// Update with committed transactions
	mp.Update([]Transaction{tx1, tx3})

	// Verify only tx2 remains
	if mp.Size() != 1 {
		t.Errorf("Expected 1 transaction after update, got %d", mp.Size())
	}

	allTxs := mp.GetAllTransactions()
	if len(allTxs) != 1 || allTxs[0].Sender != "alice" || allTxs[0].Nonce != 2 {
		t.Errorf("Expected only Alice's second tx to remain, got %+v", allTxs)
	}
}

func TestHighVolumeOperations(t *testing.T) {
	mp := NewMempool()

	// Add 1000 transactions from 10 senders
	for i := 0; i < 10; i++ {
		sender := string(rune('a' + i))
		for j := 1; j <= 100; j++ {
			_ = mp.CheckTx(Transaction{
				Sender: sender,
				Nonce:  uint64(j),
				GasFee: uint64(100 + (i * 10) + (j % 50)), // Varied fees
			})
		}
	}

	// Verify size
	if mp.Size() != 1000 {
		t.Errorf("Expected 1000 transactions, got %d", mp.Size())
	}

	// Collect in batches
	var allCollected []Transaction
	for i := 0; i < 10; i++ {
		batch := mp.CollectTxsForBlock(100)
		allCollected = append(allCollected, batch...)
	}

	// Verify all collected
	if len(allCollected) != 1000 {
		t.Errorf("Expected to collect all 1000 transactions, got %d", len(allCollected))
	}

	// Verify mempool is empty
	if mp.Size() != 0 {
		t.Errorf("Expected empty mempool after collecting all, got size %d", mp.Size())
	}
}

func TestFeeBasedPrioritization(t *testing.T) {
	mp := NewMempool()

	// Add transactions with increasing fees
	for i := 1; i <= 10; i++ {
		_ = mp.CheckTx(Transaction{
			Sender: string(rune('a' + i - 1)),
			Nonce:  1,
			GasFee: uint64(i * 100),
		})
	}

	// Collect all
	selected := mp.CollectTxsForBlock(10)

	// Verify order is by decreasing fee
	for i := 0; i < len(selected)-1; i++ {
		if selected[i].GasFee < selected[i+1].GasFee {
			t.Errorf("Transactions not ordered by fee: %d before %d",
				selected[i].GasFee, selected[i+1].GasFee)
		}
	}
}

func TestConcurrentOperations(t *testing.T) {
	mp := NewMempool()
	var wg sync.WaitGroup

	// Concurrently add transactions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 1; j <= 10; j++ {
				_ = mp.CheckTx(Transaction{
					Sender: string(rune('a' + i)),
					Nonce:  uint64(j),
					GasFee: uint64(100 + j),
				})
			}
		}(i)
	}

	wg.Wait()

	// Verify all transactions were added
	if mp.Size() != 100 {
		t.Errorf("Expected 100 transactions after concurrent addition, got %d", mp.Size())
	}

	// Concurrently collect and add
	var collected sync.Map
	wg = sync.WaitGroup{}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			txs := mp.CollectTxsForBlock(10)
			collected.Store(i, txs)
		}(i)
	}

	wg.Wait()

	// Count collected transactions
	totalCollected := 0
	collected.Range(func(_, value interface{}) bool {
		txs := value.([]Transaction)
		totalCollected += len(txs)
		return true
	})

	// Verify we collected some transactions
	if totalCollected == 0 {
		t.Error("Failed to collect any transactions concurrently")
	}

	t.Logf("Collected %d transactions concurrently", totalCollected)
	t.Logf("Remaining in mempool: %d", mp.Size())
}

func TestMempoolStressTest(t *testing.T) {
	mp := NewMempool()

	// Add a large number of transactions with random fees
	for i := 0; i < 20; i++ {
		sender := string(rune('a' + (i % 5))) // 5 different senders
		for j := 1; j <= 50; j++ {
			fee := uint64(100 + (j * 10) + (i * 5)) // Different fee pattern
			_ = mp.CheckTx(Transaction{
				Sender: sender,
				Nonce:  uint64(j),
				GasFee: fee,
			})
		}
	}

	// Verify size
	expectedSize := 5 * 50 // 5 senders, 50 transactions each
	if mp.Size() != expectedSize {
		t.Errorf("Expected %d transactions, got %d", expectedSize, mp.Size())
	}

	// Collect all in one go
	selected := mp.CollectTxsForBlock(uint(expectedSize))

	if len(selected) != expectedSize {
		t.Errorf("Expected to collect %d transactions, got %d", expectedSize, len(selected))
	}

	// Verify mempool is empty
	if mp.Size() != 0 {
		t.Errorf("Expected empty mempool after collection, got size %d", mp.Size())
	}
}

func TestCollectWithLimit(t *testing.T) {
	mp := NewMempool()

	// Add 10 transactions
	for i := 1; i <= 10; i++ {
		_ = mp.CheckTx(Transaction{
			Sender: "alice",
			Nonce:  uint64(i),
			GasFee: 100,
		})
	}

	// Collect with different limits
	testCases := []struct {
		limit    uint
		expected int
	}{
		{0, 0},
		{1, 1},
		{5, 5},
		{10, 10},
		{20, 10}, // Should only return 10 since that's all we have
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Limit_%d", tc.limit), func(t *testing.T) {
			// Create a fresh mempool for each test case
			mp := NewMempool()
			for i := 1; i <= 10; i++ {
				_ = mp.CheckTx(Transaction{
					Sender: "alice",
					Nonce:  uint64(i),
					GasFee: 100,
				})
			}

			selected := mp.CollectTxsForBlock(tc.limit)
			if len(selected) != tc.expected {
				t.Errorf("Expected %d transactions with limit %d, got %d",
					tc.expected, tc.limit, len(selected))
			}
		})
	}
}

func TestUpdateWithEmptyList(t *testing.T) {
	mp := NewMempool()

	// Add some transactions
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})

	// Update with empty list
	mp.Update([]Transaction{})

	// Verify nothing changed
	if mp.Size() != 1 {
		t.Errorf("Expected 1 transaction after empty update, got %d", mp.Size())
	}
}

func TestUpdateWithNonExistentTransactions(t *testing.T) {
	mp := NewMempool()

	// Add a transaction
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})

	// Update with non-existent transaction
	mp.Update([]Transaction{
		{Sender: "bob", Nonce: 1, GasFee: 200},
	})

	// Verify nothing changed
	if mp.Size() != 1 {
		t.Errorf("Expected 1 transaction after update with non-existent tx, got %d", mp.Size())
	}
}
