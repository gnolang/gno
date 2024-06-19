package main

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Get_Invalid(t *testing.T) {
	t.Parallel()

	// Create the command
	cmd := newRootCmd(commands.NewTestIO())
	args := []string{
		"config",
		"get",
		"--config-path",
		"",
	}

	// Run the command
	cmdErr := cmd.ParseAndRun(context.Background(), args)
	assert.ErrorContains(t, cmdErr, tryConfigInit)
}

// testSetCase outlines the single test case for config get
type testGetCase struct {
	name     string
	field    string
	verifyFn func(*config.Config, []byte)
}

// verifyGetTestTableCommon is the common test table
// verification for config set test cases
func verifyGetTestTableCommon(t *testing.T, testTable []testGetCase) {
	t.Helper()

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup the test config
			path := initializeTestConfig(t)
			args := []string{
				"config",
				"get",
				"--config-path",
				path,
			}

			// Create the command IO
			mockOut := new(bytes.Buffer)

			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))

			// Create the command
			cmd := newRootCmd(io)
			args = append(args, testCase.field)

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Make sure the config was fetched
			loadedCfg, err := config.LoadConfigFile(path)
			require.NoError(t, err)

			testCase.verifyFn(loadedCfg, mockOut.Bytes())
		})
	}
}

func unmarshalJSONCommon[T any](t *testing.T, input []byte) T {
	t.Helper()

	var output T

	require.NoError(t, json.Unmarshal(input, &output))

	return output
}

