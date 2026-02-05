package rpc

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/mock"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/status"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/events"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	e2eTestChainID            = "test-chain"
	e2eTestLatestHeight int64 = 10
	e2eTestTxHeight     int64 = 5

	e2eNodeMoniker = "test-node"
)

var (
	e2eTestTx          = types.Tx("test-tx")
	e2eTestDeliverData = []byte("deliver-result")
)

// e2eTestServer wraps the JSONRPC server for E2E testing
type e2eTestServer struct {
	listener net.Listener
	mux      *chi.Mux
	jsonrpc  *server.JSONRPC
}

// newE2ETestServer creates a new test server instance with all handlers registered
func newE2ETestServer(t *testing.T) *e2eTestServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	jsonrpc := server.NewJSONRPC(server.WithLogger(log.NewNoopLogger()))

	// Create mock deps
	stateDB := memdb.NewMemDB()
	seedStateDB(t, stateDB)

	blockStore := &mock.BlockStore{
		HeightFn: func() int64 {
			return e2eTestLatestHeight
		},
		LoadBlockMetaFn: func(h int64) *types.BlockMeta {
			header := types.Header{
				Height:  h,
				ChainID: e2eTestChainID,
			}

			if h == e2eTestTxHeight {
				header.NumTxs = 1
				header.TotalTxs = 1
			}

			return &types.BlockMeta{
				Header: header,
			}
		},
		LoadBlockFn: func(h int64) *types.Block {
			header := types.Header{
				Height:  h,
				ChainID: e2eTestChainID,
			}

			if h == e2eTestTxHeight {
				header.NumTxs = 1
				header.TotalTxs = 1
			}

			block := &types.Block{
				Header: header,
			}

			if h == e2eTestTxHeight {
				block.Data = types.Data{
					Txs: types.Txs{e2eTestTx},
				}
			}

			return block
		},
		LoadSeenCommitFn: func(h int64) *types.Commit {
			return &types.Commit{}
		},
		LoadBlockCommitFn: func(h int64) *types.Commit {
			return &types.Commit{}
		},
	}

	mempool := &mock.Mempool{
		CheckTxFn: func(tx types.Tx, cb func(abci.Response)) error {
			if cb != nil {
				cb(abci.ResponseCheckTx{})
			}

			return nil
		},
		ReapMaxTxsFn: func(maxTxs int) types.Txs {
			return types.Txs{}
		},
		SizeFn: func() int {
			return 5
		},
		TxsBytesFn: func() int64 {
			return 1024
		},
	}

	peerSet := &mock.PeerSet{
		ListFn: func() []p2p.PeerConn {
			return nil
		},
		NumInboundFn: func() uint64 {
			return 0
		},
		NumOutboundFn: func() uint64 {
			return 0
		},
	}

	peers := &mock.Peers{
		PeersFn: func() p2p.PeerSet { return peerSet },
	}

	transport := &mock.Transport{
		ListenersFn: func() []string {
			return []string{
				"tcp://127.0.0.1:26656",
			}
		},
		IsListeningFn: func() bool {
			return true
		},
		NodeInfoFn: func() p2pTypes.NodeInfo {
			return p2pTypes.NodeInfo{
				Moniker: e2eNodeMoniker,
			}
		},
	}

	appConn := &mock.AppConn{
		InfoSyncFn: func(_ abci.RequestInfo) (abci.ResponseInfo, error) {
			return abci.ResponseInfo{
				ABCIVersion:      "1.0.0",
				AppVersion:       "1.0.0",
				LastBlockHeight:  e2eTestLatestHeight,
				LastBlockAppHash: []byte("app-hash"),
			}, nil
		},
		QuerySyncFn: func(_ abci.RequestQuery) (abci.ResponseQuery, error) {
			return abci.ResponseQuery{
				Value: []byte("query-result"),
			}, nil
		},
		EchoSyncFn: func(msg string) (abci.ResponseEcho, error) {
			return abci.ResponseEcho{
				Message: msg,
			}, nil
		},
	}

	consensus := &mock.Consensus{
		GetStateFn: func() sm.State {
			return sm.State{
				LastBlockHeight: e2eTestLatestHeight - 1,
				ChainID:         e2eTestChainID,
			}
		},
		GetRoundStateDeepCopyFn: func() *cstypes.RoundState {
			return &cstypes.RoundState{
				Height: e2eTestLatestHeight,
				Round:  0,
			}
		},
		GetRoundStateSimpleFn: func() cstypes.RoundStateSimple {
			return cstypes.RoundStateSimple{
				HeightRoundStep:   "1/0/1",
				StartTime:         time.Now(),
				ProposalBlockHash: []byte("mock-hash"),
			}
		},
		GetConfigDeepCopyFn: func() *cnscfg.ConsensusConfig {
			return cnscfg.DefaultConsensusConfig()
		},
	}

	genesisDoc := &types.GenesisDoc{
		ChainID:     e2eTestChainID,
		GenesisTime: time.Now(),
	}

	// Register all handlers
	core.SetupHealth(jsonrpc)

	core.SetupStatus(jsonrpc, func() (*status.ResultStatus, error) {
		return &status.ResultStatus{
			NodeInfo: p2pTypes.NodeInfo{
				Moniker: e2eNodeMoniker,
			},
			SyncInfo: status.SyncInfo{
				LatestBlockHeight: e2eTestLatestHeight,
			},
		}, nil
	})

	core.SetupABCI(jsonrpc, appConn)
	core.SetupBlocks(jsonrpc, blockStore, stateDB)
	core.SetupConsensus(jsonrpc, consensus, stateDB, peers)
	core.SetupMempool(jsonrpc, mempool, events.NewEventSwitch())
	core.SetupNet(jsonrpc, peers, transport, genesisDoc)
	core.SetupTx(jsonrpc, blockStore, stateDB)

	mux := chi.NewMux()
	mux.Mount("/", jsonrpc.SetupRoutes(chi.NewMux()))

	return &e2eTestServer{
		listener: listener,
		mux:      mux,
		jsonrpc:  jsonrpc,
	}
}

