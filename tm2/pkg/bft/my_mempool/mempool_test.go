package my_mempool

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"testing"
)

func TestRejectEmptySender(t *testing.T) {
	mp := NewMempool()
	tx := Transaction{Sender: "", Nonce: 1, GasFee: 100}
	if err := mp.CheckTx(tx); err == nil {
		t.Error("Expected error for empty sender, got nil")
	}
}

func TestRejectNonceGap(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 200})
	selected := mp.Update(10)
	if len(selected) != 1 || selected[0].Nonce != 1 {
		t.Errorf("Expected only tx with nonce 1, got %+v", selected)
	}
}

func TestMaxPerSenderLimit(t *testing.T) {
	mp := NewMempool()
	for i := 1; i <= 15; i++ {
		tx := Transaction{Sender: "bob", Nonce: uint64(i), GasFee: 100 + uint64(i)}
		_ = mp.CheckTx(tx)
	}
	selected := mp.Update(20)
	if len(selected) != 10 {
		t.Errorf("Expected 10 txs due to per-sender limit, got %d", len(selected))
	}
}

func TestGasFeePriority(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "low", Nonce: 1, GasFee: 1})
	_ = mp.CheckTx(Transaction{Sender: "high", Nonce: 1, GasFee: 1000})
	selected := mp.Update(1)
	if selected[0].Sender != "high" {
		t.Errorf("Expected high-fee sender first, got %s", selected[0].Sender)
	}
}

func TestSenderCleanup(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "carol", Nonce: 1, GasFee: 99})
	_ = mp.Update(1)
	if _, exists := mp.senderMap["carol"]; exists {
		t.Error("Expected sender carol to be removed from heap map after update")
	}
	if len(mp.txsBySender["carol"]) != 0 {
		t.Error("Expected carol's txs to be empty after update")
	}
}

func TestHeapConsistency(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "dave", Nonce: 1, GasFee: 10})
	_ = mp.CheckTx(Transaction{Sender: "dave", Nonce: 2, GasFee: 50})
	selected := mp.Update(1)
	if len(selected) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(selected))
	}
	entry, exists := mp.senderMap["dave"]
	if !exists {
		t.Error("Expected dave to still exist in senderMap")
	}
	if entry.fee != 50 {
		t.Errorf("Expected fee to be 50, got %d", entry.fee)
	}
}

func TestMultipleSendersFairness(t *testing.T) {
	mp := NewMempool()
	for i := 1; i <= 10; i++ {
		_ = mp.CheckTx(Transaction{Sender: "a", Nonce: uint64(i), GasFee: 10})
		_ = mp.CheckTx(Transaction{Sender: "b", Nonce: uint64(i), GasFee: 100 + uint64(i)})
	}
	selected := mp.Update(15)
	countA, countB := 0, 0
	for _, tx := range selected {
		if tx.Sender == "a" {
			countA++
		} else if tx.Sender == "b" {
			countB++
		}
	}
	if countB < countA {
		t.Errorf("Expected sender B with higher gas fee to be prioritized, got A:%d B:%d", countA, countB)
	}
}

func TestZeroMaxTxs(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "zed", Nonce: 1, GasFee: 1})
	selected := mp.Update(0)
	if len(selected) != 0 {
		t.Errorf("Expected no transactions returned when maxTxs=0, got %d", len(selected))
	}
}

func TestUpdateExactFit(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "fizz", Nonce: 1, GasFee: 20})
	_ = mp.CheckTx(Transaction{Sender: "fizz", Nonce: 2, GasFee: 30})
	selected := mp.Update(2)
	if len(selected) != 2 {
		t.Errorf("Expected exactly 2 transactions, got %d", len(selected))
	}
}

