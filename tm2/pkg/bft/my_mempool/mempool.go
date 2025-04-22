package my_mempool

import (
	"container/heap"
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

// senderEntry represents the current representative of a sender
// (i.e., their transaction with the lowest nonce)
type senderEntry struct {
	sender string
	fee    uint64
	index  int
}

// senderHeap implements heap.Interface for prioritizing representatives by gas fee
type senderHeap []*senderEntry

func (h senderHeap) Len() int           { return len(h) }
func (h senderHeap) Less(i, j int) bool { return h[i].fee > h[j].fee }
func (h senderHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}
func (h *senderHeap) Push(x any) {
	e := x.(*senderEntry)
	e.index = len(*h)
	*h = append(*h, e)
}
func (h *senderHeap) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	e.index = -1
	*h = old[:n-1]
	return e
}

// Mempool structure
type Mempool struct {
	txsBySender map[string][]Transaction
	senderHeap  senderHeap
	senderMap   map[string]*senderEntry
	cachedTxs   []Transaction
	mutex       sync.RWMutex
}

// NewMempool creates a new empty mempool instance
func NewMempool() *Mempool {
	return &Mempool{
		txsBySender: make(map[string][]Transaction),
		senderHeap:  make(senderHeap, 0),
		senderMap:   make(map[string]*senderEntry),
		cachedTxs:   make([]Transaction, 0),
	}
}

// CheckTx validates and adds a transaction to the mempool
func (mp *Mempool) CheckTx(tx Transaction) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if tx.Sender == "" {
		return errors.New("sender cannot be empty")
	}

	list := mp.txsBySender[tx.Sender]
	idx := sort.Search(len(list), func(i int) bool {
		return list[i].Nonce > tx.Nonce
	})

	// Check for duplicate transaction
	if idx > 0 && idx <= len(list) && list[idx-1].Nonce == tx.Nonce {
		return nil // transaction already exists
	}

	// Insert at the correct position
	list = append(list, Transaction{})
	copy(list[idx+1:], list[idx:])
	list[idx] = tx
	mp.txsBySender[tx.Sender] = list

	// Update the heap
	minTx := list[0]
	entry, exists := mp.senderMap[tx.Sender]
	if exists {
		entry.fee = minTx.GasFee
		heap.Fix(&mp.senderHeap, entry.index)
	} else {
		entry := &senderEntry{sender: tx.Sender, fee: minTx.GasFee}
		mp.senderMap[tx.Sender] = entry
		heap.Push(&mp.senderHeap, entry)
	}

	return nil
}

// isValid checks if a transaction is valid for inclusion in a block
// This will include verification of expected nonce values in the future
func (mp *Mempool) isValid(tx Transaction) bool {
	return true
}

// selectOne selects the highest fee transaction from the mempool
func (mp *Mempool) selectOne() *Transaction {
	for mp.senderHeap.Len() > 0 {
		e := heap.Pop(&mp.senderHeap).(*senderEntry)
		sender := e.sender
		txs := mp.txsBySender[sender]

		if len(txs) == 0 {
			delete(mp.txsBySender, sender)
			delete(mp.senderMap, sender)
			continue
		}

		tx := txs[0]
		if !mp.isValid(tx) {
			heap.Push(&mp.senderHeap, e)
			continue
		}

		mp.cachedTxs = append(mp.cachedTxs, tx)

		// Remove the transaction from the list
		txs = txs[1:]

		if len(txs) == 0 {
			delete(mp.txsBySender, sender)
			delete(mp.senderMap, sender)
		} else {
			mp.txsBySender[sender] = txs
			e.fee = txs[0].GasFee
			heap.Push(&mp.senderHeap, e)
		}

		return &tx
	}
	return nil
}

// CollectTxsForBlock selects transactions for inclusion in a block
func (mp *Mempool) CollectTxsForBlock(maxTxs uint) []Transaction {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Reset the cache
	mp.cachedTxs = nil
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
			if entry, ok := mp.senderMap[tx.Sender]; ok {
				if entry.index >= 0 && entry.index < len(mp.senderHeap) {
					heap.Remove(&mp.senderHeap, entry.index)
				}
				delete(mp.senderMap, tx.Sender)
			}
		} else {
			mp.txsBySender[tx.Sender] = newList
			if entry, ok := mp.senderMap[tx.Sender]; ok {
				entry.fee = newList[0].GasFee
				heap.Push(&mp.senderHeap, entry)
			}
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

// Rollback returns all cached transactions back to the mempool
func (mp *Mempool) Rollback() {
	mp.mutex.Lock()
	cached := mp.cachedTxs
	mp.cachedTxs = nil
	mp.mutex.Unlock()

	for _, tx := range cached {
		_ = mp.CheckTx(tx)
	}
}
