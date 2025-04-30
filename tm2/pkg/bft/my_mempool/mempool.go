package my_mempool

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/appconn"
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

// GetAccountSequence retrieves the sequence number (nonce) for an account address
func GetAccountSequence(address string, queryClient appconn.Query) (uint64, error) {
	// Create the query request
	path := "auth/accounts/" + address
	reqQuery := abci.RequestQuery{
		Path: path,
		Data: nil,
	}

	// Execute the query
	resp, err := queryClient.QuerySync(reqQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to query account: %w", err)
	}

	// Parse the response to extract the sequence number
	var accountData struct {
		BaseAccount struct {
			Sequence string `json:"sequence"`
		} `json:"BaseAccount"`
	}

	if err := json.Unmarshal(resp.Value, &accountData); err != nil {
		return 0, fmt.Errorf("failed to parse account data: %w", err)
	}

	// Convert sequence string to uint64
	sequence, err := strconv.ParseUint(accountData.BaseAccount.Sequence, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid sequence number format: %w", err)
	}

	return sequence, nil
}

// QueryRealAccount retrieves account information from a remote RPC endpoint
func QueryRealAccount(address string, rpcEndpoint string) (string, error) {
	// Create an HTTP client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// Construct the URL for the query
	url := fmt.Sprintf("%s/abci_query?path=\"auth/accounts/%s\"", rpcEndpoint, address)

	// Make the HTTP request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to query account: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Print the raw response for debugging
	fmt.Printf("Raw response: %s\n", string(body))

	// Parse the JSON response
	var result struct {
		Result struct {
			Response struct {
				ResponseBase struct {
					Data string `json:"Data"`
				} `json:"ResponseBase"`
			} `json:"response"`
		} `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if data is empty
	if result.Result.Response.ResponseBase.Data == "" {
		return "", fmt.Errorf("empty response data")
	}

	// Decode the base64-encoded value
	decodedValue, err := base64.StdEncoding.DecodeString(result.Result.Response.ResponseBase.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// Check if decoded value is empty
	if len(decodedValue) == 0 {
		return "", fmt.Errorf("empty decoded data")
	}

	// Return the decoded value as a string
	return string(decodedValue), nil
}

// GetRealAccountSequence retrieves the sequence number from a real account on the testnet
func GetRealAccountSequence(address string, rpcEndpoint string) (uint64, error) {
	// Query the account
	accountJSON, err := QueryRealAccount(address, rpcEndpoint)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Account JSON for parsing: %s\n", accountJSON)

	// Parse the JSON to extract the sequence number
	var accountData struct {
		BaseAccount struct {
			Sequence string `json:"sequence"`
		} `json:"BaseAccount"`
	}

	if err := json.Unmarshal([]byte(accountJSON), &accountData); err != nil {
		return 0, fmt.Errorf("failed to parse account data: %w", err)
	}

	if accountData.BaseAccount.Sequence == "" {
		return 0, fmt.Errorf("sequence not found in response")
	}

	// Convert sequence string to uint64
	sequence, err := strconv.ParseUint(accountData.BaseAccount.Sequence, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid sequence number format: %w", err)
	}

	return sequence, nil
}
