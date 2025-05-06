package my_mempool

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
)

// Transaction represents a basic transaction structure with sender information,
// nonce for ordering, and gas fee for prioritization.
type Transaction struct {
	Sender string // Account address that initiated the transaction
	Nonce  uint64 // Sequential number to prevent replay attacks and ensure ordering
	GasFee uint64 // Fee paid for transaction execution, used for prioritization
}

// Mempool manages pending transactions before they are included in a block.
// It implements a nonce-based ordering system with gas fee prioritization.
type Mempool struct {
	txsBySender    map[string][]Transaction // Future transactions with nonces > expectedNonce
	pendingTxs     map[string]Transaction   // Transactions ready for immediate processing
	expectedNonces map[string]uint64        // Next valid nonce for each sender account
	proxyAppConn   appconn.Mempool          // Interface to query application state
	mutex          sync.RWMutex             // Protects concurrent access to mempool state
}

// NewMempool creates a new mempool instance with the provided application connection.
func NewMempool(proxyAppConn appconn.Mempool) *Mempool {
	return &Mempool{
		txsBySender:    make(map[string][]Transaction),
		pendingTxs:     make(map[string]Transaction),
		expectedNonces: make(map[string]uint64),
		proxyAppConn:   proxyAppConn,
	}
}

// AddTx validates and adds a transaction to the mempool.
// Returns an error if the transaction is invalid or has a nonce that is too low.
func (mp *Mempool) AddTx(tx Transaction) error {
	if tx.Sender == "" {
		return errors.New("sender cannot be empty")
	}

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Fetch account sequence from application state if not cached
	if _, ok := mp.expectedNonces[tx.Sender]; !ok {
		seq, err := mp.getAccountSequence(tx.Sender)
		if err != nil {
			return fmt.Errorf("failed to get expected nonce: %w", err)
		}
		mp.expectedNonces[tx.Sender] = seq
	}

	// Reject transactions with nonces lower than expected (already processed)
	if tx.Nonce < mp.expectedNonces[tx.Sender] {
		return fmt.Errorf("tx nonce too low (expected %d, got %d)", mp.expectedNonces[tx.Sender], tx.Nonce)
	}

	// If nonce matches expected nonce, add to pendingTxs for immediate processing
	if tx.Nonce == mp.expectedNonces[tx.Sender] {
		mp.pendingTxs[tx.Sender] = tx
		return nil
	}

	// Store future transactions in sorted order by nonce
	txList := mp.txsBySender[tx.Sender]
	idx := findInsertIndex(txList, tx.Nonce)

	if idx < len(txList) {
		// Insert in the middle
		txList = append(txList[:idx+1], txList[idx:]...)
		txList[idx] = tx
	} else {
		// Append at the end
		txList = append(txList, tx)
	}
	mp.txsBySender[tx.Sender] = txList
	return nil
}

// findInsertIndex uses binary search to find the correct insertion index
// for a transaction with the given nonce in a sorted transaction list.
func findInsertIndex(txList []Transaction, nonce uint64) int {
	return sort.Search(len(txList), func(i int) bool {
		return txList[i].Nonce >= nonce
	})
}

// promoteReadyTx checks if there's a transaction in txsBySender with the given nonce
// and moves it to pendingTxs if found. Returns true if a transaction was promoted.
func (mp *Mempool) promoteReadyTx(sender string, nonce uint64) bool {
	txList := mp.txsBySender[sender]
	if len(txList) == 0 {
		return false
	}

	idx := findInsertIndex(txList, nonce)
	if idx < len(txList) && txList[idx].Nonce == nonce {
		// Found a ready transaction, move it to pendingTxs
		mp.pendingTxs[sender] = txList[idx]

		// Remove from txsBySender
		mp.txsBySender[sender] = append(txList[:idx], txList[idx+1:]...)
		if len(mp.txsBySender[sender]) == 0 {
			delete(mp.txsBySender, sender)
		}
		return true
	}
	return false
}

