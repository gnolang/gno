package my_mempool

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

// Tx defines the interface for transactions that can be stored in the mempool.
// Each transaction must provide methods to access its key properties.
type Tx interface {
	Hash() []byte           // Unique identifier of the transaction
	Sender() crypto.Address // Address of the transaction sender
	Sequence() uint64       // Nonce/sequence number for ordering transactions from the same sender
	Price() uint64          // Gas price offered by the transaction
	Size() uint64           // Size of the transaction in bytes
}

// Mempool manages pending transactions organized by sender accounts.
// It ensures transactions are executed in the correct sequence order.
type Mempool struct {
	accounts     sync.Map        // map[crypto.Address]*account - Thread-safe map of sender accounts
	proxyAppConn appconn.Mempool // Connection to the underlying application
}

// account tracks the state of a single sender's transactions in the mempool.
type account struct {
	pending []Tx         // Transactions ready for immediate execution (contiguous nonces)
	queued  []Tx         // Future transactions waiting for their nonce to become current
	nonce   uint64       // Next expected nonce for this account
	mux     sync.RWMutex // Protects concurrent access to account data
}

// NewMempool creates a new mempool instance with the provided application connection.
func NewMempool(proxyAppConn appconn.Mempool) *Mempool {
	return &Mempool{
		proxyAppConn: proxyAppConn,
	}
}

// AddTx adds a new transaction to the mempool.
// It validates the transaction's nonce and places it in either the pending
// or queued list depending on its sequence number.
func (mp *Mempool) AddTx(tx Tx) error {
	// First check transaction with application
	req := abci.RequestCheckTx{
		Tx: tx.Hash(),
	}

	// Send the transaction to the application for validation
	// We use the asynchronous version to avoid blocking
	reqRes := mp.proxyAppConn.CheckTxAsync(req)

	// Wait for validation response
	reqRes.Wait()

	// Check if validation was successful
	res, ok := reqRes.Response.(abci.ResponseCheckTx)
	if !ok {
		return fmt.Errorf("invalid ABCI response type")
	}

	// If application rejected the transaction, return error
	if res.Error != nil {
		return fmt.Errorf("transaction rejected by application: %s", res.Error)
	}

	// Now proceed with mempool logic
	sender := tx.Sender()
	seq := tx.Sequence()

	// Load or initialize account
	accAny, exists := mp.accounts.Load(sender)
	var acc *account

	if !exists {
		// If this is a new sender, get their sequence from the application
		nonce, err := mp.getAccountSequence(sender.String())
		if err != nil {
			return fmt.Errorf("failed to get account sequence: %w", err)
		}
		acc = &account{nonce: nonce}
		mp.accounts.Store(sender, acc)
	} else {
		acc = accAny.(*account)
	}

	acc.mux.Lock()
	defer acc.mux.Unlock()

	switch {
	case seq < acc.nonce:
		return fmt.Errorf("tx nonce too low (expected %d, got %d)", acc.nonce, seq)

	case seq == acc.nonce:
		acc.pending = append(acc.pending, tx)
		acc.nonce++
		mp.promoteReadyTxs(acc)

	default:
		// Insert into queued in nonce-sorted order
		index := sort.Search(len(acc.queued), func(i int) bool {
			return acc.queued[i].Sequence() >= seq
		})
		acc.queued = append(acc.queued, nil)
		copy(acc.queued[index+1:], acc.queued[index:])
		acc.queued[index] = tx
	}

	return nil
}

// promoteReadyTxs checks the queued transactions and moves contiguous ones to pending.
// This ensures that all transactions with sequential nonces become available
// for execution as soon as their prerequisites are met.
func (mp *Mempool) promoteReadyTxs(acc *account) {
	for len(acc.queued) > 0 {
		next := acc.queued[0]
		if next.Sequence() != acc.nonce {
			break
		}
		acc.pending = append(acc.pending, next)
		acc.queued = acc.queued[1:]
		acc.nonce++
	}
}

// Pending returns a map of sender address to their pending transaction list.
// These transactions are ready for inclusion in the next block.
func (mp *Mempool) Pending() map[crypto.Address][]Tx {
	pendingMap := make(map[crypto.Address][]Tx)

	mp.accounts.Range(func(key, value any) bool {
		addr := key.(crypto.Address)
		acc := value.(*account)

		acc.mux.RLock()
		if len(acc.pending) > 0 {
			pendingMap[addr] = append([]Tx(nil), acc.pending...)
		}
		acc.mux.RUnlock()
		return true
	})

	return pendingMap
}

