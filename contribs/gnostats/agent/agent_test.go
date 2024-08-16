package agent

import (
	"context"
	"fmt"
	mrand "math/rand"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	"github.com/gnolang/gnostats/proto"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

var random = mrand.New(mrand.NewSource(time.Now().UnixNano()))

// mockHubClient is a mock of the hubClient used for testing
type mockHubClient struct {
	mock.Mock
	static chan *proto.StaticInfo
}

// Register pushes StaticInfo onto a channel accessible by the tests
func (m *mockHubClient) Register(ctx context.Context, in *proto.StaticInfo, _ ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in)
	m.static <- in
	return nil, args.Error(0)
}

// PushData returns a mockPushDataClient
func (m *mockHubClient) PushData(ctx context.Context, _ ...grpc.CallOption) (proto.Hub_PushDataClient, error) {
	args := m.Called(ctx)
	return args.Get(0).(proto.Hub_PushDataClient), args.Error(1)
}

// GetDataStream is not used in this test file
func (m *mockHubClient) GetDataStream(_ context.Context, _ *emptypb.Empty, _ ...grpc.CallOption) (proto.Hub_GetDataStreamClient, error) {
	panic("should never happen")
}

// mockPushDataClient is a mock of the PushDataClient used for testing
type mockPushDataClient struct {
	mock.Mock
	dynamic chan *proto.DynamicInfo
}

// Send pushes DynamicInfo onto a channel accessible by the tests
func (m *mockPushDataClient) Send(out *proto.DynamicInfo) error {
	args := m.Called(out)
	m.dynamic <- out
	return args.Error(0)
}

// The following methods won't be used in this test file
func (m *mockPushDataClient) CloseAndRecv() (*emptypb.Empty, error) { panic("should never happen") }
func (m *mockPushDataClient) Header() (metadata.MD, error)          { panic("should never happen") }
func (m *mockPushDataClient) Trailer() metadata.MD                  { panic("should never happen") }
func (m *mockPushDataClient) CloseSend() error                      { panic("should never happen") }
func (m *mockPushDataClient) Context() context.Context              { panic("should never happen") }
func (m *mockPushDataClient) SendMsg(msg any) error                 { panic("should never happen") }
func (m *mockPushDataClient) RecvMsg(msg any) error                 { panic("should never happen") }

// Helpers that generate random string and int
func randomIntInRange(t *testing.T, min, max int) int {
	t.Helper()

	return random.Intn(max-min+1) + min
}

func randomStringOfLength(t *testing.T, length int) string {
	t.Helper()

	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

	randBytes := make([]byte, length)
	for i := range randBytes {
		randBytes[i] = charset[random.Intn(len(charset))]
	}

	return string(randBytes)
}

func randomStringOfLengthInRange(t *testing.T, min, max int) string {
	t.Helper()

	return randomStringOfLength(t, randomIntInRange(t, min, max))
}

func randomNodeInfo(t *testing.T) p2p.NodeInfo {
	t.Helper()

	goos := []string{"aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "windows"}
	goarch := []string{"386", "amd64", "arm", "arm64", "mips", "mips64", "mips64le", "mipsle", "ppc64", "ppc64le", "riscv64", "s390x", "wasm"}

	return p2p.NodeInfo{
		Moniker: randomStringOfLengthInRange(t, 1, 128),
		NetAddress: p2p.NewNetAddress(
			crypto.ID(randomStringOfLengthInRange(t, 64, 128)),
			mockNetAddr{
				network: randomStringOfLengthInRange(t, 3, 6),
				str: fmt.Sprintf(
					"%d.%d.%d.%d",
					randomIntInRange(t, 1, 255),
					randomIntInRange(t, 0, 255),
					randomIntInRange(t, 0, 255),
					randomIntInRange(t, 0, 255),
				),
			},
		),
		Other: p2p.NodeInfoOther{
			OS:       goos[randomIntInRange(t, 0, len(goos)-1)],
			Arch:     goarch[randomIntInRange(t, 0, len(goarch)-1)],
			Location: "",
		},
	}

}

// Helper that generates a valid random RPC batch result
func getRandomBatchResults(t *testing.T) []any {
	t.Helper()

	// Generate peers for NetInfo request
	peers := make([]ctypes.Peer, randomIntInRange(t, 1, 32))
	for i := range peers {
		peers[i] = ctypes.Peer{NodeInfo: randomNodeInfo(t)}
	}

	return []any{
		&ctypes.ResultStatus{NodeInfo: randomNodeInfo(t)},
		&ctypes.ResultNetInfo{Peers: peers},
		&ctypes.ResultUnconfirmedTxs{Total: randomIntInRange(t, 0, 100)},

		&ctypes.ResultBlock{
			Block: &types.Block{
				Header: types.Header{
					Height:          int64(randomIntInRange(t, 1, 10000000)),
					Time:            time.Now(),
					ProposerAddress: crypto.Address{},
				},
			},
		},

		&ctypes.ResultBlockResults{
			Results: &state.ABCIResponses{
				DeliverTxs: []abci.ResponseDeliverTx{{
					GasUsed:   int64(randomIntInRange(t, 5, 1000)),
					GasWanted: int64(randomIntInRange(t, 5, 1000)),
				}},
			},
		},
	}
}

func TestAgent_E2E(t *testing.T) {
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

	// Setup gRPC mocks
	mockHub := new(mockHubClient)
	mockHub.static = make(chan *proto.StaticInfo)

	mockStream := new(mockPushDataClient)
	mockStream.dynamic = make(chan *proto.DynamicInfo)

	mockHub.On("Register", mock.AnythingOfType("*context.cancelCtx"), mock.AnythingOfType("*proto.StaticInfo")).Return(nil, nil)
	mockHub.On("PushData", mock.AnythingOfType("*context.cancelCtx")).Return(mockStream, nil)
	mockStream.On("Send", mock.AnythingOfType("*proto.DynamicInfo")).Return(nil)

	// Inject both mocks of the clients into a new agent
	agent := NewAgent(config{
		hClient:      mockHub,
		rClient:      mockCaller,
		pollInterval: 20 * time.Millisecond,
	})

	// Setup a first random batch result
	results := getRandomBatchResults(t)
	status := results[0].(*ctypes.ResultStatus)
	mockBatch.On("Send", mock.Anything).Return(results, nil)

	// Test if registering with the Hub works as expected
	go agent.Start(context.Background())
	static := <-mockHub.static
	osVersion := fmt.Sprintf("%s - %s", status.NodeInfo.Other.OS, status.NodeInfo.Other.Arch)
	compareStatusRespToStaticInfo(t, status, osVersion, static)

	// Test if the first five data pushes to the Hub work as expected
	for i := 0; i < 5; i++ {
		dynamic := <-mockStream.dynamic
		compareBatchResultToDynamicInfo(t, results, dynamic)

		results = getRandomBatchResults(t)
		mockBatch.On("Send").Unset() // Clear previous expected results
		mockBatch.On("Send", mock.Anything).Return(results, nil)
	}

	agent.Stop()
}