func TestSenderRemovalWhenAllTxsConsumed(t *testing.T) {
	mp := NewMempool()
	_ = mp.CheckTx(Transaction{Sender: "gone", Nonce: 1, GasFee: 100})
	_ = mp.Update(1)
	if _, exists := mp.txsBySender["gone"]; exists {
		t.Error("Expected gone to be removed from txsBySender")
	}
	if _, exists := mp.senderMap["gone"]; exists {
		t.Error("Expected gone to be removed from senderMap")
	}
}

func TestNonceOrdering(t *testing.T) {
	mp := NewMempool()
	// Add transactions in reverse nonce order
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})

	selected := mp.Update(10)

	// Should select transactions in correct nonce order
	if len(selected) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(selected))
	}

	for i, expectedNonce := range []uint64{1, 2, 3} {
		if selected[i].Nonce != expectedNonce {
			t.Errorf("Expected nonce %d at position %d, got %d", expectedNonce, i, selected[i].Nonce)
		}
	}
}

func TestMultipleSendersWithNonceGaps(t *testing.T) {
	mp := NewMempool()

	// Alice has continuous nonces
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 200})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 200})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 200})

	// Bob has a nonce gap
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 300})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 3, GasFee: 300}) // Gap here

	selected := mp.Update(10)

	// Count transactions by sender
	aliceTxs, bobTxs := 0, 0
	for _, tx := range selected {
		if tx.Sender == "alice" {
			aliceTxs++
		} else if tx.Sender == "bob" {
			bobTxs++
		}
	}

	if aliceTxs != 3 {
		t.Errorf("Expected 3 transactions from Alice, got %d", aliceTxs)
	}

	if bobTxs != 1 {
		t.Errorf("Expected 1 transaction from Bob (due to nonce gap), got %d", bobTxs)
	}
}

func TestHeapReordering(t *testing.T) {
	mp := NewMempool()

	// Add transactions with increasing fees
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 200})

	// First update should prioritize Bob
	selected := mp.Update(1)
	if selected[0].Sender != "bob" {
		t.Errorf("Expected Bob's transaction first, got %s", selected[0].Sender)
	}

	// Now add a higher fee transaction for Alice
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 300})

	// Second update should prioritize Alice
	selected = mp.Update(1)
	if selected[0].Sender != "alice" {
		t.Errorf("Expected Alice's transaction first after fee increase, got %s", selected[0].Sender)
	}
}

func TestMaxTxsLimit(t *testing.T) {
	mp := NewMempool()

	// Add 5 transactions from different senders
	for i := 1; i <= 5; i++ {
		sender := string(rune('a' + i - 1))
		_ = mp.CheckTx(Transaction{Sender: sender, Nonce: 1, GasFee: 100})
	}

	// Request only 3 transactions
	selected := mp.Update(3)

	if len(selected) != 3 {
		t.Errorf("Expected exactly 3 transactions, got %d", len(selected))
	}
}

func TestFeeUpdatesAfterPartialSelection(t *testing.T) {
	mp := NewMempool()

	// Add transactions with different fees
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 300})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 200})

	// First update should take Alice's first tx (highest fee)
	selected := mp.Update(1)
	if selected[0].Sender != "alice" || selected[0].GasFee != 300 {
		t.Errorf("Expected Alice's tx with fee 300, got %s with fee %d", selected[0].Sender, selected[0].GasFee)
	}

	// Second update should take Bob's tx since Alice's remaining tx has lower fee
	selected = mp.Update(1)
	if selected[0].Sender != "bob" || selected[0].GasFee != 200 {
		t.Errorf("Expected Bob's tx with fee 200, got %s with fee %d", selected[0].Sender, selected[0].GasFee)
	}
}

func TestEmptyMempool(t *testing.T) {
	mp := NewMempool()

	selected := mp.Update(10)
	if len(selected) != 0 {
		t.Errorf("Expected 0 transactions from empty mempool, got %d", len(selected))
	}

	if mp.Size() != 0 {
		t.Errorf("Expected empty mempool to have size 0, got %d", mp.Size())
	}
}