func (s *e2eTestServer) start() {
	go func() {
		_ = http.Serve(s.listener, s.mux)
	}()
}

func (s *e2eTestServer) stop() {
	_ = s.listener.Close()
}

func (s *e2eTestServer) httpAddress() string {
	return fmt.Sprintf("http://%s", s.listener.Addr().String())
}

func (s *e2eTestServer) wsAddress() string {
	return fmt.Sprintf("ws://%s/websocket", s.listener.Addr().String())
}

// seedStateDB seeds the memory DB with default E2E state (predefined)
func seedStateDB(t *testing.T, stateDB *memdb.MemDB) {
	t.Helper()

	var (
		consensusParams = types.DefaultConsensusParams()
		paramsInfo      = sm.ConsensusParamsInfo{
			ConsensusParams:   consensusParams,
			LastHeightChanged: e2eTestLatestHeight,
		}
	)

	require.NoError(
		t,
		stateDB.Set(
			[]byte(fmt.Sprintf("consensusParamsKey:%x", e2eTestLatestHeight)),
			paramsInfo.Bytes(),
		),
	)

	var (
		// Generate a validator secret
		privKey = ed25519.GenPrivKeyFromSecret([]byte("e2e-validator"))

		valSet = types.NewValidatorSet([]*types.Validator{
			types.NewValidator(privKey.PubKey(), 10),
		})

		valInfo = sm.ValidatorsInfo{
			ValidatorSet:      valSet,
			LastHeightChanged: e2eTestLatestHeight,
		}
	)

	require.NoError(
		t,
		stateDB.Set(
			[]byte(fmt.Sprintf("validatorsKey:%x", e2eTestLatestHeight)),
			valInfo.Bytes(),
		),
	)

	txIndex := sm.TxResultIndex{
		BlockNum: e2eTestTxHeight,
		TxIndex:  0,
	}
	require.NoError(
		t,
		stateDB.Set(
			sm.CalcTxResultKey(e2eTestTx.Hash()),
			txIndex.Bytes(),
		),
	)

	abciResponses := &sm.ABCIResponses{
		DeliverTxs: []abci.ResponseDeliverTx{
			{
				ResponseBase: abci.ResponseBase{
					Data: e2eTestDeliverData,
				},
			},
		},
	}

	require.NoError(
		t,
		stateDB.Set(
			sm.CalcABCIResponsesKey(e2eTestTxHeight),
			abciResponses.Bytes(),
		),
	)
}

type e2eClientType string

const (
	clientTypeHTTP e2eClientType = "http"
	clientTypeWS   e2eClientType = "ws"
)

