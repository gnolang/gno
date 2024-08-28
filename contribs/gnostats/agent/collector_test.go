package agent

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gnostats/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRPCClient is a mock of the rpcClient used for testing
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) NewBatch() rpcBatch {
	args := m.Called()
	return args.Get(0).(rpcBatch)
}

// MockRPCBatch is a mock of the rpcBatch used for testing
type MockRPCBatch struct {
	mock.Mock

	sendLatency time.Duration // Duration used to simulate a latency when calling Send()
}

func (m *MockRPCBatch) Status() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRPCBatch) Validators() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRPCBatch) NetInfo() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRPCBatch) NumUnconfirmedTxs() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRPCBatch) Block(height *uint64) error {
	args := m.Called(height)
	return args.Error(0)
}

func (m *MockRPCBatch) BlockResults(height *uint64) error {
	args := m.Called(height)
	return args.Error(0)
}

// Send mock method supporting latency simulation to test context timeout
func (m *MockRPCBatch) Send(ctx context.Context) ([]any, error) {
	args := m.Called(ctx)

	select {
	case <-ctx.Done():
		return args.Get(0).([]any), ctx.Err()
	case <-time.After(m.sendLatency):
		return args.Get(0).([]any), args.Error(1)
	}
}

// mockNetAddr is a mock used for p2p.NewNetAddress()
type mockNetAddr struct {
	network string
	str     string
}

func (m mockNetAddr) Network() string { return m.network }
func (m mockNetAddr) String() string  { return m.str }

// Helper that generates a valid RPC batch result
func getBatchResults(t *testing.T) []any {
	t.Helper()

	// Generate peers for NetInfo request
	peers := make([]p2p.NodeInfo, 3)
	for i, info := range []struct{ moniker, address string }{
		{"peer1", "1.1.1.1"},
		{"peer2", "2.2.2.2"},
		{"peer3", "3.3.3.3"},
	} {
		peers[i].Moniker = info.moniker
		peers[i].NetAddress = p2p.NewNetAddress(
			crypto.ID(info.moniker),
			mockNetAddr{
				network: "tcp",
				str:     info.address,
			},
		)
	}

	// Generate validators for Validators request
	validators := make([]*types.Validator, 3)
	for i, pubKey := range []crypto.PubKey{
		ed25519.GenPrivKeyFromSecret([]byte("validator1")).PubKey(),
		ed25519.GenPrivKeyFromSecret([]byte("validator2")).PubKey(),
		ed25519.GenPrivKeyFromSecret([]byte("validator3")).PubKey(),
	} {
		validators[i] = types.NewValidator(pubKey, 42)
	}

	return []any{
		&ctypes.ResultStatus{
			NodeInfo: p2p.NodeInfo{
				Moniker: "self",
				NetAddress: p2p.NewNetAddress(
					crypto.ID("self"),
					mockNetAddr{
						network: "tcp",
						str:     "0.0.0.0",
					},
				),
			},
			ValidatorInfo: ctypes.ValidatorInfo{
				Address:     validators[2].Address,
				PubKey:      validators[2].PubKey,
				VotingPower: validators[2].VotingPower,
			},
		},

		&ctypes.ResultValidators{
			Validators: validators,
		},

		&ctypes.ResultNetInfo{
			Peers: []ctypes.Peer{
				{NodeInfo: peers[0]},
				{NodeInfo: peers[1]},
				{NodeInfo: peers[2]},
			},
		},

		&ctypes.ResultUnconfirmedTxs{Total: 5},

		&ctypes.ResultBlock{
			Block: &types.Block{
				Header: types.Header{
					Height:          10,
					Time:            time.Now(),
					ProposerAddress: crypto.Address{},
				},
			},
		},

		&ctypes.ResultBlockResults{
			Results: &state.ABCIResponses{
				DeliverTxs: []abci.ResponseDeliverTx{
					{GasUsed: 100, GasWanted: 200},
					{GasUsed: 42, GasWanted: 24},
				},
			},
		},
	}
}

