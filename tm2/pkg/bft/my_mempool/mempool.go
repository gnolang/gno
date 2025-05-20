package my_mempool

import (
	"fmt"
	"sync"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Mempool implements a FIFO transaction queue for blockchain consensus.
// It maintains an ordered list of valid transactions ready for consensus inclusion.
type Mempool struct {
	txMap        map[string]txEntry // Transactions indexed by hash string
	txHashes     []string           // Ordered list preserving FIFO sequence
	mutex        sync.RWMutex       // Synchronizes concurrent access
	proxyAppConn appconn.Mempool    // Connection to the underlying application
	txsBytes     int64              // Total size of all transactions in bytes
}

// txEntry encapsulates a transaction with its associated metadata.
type txEntry struct {
	tx        types.Tx  // Transaction data
	gasWanted int64     // Gas amount requested by CheckTx
	timestamp time.Time // Time when transaction was added
}

// NewMempool creates and initializes a new Mempool with the provided application connection.
func NewMempool(proxyAppConn appconn.Mempool) *Mempool {
	return &Mempool{
		proxyAppConn: proxyAppConn,
		txMap:        make(map[string]txEntry),
		txHashes:     make([]string, 0, 1024), // Pre-allocate memory for efficiency
	}
}

// AddTx validates and adds a transaction to the mempool.
// Returns error if the transaction is invalid or already exists.
func (mp *Mempool) AddTx(tx types.Tx) error {
	txHash := tx.Hash()
	hashStr := string(txHash)

	mp.mutex.RLock()
	_, exists := mp.txMap[hashStr]
	mp.mutex.RUnlock()

	if exists {
		return fmt.Errorf("transaction already exists in mempool")
	}

	// Validate transaction with application
	req := abci.RequestCheckTx{
		Tx: tx,
	}

	reqRes := mp.proxyAppConn.CheckTxAsync(req)
	reqRes.Wait()

	res, ok := reqRes.Response.(abci.ResponseCheckTx)
	if !ok {
		return fmt.Errorf("invalid ABCI response type")
	}

	if res.Error != nil {
		return fmt.Errorf("transaction rejected by application: %s", res.Error)
	}

	// Store the validated transaction
	entry := txEntry{
		tx:        tx,
		gasWanted: res.GasWanted,
		timestamp: time.Now(),
	}

	txSize := int64(len(tx))
	mp.mutex.Lock()
	mp.txMap[hashStr] = entry
	mp.txHashes = append(mp.txHashes, hashStr)
	mp.txsBytes += txSize
	mp.mutex.Unlock()

	return nil
}

// RemoveTx removes a transaction from the mempool by its hash.
// This operation maintains the FIFO order of remaining transactions.
func (mp *Mempool) RemoveTx(hash []byte) {
	hashStr := string(hash)

	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	entry, exists := mp.txMap[hashStr]
	if !exists {
		return
	}

	delete(mp.txMap, hashStr)
	mp.txsBytes -= int64(len(entry.tx))

	for i, h := range mp.txHashes {
		if h == hashStr {
			mp.txHashes = append(mp.txHashes[:i], mp.txHashes[i+1:]...)
			break
		}
	}
}

// Update synchronizes the mempool state by removing transactions that were
// committed in a block.
func (mp *Mempool) Update(committed []types.Tx) {
	for _, tx := range committed {
		mp.RemoveTx(tx.Hash())
	}
}

// ReapMaxBytesMaxGas selects transactions from the mempool that fit within
// the specified gas and byte limits, maintaining their original FIFO order.
func (mp *Mempool) ReapMaxBytesMaxGas(maxBytes, maxGas int64) []types.Tx {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	if maxBytes <= 0 {
		return nil
	}

	return mp.iterateTransactions(func(entry txEntry, totalGas, totalBytes int64) bool {
		txSize := int64(len(entry.tx))

		if totalBytes+txSize > maxBytes {
			return false
		}

		if maxGas > 0 && totalGas+entry.gasWanted > maxGas {
			return false
		}

		return true
	})
}

// Content returns all transactions currently in the mempool in FIFO order.
func (mp *Mempool) Content() []types.Tx {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()

	return mp.iterateTransactions(nil)
}

// Flush empties the mempool, removing all pending transactions.
func (mp *Mempool) Flush() {
	mp.mutex.Lock()
	mp.txMap = make(map[string]txEntry)
	mp.txHashes = make([]string, 0, 1024)
	mp.txsBytes = 0
	mp.mutex.Unlock()
}

// iterateTransactions processes mempool transactions applying optional limit criteria.
// The limitFunc parameter allows flexible filtering based on transaction attributes.
func (mp *Mempool) iterateTransactions(limitFunc func(entry txEntry, totalGas, totalBytes int64) bool) []types.Tx {
	var (
		txs        = make([]types.Tx, 0, len(mp.txHashes))
		totalGas   int64
		totalBytes int64
	)

	for _, hashStr := range mp.txHashes {
		entry, exists := mp.txMap[hashStr]
		if !exists {
			continue
		}

		txSize := int64(len(entry.tx))

		if limitFunc != nil && !limitFunc(entry, totalGas, totalBytes) {
			break
		}

		txs = append(txs, entry.tx)
		totalGas += entry.gasWanted
		totalBytes += txSize
	}

	return txs
}

// GetTx retrieves a transaction by its hash if present in the mempool.
// Returns the transaction and a boolean indicating if it was found.
func (mp *Mempool) GetTx(hash []byte) (types.Tx, bool) {
	mp.mutex.RLock()
	entry, exists := mp.txMap[string(hash)]
	mp.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	return entry.tx, true
}

// Size returns the number of transactions in the mempool.
func (mp *Mempool) Size() int {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return len(mp.txHashes)
}

// TxsBytes returns the total size of all transactions in the mempool in bytes.
func (mp *Mempool) TxsBytes() int64 {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	return mp.txsBytes
}
