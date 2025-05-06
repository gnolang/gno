package my_mempool

import (
	"fmt"
	"sync"
	"testing"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMempoolClient implements the appconn.Mempool interface for testing
type MockMempoolClient struct {
	mock.Mock
}

func (m *MockMempoolClient) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMempoolClient) FlushAsync() *abcicli.ReqRes {
	args := m.Called()
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockMempoolClient) CheckTxAsync(req abci.RequestCheckTx) *abcicli.ReqRes {
	args := m.Called(req)
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockMempoolClient) CheckTxSync(tx []byte) (abci.ResponseCheckTx, error) {
	args := m.Called(tx)
	return args.Get(0).(abci.ResponseCheckTx), args.Error(1)
}

func (m *MockMempoolClient) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	args := m.Called(req)
	return args.Get(0).(abci.ResponseQuery), args.Error(1)
}

func (m *MockMempoolClient) FlushSync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMempoolClient) SetResponseCallback(cb abcicli.Callback) {
	m.Called(cb)
}

// Helper function to create a mock mempool client with default sequence responses
func newMockMempoolClient() *MockMempoolClient {
	mockClient := new(MockMempoolClient)

	// Setup default response for account sequence queries
	mockClient.On("QuerySync", mock.Anything).Return(abci.ResponseQuery{
		Value: []byte(`{"BaseAccount":{"sequence":"1"}}`),
	}, nil)

	return mockClient
}

// Helper function to create a mempool with a mock client for testing
func newTestMempool() (*Mempool, *MockMempoolClient) {
	mockClient := newMockMempoolClient()
	return NewMempool(mockClient), mockClient
}

func TestBasicInsertAndRetrieve(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions
	err := mp.AddTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	assert.NoError(t, err)

	err = mp.AddTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 200})
	assert.NoError(t, err)

	// Check size
	assert.Equal(t, 2, mp.Size(), "Expected mempool size to be 2")

	// Check all transactions
	allTxs := mp.GetAllTransactions()
	assert.Len(t, allTxs, 2, "Expected 2 transactions in total")

	// Check transactions by sender
	aliceTxs := mp.GetTransactionsBySender("alice")
	assert.Len(t, aliceTxs, 1, "Expected 1 transaction for Alice")
	assert.Equal(t, uint64(1), aliceTxs[0].Nonce, "Expected Alice's transaction to have nonce 1")
}

func TestInsertionOrder(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions in non-sequential order
	mp.AddTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 150})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 120})

	// Check that they're stored in nonce order
	txs := mp.GetTransactionsBySender("alice")
	assert.Len(t, txs, 3, "Expected 3 transactions for Alice")

	for i, nonce := range []uint64{1, 2, 3} {
		assert.Equal(t, nonce, txs[i].Nonce,
			fmt.Sprintf("Expected transaction at index %d to have nonce %d", i, nonce))
	}
}

func TestSelectionOrder(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions for multiple senders
	mp.AddTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 150})
	mp.AddTx(Transaction{Sender: "bob", Nonce: 1, GasFee: 200})
	mp.AddTx(Transaction{Sender: "charlie", Nonce: 1, GasFee: 150})

	// Collect transactions
	selected := mp.CollectTxsForBlock(10)

	// Expected order: bob (highest fee), charlie, alice, then alice's second tx
	assert.Len(t, selected, 4, "Expected 4 transactions to be selected")

	expectedOrder := []struct {
		sender string
		nonce  uint64
	}{
		{"bob", 1},     // Highest fee
		{"charlie", 1}, // Second highest fee
		{"alice", 1},   // Lowest fee
		{"alice", 2},   // Next nonce for alice
	}

	for i, expected := range expectedOrder {
		assert.Equal(t, expected.sender, selected[i].Sender,
			fmt.Sprintf("Expected sender %s at position %d", expected.sender, i))
		assert.Equal(t, expected.nonce, selected[i].Nonce,
			fmt.Sprintf("Expected nonce %d at position %d", expected.nonce, i))
	}
}

func TestUpdate(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions
	tx1 := Transaction{Sender: "alice", Nonce: 1, GasFee: 100}
	tx2 := Transaction{Sender: "alice", Nonce: 2, GasFee: 150}
	tx3 := Transaction{Sender: "bob", Nonce: 1, GasFee: 200}

	mp.AddTx(tx1)
	mp.AddTx(tx2)
	mp.AddTx(tx3)

	// Update with committed transactions
	mp.Update([]Transaction{tx1, tx3})

	// Check remaining transactions
	assert.Equal(t, 1, mp.Size(), "Expected 1 transaction after update")

	remaining := mp.GetAllTransactions()
	assert.Len(t, remaining, 1, "Expected 1 transaction remaining")
	assert.Equal(t, "alice", remaining[0].Sender, "Expected remaining transaction to be from Alice")
	assert.Equal(t, uint64(2), remaining[0].Nonce, "Expected remaining transaction to have nonce 2")
}