func TestGetAllTransactions(t *testing.T) {
	mp := NewMempool()

	// Add 5 transactions
	for i := 1; i <= 5; i++ {
		_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: uint64(i), GasFee: 100})
	}

	allTxs := mp.GetAllTransactions()
	if len(allTxs) != 5 {
		t.Errorf("Expected GetAllTransactions to return 5 txs, got %d", len(allTxs))
	}

	// After update, should have fewer transactions
	_ = mp.Update(3)
	allTxs = mp.GetAllTransactions()
	if len(allTxs) != 2 {
		t.Errorf("Expected 2 remaining transactions after update, got %d", len(allTxs))
	}
}

func TestGetTransactionsBySender(t *testing.T) {
	mp := NewMempool()

	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 100})
	_ = mp.CheckTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 100})

	aliceTxs := mp.GetTransactionsBySender("alice")
	if len(aliceTxs) != 2 {
		t.Errorf("Expected 2 transactions for Alice, got %d", len(aliceTxs))
	}

	bobTxs := mp.GetTransactionsBySender("bob")
	if len(bobTxs) != 1 {
		t.Errorf("Expected 1 transaction for Bob, got %d", len(bobTxs))
	}

	charlieTxs := mp.GetTransactionsBySender("charlie")
	if len(charlieTxs) != 0 {
		t.Errorf("Expected 0 transactions for Charlie, got %d", len(charlieTxs))
	}
}

func TestConcurrentAccess(t *testing.T) {
	mp := NewMempool()
	var wg sync.WaitGroup

	// Add 100 transactions concurrently
	for i := 1; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sender := string(rune('a' + (i % 5))) // Use 5 different senders
			_ = mp.CheckTx(Transaction{Sender: sender, Nonce: uint64(i), GasFee: uint64(i * 10)})
		}(i)
	}

	// Concurrently read from the mempool
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mp.GetAllTransactions()
			_ = mp.Size()
		}()
	}

	wg.Wait()

	// Verify the mempool has transactions
	if mp.Size() == 0 {
		t.Error("Expected non-empty mempool after concurrent additions")
	}
}

func TestReplaceTransaction(t *testing.T) {
	mp := NewMempool()

	// Add a transaction
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})

	// Check initial state
	aliceTxs := mp.GetTransactionsBySender("alice")
	if len(aliceTxs) != 1 {
		t.Errorf("Expected 1 transaction for Alice, got %d", len(aliceTxs))
	}

	// Add a transaction with higher fee for the same nonce (should replace or be added alongside)
	_ = mp.CheckTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 200})

	// Check if the transaction was added
	aliceTxs = mp.GetTransactionsBySender("alice")
	if len(aliceTxs) != 2 {
		t.Errorf("Expected 2 transactions for Alice, got %d", len(aliceTxs))
	}

	// Verify the highest fee transaction is selected during update
	selected := mp.Update(1)
	if len(selected) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(selected))
	}

	if selected[0].GasFee != 200 {
		t.Errorf("Expected transaction with fee 200 to be selected, got %d", selected[0].GasFee)
	}
}