func compareBatchResultToDynamicInfo(t *testing.T, results []any, dynamicInfo *proto.DynamicInfo) {
	t.Helper()

	var (
		status     = results[0].(*ctypes.ResultStatus)
		validators = results[1].(*ctypes.ResultValidators)
		netInfo    = results[2].(*ctypes.ResultNetInfo)
		uncTxs     = results[3].(*ctypes.ResultUnconfirmedTxs)
		blk        = results[4].(*ctypes.ResultBlock)
		blkRes     = results[5].(*ctypes.ResultBlockResults)
	)

	assert.Equal(t, dynamicInfo.Address, status.NodeInfo.ID().String())
	assert.Equal(t, dynamicInfo.Moniker, status.NodeInfo.Moniker)

	isValidator := false
	for _, validator := range validators.Validators {
		if validator.Address.Compare(status.ValidatorInfo.Address) == 0 {
			isValidator = true
		}
	}
	assert.Equal(t, dynamicInfo.IsValidator, isValidator)

	assert.NotNil(t, dynamicInfo.NetInfo)
	assert.Equal(t, dynamicInfo.NetInfo.P2PAddress, status.NodeInfo.NetAddress.String())
	assert.Equal(t, len(dynamicInfo.NetInfo.Peers), len(netInfo.Peers))
	for i, peer := range dynamicInfo.NetInfo.Peers {
		assert.Equal(t, peer.Moniker, netInfo.Peers[i].NodeInfo.Moniker)
		assert.NotNil(t, netInfo.Peers[i].NodeInfo.NetAddress)
		assert.Equal(t, peer.P2PAddress, netInfo.Peers[i].NodeInfo.NetAddress.String())
	}

	assert.Equal(t, dynamicInfo.PendingTxs, uint64(uncTxs.Total))

	assert.NotNil(t, dynamicInfo.BlockInfo)
	assert.Equal(t, dynamicInfo.BlockInfo.Number, uint64(blk.Block.Height))
	assert.Equal(t, dynamicInfo.BlockInfo.Timestamp, uint64(blk.Block.Time.Unix()))

	var gasUsed, gasWanted uint64
	for _, deliverTx := range blkRes.Results.DeliverTxs {
		gasUsed += uint64(deliverTx.GasUsed)
		gasWanted += uint64(deliverTx.GasWanted)
	}
	assert.Equal(t, dynamicInfo.BlockInfo.GasUsed, gasUsed)
	assert.Equal(t, dynamicInfo.BlockInfo.GasWanted, gasWanted)
	assert.Equal(t, dynamicInfo.BlockInfo.Proposer, blk.Block.ProposerAddress.ID().String())
}

func TestCollector_DynamicSuccess(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(nil)
	mockBatch.On("Validators").Return(nil)
	mockBatch.On("NetInfo").Return(nil)
	mockBatch.On("NumUnconfirmedTxs").Return(nil)
	mockBatch.On("Block", (*uint64)(nil)).Return(nil)
	mockBatch.On("BlockResults", (*uint64)(nil)).Return(nil)

	// Get predefined RPC batch results
	ctx := context.Background()
	results := getBatchResults(t)
	mockBatch.On("Send", ctx).Return(results, nil)

	// Call the actual method to test (CollectDynamic)
	c := &collector{caller: mockCaller}
	dynamicInfo, err := c.CollectDynamic(ctx)

	// Assert that all expectations were met
	assert.NoError(t, err)
	assert.NotNil(t, dynamicInfo)
	compareBatchResultToDynamicInfo(t, results, dynamicInfo)
}

func TestCollector_DynamicFail(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(fmt.Errorf("status error"))

	// Call the actual method to test (CollectDynamic)
	c := &collector{caller: mockCaller}
	dynamicInfo, err := c.CollectDynamic(context.Background())

	// Assert that all expectations were met
	assert.Error(t, err)
	assert.Nil(t, dynamicInfo)
}

func TestCollector_DynamicTimeout(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(nil)
	mockBatch.On("Validators").Return(nil)
	mockBatch.On("NetInfo").Return(nil)
	mockBatch.On("NumUnconfirmedTxs").Return(nil)
	mockBatch.On("Block", (*uint64)(nil)).Return(nil)
	mockBatch.On("BlockResults", (*uint64)(nil)).Return(nil)

	// Set up a context and a sendLatency that will trigger a timeout
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
	mockBatch.sendLatency = time.Second
	mockBatch.On("Send", ctx).Return([]any{})

	// Call the actual method to test (CollectDynamic)
	c := &collector{caller: mockCaller}
	dynamicInfo, err := c.CollectDynamic(ctx)

	// Assert that the context timed out
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, dynamicInfo)
}

func compareStatusRespToStaticInfo(t *testing.T, status *ctypes.ResultStatus, staticInfo *proto.StaticInfo) {
	t.Helper()

	assert.Equal(t, staticInfo.Address, status.NodeInfo.ID().String())
	assert.Equal(t, staticInfo.GnoVersion, status.NodeInfo.Version)
	assert.Equal(t, staticInfo.OsVersion, fmt.Sprintf("%s - %s", runtime.GOOS, runtime.GOARCH))
}

func TestCollector_StaticSuccess(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(nil)

	// Get predefined RPC batch results
	ctx := context.Background()
	results := getBatchResults(t)
	mockBatch.On("Send", ctx).Return(results, nil)

	// Call the actual method to test (CollectStatic)
	c := &collector{caller: mockCaller}
	staticInfo, err := c.CollectStatic(ctx)

	// Assert that all expectations were met
	assert.NoError(t, err)
	assert.NotNil(t, staticInfo)
	status := results[0].(*ctypes.ResultStatus)
	compareStatusRespToStaticInfo(t, status, staticInfo)
}

func TestCollector_StaticFail(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	ctx := context.Background()
	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(fmt.Errorf("status error"))

	// Call the actual method to test (CollectStatic)
	c := &collector{caller: mockCaller}
	staticInfo, err := c.CollectStatic(ctx)

	// Assert that all expectations were met
	assert.Error(t, err)
	assert.Nil(t, staticInfo)
}

func TestCollector_StaticTimeout(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(nil)

	// Set up a context and a sendLatency that will trigger a timeout
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond)
	mockBatch.sendLatency = time.Second
	mockBatch.On("Send", ctx).Return([]any{})

	// Call the actual method to test (CollectStatic)
	c := &collector{caller: mockCaller}
	staticInfo, err := c.CollectStatic(ctx)

	// Assert that the context timed out
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, staticInfo)
}
