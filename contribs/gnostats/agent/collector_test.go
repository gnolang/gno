package agent

import (
	"context"
	"fmt"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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
				Other: p2p.NodeInfoOther{
					OS:       "plan9",
					Arch:     "ppc64",
					Location: "",
				},
			},
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
				DeliverTxs: []abci.ResponseDeliverTx{{
					GasUsed:   100,
					GasWanted: 200,
				}},
			},
		},
	}
}

func compareBatchResultToDynamicInfo(t *testing.T, results []any, dynamicInfo *proto.DynamicInfo) {
	t.Helper()

	var (
		status  = results[0].(*ctypes.ResultStatus)
		netInfo = results[1].(*ctypes.ResultNetInfo)
		uncTxs  = results[2].(*ctypes.ResultUnconfirmedTxs)
		blk     = results[3].(*ctypes.ResultBlock)
		blkRes  = results[4].(*ctypes.ResultBlockResults)
	)

	assert.Equal(t, dynamicInfo.Address, status.NodeInfo.ID().String())
	assert.Equal(t, dynamicInfo.Moniker, status.NodeInfo.Moniker)
	assert.Equal(t, dynamicInfo.IsValidator, status.ValidatorInfo.Address.ID().String() != "")

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
	assert.Equal(t, dynamicInfo.BlockInfo.GasUsed, uint64(blkRes.Results.DeliverTxs[0].GasUsed))
	assert.Equal(t, dynamicInfo.BlockInfo.GasWanted, uint64(blkRes.Results.DeliverTxs[0].GasWanted))
	assert.Equal(t, dynamicInfo.BlockInfo.Proposer, blk.Block.ProposerAddress.ID().String())
}

func TestCollector_DynamicSuccess(t *testing.T) {
	t.Parallel()

	// Setup RPC mocks
	mockCaller := new(MockRPCClient)
	mockBatch := new(MockRPCBatch)

	mockCaller.On("NewBatch").Return(mockBatch)
	mockBatch.On("Status").Return(nil)
	mockBatch.On("NetInfo").Return(nil)
	mockBatch.On("NumUnconfirmedTxs").Return(nil)
	mockBatch.On("Block", (*uint64)(nil)).Return(nil)
	mockBatch.On("BlockResults", (*uint64)(nil)).Return(nil)

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

func compareStatusRespToStaticInfo(t *testing.T, status *ctypes.ResultStatus, osExpected string, staticInfo *proto.StaticInfo) {
	t.Helper()

	assert.Equal(t, staticInfo.Address, status.NodeInfo.ID().String())
	assert.Equal(t, staticInfo.GnoVersion, status.NodeInfo.Version)
	assert.Equal(t, staticInfo.OsVersion, osExpected)
	assert.Equal(t, staticInfo.Location, status.NodeInfo.Other.Location)
}

func TestCollector_StaticSuccess(t *testing.T) {
	t.Parallel()

	// Get predefined OS and Arch values
	var (
		status = getBatchResults(t)[0].(*ctypes.ResultStatus)
		os     = status.NodeInfo.Other.OS
		arch   = status.NodeInfo.Other.Arch
	)

	// Setup multiple test cases (variation of the OsVersion)
	testCases := []struct {
		name     string
		os       string
		arch     string
		expected string
	}{
		{"Both OS and Arch", os, arch, fmt.Sprintf("%s - %s", os, arch)},
		{"Only OS", os, "", os},
		{"Only Arch", "", arch, ""},
		{"Neither OS or Arch", "", "", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup RPC mocks
			mockCaller := new(MockRPCClient)
			mockBatch := new(MockRPCBatch)

			mockCaller.On("NewBatch").Return(mockBatch)
			mockBatch.On("Status").Return(nil)

			// Get predefined RPC batch results
			results := getBatchResults(t)
			status := results[0].(*ctypes.ResultStatus)

			// Override OS and Arch in the Status RPC results
			status.NodeInfo.Other.OS = testCase.os
			status.NodeInfo.Other.Arch = testCase.arch
			ctx := context.Background()
			mockBatch.On("Send", ctx).Return(results, nil)

			// Call the actual method to test (CollectStatic)
			c := &collector{caller: mockCaller}
			staticInfo, err := c.CollectStatic(ctx)

			// Assert that all expectations were met
			assert.NoError(t, err)
			assert.NotNil(t, staticInfo)
			compareStatusRespToStaticInfo(t, status, testCase.expected, staticInfo)
		})
	}
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
