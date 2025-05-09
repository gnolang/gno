// Test skeleton for the Mempool implementation
package my_mempool_test

import (
	"testing"

	types "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	. "github.com/gnolang/gno/tm2/pkg/bft/my_mempool"

	abcicli "github.com/gnolang/gno/tm2/pkg/bft/abci/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTx struct {
	hash     []byte
	sender   crypto.Address
	sequence uint64
	price    uint64
	size     uint64
}

func (m *MockTx) Hash() []byte           { return m.hash }
func (m *MockTx) Sender() crypto.Address { return m.sender }
func (m *MockTx) Sequence() uint64       { return m.sequence }
func (m *MockTx) Price() uint64          { return m.price }
func (m *MockTx) Size() uint64           { return m.size }

type MockAppConn struct{ mock.Mock }

func (m *MockAppConn) QuerySync(req types.RequestQuery) (types.ResponseQuery, error) {
	args := m.Called(req)
	return args.Get(0).(types.ResponseQuery), args.Error(1)
}

func (m *MockAppConn) CheckTxAsync(req types.RequestCheckTx) *abcicli.ReqRes {
	args := m.Called(req)
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockAppConn) Error() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAppConn) FlushAsync() *abcicli.ReqRes {
	args := m.Called()
	return args.Get(0).(*abcicli.ReqRes)
}

func (m *MockAppConn) FlushSync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAppConn) SetResponseCallback(cb abcicli.Callback) {
	m.Called(cb)
}

func newTestMempool(t *testing.T, initialSeq uint64) (*Mempool, *MockAppConn, crypto.Address) {
	mockApp := new(MockAppConn)
	sender := crypto.Address([]byte("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"))

	resp := types.ResponseQuery{
		Value: []byte(`{"BaseAccount":{"sequence":"` + string(rune(initialSeq+'0')) + `"}}`),
	}
	mockApp.On("QuerySync", mock.Anything).Return(resp, nil)

	mp := NewMempool(mockApp)
	return mp, mockApp, sender
}

func TestFullTransactionFlow(t *testing.T) {
	mp, _, sender := newTestMempool(t, 1)

	tx1 := &MockTx{hash: []byte("tx1"), sender: sender, sequence: 1, price: 300, size: 10}
	tx2 := &MockTx{hash: []byte("tx2"), sender: sender, sequence: 2, price: 200, size: 10}
	tx3 := &MockTx{hash: []byte("tx3"), sender: sender, sequence: 3, price: 150, size: 10}
	tx4 := &MockTx{hash: []byte("tx4"), sender: sender, sequence: 4, price: 100, size: 10}

	// Dodajemo tx2 i tx4 bez prethodnih nonce-ova → očekuje se da idu u queued
	_ = mp.AddTx(tx2)
	_ = mp.AddTx(tx4)
	assert.Len(t, mp.GetPendingBySender(sender), 0)
	assert.Len(t, mp.GetQueuedTxs(sender), 2)

	// Dodajemo tx1 → on se ubacuje u pending, i odmah se promoviše tx2
	_ = mp.AddTx(tx1)
	pending := mp.GetPendingBySender(sender)
	assert.Len(t, pending, 2)
	assert.Equal(t, tx1.Hash(), pending[0].Hash())
	assert.Equal(t, tx2.Hash(), pending[1].Hash())
	assert.Len(t, mp.GetQueuedTxs(sender), 1) // tx4 ostaje

	// Dodajemo tx3 → sada se automatski promoviše tx3 i odmah zatim tx4
	_ = mp.AddTx(tx3)
	pending = mp.GetPendingBySender(sender)
	assert.Len(t, pending, 4)
	assert.Len(t, mp.GetQueuedTxs(sender), 0)

	// Provera redosleda u pending
	expected := [][]byte{tx1.Hash(), tx2.Hash(), tx3.Hash(), tx4.Hash()}
	for i, tx := range pending {
		assert.Equal(t, expected[i], tx.Hash())
	}

	// Proveri Content
	content := mp.Content()
	assert.Len(t, content, 4)

	// Uklanjamo tx2 ručno
	mp.RemoveTx(sender, tx2.Hash())
	assert.Len(t, mp.GetTxsBySender(sender), 3)

	// Pozivamo Update sa tx1 i tx3 kao commit-ovanima → nonce se pomera, tx4 ostaje
	mp.Update([]Tx{tx1, tx3})
	expectedNonce, _ := mp.GetExpectedNonce(sender)
	assert.Equal(t, uint64(5), expectedNonce)

	remaining := mp.GetTxsBySender(sender)
	assert.Len(t, remaining, 1)
	assert.Equal(t, tx4.Hash(), remaining[0].Hash())
}

func TestFlushAndSize(t *testing.T) {
	mp, _, sender := newTestMempool(t, 1)
	tx := &MockTx{hash: []byte("flush"), sender: sender, sequence: 1, price: 100, size: 10}
	_ = mp.AddTx(tx)
	assert.Equal(t, 1, mp.Size())
	mp.Flush()
	assert.Equal(t, 0, mp.Size())
}

func TestNonceReject(t *testing.T) {
	mp, _, sender := newTestMempool(t, 5)
	tx := &MockTx{hash: []byte("old"), sender: sender, sequence: 3, price: 100, size: 10}
	err := mp.AddTx(tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nonce too low")
}