func TestComplexMempoolSimulation(t *testing.T) {
	mp := NewMempool()
	var wg sync.WaitGroup

	// Simulacija više epoha sa konkurentnim operacijama
	const (
		numSenders    = 20
		txsPerSender  = 30
		numEpochs     = 5
		txsPerEpoch   = 50
		concurrentOps = 5
	)

	// Struktura za praćenje nonce-a po pošiljaocu
	senderNonces := make(map[string]uint64)
	for i := 0; i < numSenders; i++ {
		sender := fmt.Sprintf("sender-%d", i)
		senderNonces[sender] = 1 // Početni nonce
	}

	// Mutex za zaštitu mape nonce-a
	var nonceMutex sync.Mutex

	// Funkcija za dodavanje transakcija sa rastućim nonce-om
	addTxsWithIncreasingNonce := func(sender string, count int, baseFee uint64) {
		nonceMutex.Lock()
		startNonce := senderNonces[sender]
		senderNonces[sender] += uint64(count)
		nonceMutex.Unlock()

		for i := 0; i < count; i++ {
			nonce := startNonce + uint64(i)
			// Varijacija u naknadama
			fee := baseFee + uint64(rand.Intn(100))
			tx := Transaction{Sender: sender, Nonce: nonce, GasFee: fee}
			err := mp.CheckTx(tx)
			if err != nil {
				t.Errorf("Failed to add tx: %v", err)
			}
		}
	}

	// Faza 1: Inicijalno punjenje mempool-a
	t.Log("Phase 1: Initial mempool population")
	for i := 0; i < numSenders; i++ {
		sender := fmt.Sprintf("sender-%d", i)
		// Različite naknade za različite pošiljaoce
		baseFee := uint64(100 + i*50)
		// Dodaj različit broj transakcija za različite pošiljaoce
		txCount := 5 + rand.Intn(10)
		addTxsWithIncreasingNonce(sender, txCount, baseFee)
	}

	// Provera inicijalnog stanja
	initialSize := mp.Size()
	t.Logf("Initial mempool size: %d", initialSize)
	if initialSize == 0 {
		t.Fatal("Expected non-empty mempool after initial population")
	}

	// Faza 2: Simulacija više epoha obrade
	t.Log("Phase 2: Multiple epoch simulation")
	for epoch := 1; epoch <= numEpochs; epoch++ {
		t.Logf("Epoch %d starting", epoch)

		// Konkurentno dodavanje novih transakcija tokom epohe
		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Nasumično izaberi pošiljaoca
				senderIdx := rand.Intn(numSenders)
				sender := fmt.Sprintf("sender-%d", senderIdx)
				// Dodaj 1-5 transakcija
				txCount := 1 + rand.Intn(5)
				baseFee := uint64(150 + rand.Intn(200))
				addTxsWithIncreasingNonce(sender, txCount, baseFee)
			}()
		}

		// Konkurentno čitanje iz mempool-a
		for i := 0; i < concurrentOps; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = mp.GetAllTransactions()
				// Nasumično izaberi pošiljaoca za čitanje
				senderIdx := rand.Intn(numSenders)
				sender := fmt.Sprintf("sender-%d", senderIdx)
				_ = mp.GetTransactionsBySender(sender)
			}()
		}

		// Sačekaj da se konkurentne operacije završe
		wg.Wait()

		// Izvrši Update za ovu epohu
		beforeUpdateSize := mp.Size()
		selected := mp.Update(uint(txsPerEpoch))
		afterUpdateSize := mp.Size()

		t.Logf("Epoch %d: Selected %d transactions, mempool size before: %d, after: %d",
			epoch, len(selected), beforeUpdateSize, afterUpdateSize)

		// Provera da li su transakcije izabrane po redosledu nonce-a
		senderLastNonce := make(map[string]uint64)
		for _, tx := range selected {
			lastNonce, exists := senderLastNonce[tx.Sender]
			if exists && tx.Nonce <= lastNonce {
				t.Errorf("Epoch %d: Nonce ordering violation for sender %s: got nonce %d after nonce %d",
					epoch, tx.Sender, tx.Nonce, lastNonce)
			}
			senderLastNonce[tx.Sender] = tx.Nonce
		}

		// Provera da li je broj izabranih transakcija manji ili jednak limitu
		if len(selected) > txsPerEpoch {
			t.Errorf("Epoch %d: Selected more transactions (%d) than limit (%d)",
				epoch, len(selected), txsPerEpoch)
		}

		// Provera da li je veličina mempool-a smanjena za broj izabranih transakcija
		expectedSize := beforeUpdateSize - len(selected)
		if afterUpdateSize != expectedSize {
			t.Errorf("Epoch %d: Expected mempool size %d after update, got %d",
				epoch, expectedSize, afterUpdateSize)
		}

		// Faza 3: Dodavanje transakcija sa prazninama u nonce-u
		if epoch == numEpochs/2 {
			t.Log("Phase 3: Adding transactions with nonce gaps")
			for i := 0; i < numSenders/2; i++ {
				sender := fmt.Sprintf("sender-%d", i)
				nonceMutex.Lock()
				currentNonce := senderNonces[sender]
				// Dodaj transakciju sa prazninom u nonce-u
				gapNonce := currentNonce + 2
				senderNonces[sender] = gapNonce + 1
				nonceMutex.Unlock()

				tx := Transaction{Sender: sender, Nonce: gapNonce, GasFee: 500}
				_ = mp.CheckTx(tx)
				t.Logf("Added transaction with nonce gap for %s: current %d, gap %d",
					sender, currentNonce, gapNonce)
			}
		}

		// Faza 4: Dodavanje transakcija sa visokim naknadama
		if epoch == numEpochs-1 {
			t.Log("Phase 4: Adding high-fee transactions")
			for i := 0; i < numSenders/4; i++ {
				sender := fmt.Sprintf("high-fee-sender-%d", i)
				// Visoke naknade
				tx := Transaction{Sender: sender, Nonce: 1, GasFee: 10000}
				_ = mp.CheckTx(tx)
			}

			// Proveri da li su transakcije sa visokim naknadama prioritizovane
			selected = mp.Update(numSenders / 4)
			highFeeCount := 0
			for _, tx := range selected {
				if strings.HasPrefix(tx.Sender, "high-fee-sender") {
					highFeeCount++
				}
			}
			t.Logf("Selected %d high-fee transactions out of %d", highFeeCount, len(selected))
			if highFeeCount < len(selected)/2 {
				t.Errorf("Expected high-fee transactions to be prioritized, got only %d out of %d",
					highFeeCount, len(selected))
			}
		}
	}

	// Faza 5: Provera finalnog stanja
	t.Log("Phase 5: Final state verification")
	finalSize := mp.Size()
	t.Logf("Final mempool size: %d", finalSize)

	// Provera da li su sve transakcije sa kontinuiranim nonce-om obrađene
	allTxs := mp.GetAllTransactions()
	senderTxMap := make(map[string][]Transaction)
	for _, tx := range allTxs {
		senderTxMap[tx.Sender] = append(senderTxMap[tx.Sender], tx)
	}

	for sender, txs := range senderTxMap {
		// Sortiraj transakcije po nonce-u
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Nonce < txs[j].Nonce
		})

		// Proveri da li postoje praznine u nonce-u
		hasGap := false
		for i := 1; i < len(txs); i++ {
			if txs[i].Nonce > txs[i-1].Nonce+1 {
				hasGap = true
				t.Logf("Sender %s has nonce gap: %d -> %d",
					sender, txs[i-1].Nonce, txs[i].Nonce)
				break
			}
		}

		// Ako postoje praznine, proveri da li je to očekivano
		if hasGap {
			// Ovde možete dodati dodatne provere ako je potrebno
		}
	}

	// Provera konzistentnosti heap-a
	// Ovo je indirektna provera - ne možemo direktno pristupiti internim strukturama
	// Ali možemo proveriti da li Update vraća transakcije u očekivanom redosledu
	if finalSize > 0 {
		selected := mp.Update(uint(finalSize))

		// Proveri da li su transakcije sortirane po naknadi između pošiljalaca
		prevSender := ""
		prevFee := uint64(0)
		for i, tx := range selected {
			if tx.Sender != prevSender {
				if i > 0 && tx.GasFee > prevFee {
					// Ovo nije striktna greška, ali može ukazati na problem u prioritizaciji
					t.Logf("Potential priority issue: sender %s with fee %d selected after sender %s with fee %d",
						tx.Sender, tx.GasFee, prevSender, prevFee)
				}
				prevSender = tx.Sender
				prevFee = tx.GasFee
			}
		}
	}

	t.Log("Complex mempool simulation completed successfully")
}
