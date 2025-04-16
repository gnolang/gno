package my_mempool

import (
	"container/heap"
	"errors"
	"sort"
	"sync"
)

// senderEntry represents an entry in the sender heap
type senderEntry struct {
	sender string
	fee    uint64
	index  int // required for heap.Fix
}

// senderHeap implements heap.Interface for prioritizing senders by fee
type senderHeap []*senderEntry

func (h senderHeap) Len() int { return len(h) }
func (h senderHeap) Less(i, j int) bool {
	// inverted for max-heap by fee
	return h[i].fee > h[j].fee
}
func (h senderHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *senderHeap) Push(x any) {
	entry := x.(*senderEntry)
	entry.index = len(*h)
	*h = append(*h, entry)
}

func (h *senderHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1 // safety measure
	*h = old[0 : n-1]
	return item
}

// Transaction represents a basic transaction structure
type Transaction struct {
	Sender string
	Nonce  uint64
	GasFee uint64
}

// Mempool stores transactions grouped by sender address
type Mempool struct {
	txsBySender map[string][]Transaction
	senderHeap  senderHeap
	senderMap   map[string]*senderEntry
	mutex       sync.RWMutex
}

// NewMempool creates a new mempool instance
func NewMempool() *Mempool {
	return &Mempool{
		txsBySender: make(map[string][]Transaction),
		senderHeap:  make(senderHeap, 0),
		senderMap:   make(map[string]*senderEntry),
	}
}

// CheckTx validates and adds a transaction to the mempool
func (mp *Mempool) CheckTx(tx Transaction) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if tx.Sender == "" {
		return errors.New("sender address cannot be empty")
	}

	mp.txsBySender[tx.Sender] = append(mp.txsBySender[tx.Sender], tx)

	// Update heap
	entry, exists := mp.senderMap[tx.Sender]
	if exists {
		if tx.GasFee > entry.fee {
			entry.fee = tx.GasFee
			heap.Fix(&mp.senderHeap, entry.index)
		}
	} else {
		entry = &senderEntry{
			sender: tx.Sender,
			fee:    tx.GasFee,
		}
		mp.senderMap[tx.Sender] = entry
		heap.Push(&mp.senderHeap, entry)
	}

	return nil
}

// Update selects and removes transactions from the mempool based on priority
// Returns selected transactions up to maxTxs limit
func (mp *Mempool) Update(maxTxs uint) []Transaction {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Early return if maxTxs is 0
	if maxTxs == 0 {
		return []Transaction{}
	}

	selectedTxs := make([]Transaction, 0, maxTxs)
	txsToRemove := make(map[string][]int)

	// Step 1: Create a copy of the heap (to avoid mutating the main heap during update)
	heapCopy := make(senderHeap, len(mp.senderHeap))
	copy(heapCopy, mp.senderHeap)
	heap.Init(&heapCopy)

	const maxPerSender = 10

	for heapCopy.Len() > 0 && uint(len(selectedTxs)) < maxTxs {
		entry := heap.Pop(&heapCopy).(*senderEntry)
		sender := entry.sender
		txs := mp.txsBySender[sender]

		if len(txs) == 0 {
			continue
		}

		// Sort by nonce
		sort.Slice(txs, func(i, j int) bool {
			return txs[i].Nonce < txs[j].Nonce
		})

		var senderTxs []Transaction
		var indicesToRemove []int

		expectedNonce := txs[0].Nonce
		count := 0

		// Calculate how many more transactions we can add
		remainingCapacity := int(maxTxs) - len(selectedTxs)
		// Limit by both maxPerSender and remaining capacity
		maxToAdd := min(maxPerSender, remainingCapacity)

		for i, tx := range txs {
			if count >= maxToAdd {
				break
			}
			if tx.Nonce != expectedNonce {
				break
			}
			senderTxs = append(senderTxs, tx)
			indicesToRemove = append(indicesToRemove, i)
			expectedNonce++
			count++
		}

		selectedTxs = append(selectedTxs, senderTxs...)
		txsToRemove[sender] = indicesToRemove
	}

	// Step 2: Remove transactions and update the heap
	for sender, indices := range txsToRemove {
		txs := mp.txsBySender[sender]
		for i := len(indices) - 1; i >= 0; i-- {
			idx := indices[i]
			if idx >= len(txs) {
				continue
			}
			txs = append(txs[:idx], txs[idx+1:]...)
		}
		mp.txsBySender[sender] = txs

		if len(txs) == 0 {
			delete(mp.txsBySender, sender)
			if entry, exists := mp.senderMap[sender]; exists {
				if entry.index >= 0 && entry.index < len(mp.senderHeap) {
					heap.Remove(&mp.senderHeap, entry.index)
				}
				delete(mp.senderMap, sender)
			}
			continue
		}

		// Otherwise: update fee and position in heap
		newMax := txs[0].GasFee
		for _, tx := range txs {
			if tx.GasFee > newMax {
				newMax = tx.GasFee
			}
		}
		if entry, exists := mp.senderMap[sender]; exists {
			entry.fee = newMax
			if entry.index >= 0 && entry.index < len(mp.senderHeap) {
				heap.Fix(&mp.senderHeap, entry.index)
			}
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