func TestCollectWithLimit(t *testing.T) {
	mp, _ := newTestMempool()

	// Add 10 transactions for Alice
	for i := 1; i <= 10; i++ {
		mp.AddTx(Transaction{
			Sender: "alice",
			Nonce:  uint64(i),
			GasFee: 100,
		})
	}

	// Collect with limit of 5
	selected := mp.CollectTxsForBlock(5)
	assert.Len(t, selected, 5, "Expected 5 transactions to be selected")

	// Check remaining count
	assert.Equal(t, 5, mp.Size(), "Expected 5 transactions remaining in mempool")

	// Collect remaining with higher limit
	selected = mp.CollectTxsForBlock(10)
	assert.Len(t, selected, 5, "Expected 5 more transactions to be selected")

	// Mempool should be empty now
	assert.Equal(t, 0, mp.Size(), "Expected mempool to be empty")
}

func TestConcurrentOperations(t *testing.T) {
	mp, _ := newTestMempool()
	var wg sync.WaitGroup

	// Concurrently add transactions from multiple senders
	for i := 0; i < 5; i++ {
		sender := string(rune('a' + i))
		wg.Add(1)
		go func(sender string) {
			defer wg.Done()
			for j := 1; j <= 10; j++ {
				mp.AddTx(Transaction{
					Sender: sender,
					Nonce:  uint64(j),
					GasFee: uint64(100 + j),
				})
			}
		}(sender)
	}

	wg.Wait()

	// Verify all transactions were added
	assert.Equal(t, 50, mp.Size(), "Expected 50 transactions after concurrent addition")

	// Collect all transactions
	allTxs := mp.CollectTxsForBlock(50)
	assert.Len(t, allTxs, 50, "Expected to collect all 50 transactions")

	// Mempool should be empty
	assert.Equal(t, 0, mp.Size(), "Expected mempool to be empty after collection")
}

func TestAccountSequenceRetrieval(t *testing.T) {
	mockClient := new(MockMempoolClient)
	mp := NewMempool(mockClient)

	address := "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"
	req := abci.RequestQuery{Path: "auth/accounts/" + address}

	// Setup mock response
	mockResponse := `{
		"BaseAccount": {
			"address": "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2",
			"sequence": "8"
		}
	}`

	mockClient.On("QuerySync", req).Return(abci.ResponseQuery{
		Value: []byte(mockResponse),
	}, nil)

	// Test adding a transaction that requires sequence lookup
	err := mp.AddTx(Transaction{Sender: address, Nonce: 8, GasFee: 100})
	assert.NoError(t, err, "Expected no error when adding transaction with correct nonce")

	// Verify the mock was called
	mockClient.AssertExpectations(t)

	// Try adding a transaction with too low nonce
	err = mp.AddTx(Transaction{Sender: address, Nonce: 7, GasFee: 100})
	assert.Error(t, err, "Expected error when adding transaction with nonce too low")
	assert.Contains(t, err.Error(), "tx nonce too low", "Error should mention nonce too low")
}

func TestNonceOrdering(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions with non-sequential nonces
	mp.AddTx(Transaction{Sender: "alice", Nonce: 5, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 3, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 1, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 4, GasFee: 100})
	mp.AddTx(Transaction{Sender: "alice", Nonce: 2, GasFee: 100})

	// Check they're stored in order
	txs := mp.GetTransactionsBySender("alice")
	assert.Len(t, txs, 5, "Expected 5 transactions for Alice")

	for i := 0; i < len(txs); i++ {
		assert.Equal(t, uint64(i+1), txs[i].Nonce,
			fmt.Sprintf("Expected transaction at index %d to have nonce %d", i, i+1))
	}

	// Collect transactions - should come out in nonce order
	selected := mp.CollectTxsForBlock(5)
	assert.Len(t, selected, 5, "Expected to collect all 5 transactions")

	for i := 0; i < len(selected); i++ {
		assert.Equal(t, uint64(i+1), selected[i].Nonce,
			fmt.Sprintf("Expected collected transaction at index %d to have nonce %d", i, i+1))
	}
}

