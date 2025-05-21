package my_mempool

import (
	"fmt"
	"testing"
	"time"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAppConn struct {
	mock.Mock
}

func (m *MockAppConn) CheckTxAsync(req abci.RequestCheckTx) *abcicli.ReqRes {
	args := m.Called(req)
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockAppConn) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAppConn) SetResponseCallback(cb abcicli.Callback) {
	m.Called(cb)
}

func (m *MockAppConn) FlushAsync() *abcicli.ReqRes {
	args := m.Called()
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockAppConn) FlushSync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAppConn) QuerySync(req abci.RequestQuery) (abci.ResponseQuery, error) {
	args := m.Called(req)
	return args.Get(0).(abci.ResponseQuery), args.Error(1)
}

// Helper to create test transactions with predictable hashes
func makeTx(data string) types.Tx {
	return types.Tx([]byte(data))
}

func TestNewMempool(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	mp := NewMempool(mockConn)

	assert.NotNil(t, mp, "NewMempool returned nil")
	assert.Equal(t, 0, mp.Size(), "Expected empty mempool")
	assert.Equal(t, int64(0), mp.TxsBytes(), "Expected 0 bytes")
}

func TestAddTx(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{
		GasWanted: 10,
	})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)
	tx := makeTx("test_tx")
	err := mp.AddTx(tx)

	assert.NoError(t, err, "AddTx should not fail")
	assert.Equal(t, 1, mp.Size(), "Expected size 1")
	assert.Equal(t, int64(len(tx)), mp.TxsBytes(), "Expected bytes to match tx length")
	mockConn.AssertCalled(t, "CheckTxAsync", mock.Anything)
}

func TestAddDuplicateTx(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{
		GasWanted: 10,
	})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)
	tx := makeTx("duplicate_tx")

	err := mp.AddTx(tx)
	assert.NoError(t, err, "First AddTx should succeed")

	err = mp.AddTx(tx)
	assert.Error(t, err, "Second AddTx should fail")
	assert.Contains(t, err.Error(), "already exists", "Error should mention duplicate")

	assert.Equal(t, 1, mp.Size(), "Expected size to remain 1")
}

func TestAddInvalidTx(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response with error
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	errResp := abci.ResponseCheckTx{}
	errResp.ResponseBase.Error = abci.StringError("unauthorized")
	mockReqRes.SetResponse(errResp)
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)
	tx := makeTx("invalid_tx")

	err := mp.AddTx(tx)
	assert.Error(t, err, "AddTx should fail for invalid transaction")
	assert.Contains(t, err.Error(), "rejected", "Error should mention rejection")

	assert.Equal(t, 0, mp.Size(), "Mempool size should be 0")
}

func TestRemoveTxViaUpdate(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Set up expectations for CheckTxAsync calls
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx1 := makeTx("tx1")
	tx2 := makeTx("tx2")

	mp.AddTx(tx1)
	mp.AddTx(tx2)

	if mp.Size() != 2 {
		t.Fatalf("Expected 2 txs, got %d", mp.Size())
	}

	// Instead of calling RemoveTx, we now use Update
	mp.Update([]types.Tx{tx1})

	if mp.Size() != 1 {
		t.Errorf("Expected 1 tx after update, got %d", mp.Size())
	}

	foundTx, exists := mp.GetTx(tx1.Hash())
	if exists {
		t.Errorf("Tx1 should be removed, but was found: %v", foundTx)
	}

	_, exists = mp.GetTx(tx2.Hash())
	if !exists {
		t.Errorf("Tx2 should exist but was not found")
	}
}

func TestUpdateWithNonexistentTx(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Set up expectations for CheckTxAsync calls
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx := makeTx("existing_tx")
	mp.AddTx(tx)
	initialSize := mp.Size()

	nonexistentTx := makeTx("nonexistent_tx")
	mp.Update([]types.Tx{nonexistentTx})

	if mp.Size() != initialSize {
		t.Errorf("Expected size to remain %d, got %d", initialSize, mp.Size())
	}
}