// RemoveTx removes a transaction from the sender's pending or queued lists by hash.
// Used when transactions are included in a block or become invalid.
func (mp *Mempool) RemoveTx(sender crypto.Address, hash []byte) {
	accAny, ok := mp.accounts.Load(sender)
	if !ok {
		return
	}

	acc := accAny.(*account)
	acc.mux.Lock()
	defer acc.mux.Unlock()

	// Helper for removing a tx by hash from slice
	removeByHash := func(txs []Tx) []Tx {
		for i, tx := range txs {
			if bytes.Equal(tx.Hash(), hash) {
				return append(txs[:i], txs[i+1:]...)
			}
		}
		return txs
	}

	acc.pending = removeByHash(acc.pending)
	acc.queued = removeByHash(acc.queued)

	if len(acc.pending) == 0 && len(acc.queued) == 0 {
		mp.accounts.Delete(sender)
	}
}

// Update removes committed transactions and promotes ready ones from queued.
// Called after a block is committed to keep the mempool state in sync.
func (mp *Mempool) Update(committed []Tx) {
	for _, tx := range committed {
		sender := tx.Sender()
		hash := tx.Hash()

		mp.RemoveTx(sender, hash)
	}
}

// Tx returns the transaction with the given hash, if present in the mempool.
func (mp *Mempool) Tx(hash []byte) Tx {
	var result Tx

	mp.accounts.Range(func(_, value any) bool {
		acc := value.(*account)

		acc.mux.RLock()
		defer acc.mux.RUnlock()

		for _, list := range [][]Tx{acc.pending, acc.queued} {
			for _, tx := range list {
				if bytes.Equal(tx.Hash(), hash) {
					result = tx
					return false
				}
			}
		}
		return true
	})

	return result
}

// Content returns all transactions in the mempool (pending + queued).
func (mp *Mempool) Content() []Tx {
	var all []Tx
	mp.accounts.Range(func(_, value any) bool {
		acc := value.(*account)
		acc.mux.RLock()
		all = append(all, acc.pending...)
		all = append(all, acc.queued...)
		acc.mux.RUnlock()
		return true
	})
	return all
}

// Flush removes all transactions from the mempool.
// Typically used during blockchain resets or major state changes.
func (mp *Mempool) Flush() {
	mp.accounts.Range(func(key, value any) bool {
		mp.accounts.Delete(key)
		return true
	})
}

// getAccountSequence queries the app for the current sequence of the account.
// This establishes the baseline nonce for new accounts in the mempool.
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

// GetQueuedTxs returns all queued transactions for a given address.
// Used primarily for testing and debugging.
func (mp *Mempool) GetQueuedTxs(address crypto.Address) []Tx {
	accAny, exists := mp.accounts.Load(address)
	if !exists {
		return nil
	}

	acc := accAny.(*account)
	acc.mux.RLock()
	defer acc.mux.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]Tx, len(acc.queued))
	copy(result, acc.queued)
	return result
}

// Size returns the total number of transactions in the mempool.
func (mp *Mempool) Size() int {
	total := 0
	mp.accounts.Range(func(_, value any) bool {
		acc := value.(*account)
		acc.mux.RLock()
		total += len(acc.pending) + len(acc.queued)
		acc.mux.RUnlock()
		return true
	})
	return total
}

// GetTxsBySender returns all transactions (pending + queued) for a given sender.
func (mp *Mempool) GetTxsBySender(sender crypto.Address) []Tx {
	accAny, ok := mp.accounts.Load(sender)
	if !ok {
		return nil
	}

	acc := accAny.(*account)
	acc.mux.RLock()
	defer acc.mux.RUnlock()

	txs := make([]Tx, 0, len(acc.pending)+len(acc.queued))
	txs = append(txs, acc.pending...)
	txs = append(txs, acc.queued...)
	return txs
}

// GetPendingBySender returns only the pending transactions for a given sender.
// These are transactions ready for immediate execution.
func (mp *Mempool) GetPendingBySender(sender crypto.Address) []Tx {
	accAny, ok := mp.accounts.Load(sender)
	if !ok {
		return nil
	}

	acc := accAny.(*account)
	acc.mux.RLock()
	defer acc.mux.RUnlock()

	copied := make([]Tx, len(acc.pending))
	copy(copied, acc.pending)
	return copied
}

// GetExpectedNonce returns the current expected nonce for the given sender.
// This is the sequence number that the next transaction should have.
func (mp *Mempool) GetExpectedNonce(sender crypto.Address) (uint64, bool) {
	accAny, ok := mp.accounts.Load(sender)
	if !ok {
		return 0, false
	}
	acc := accAny.(*account)
	acc.mux.RLock()
	defer acc.mux.RUnlock()
	return acc.nonce, true
}
