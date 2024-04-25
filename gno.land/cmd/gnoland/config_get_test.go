package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
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
	assert.ErrorContains(t, cmdErr, "unable to load config")
}

// testSetCase outlines the single test case for config get
type testGetCase struct {
	name     string
	field    string
	verifyFn func(*config.Config, string)
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

			testCase.verifyFn(loadedCfg, mockOut.String())
		})
	}
}

func TestConfig_Get_Base(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir fetched",
			"home",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.RootDir, value)
			},
		},
		{
			"proxy app fetched",
			"proxy_app",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ProxyApp, value)
			},
		},
		{
			"moniker fetched",
			"moniker",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.Moniker, value)
			},
		},
		{
			"fast sync mode fetched",
			"fast_sync",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, loadedCfg.FastSyncMode, boolVal)
			},
		},
		{
			"db backend fetched",
			"db_backend",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.DBBackend, value)
			},
		},
		{
			"db path fetched",
			"db_dir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.DBPath, value)
			},
		},
		{
			"validator key fetched",
			"priv_validator_key_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorKey, value)
			},
		},
		{
			"validator state file fetched",
			"priv_validator_state_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorState, value)
			},
		},
		{
			"validator listen addr fetched",
			"priv_validator_laddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorListenAddr, value)
			},
		},
		{
			"node key path fetched",
			"node_key_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.NodeKey, value)
			},
		},
		{
			"abci fetched",
			"abci",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ABCI, value)
			},
		},
		{
			"profiling listen address fetched",
			"prof_laddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ProfListenAddress, value)
			},
		},
		{
			"filter peers flag fetched",
			"filter_peers",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, loadedCfg.FilterPeers, boolVal)
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
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.RootDir)
			},
		},
		{
			"WAL path updated",
			"consensus.wal_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.WALPath)
			},
		},
		{
			"propose timeout updated",
			"consensus.timeout_propose",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPropose.String())
			},
		},
		{
			"propose timeout delta updated",
			"consensus.timeout_propose_delta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutProposeDelta.String())
			},
		},
		{
			"prevote timeout updated",
			"consensus.timeout_prevote",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevote.String())
			},
		},
		{
			"prevote timeout delta updated",
			"consensus.timeout_prevote_delta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevoteDelta.String())
			},
		},
		{
			"precommit timeout updated",
			"consensus.timeout_precommit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommit.String())
			},
		},
		{
			"precommit timeout delta updated",
			"consensus.timeout_precommit_delta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommitDelta.String())
			},
		},
		{
			"commit timeout updated",
			"consensus.timeout_commit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutCommit.String())
			},
		},
		{
			"skip commit timeout toggle updated",
			"consensus.skip_timeout_commit",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.Consensus.SkipTimeoutCommit)
			},
		},
		{
			"create empty blocks toggle updated",
			"consensus.create_empty_blocks",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)
				assert.Equal(t, boolVal, loadedCfg.Consensus.CreateEmptyBlocks)
			},
		},
		{
			"create empty blocks interval updated",
			"consensus.create_empty_blocks_interval",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.CreateEmptyBlocksInterval.String())
			},
		},
		{
			"peer gossip sleep duration updated",
			"consensus.peer_gossip_sleep_duration",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerGossipSleepDuration.String())
			},
		},
		{
			"peer query majority sleep duration updated",
			"consensus.peer_query_maj23_sleep_duration",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerQueryMaj23SleepDuration.String())
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
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.TxEventStore.EventStoreType)
			},
		},
		{
			"event store params updated",
			"tx_event_store.event_store_params",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.TxEventStore.Params))
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
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.RootDir)
			},
		},
		{
			"listen address updated",
			"p2p.laddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ListenAddress)
			},
		},
		{
			"external address updated",
			"p2p.external_address",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ExternalAddress)
			},
		},
		{
			"seeds updated",
			"p2p.seeds",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.Seeds)
			},
		},
		{
			"persistent peers updated",
			"p2p.persistent_peers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PersistentPeers)
			},
		},
		{
			"upnp toggle updated",
			"p2p.upnp",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.UPNP)
			},
		},
		{
			"max inbound peers updated",
			"p2p.max_num_inbound_peers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumInboundPeers))
			},
		},
		{
			"max outbound peers updated",
			"p2p.max_num_outbound_peers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumOutboundPeers))
			},
		},
		{
			"flush throttle timeout updated",
			"p2p.flush_throttle_timeout",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.FlushThrottleTimeout.String())
			},
		},
		{
			"max package payload size updated",
			"p2p.max_packet_msg_payload_size",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxPacketMsgPayloadSize))
			},
		},
		{
			"send rate updated",
			"p2p.send_rate",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.SendRate))
			},
		},
		{
			"receive rate updated",
			"p2p.recv_rate",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.RecvRate))
			},
		},
		{
			"pex reactor toggle updated",
			"p2p.pex",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.PexReactor)
			},
		},
		{
			"seed mode updated",
			"p2p.seed_mode",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.SeedMode)
			},
		},
		{
			"private peer IDs updated",
			"p2p.private_peer_ids",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PrivatePeerIDs)
			},
		},
		{
			"allow duplicate IP updated",
			"p2p.allow_duplicate_ip",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.AllowDuplicateIP)
			},
		},
		{
			"handshake timeout updated",
			"p2p.handshake_timeout",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.HandshakeTimeout.String())
			},
		},
		{
			"dial timeout updated",
			"p2p.dial_timeout",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.DialTimeout.String())
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
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.RootDir)
			},
		},
		{
			"listen address updated",
			"rpc.laddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.ListenAddress)
			},
		},
		{
			"CORS Allowed Origins updated",
			"rpc.cors_allowed_origins",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedOrigins))
			},
		},
		{
			"CORS Allowed Methods updated",
			"rpc.cors_allowed_methods",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedMethods))
			},
		},
		{
			"CORS Allowed Headers updated",
			"rpc.cors_allowed_headers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedHeaders))
			},
		},
		{
			"GRPC listen address updated",
			"rpc.grpc_laddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.GRPCListenAddress)
			},
		},
		{
			"GRPC max open connections updated",
			"rpc.grpc_max_open_connections",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.GRPCMaxOpenConnections))
			},
		},
		{
			"unsafe value updated",
			"rpc.unsafe",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.RPC.Unsafe)
			},
		},
		{
			"rpc max open connections updated",
			"rpc.max_open_connections",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxOpenConnections))
			},
		},
		{
			"tx commit broadcast timeout updated",
			"rpc.timeout_broadcast_tx_commit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TimeoutBroadcastTxCommit.String())
			},
		},
		{
			"max body bytes updated",
			"rpc.max_body_bytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxBodyBytes))
			},
		},
		{
			"max header bytes updated",
			"rpc.max_header_bytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxHeaderBytes))
			},
		},
		{
			"TLS cert file updated",
			"rpc.tls_cert_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSCertFile)
			},
		},
		{
			"TLS key file updated",
			"rpc.tls_key_file",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSKeyFile)
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
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.RootDir)
			},
		},
		{
			"recheck flag updated",
			"mempool.recheck",
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Recheck)
			},
		},
		{
			"broadcast flag updated",
			"mempool.broadcast",
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Broadcast)
			},
		},
		{
			"WAL path updated",
			"mempool.wal_dir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.WalPath)
			},
		},
		{
			"size updated",
			"mempool.size",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.Size))
			},
		},
		{
			"max pending txs bytes updated",
			"mempool.max_pending_txs_bytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.MaxPendingTxsBytes))
			},
		},
		{
			"cache size updated",
			"mempool.cache_size",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.CacheSize))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}