func TestUpdate(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx1 := makeTx("tx1")
	tx2 := makeTx("tx2")
	tx3 := makeTx("tx3")

	mp.AddTx(tx1)
	mp.AddTx(tx2)
	mp.AddTx(tx3)

	if mp.Size() != 3 {
		t.Fatalf("Expected 3 txs, got %d", mp.Size())
	}

	committed := []types.Tx{tx1, tx3}
	mp.Update(committed)

	if mp.Size() != 1 {
		t.Errorf("Expected 1 tx after update, got %d", mp.Size())
	}

	foundTx, exists := mp.GetTx(tx2.Hash())
	if !exists {
		t.Errorf("Tx2 should still exist but was not found")
	}

	if string(foundTx) != string(tx2) {
		t.Errorf("Expected tx %s, got %s", tx2, foundTx)
	}
}

func TestPending(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{
		GasWanted: 20,
	})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx1 := makeTx("tx1")              // 3 bytes
	tx2 := makeTx("tx2")              // 3 bytes
	tx3 := makeTx("long_transaction") // 16 bytes

	mp.AddTx(tx1)
	mp.AddTx(tx2)
	mp.AddTx(tx3)

	// Test bytes limit
	txs := mp.Pending(6, -1)
	if len(txs) != 2 {
		t.Errorf("Expected 2 txs with 6 byte limit, got %d", len(txs))
	}

	mp.Flush()
	mp.AddTx(tx1) // 20 gas
	mp.AddTx(tx2) // 20 gas

	txs = mp.Pending(100, 30)
	if len(txs) != 1 {
		t.Errorf("Expected 1 tx with 30 gas limit, got %d", len(txs))
	}

	// Test zero byte limit
	txs = mp.Pending(0, 100)
	if txs != nil {
		t.Errorf("Expected nil result with 0 byte limit, got %v", txs)
	}
}

func TestFIFOOrdering(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx1 := makeTx("first")
	tx2 := makeTx("second")
	tx3 := makeTx("third")

	mp.AddTx(tx1)
	mp.AddTx(tx2)
	mp.AddTx(tx3)

	txs := mp.Content()

	if len(txs) != 3 {
		t.Fatalf("Expected 3 txs, got %d", len(txs))
	}

	if string(txs[0]) != string(tx1) {
		t.Errorf("Expected first tx to be %s, got %s", tx1, txs[0])
	}

	if string(txs[1]) != string(tx2) {
		t.Errorf("Expected second tx to be %s, got %s", tx2, txs[1])
	}

	if string(txs[2]) != string(tx3) {
		t.Errorf("Expected third tx to be %s, got %s", tx3, txs[2])
	}
}

func TestContent(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	txs := mp.Content()
	if len(txs) != 0 {
		t.Errorf("Expected empty content for empty mempool, got %d txs", len(txs))
	}

	tx1 := makeTx("content_test")
	mp.AddTx(tx1)

	txs = mp.Content()
	if len(txs) != 1 {
		t.Errorf("Expected 1 tx, got %d", len(txs))
	}
}

func TestFlush(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx1 := makeTx("flush_test1")
	tx2 := makeTx("flush_test2")

	mp.AddTx(tx1)
	mp.AddTx(tx2)

	if mp.Size() != 2 {
		t.Fatalf("Expected 2 txs, got %d", mp.Size())
	}

	mp.Flush()

	if mp.Size() != 0 {
		t.Errorf("Expected empty mempool after flush, got %d txs", mp.Size())
	}

	if mp.TxsBytes() != 0 {
		t.Errorf("Expected 0 bytes after flush, got %d", mp.TxsBytes())
	}
}