// selectBestReadyTx selects the transaction with the highest gas fee from all
// pending transactions that are ready to be processed.
// Returns nil if no transactions are ready.
func (mp *Mempool) selectBestReadyTx() *Transaction {
	if len(mp.pendingTxs) == 0 {
		return nil
	}

	var bestTx *Transaction
	var bestSender string

	// Select the pending transaction with the highest gas fee
	for sender, tx := range mp.pendingTxs {
		if bestTx == nil || tx.GasFee > bestTx.GasFee {
			txCopy := tx // Make a copy to avoid issues with map iteration
			bestTx = &txCopy
			bestSender = sender
		}
	}

	if bestTx == nil {
		return nil
	}

	// Remove the selected transaction from pendingTxs
	delete(mp.pendingTxs, bestSender)

	// Update the expected nonce for this sender
	mp.expectedNonces[bestSender]++

	// Check if we have another transaction from this sender with the new expected nonce
	mp.promoteReadyTx(bestSender, mp.expectedNonces[bestSender])

	return bestTx
}

// CollectTxsForBlock selects transactions for inclusion in a block based on
// correct nonce ordering and gas fee prioritization.
// Returns at most maxTxs transactions.
func (mp *Mempool) CollectTxsForBlock(maxTxs uint) []Transaction {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	selected := make([]Transaction, 0, maxTxs)

	for uint(len(selected)) < maxTxs {
		tx := mp.selectBestReadyTx()
		if tx == nil {
			break
		}
		selected = append(selected, *tx)
	}

	return selected
}

// Update processes committed transactions and removes them from the mempool.
// It also updates the expected nonces for affected senders and promotes
// any transactions that become ready as a result.
func (mp *Mempool) Update(committed []Transaction) {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	for _, tx := range committed {
		// Check if this transaction is in pendingTxs
		if pendingTx, ok := mp.pendingTxs[tx.Sender]; ok && pendingTx.Nonce == tx.Nonce {
			delete(mp.pendingTxs, tx.Sender)
		}

		// Also remove from txsBySender if it exists there
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

		// Update expected nonce if this transaction's nonce matches the current expected nonce
		if tx.Nonce == mp.expectedNonces[tx.Sender] {
			mp.expectedNonces[tx.Sender]++
			mp.promoteReadyTx(tx.Sender, mp.expectedNonces[tx.Sender])
		}
	}
}

// Size returns the total number of transactions in the mempool.
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	count := len(mp.pendingTxs)
	for _, txs := range mp.txsBySender {
		count += len(txs)
	}
	return count
}

// GetTransactionsBySender returns all transactions from a specific sender.
// Transactions are sorted by nonce in ascending order.
func (mp *Mempool) GetTransactionsBySender(sender string) []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	result := []Transaction{}

	// Add pending transaction if exists
	if pendingTx, ok := mp.pendingTxs[sender]; ok {
		result = append(result, pendingTx)
	}

	// Add other transactions
	result = append(result, mp.txsBySender[sender]...)

	return result
}

// GetAllTransactions returns all transactions currently in the mempool.
func (mp *Mempool) GetAllTransactions() []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	all := make([]Transaction, 0, len(mp.pendingTxs))

	// Add all pending transactions
	for _, tx := range mp.pendingTxs {
		all = append(all, tx)
	}

	// Add all other transactions
	for _, txs := range mp.txsBySender {
		all = append(all, txs...)
	}

	return all
}

// getAccountSequence retrieves the sequence number (nonce) for an account address
// by querying the application state through ABCI interface.
func (mp *Mempool) getAccountSequence(address string) (uint64, error) {
	path := "auth/accounts/" + address
	reqQuery := abci.RequestQuery{Path: path}

	resp, err := mp.proxyAppConn.QuerySync(reqQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query account: %w", err)
	}

	var accountData struct {
		BaseAccount struct {
			Sequence string `json:"sequence"`
		} `json:"BaseAccount"`
	}

	if err := json.Unmarshal(resp.Value, &accountData); err != nil {
		return 0, fmt.Errorf("failed to parse account data: %w", err)
	}

	sequence, err := strconv.ParseUint(accountData.BaseAccount.Sequence, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid sequence number format: %w", err)
	}

	return sequence, nil
}
