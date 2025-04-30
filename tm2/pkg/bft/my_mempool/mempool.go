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

// Mempool structure
type Mempool struct {
	txsBySender map[string][]Transaction
	mutex       sync.RWMutex
}

// NewMempool creates a new empty mempool instance
func NewMempool() *Mempool {
	return &Mempool{
		txsBySender: make(map[string][]Transaction),
	}
}

// AddTx validates and adds a transaction to the mempool.
// Transactions for each sender are kept sorted by nonce (ascending).
func (mp *Mempool) AddTx(tx Transaction) error {
	if tx.Sender == "" {
		return errors.New("sender cannot be empty")
	}

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	txList := mp.txsBySender[tx.Sender]
	idx := findInsertIndex(txList, tx.Nonce)

	// Check if transaction with the same nonce exists
	if idx < len(txList) && txList[idx].Nonce == tx.Nonce {
		return nil // already exists
	}

	// Insert transaction at the correct position
	txList = append(txList, Transaction{}) // increase slice size
	copy(txList[idx+1:], txList[idx:])     // shift elements
	txList[idx] = tx                       // insert new tx
	mp.txsBySender[tx.Sender] = txList     // update map
	return nil
}

// findInsertIndex uses binary search to find the insertion index
func findInsertIndex(txList []Transaction, nonce uint64) int {
	return sort.Search(len(txList), func(i int) bool {
		return txList[i].Nonce >= nonce
	})
}

// isValid checks if a transaction is valid for inclusion in a block
// This will be implemented later with more complex validation logic
func (mp *Mempool) isValid(tx Transaction) bool {
	return true
}

// selectOne selects the best transaction based on gas fee
// Returns the transaction with the highest gas fee from all valid transactions
func (mp *Mempool) selectOne() *Transaction {
	var bestTx *Transaction
	var bestSender string

	for sender, txs := range mp.txsBySender {
		if len(txs) == 0 {
			continue
		}
		tx := txs[0]
		if !mp.isValid(tx) {
			continue
		}
		if bestTx == nil || tx.GasFee > bestTx.GasFee {
			bestTx = &tx
			bestSender = sender
		}
	}

	if bestTx == nil {
		return nil
	}

	// Remove the transaction from the sender's list
	mp.txsBySender[bestSender] = mp.txsBySender[bestSender][1:]
	if len(mp.txsBySender[bestSender]) == 0 {
		delete(mp.txsBySender, bestSender)
	}

	return bestTx
}

// CollectTxsForBlock selects transactions for inclusion in a block
// Used primarily for testing purposes
func (mp *Mempool) CollectTxsForBlock(maxTxs uint) []Transaction {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	selected := make([]Transaction, 0, maxTxs)

	for uint(len(selected)) < maxTxs {
		tx := mp.selectOne()
		if tx == nil {
			break
		}
		selected = append(selected, *tx)
	}

	return selected
}

// Update processes committed transactions and removes them from the mempool
// This is typically called after transactions have been included in a block
func (mp *Mempool) Update(committed []Transaction) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for _, tx := range committed {
		txs := mp.txsBySender[tx.Sender]
		newList := make([]Transaction, 0, len(txs))
		for _, existing := range txs {
			if existing.Nonce != tx.Nonce {
				newList = append(newList, existing)
			}
		}
		if len(newList) == 0 {
			delete(mp.txsBySender, tx.Sender)
		} else {
			mp.txsBySender[tx.Sender] = newList
		}
	}
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

// GetTransactionsBySender returns all transactions from a specific sender
// Transactions are sorted by nonce in ascending order
func (mp *Mempool) GetTransactionsBySender(sender string) []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.txsBySender[sender]
}

// GetAllTransactions returns all transactions currently in the mempool
func (mp *Mempool) GetAllTransactions() []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	all := []Transaction{}
	for _, txs := range mp.txsBySender {
		all = append(all, txs...)
	}
	return all
}