func TestGetTx(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx := makeTx("get_tx_test")
	mp.AddTx(tx)

	foundTx, exists := mp.GetTx(tx.Hash())

	if !exists {
		t.Fatal("Expected tx to exist, but it wasn't found")
	}

	if string(foundTx) != string(tx) {
		t.Errorf("Expected tx %s, got %s", tx, foundTx)
	}

	nonexistentTx := makeTx("nonexistent")
	_, exists = mp.GetTx(nonexistentTx.Hash())

	if exists {
		t.Error("Found tx that shouldn't exist")
	}
}

func TestSizeAndTxsBytes(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Setup mock response
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)
	mp := NewMempool(mockConn)

	if mp.Size() != 0 {
		t.Errorf("Expected size 0, got %d", mp.Size())
	}

	if mp.TxsBytes() != 0 {
		t.Errorf("Expected 0 bytes, got %d", mp.TxsBytes())
	}

	tx1 := makeTx("size_test1") // 10 bytes
	tx2 := makeTx("size_test2") // 10 bytes

	mp.AddTx(tx1)

	if mp.Size() != 1 {
		t.Errorf("Expected size 1, got %d", mp.Size())
	}

	if mp.TxsBytes() != int64(len(tx1)) {
		t.Errorf("Expected %d bytes, got %d", len(tx1), mp.TxsBytes())
	}

	mp.AddTx(tx2)
	expectedBytes := int64(len(tx1) + len(tx2))

	if mp.Size() != 2 {
		t.Errorf("Expected size 2, got %d", mp.Size())
	}

	if mp.TxsBytes() != expectedBytes {
		t.Errorf("Expected %d bytes, got %d", expectedBytes, mp.TxsBytes())
	}
}

