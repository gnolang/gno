package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gnolang/gnostats/proto"
	"github.com/rs/xid"
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

// generateDataPoints generates dummy data points
func generateDataPoints(t *testing.T, count int) []*proto.DataPoint {
	t.Helper()

	dataPoints := make([]*proto.DataPoint, count)

	for i := 0; i < count; i++ {
		dataPoints[i] = &proto.DataPoint{
			StaticInfo: &proto.StaticInfo{
				Address: fmt.Sprintf("address %d", i),
			},
		}
	}

	return dataPoints
}

func TestHub_GetDataStream(t *testing.T) {
	t.Parallel()

	var (
		hub    = NewHub()
		client = newMockHubClient(t, hub)

		dataPoints   = generateDataPoints(t, 100)
		receivedData []*proto.DataPoint

		sendCh = make(chan *proto.DataPoint, len(dataPoints))
		sendID = xid.New()

		mockSubscriptions = &mockSubscriptions{
			subscribeFn: func() (xid.ID, dataStream) {
				return sendID, sendCh
			},

			unsubscribeFn: func(id xid.ID) {
				require.Equal(t, sendID, id)
			},
		}
	)

	// Set the subs handler
	hub.subs = mockSubscriptions

	// Preload the channel with data points
	for _, dataPoint := range dataPoints {
		sendCh <- dataPoint
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Initiate the stream
	stream, err := client.GetDataStream(ctx, nil)
	require.NoError(t, err)

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		defer require.NoError(t, stream.CloseSend())

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				return
			default:
				data, err := stream.Recv()
				require.NoError(t, err)

				receivedData = append(receivedData, data)

				if len(receivedData) == len(dataPoints) {
					return
				}
			}
		}
	}()

	wg.Wait()

	// Verify the received data
	require.Len(t, receivedData, len(dataPoints))

	for index, dataPoint := range receivedData {
		require.Equal(
			t,
			dataPoints[index].StaticInfo.Address,
			dataPoint.StaticInfo.Address,
		)
	}
}

// generateStaticInfo generates dummy static info
func generateStaticInfo(t *testing.T, count int) []*proto.StaticInfo {
	t.Helper()

	info := make([]*proto.StaticInfo, count)

	for i := 0; i < count; i++ {
		info[i] = &proto.StaticInfo{
			Address:    fmt.Sprintf("address %d", i),
			GnoVersion: fmt.Sprintf("gno version %d", i),
			OsVersion:  fmt.Sprintf("os version %d", i),
		}
	}

	return info
}

// generateDynamicInfo generates dummy dynamic info
func generateDynamicInfo(t *testing.T, count int) []*proto.DynamicInfo {
	t.Helper()

	info := make([]*proto.DynamicInfo, count)

	for i := 0; i < count; i++ {
		info[i] = &proto.DynamicInfo{
			Address:     fmt.Sprintf("address %d", i),
			Moniker:     fmt.Sprintf("moniker %d", i),
			IsValidator: true,
			NetInfo: &proto.NetInfo{
				P2PAddress: fmt.Sprintf("p2p address %d", i),
				Peers:      make([]*proto.PeerInfo, 0),
			},
			BlockInfo: &proto.BlockInfo{
				Number: 1,
			},
		}
	}

	return info
}

func TestHub_PushData(t *testing.T) {
	t.Parallel()

	t.Run("unregistered agent", func(t *testing.T) {
		t.Parallel()

		var (
			hub    = NewHub()
			client = newMockHubClient(t, hub)

			dynamicInfo = generateDynamicInfo(t, 1)[0]

			receivedData []*proto.DataPoint

			mockSubscriptions = &mockSubscriptions{
				notifyFn: func(data *proto.DataPoint) {
					receivedData = append(receivedData, data)
				},
			}
		)

		// Set the subs handler
		hub.subs = mockSubscriptions

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Initiate the stream
		stream, err := client.PushData(ctx)
		require.NoError(t, err)

		// Attempt to send data
		require.NoError(t, stream.Send(dynamicInfo))
		require.NoError(t, stream.CloseSend())

		// Make sure nothing was sent out
		assert.Len(t, receivedData, 0)
	})

	t.Run("valid agent", func(t *testing.T) {
		t.Parallel()

		var (
			hub    = NewHub()
			client = newMockHubClient(t, hub)

			staticInfo  = generateStaticInfo(t, 100)
			dynamicInfo = generateDynamicInfo(t, len(staticInfo))

			receivedData []*proto.DataPoint
			receivedCh   = make(chan struct{}, 1)

			mockSubscriptions = &mockSubscriptions{
				notifyFn: func(data *proto.DataPoint) {
					receivedData = append(receivedData, data)

					if len(receivedData) == len(dynamicInfo) {
						// We need to explicitly trigger a reception check,
						// since gRPC stream.Sends are async under the hood
						receivedCh <- struct{}{}
					}
				},
			}
		)

		// Set the subs handler
		hub.subs = mockSubscriptions

		// Register the agents
		for _, staticInfo := range staticInfo {
			_, err := client.Register(context.Background(), staticInfo)
			require.NoError(t, err)
		}

		ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFn()

		// Initiate the stream
		stream, err := client.PushData(ctx)
		require.NoError(t, err)

		// Send the dynamic info.
		// This call, albeit seemingly synchronous
		// does not immediately make the data
		// available on stream.Recv()
		for _, info := range dynamicInfo {
			require.NoError(t, stream.Send(info))
		}

		require.NoError(t, stream.CloseSend())

		// Make sure all data was properly received
		select {
		case <-time.After(5 * time.Second):
		case <-receivedCh:
		}

		// Make sure the correct data points were sent out
		require.Len(t, receivedData, len(dynamicInfo))

		for index, dataPoint := range receivedData {
			assert.Equal(t, staticInfo[index].Address, dataPoint.StaticInfo.Address)
			assert.Equal(t, staticInfo[index].OsVersion, dataPoint.StaticInfo.OsVersion)
			assert.Equal(t, staticInfo[index].GnoVersion, dataPoint.StaticInfo.GnoVersion)

			assert.Equal(t, dynamicInfo[index].Address, dataPoint.DynamicInfo.Address)
			assert.Equal(t, dynamicInfo[index].Moniker, dataPoint.DynamicInfo.Moniker)
			assert.Equal(t, dynamicInfo[index].IsValidator, dataPoint.DynamicInfo.IsValidator)
			assert.Equal(t, dynamicInfo[index].NetInfo.P2PAddress, dataPoint.DynamicInfo.NetInfo.P2PAddress)
			assert.Equal(t, dynamicInfo[index].BlockInfo.Number, dataPoint.DynamicInfo.BlockInfo.Number)
		}
	})
}