func TestConfig_Get_Base(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir fetched",
			"home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RootDir, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"proxy app fetched",
			"proxy_app",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProxyApp, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"moniker fetched",
			"moniker",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Moniker, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"fast sync mode fetched",
			"fast_sync",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.FastSyncMode, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"db backend fetched",
			"db_backend",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBBackend, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"db path fetched",
			"db_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBPath, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"validator key fetched",
			"priv_validator_key_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.PrivValidatorKey, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"validator state file fetched",
			"priv_validator_state_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.PrivValidatorState, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"validator listen addr fetched",
			"priv_validator_laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.PrivValidatorListenAddr, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"node key path fetched",
			"node_key_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.NodeKey, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"abci fetched",
			"abci",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ABCI, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"profiling listen address fetched",
			"prof_laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProfListenAddress, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"filter peers flag fetched",
			"filter_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.FilterPeers, unmarshalJSONCommon[bool](t, value))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_Consensus(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir updated",
			"consensus.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.RootDir, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"WAL path updated",
			"consensus.wal_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.WALPath, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"propose timeout updated",
			"consensus.timeout_propose",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutPropose,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"propose timeout delta updated",
			"consensus.timeout_propose_delta",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutProposeDelta,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"prevote timeout updated",
			"consensus.timeout_prevote",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutPrevote,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"prevote timeout delta updated",
			"consensus.timeout_prevote_delta",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutPrevoteDelta,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"precommit timeout updated",
			"consensus.timeout_precommit",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutPrecommit,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"precommit timeout delta updated",
			"consensus.timeout_precommit_delta",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutPrecommitDelta,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"commit timeout updated",
			"consensus.timeout_commit",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.TimeoutCommit,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"skip commit timeout toggle updated",
			"consensus.skip_timeout_commit",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.SkipTimeoutCommit,
					unmarshalJSONCommon[bool](t, value),
				)
			},
		},
		{
			"create empty blocks toggle updated",
			"consensus.create_empty_blocks",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.CreateEmptyBlocks,
					unmarshalJSONCommon[bool](t, value),
				)
			},
		},
		{
			"create empty blocks interval updated",
			"consensus.create_empty_blocks_interval",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.CreateEmptyBlocksInterval,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"peer gossip sleep duration updated",
			"consensus.peer_gossip_sleep_duration",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.PeerGossipSleepDuration,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
		{
			"peer query majority sleep duration updated",
			"consensus.peer_query_maj23_sleep_duration",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.Consensus.PeerQueryMaj23SleepDuration,
					unmarshalJSONCommon[time.Duration](t, value),
				)
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_Events(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"event store type updated",
			"tx_event_store.event_store_type",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.TxEventStore.EventStoreType,
					unmarshalJSONCommon[string](t, value),
				)
			},
		},
		{
			"event store params updated",
			"tx_event_store.event_store_params",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.TxEventStore.Params,
					unmarshalJSONCommon[types.EventStoreParams](t, value),
				)
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_P2P(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir updated",
			"p2p.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.RootDir, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"listen address updated",
			"p2p.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ListenAddress, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"external address updated",
			"p2p.external_address",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ExternalAddress, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"seeds updated",
			"p2p.seeds",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.Seeds, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"persistent peers updated",
			"p2p.persistent_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PersistentPeers, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"upnp toggle updated",
			"p2p.upnp",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.UPNP, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"max inbound peers updated",
			"p2p.max_num_inbound_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxNumInboundPeers, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"max outbound peers updated",
			"p2p.max_num_outbound_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxNumOutboundPeers, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"flush throttle timeout updated",
			"p2p.flush_throttle_timeout",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.FlushThrottleTimeout, unmarshalJSONCommon[time.Duration](t, value))
			},
		},
		{
			"max package payload size updated",
			"p2p.max_packet_msg_payload_size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxPacketMsgPayloadSize, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"send rate updated",
			"p2p.send_rate",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.SendRate, unmarshalJSONCommon[int64](t, value))
			},
		},
		{
			"receive rate updated",
			"p2p.recv_rate",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.RecvRate, unmarshalJSONCommon[int64](t, value))
			},
		},
		{
			"pex reactor toggle updated",
			"p2p.pex",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PexReactor, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"seed mode updated",
			"p2p.seed_mode",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.SeedMode, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"private peer IDs updated",
			"p2p.private_peer_ids",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PrivatePeerIDs, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"allow duplicate IP updated",
			"p2p.allow_duplicate_ip",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.AllowDuplicateIP, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"handshake timeout updated",
			"p2p.handshake_timeout",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.HandshakeTimeout, unmarshalJSONCommon[time.Duration](t, value))
			},
		},
		{
			"dial timeout updated",
			"p2p.dial_timeout",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.DialTimeout, unmarshalJSONCommon[time.Duration](t, value))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_RPC(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir updated",
			"rpc.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.RootDir, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"listen address updated",
			"rpc.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.ListenAddress, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"CORS Allowed Origins updated",
			"rpc.cors_allowed_origins",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.CORSAllowedOrigins, unmarshalJSONCommon[[]string](t, value))
			},
		},
		{
			"CORS Allowed Methods updated",
			"rpc.cors_allowed_methods",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.CORSAllowedMethods, unmarshalJSONCommon[[]string](t, value))
			},
		},
		{
			"CORS Allowed Headers updated",
			"rpc.cors_allowed_headers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.CORSAllowedHeaders, unmarshalJSONCommon[[]string](t, value))
			},
		},
		{
			"GRPC listen address updated",
			"rpc.grpc_laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.GRPCListenAddress, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"GRPC max open connections updated",
			"rpc.grpc_max_open_connections",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.GRPCMaxOpenConnections, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"unsafe value updated",
			"rpc.unsafe",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.Unsafe, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"rpc max open connections updated",
			"rpc.max_open_connections",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.MaxOpenConnections, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"tx commit broadcast timeout updated",
			"rpc.timeout_broadcast_tx_commit",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.TimeoutBroadcastTxCommit, unmarshalJSONCommon[time.Duration](t, value))
			},
		},
		{
			"max body bytes updated",
			"rpc.max_body_bytes",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.MaxBodyBytes, unmarshalJSONCommon[int64](t, value))
			},
		},
		{
			"max header bytes updated",
			"rpc.max_header_bytes",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.MaxHeaderBytes, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"TLS cert file updated",
			"rpc.tls_cert_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.TLSCertFile, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"TLS key file updated",
			"rpc.tls_key_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.TLSKeyFile, unmarshalJSONCommon[string](t, value))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_Mempool(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir updated",
			"mempool.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.RootDir, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"recheck flag updated",
			"mempool.recheck",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Recheck, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"broadcast flag updated",
			"mempool.broadcast",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Broadcast, unmarshalJSONCommon[bool](t, value))
			},
		},
		{
			"WAL path updated",
			"mempool.wal_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.WalPath, unmarshalJSONCommon[string](t, value))
			},
		},
		{
			"size updated",
			"mempool.size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Size, unmarshalJSONCommon[int](t, value))
			},
		},
		{
			"max pending txs bytes updated",
			"mempool.max_pending_txs_bytes",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.MaxPendingTxsBytes, unmarshalJSONCommon[int64](t, value))
			},
		},
		{
			"cache size updated",
			"mempool.cache_size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.CacheSize, unmarshalJSONCommon[int](t, value))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}