func TestEmptyMempool(t *testing.T) {
	mp, _ := newTestMempool()

	// Try to collect from empty mempool
	selected := mp.CollectTxsForBlock(10)
	assert.Empty(t, selected, "Expected no transactions from empty mempool")

	// Try to get transactions for non-existent sender
	txs := mp.GetTransactionsBySender("nobody")
	assert.Empty(t, txs, "Expected no transactions for non-existent sender")
}

func TestFeePrioritization(t *testing.T) {
	mp, _ := newTestMempool()

	// Add transactions with same nonce but different fees
	senders := []string{"alice", "bob", "charlie", "dave", "eve"}
	fees := []uint64{300, 100, 500, 200, 400}

	for i, sender := range senders {
		mp.AddTx(Transaction{Sender: sender, Nonce: 1, GasFee: fees[i]})
	}

	// Collect all transactions
	selected := mp.CollectTxsForBlock(5)
	assert.Len(t, selected, 5, "Expected to collect all 5 transactions")

	// Should be ordered by decreasing fee
	expectedOrder := []string{"charlie", "eve", "alice", "dave", "bob"}
	for i, sender := range expectedOrder {
		assert.Equal(t, sender, selected[i].Sender,
			fmt.Sprintf("Expected sender %s at position %d", sender, i))
	}
}

func TestGetAccountSequence(t *testing.T) {
	mockClient := new(MockMempoolClient)
	mp := NewMempool(mockClient)

	address := "g1e6gxg5tvc55mwsn7t7dymmlasratv7mkv0rap2"
	req := abci.RequestQuery{Path: "auth/accounts/" + address}

	// Setup mock response with invalid JSON
	mockClient.On("QuerySync", req).Return(abci.ResponseQuery{
		Value: []byte(`invalid json`),
	}, nil)

	// This should fail to parse the response
	_, err := mp.getAccountSequence(address)
	assert.Error(t, err, "Expected error when parsing invalid JSON")
	assert.Contains(t, err.Error(), "failed to parse account data", "Error should mention parsing failure")
}

func TestAddTxOnlyFetchesNonceOnce(t *testing.T) {
	mockClient := new(MockMempoolClient)
	mp := NewMempool(mockClient)

	address := "g1abc"
	req := abci.RequestQuery{Path: "auth/accounts/" + address}

	mockClient.On("QuerySync", req).Return(abci.ResponseQuery{
		Value: []byte(`{"BaseAccount":{"sequence":"2"}}`),
	}, nil).Once()

	err1 := mp.AddTx(Transaction{Sender: address, Nonce: 3, GasFee: 100})
	err2 := mp.AddTx(Transaction{Sender: address, Nonce: 4, GasFee: 100})
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	mockClient.AssertNumberOfCalls(t, "QuerySync", 1)
}

func TestRejectTxWithTooLowNonce(t *testing.T) {
	mockClient := new(MockMempoolClient)
	mp := NewMempool(mockClient)

	address := "g1reject"
	req := abci.RequestQuery{Path: "auth/accounts/" + address}

	mockClient.On("QuerySync", req).Return(abci.ResponseQuery{
		Value: []byte(`{"BaseAccount":{"sequence":"5"}}`),
	}, nil)

	err := mp.AddTx(Transaction{Sender: address, Nonce: 3, GasFee: 50})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tx nonce too low")
}

func TestUpdateIncrementsExpectedNonce(t *testing.T) {
	mockClient := new(MockMempoolClient)
	mp := NewMempool(mockClient)

	address := "g1update"
	req := abci.RequestQuery{Path: "auth/accounts/" + address}

	mockClient.On("QuerySync", req).Return(abci.ResponseQuery{
		Value: []byte(`{"BaseAccount":{"sequence":"10"}}`),
	}, nil)

	// Dodaj jednu transakciju
	_ = mp.AddTx(Transaction{Sender: address, Nonce: 10, GasFee: 100})
	_ = mp.AddTx(Transaction{Sender: address, Nonce: 11, GasFee: 100})

	// Update sa prvim tx
	mp.Update([]Transaction{{Sender: address, Nonce: 10, GasFee: 100}})

	assert.Equal(t, uint64(11), mp.expectedNonces[address], "Expected nonce should now be 11")

	// Drugi update
	mp.Update([]Transaction{{Sender: address, Nonce: 11, GasFee: 100}})
	assert.Equal(t, uint64(12), mp.expectedNonces[address], "Expected nonce should now be 12")
}
