package my_mempool

import (
	"errors"
	"sort"
	"sync"
)

// Transaction represents a basic transaction structure
type Transaction struct {
	Sender string
	Nonce  uint64
	GasFee uint64
}

// Mempool stores transactions grouped by sender address
type Mempool struct {
	txsBySender          map[string][]Transaction
	highestFeeTxBySender map[string]Transaction
	mutex                sync.RWMutex
}

// NewMempool creates a new mempool instance
func NewMempool() *Mempool {
	return &Mempool{
		txsBySender:          make(map[string][]Transaction),
		highestFeeTxBySender: make(map[string]Transaction),
	}
}

// CheckTx validates and adds a transaction to the mempool
func (mp *Mempool) CheckTx(tx Transaction) error {
	// Acquire write lock
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Basic validation
	if tx.Sender == "" {
		return errors.New("sender address cannot be empty")
	}

	// Add transaction to the mempool
	mp.txsBySender[tx.Sender] = append(mp.txsBySender[tx.Sender], tx)

	// Update highest fee transaction for this sender if needed
	currentHighest, exists := mp.highestFeeTxBySender[tx.Sender]
	if !exists || tx.GasFee > currentHighest.GasFee {
		mp.highestFeeTxBySender[tx.Sender] = tx
	}

	return nil
}

// Update selects and removes a specified number of transactions from the mempool
// prioritizing transactions with the highest gas fees
func (mp *Mempool) Update(maxTxs uint) []Transaction {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	selectedTxs := make([]Transaction, 0, maxTxs)
	txsToRemove := make(map[string][]int)

	// Step 1: Sort senders by their highest fee transaction
	type senderWithHighestFee struct {
		sender    string
		highestTx Transaction
	}

	sendersByFee := make([]senderWithHighestFee, 0, len(mp.highestFeeTxBySender))
	for sender, tx := range mp.highestFeeTxBySender {
		sendersByFee = append(sendersByFee, senderWithHighestFee{
			sender:    sender,
			highestTx: tx,
		})
	}

	// Sort senders by gas fee in descending order
	sort.Slice(sendersByFee, func(i, j int) bool {
		return sendersByFee[i].highestTx.GasFee > sendersByFee[j].highestTx.GasFee
	})

	// Step 2: Process transactions from senders with highest fees first
	for _, senderInfo := range sendersByFee {
		sender := senderInfo.sender
		txs := mp.txsBySender[sender]

		if len(txs) == 0 {
			continue
		}

		// Sort transactions by nonce in ascending order
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Nonce < txs[j].Nonce
		})

		// Add transactions from this sender until we reach maxTxs
		for i, tx := range txs {
			if uint(len(selectedTxs)) >= maxTxs {
				break
			}

			selectedTxs = append(selectedTxs, tx)

			// Track which transactions to remove
			if _, exists := txsToRemove[sender]; !exists {
				txsToRemove[sender] = []int{}
			}
			txsToRemove[sender] = append(txsToRemove[sender], i)
		}

		if uint(len(selectedTxs)) >= maxTxs {
			break
		}
	}

	// Remove selected transactions from the mempool
	for sender, indices := range txsToRemove {
		// Remove in reverse order to avoid index shifting problems
		for i := len(indices) - 1; i >= 0; i-- {
			idx := indices[i]
			txs := mp.txsBySender[sender]

			// Check if we're removing the highest fee transaction
			if mp.highestFeeTxBySender[sender].Nonce == txs[idx].Nonce &&
				mp.highestFeeTxBySender[sender].GasFee == txs[idx].GasFee {
				// We need to find a new highest fee transaction or delete the entry
				delete(mp.highestFeeTxBySender, sender)
			}

			mp.txsBySender[sender] = append(txs[:idx], txs[idx+1:]...)
		}

		// If we deleted the highest fee transaction, find a new one
		if _, exists := mp.highestFeeTxBySender[sender]; !exists && len(mp.txsBySender[sender]) > 0 {
			// Find the new highest fee transaction
			highest := mp.txsBySender[sender][0]
			for _, tx := range mp.txsBySender[sender] {
				if tx.GasFee > highest.GasFee {
					highest = tx
				}
			}
			mp.highestFeeTxBySender[sender] = highest
		}

		// If no transactions left for this sender, remove the sender entry
		if len(mp.txsBySender[sender]) == 0 {
			delete(mp.txsBySender, sender)
			delete(mp.highestFeeTxBySender, sender)
		}
	}

	return selectedTxs
}

// GetAllTransactions returns all transactions in the mempool
func (mp *Mempool) GetAllTransactions() []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	allTxs := []Transaction{}
	for _, txs := range mp.txsBySender {
		allTxs = append(allTxs, txs...)
	}

	return allTxs
}

// GetTransactionsBySender returns all transactions for a specific sender
func (mp *Mempool) GetTransactionsBySender(sender string) []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	return mp.txsBySender[sender]
}

// GetHighestFeeTxBySender returns the transaction with the highest gas fee for a specific sender
func (mp *Mempool) GetHighestFeeTxBySender(sender string) (Transaction, bool) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	tx, exists := mp.highestFeeTxBySender[sender]
	return tx, exists
}

// Size returns the total number of transactions in the mempool
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	count := 0
	for _, txs := range mp.txsBySender {
		count += len(txs)
	}

	return count
}