func TestConcurrentAccess(t *testing.T) {
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	// Since there will be many calls, set unlimited number of expectations
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes).Maybe()

	mp := NewMempool(mockConn)

	done := make(chan bool)

	// Add transactions concurrently
	go func() {
		for i := 0; i < 50; i++ {
			tx := makeTx(fmt.Sprintf("concurrent_tx_%d", i))
			mp.AddTx(tx)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Read mempool state concurrently
	go func() {
		for i := 0; i < 100; i++ {
			_ = mp.Size()
			_ = mp.TxsBytes()
			_ = mp.Content()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both goroutines to finish
	<-done
	<-done

	// Verify mempool is in consistent state
	var totalBytes int64
	for _, tx := range mp.Content() {
		totalBytes += int64(len(tx))
	}

	if mp.Size() != len(mp.Content()) {
		t.Errorf("Size mismatch: Size()=%d, Content()=%d", mp.Size(), len(mp.Content()))
	}

	if mp.TxsBytes() != totalBytes {
		t.Errorf("Bytes mismatch: TxsBytes()=%d, actual=%d", mp.TxsBytes(), totalBytes)
	}
}

func TestFullMempoolLifecycleWithTwoMempools(t *testing.T) {
	// Create two mempools with identical state
	mockConn1 := new(MockAppConn)
	mockConn2 := new(MockAppConn)

	// Set expectations for both mocks
	mockConn1.On("Error").Return(nil)
	mockConn2.On("Error").Return(nil)

	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn1.On("CheckTxAsync", mock.Anything).Return(mockReqRes)
	mockConn2.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp1 := NewMempool(mockConn1)
	mp2 := NewMempool(mockConn2)

	// Add the same transactions to both mempools
	txs := []types.Tx{
		makeTx("tx1"), // 3 bytes
		makeTx("tx2"), // 3 bytes
		makeTx("tx3"), // 3 bytes
		makeTx("tx4"), // 3 bytes
		makeTx("tx5"), // 3 bytes
	}

	for _, tx := range txs {
		err1 := mp1.AddTx(tx)
		err2 := mp2.AddTx(tx)
		assert.NoError(t, err1)
		assert.NoError(t, err2)
	}

	// Verify that both mempools are in identical state
	assert.Equal(t, mp1.Size(), mp2.Size())
	assert.Equal(t, mp1.TxsBytes(), mp2.TxsBytes())

	// Get transactions for a block from the first mempool (only first 3)
	bytesLimit := int64(9) // Limit of 9 bytes (3 transactions of 3 bytes each)
	selectedTxs := mp1.Pending(bytesLimit, -1)

	// Verify we received exactly 3 transactions
	assert.Equal(t, 3, len(selectedTxs))

	// Update the first mempool with the selected transactions (simulate commit)
	mp1.Update(selectedTxs)

	// Expect first mempool to have 2 fewer transactions
	assert.Equal(t, 2, mp1.Size())

	// Second mempool still has all transactions
	assert.Equal(t, 5, mp2.Size())

	// Now update the second mempool with the same transactions
	mp2.Update(selectedTxs)

	// Expect both mempools to have the same number of transactions
	assert.Equal(t, mp1.Size(), mp2.Size())
	assert.Equal(t, mp1.TxsBytes(), mp2.TxsBytes())

	// Verify that the remaining transactions are identical in both mempools
	mp1Content := mp1.Content()
	mp2Content := mp2.Content()

	assert.Equal(t, len(mp1Content), len(mp2Content))

	// Additional check - the same transactions remain in both mempools
	for i := 0; i < len(mp1Content); i++ {
		assert.Equal(t, string(mp1Content[i]), string(mp2Content[i]))
	}

	// Verify that FIFO order is preserved after Update operation
	// Remaining transactions should be tx4 and tx5 in that order
	if len(mp1Content) == 2 {
		assert.Equal(t, "tx4", string(mp1Content[0]))
		assert.Equal(t, "tx5", string(mp1Content[1]))
	}
}

func TestAddAfterFlush(t *testing.T) {
	mockConn := new(MockAppConn)
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)
	tx := makeTx("flush_me")

	err := mp.AddTx(tx)
	assert.NoError(t, err)
	assert.Equal(t, 1, mp.Size())

	mp.Flush()
	assert.Equal(t, 0, mp.Size())

	err = mp.AddTx(tx)
	assert.NoError(t, err)
	assert.Equal(t, 1, mp.Size())
}

func TestUpdatePreservesOrder(t *testing.T) {
	// Create and configure mock
	mockConn := new(MockAppConn)
	mockConn.On("Error").Return(nil)

	// Set up mock response with proper structure
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{
		GasWanted: 10,
	})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes).Times(4)

	// Create mempool and transactions
	mp := NewMempool(mockConn)
	tx1 := makeTx("a")
	tx2 := makeTx("b")
	tx3 := makeTx("c")
	tx4 := makeTx("d")

	// Add transactions in order
	err1 := mp.AddTx(tx1)
	err2 := mp.AddTx(tx2)
	err3 := mp.AddTx(tx3)
	err4 := mp.AddTx(tx4)

	// Verify all transactions were added successfully
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)
	assert.Equal(t, 4, mp.Size(), "Should have 4 transactions before update")

	// Remove two transactions (b and d)
	mp.Update([]types.Tx{tx2, tx4})

	// Verify result
	txs := mp.Content()
	assert.Equal(t, 2, len(txs), "Should have 2 transactions after update")

	// Check that order is preserved (a, c)
	if len(txs) == 2 {
		assert.Equal(t, string(tx1), string(txs[0]), "First tx should be 'a'")
		assert.Equal(t, string(tx3), string(txs[1]), "Second tx should be 'c'")
	}

	// Add logging to verify test execution
	t.Logf("Test completed successfully: mempool contains %d transactions after update", len(txs))
}

func TestGetTxNotFound(t *testing.T) {
	mockConn := new(MockAppConn)
	mockReqRes := abcicli.NewReqRes(abci.RequestCheckTx{})
	mockReqRes.SetResponse(abci.ResponseCheckTx{GasWanted: 10})
	mockConn.On("CheckTxAsync", mock.Anything).Return(mockReqRes)

	mp := NewMempool(mockConn)

	tx := makeTx("some_tx")
	_, exists := mp.GetTx(tx.Hash())

	assert.False(t, exists, "Expected transaction not to be found")
}
