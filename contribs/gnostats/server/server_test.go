package server

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/gnolang/gnostats/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setupGRPCServer sets up the gRPC server with the specified callbacks
func setupGRPCServer(t *testing.T, cbs ...func(server *grpc.Server)) *bufconn.Listener {
	t.Helper()

	var (
		grpcServer = grpc.NewServer()
		listener   = bufconn.Listen(1024 * 1024)
	)

	for _, cb := range cbs {
		cb(grpcServer)
	}

	go func() {
		_ = grpcServer.Serve(listener)
	}()

	t.Cleanup(func() {
		grpcServer.Stop()
	})

	return listener
}

// newMockHubConn spins up a gRPC server for the passed in Hub,
// and returns a dialed client connection
func newMockHubConn(t *testing.T, hub proto.HubServer) *grpc.ClientConn {
	t.Helper()

	listener := setupGRPCServer(t, func(server *grpc.Server) {
		proto.RegisterHubServer(server, hub)
	})

	conn, err := grpc.NewClient(
		"passthrough://bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return listener.Dial()
		}),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return conn
}

// newMockHubClient creates a new Hub client
func newMockHubClient(t *testing.T, hub *Hub) proto.HubClient {
	t.Helper()

	return proto.NewHubClient(newMockHubConn(t, hub))
}

func TestHub_Register_Invalid(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name        string
		request     *proto.StaticInfo
		expectedErr error
	}{
		{
			"invalid address",
			&proto.StaticInfo{
				Address: "",
			},
			errInvalidInfoAddress,
		},
		{
			"invalid gno version",
			&proto.StaticInfo{
				Address:    "random",
				GnoVersion: "",
			},
			errInvalidInfoGnoVersion,
		},
		{
			"invalid OS version",
			&proto.StaticInfo{
				Address:    "random",
				GnoVersion: "random",
				OsVersion:  "",
			},
			errInvalidInfoOSVersion,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				hub    = NewHub()
				client = newMockHubClient(t, hub)
			)

			ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*10)
			defer cancelFn()

			_, err := client.Register(ctx, testCase.request)
			assert.ErrorContains(t, err, testCase.expectedErr.Error())
		})
	}
}

func TestHub_Register(t *testing.T) {
	t.Parallel()

	var (
		hub    = NewHub()
		client = newMockHubClient(t, hub)

		request = &proto.StaticInfo{
			Address:    "random",
			GnoVersion: "random",
			OsVersion:  "random",
		}
	)

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFn()

	// Register the agent
	_, err := client.Register(ctx, request)
	require.NoError(t, err)

	// Make sure the agent is present
	agentRaw, exists := hub.agents.Load(request.Address)
	require.True(t, exists)

	agent, ok := agentRaw.(*proto.StaticInfo)
	require.True(t, ok)

	assert.Equal(t, request.Address, agent.Address)
	assert.Equal(t, request.GnoVersion, agent.GnoVersion)
	assert.Equal(t, request.OsVersion, agent.OsVersion)
}
