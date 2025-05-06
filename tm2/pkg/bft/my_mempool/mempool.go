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
	Sender string
	Nonce  uint64
	GasFee uint64
}

// Mempool manages pending transactions before they are included in a block.
// It organizes transactions by sender and maintains expected nonces to ensure
// transactions are processed in the correct order.
type Mempool struct {
	txsBySender    map[string][]Transaction // Transactions organized by sender address
	expectedNonces map[string]uint64        // Expected next nonce for each sender
	proxyAppConn   appconn.Mempool          // Connection to the application for queries
	mutex          sync.RWMutex             // Mutex for thread-safe operations
}

// NewMempool creates a new mempool instance with the provided application connection.
func NewMempool(proxyAppConn appconn.Mempool) *Mempool {
	return &Mempool{
		txsBySender:    make(map[string][]Transaction),
		expectedNonces: make(map[string]uint64),
		proxyAppConn:   proxyAppConn,
	}
}

// AddTx validates and adds a transaction to the mempool.
// Transactions for each sender are kept sorted by nonce (ascending).
// Returns an error if the transaction is invalid or has a nonce that is too low.
func (mp *Mempool) AddTx(tx Transaction) error {
	if tx.Sender == "" {
		return errors.New("sender cannot be empty")
	}

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// If we don't have a cached nonce for this sender, fetch it from the application
	if _, ok := mp.expectedNonces[tx.Sender]; !ok {
		seq, err := mp.getAccountSequence(tx.Sender)
		if err != nil {
			return fmt.Errorf("failed to get expected nonce: %w", err)
		}
		mp.expectedNonces[tx.Sender] = seq
	}

	// Reject transactions with nonces lower than expected
	if tx.Nonce < mp.expectedNonces[tx.Sender] {
		return fmt.Errorf("tx nonce too low (expected %d, got %d)", mp.expectedNonces[tx.Sender], tx.Nonce)
	}

	// Insert into sorted array by nonce
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

// selectBestReadyTx selects the transaction with the highest gas fee from all
// transactions that have the correct nonce (ready to be processed).
// Returns nil if no transactions are ready.
func (mp *Mempool) selectBestReadyTx() *Transaction {
	var bestTx *Transaction
	var bestSender string

	for sender, txs := range mp.txsBySender {
		if len(txs) == 0 {
			continue
		}

		// Get the expected nonce for this sender
		expectedNonce, exists := mp.expectedNonces[sender]
		if !exists {
			// If we don't have the expected nonce in cache, skip this sender
			continue
		}

		// Check if the first transaction has the correct nonce
		tx := txs[0]
		if tx.Nonce != expectedNonce {
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

	// Remove the selected transaction from the sender's list
	mp.txsBySender[bestSender] = mp.txsBySender[bestSender][1:]
	if len(mp.txsBySender[bestSender]) == 0 {
		delete(mp.txsBySender, bestSender)
	}

	// Update the expected nonce for this sender
	mp.expectedNonces[bestSender]++

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
// This is typically called after transactions have been included in a block.
// It also updates the expected nonces for affected senders.
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

		// Update expected nonce if this transaction's nonce matches the current expected nonce
		if expectedNonce, exists := mp.expectedNonces[tx.Sender]; exists && tx.Nonce == expectedNonce {
			mp.expectedNonces[tx.Sender] = tx.Nonce + 1
		}
	}
}

// Size returns the total number of transactions in the mempool.
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	count := 0
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
	return mp.txsBySender[sender]
}

// GetAllTransactions returns all transactions currently in the mempool.
func (mp *Mempool) GetAllTransactions() []Transaction {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	all := []Transaction{}
	for _, txs := range mp.txsBySender {
		all = append(all, txs...)
	}
	return all
}

// getAccountSequence retrieves the sequence number (nonce) for an account address
// by querying the application state.
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