func TestE2E_SingleRequests(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name     string
		verifyFn func(t *testing.T, c *client.RPCClient)
	}{
		{
			name: "health",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.Health(context.Background())
				require.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "status",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.Status(context.Background(), nil)
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, e2eNodeMoniker, result.NodeInfo.Moniker)
				assert.Equal(t, e2eTestLatestHeight, result.SyncInfo.LatestBlockHeight)
			},
		},
		{
			name: "abci_info",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.ABCIInfo(context.Background())
				require.NoError(t, err)
				assert.NotNil(t, result)
			},
		},
		{
			name: "abci_query",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.ABCIQuery(context.Background(), "/test", []byte("data"))
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, []byte("query-result"), result.Response.Value)
			},
		},
		{
			name: "net_info",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.NetInfo(context.Background())
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.True(t, result.Listening)
			},
		},
		{
			name: "genesis",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.Genesis(context.Background())
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, e2eTestChainID, result.Genesis.ChainID)
			},
		},
		{
			name: "block",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				height := e2eTestTxHeight
				result, err := c.Block(context.Background(), &height)
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, height, result.Block.Height)
			},
		},
		{
			name: "blockchain",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.BlockchainInfo(context.Background(), 1, e2eTestLatestHeight)
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, e2eTestLatestHeight, result.LastHeight)
			},
		},
		{
			name: "block_results",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				height := e2eTestTxHeight
				result, err := c.BlockResults(context.Background(), &height)
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, height, result.Height)
				require.NotNil(t, result.Results)
				require.Len(t, result.Results.DeliverTxs, 1)
				assert.Equal(t, e2eTestDeliverData, result.Results.DeliverTxs[0].Data)
			},
		},
		{
			name: "commit",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				height := e2eTestTxHeight
				result, err := c.Commit(context.Background(), &height)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.True(t, result.CanonicalCommit)
			},
		},
		{
			name: "validators",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.Validators(context.Background(), nil)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, e2eTestLatestHeight, result.BlockHeight)
				require.Len(t, result.Validators, 1)
			},
		},
		{
			name: "consensus_state",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.ConsensusState(context.Background())
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, []byte("mock-hash"), result.RoundState.ProposalBlockHash)
			},
		},
		{
			name: "consensus_params",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.ConsensusParams(context.Background(), nil)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, e2eTestLatestHeight, result.BlockHeight)
				require.NotNil(t, result.ConsensusParams.Block)
				assert.Equal(t, types.DefaultConsensusParams().Block.MaxTxBytes, result.ConsensusParams.Block.MaxTxBytes)
			},
		},
		{
			name: "dump_consensus_state",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.DumpConsensusState(context.Background())
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.RoundState)
			},
		},
		{
			name: "unconfirmed_txs",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.UnconfirmedTxs(context.Background(), 10)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 5, result.Total)
			},
		},
		{
			name: "num_unconfirmed_txs",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.NumUnconfirmedTxs(context.Background())
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, 5, result.Count)
			},
		},
		{
			name: "tx",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.Tx(context.Background(), e2eTestTx.Hash())
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.Equal(t, e2eTestTxHeight, result.Height)
				assert.Equal(t, uint32(0), result.Index)
				assert.Equal(t, e2eTestTx, result.Tx)
				assert.Equal(t, e2eTestDeliverData, result.TxResult.Data)
			},
		},
		{
			name: "broadcast_tx_async",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.BroadcastTxAsync(context.Background(), []byte("test-tx"))
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Hash)
			},
		},
		{
			name: "broadcast_tx_sync",
			verifyFn: func(t *testing.T, c *client.RPCClient) {
				t.Helper()

				result, err := c.BroadcastTxSync(context.Background(), []byte("test-tx"))
				require.NoError(t, err)

				assert.NotNil(t, result)
				assert.NotEmpty(t, result.Hash)
			},
		},
	}

	clientTypes := []e2eClientType{
		clientTypeHTTP,
		clientTypeWS,
	}

	for _, testCase := range testTable {
		for _, clientType := range clientTypes {
			t.Run(fmt.Sprintf("%s/%s", testCase.name, clientType), func(t *testing.T) {
				t.Parallel()

				// Each subtest gets its own server and client
				srv := newE2ETestServer(t)

				srv.start()
				defer srv.stop()

				var (
					rpcClient *client.RPCClient
					err       error
				)

				if clientType == clientTypeHTTP {
					rpcClient, err = client.NewHTTPClient(srv.httpAddress())
				} else {
					rpcClient, err = client.NewWSClient(srv.wsAddress())
				}

				require.NoError(t, err)

				defer func() {
					require.NoError(t, rpcClient.Close())
				}()

				testCase.verifyFn(t, rpcClient)
			})
		}
	}
}

func TestE2E_BatchRequests_HTTP(t *testing.T) {
	t.Parallel()

	// Create and start the server
	srv := newE2ETestServer(t)
	srv.start()
	defer srv.stop()

	// Create HTTP client
	httpClient, err := client.NewHTTPClient(srv.httpAddress())
	require.NoError(t, err)

	defer func() {
		_ = httpClient.Close()
	}()

	// Create batch
	batch := httpClient.NewBatch()

	// Add multiple requests to the batch
	batch.Health()
	batch.Status()
	batch.NetInfo()
	batch.Genesis()

	require.Equal(t, 4, batch.Count())

	// Send batch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := batch.Send(ctx)
	require.NoError(t, err)
	require.Len(t, results, 4)

	// Verify batch is cleared after send
	assert.Equal(t, 0, batch.Count())

	// Verify results
	for i, result := range results {
		assert.NotNil(t, result, "result %d should not be nil", i)
	}
}

func TestE2E_BatchRequests_WS(t *testing.T) {
	t.Parallel()

	// Create and start the server
	srv := newE2ETestServer(t)
	srv.start()
	defer srv.stop()

	// Create WS client
	wsClient, err := client.NewWSClient(srv.wsAddress())
	require.NoError(t, err)
	defer func() {
		_ = wsClient.Close()
	}()

	// Create batch
	batch := wsClient.NewBatch()

	// Add multiple requests to the batch
	batch.Health()
	batch.Status()
	batch.NetInfo()
	batch.Genesis()

	require.Equal(t, 4, batch.Count())

	// Send batch
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := batch.Send(ctx)
	require.NoError(t, err)
	require.Len(t, results, 4)

	// Verify batch is cleared after send
	assert.Equal(t, 0, batch.Count())

	// Verify results
	for i, result := range results {
		assert.NotNil(t, result, "result %d should not be nil", i)
	}
}

func TestE2E_BatchRequests_MixedEndpoints(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		clientType e2eClientType
	}{
		{
			name:       "http",
			clientType: clientTypeHTTP,
		},
		{
			name:       "ws",
			clientType: clientTypeWS,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create and start the server
			srv := newE2ETestServer(t)
			srv.start()
			defer srv.stop()

			var (
				rpcClient *client.RPCClient
				err       error
			)

			if testCase.clientType == clientTypeHTTP {
				rpcClient, err = client.NewHTTPClient(srv.httpAddress())
			} else {
				rpcClient, err = client.NewWSClient(srv.wsAddress())
			}

			require.NoError(t, err)

			defer func() {
				_ = rpcClient.Close()
			}()

			// Create batch with mixed endpoints
			batch := rpcClient.NewBatch()

			height := e2eTestTxHeight

			batch.Health()
			batch.Status()
			batch.ABCIQuery("/test", []byte("data"))
			batch.Block(&height)

			require.Equal(t, 4, batch.Count())

			// Send batch
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			results, err := batch.Send(ctx)
			require.NoError(t, err)
			require.Len(t, results, 4)

			// Verify all results are non-nil
			for i, result := range results {
				assert.NotNil(t, result, "result %d should not be nil", i)
			}
		})
	}
}

func TestE2E_SequentialRequests(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		clientType e2eClientType
	}{
		{
			name:       "http",
			clientType: clientTypeHTTP,
		},
		{
			name:       "ws",
			clientType: clientTypeWS,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create and start the server
			srv := newE2ETestServer(t)
			srv.start()
			defer srv.stop()

			var (
				rpcClient *client.RPCClient
				err       error
			)

			if testCase.clientType == clientTypeHTTP {
				rpcClient, err = client.NewHTTPClient(srv.httpAddress())
			} else {
				rpcClient, err = client.NewWSClient(srv.wsAddress())
			}

			require.NoError(t, err)

			defer func() {
				_ = rpcClient.Close()
			}()

			ctx := context.Background()

			// Send multiple sequential requests
			for i := 0; i < 5; i++ {
				// Health check
				healthResult, err := rpcClient.Health(ctx)
				require.NoError(t, err)

				assert.NotNil(t, healthResult)

				// Status check
				statusResult, err := rpcClient.Status(ctx, nil)
				require.NoError(t, err)

				assert.NotNil(t, statusResult)
				assert.Equal(t, e2eNodeMoniker, statusResult.NodeInfo.Moniker)
			}
		})
	}
}

func TestE2E_ClientReuse(t *testing.T) {
	t.Parallel()

	// Create and start the server
	srv := newE2ETestServer(t)
	srv.start()
	defer srv.stop()

	// Test with WS client (most relevant for connection reuse)
	wsClient, err := client.NewWSClient(srv.wsAddress())

	require.NoError(t, err)

	defer func() {
		_ = wsClient.Close()
	}()

	ctx := context.Background()

	// Send multiple batches using the same client
	for i := 0; i < 3; i++ {
		batch := wsClient.NewBatch()
		batch.Health()
		batch.Status()

		results, err := batch.Send(ctx)
		require.NoError(t, err, "batch %d failed", i)
		require.Len(t, results, 2, "batch %d should have 2 results", i)

		for j, result := range results {
			assert.NotNil(t, result, "batch %d result %d should not be nil", i, j)
		}
	}
}
