package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/bft/config"
	"github.com/gnolang/gno/tm2/pkg/bft/state/eventstore/file"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// initializeTestConfig initializes a default configuration
// at a temporary path
func initializeTestConfig(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.toml")
	cfg := config.DefaultConfig()

	require.NoError(t, config.WriteConfigFile(path, cfg))

	return path
}

// testSetCase outlines the single test case for config set
type testSetCase struct {
	name     string
	flags    []string
	verifyFn func(*config.Config, string)
}

// verifySetTestTableCommon is the common test table
// verification for config set test cases
func verifySetTestTableCommon(t *testing.T, testTable []testSetCase) {
	t.Helper()

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Setup the test config
			path := initializeTestConfig(t)
			args := []string{
				"config",
				"set",
				"--config-path",
				path,
			}

			// Create the command
			cmd := newRootCmd(commands.NewTestIO())
			args = append(args, testCase.flags...)

			// Run the command
			cmdErr := cmd.ParseAndRun(context.Background(), args)
			require.NoError(t, cmdErr)

			// Make sure the config was updated
			loadedCfg, err := config.LoadConfigFile(path)
			require.NoError(t, err)

			testCase.verifyFn(loadedCfg, testCase.flags[len(testCase.flags)-1])
		})
	}
}

func TestConfig_Set_Invalid(t *testing.T) {
	t.Parallel()

	t.Run("invalid config path", func(t *testing.T) {
		t.Parallel()

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"set",
			"--config-path",
			"",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to load config")
	})

	t.Run("invalid config change", func(t *testing.T) {
		t.Parallel()

		// Setup the test config
		path := initializeTestConfig(t)

		// Create the command
		cmd := newRootCmd(commands.NewTestIO())
		args := []string{
			"config",
			"set",
			"--config-path",
			path,
			"db_backend",
			"random db backend",
		}

		// Run the command
		cmdErr := cmd.ParseAndRun(context.Background(), args)
		assert.ErrorContains(t, cmdErr, "unable to validate config")
	})
}

func TestConfig_Set_Base(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"root dir updated",
			[]string{
				"home",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RootDir)
			},
		},
		{
			"proxy app updated",
			[]string{
				"proxy_app",
				"example proxy app",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProxyApp)
			},
		},
		{
			"moniker updated",
			[]string{
				"moniker",
				"example moniker",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Moniker)
			},
		},
		{
			"fast sync mode updated",
			[]string{
				"fast_sync",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.FastSyncMode)
			},
		},
		{
			"db backend updated",
			[]string{
				"db_backend",
				db.GoLevelDBBackend.String(),
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBBackend)
			},
		},
		{
			"db path updated",
			[]string{
				"db_dir",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBPath)
			},
		},
		{
			"genesis path updated",
			[]string{
				"genesis_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Genesis)
			},
		},
		{
			"validator key updated",
			[]string{
				"priv_validator_key_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorKey)
			},
		},
		{
			"validator state file updated",
			[]string{
				"priv_validator_state_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorState)
			},
		},
		{
			"validator listen addr updated",
			[]string{
				"priv_validator_laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorListenAddr)
			},
		},
		{
			"node key path updated",
			[]string{
				"node_key_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.NodeKey)
			},
		},
		{
			"abci updated",
			[]string{
				"abci",
				config.LocalABCI,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ABCI)
			},
		},
		{
			"profiling listen address updated",
			[]string{
				"prof_laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProfListenAddress)
			},
		},
		{
			"filter peers flag updated",
			[]string{
				"filter_peers",
				"true",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.FilterPeers)
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}

func TestConfig_Set_Consensus(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"root dir updated",
			[]string{
				"consensus.home",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.RootDir)
			},
		},
		{
			"WAL path updated",
			[]string{
				"consensus.wal_file",
				"example WAL path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.WALPath)
			},
		},
		{
			"propose timeout updated",
			[]string{
				"consensus.timeout_propose",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPropose.String())
			},
		},
		{
			"propose timeout delta updated",
			[]string{
				"consensus.timeout_propose_delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutProposeDelta.String())
			},
		},
		{
			"prevote timeout updated",
			[]string{
				"consensus.timeout_prevote",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevote.String())
			},
		},
		{
			"prevote timeout delta updated",
			[]string{
				"consensus.timeout_prevote_delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevoteDelta.String())
			},
		},
		{
			"precommit timeout updated",
			[]string{
				"consensus.timeout_precommit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommit.String())
			},
		},
		{
			"precommit timeout delta updated",
			[]string{
				"consensus.timeout_precommit_delta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommitDelta.String())
			},
		},
		{
			"commit timeout updated",
			[]string{
				"consensus.timeout_commit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutCommit.String())
			},
		},
		{
			"skip commit timeout toggle updated",
			[]string{
				"consensus.skip_timeout_commit",
				"true",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.Consensus.SkipTimeoutCommit)
			},
		},
		{
			"create empty blocks toggle updated",
			[]string{
				"consensus.create_empty_blocks",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)
				assert.Equal(t, boolVal, loadedCfg.Consensus.CreateEmptyBlocks)
			},
		},
		{
			"create empty blocks interval updated",
			[]string{
				"consensus.create_empty_blocks_interval",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.CreateEmptyBlocksInterval.String())
			},
		},
		{
			"peer gossip sleep duration updated",
			[]string{
				"consensus.peer_gossip_sleep_duration",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerGossipSleepDuration.String())
			},
		},
		{
			"peer query majority sleep duration updated",
			[]string{
				"consensus.peer_query_maj23_sleep_duration",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerQueryMaj23SleepDuration.String())
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}

func TestConfig_Set_Events(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"event store type updated",
			[]string{
				"tx_event_store.event_store_type",
				file.EventStoreType,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.TxEventStore.EventStoreType)
			},
		},
		{
			"event store params updated",
			[]string{
				"tx_event_store.event_store_params",
				"key1=value1,key2=value2",
			},
			func(loadedCfg *config.Config, value string) {
				val, ok := loadedCfg.TxEventStore.Params["key1"]
				assert.True(t, ok)
				assert.Equal(t, "value1", val)

				val, ok = loadedCfg.TxEventStore.Params["key2"]
				assert.True(t, ok)
				assert.Equal(t, "value2", val)
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}

func TestConfig_Set_P2P(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"root dir updated",
			[]string{
				"p2p.home",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"p2p.laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ListenAddress)
			},
		},
		{
			"external address updated",
			[]string{
				"p2p.external_address",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ExternalAddress)
			},
		},
		{
			"seeds updated",
			[]string{
				"p2p.seeds",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.Seeds)
			},
		},
		{
			"persistent peers updated",
			[]string{
				"p2p.persistent_peers",
				"nodeID@0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PersistentPeers)
			},
		},
		{
			"upnp toggle updated",
			[]string{
				"p2p.upnp",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.UPNP)
			},
		},
		{
			"max inbound peers updated",
			[]string{
				"p2p.max_num_inbound_peers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumInboundPeers))
			},
		},
		{
			"max outbound peers updated",
			[]string{
				"p2p.max_num_outbound_peers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumOutboundPeers))
			},
		},
		{
			"flush throttle timeout updated",
			[]string{
				"p2p.flush_throttle_timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.FlushThrottleTimeout.String())
			},
		},
		{
			"max package payload size updated",
			[]string{
				"p2p.max_packet_msg_payload_size",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxPacketMsgPayloadSize))
			},
		},
		{
			"send rate updated",
			[]string{
				"p2p.send_rate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.SendRate))
			},
		},
		{
			"receive rate updated",
			[]string{
				"p2p.recv_rate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.RecvRate))
			},
		},
		{
			"pex reactor toggle updated",
			[]string{
				"p2p.pex",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.PexReactor)
			},
		},
		{
			"seed mode updated",
			[]string{
				"p2p.seed_mode",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.SeedMode)
			},
		},
		{
			"private peer IDs updated",
			[]string{
				"p2p.private_peer_ids",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PrivatePeerIDs)
			},
		},
		{
			"allow duplicate IPs updated",
			[]string{
				"p2p.allow_duplicate_ip",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.AllowDuplicateIP)
			},
		},
		{
			"handshake timeout updated",
			[]string{
				"p2p.handshake_timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.HandshakeTimeout.String())
			},
		},
		{
			"dial timeout updated",
			[]string{
				"p2p.dial_timeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.DialTimeout.String())
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}

func TestConfig_Set_RPC(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"root dir updated",
			[]string{
				"rpc.home",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"rpc.laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.ListenAddress)
			},
		},
		{
			"CORS Allowed Origins updated",
			[]string{
				"rpc.cors_allowed_origins",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedOrigins)
			},
		},
		{
			"CORS Allowed Methods updated",
			[]string{
				"rpc.cors_allowed_methods",
				"POST,GET",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedMethods)
			},
		},
		{
			"CORS Allowed Headers updated",
			[]string{
				"rpc.cors_allowed_headers",
				"*",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedHeaders)
			},
		},
		{
			"GRPC listen address updated",
			[]string{
				"rpc.grpc_laddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.GRPCListenAddress)
			},
		},
		{
			"GRPC max open connections updated",
			[]string{
				"rpc.grpc_max_open_connections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.GRPCMaxOpenConnections))
			},
		},
		{
			"unsafe value updated",
			[]string{
				"rpc.unsafe",
				"true",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.RPC.Unsafe)
			},
		},
		{
			"rpc max open connections updated",
			[]string{
				"rpc.max_open_connections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxOpenConnections))
			},
		},
		{
			"tx commit broadcast timeout updated",
			[]string{
				"rpc.timeout_broadcast_tx_commit",
				(time.Second * 10).String(),
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TimeoutBroadcastTxCommit.String())
			},
		},
		{
			"max body bytes updated",
			[]string{
				"rpc.max_body_bytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxBodyBytes))
			},
		},
		{
			"max header bytes updated",
			[]string{
				"rpc.max_header_bytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxHeaderBytes))
			},
		},
		{
			"TLS cert file updated",
			[]string{
				"rpc.tls_cert_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSCertFile)
			},
		},
		{
			"TLS key file updated",
			[]string{
				"rpc.tls_key_file",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSKeyFile)
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}

func TestConfig_Set_Mempool(t *testing.T) {
	t.Parallel()

	testTable := []testSetCase{
		{
			"root dir updated",
			[]string{
				"mempool.home",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.RootDir)
			},
		},
		{
			"recheck flag updated",
			[]string{
				"mempool.recheck",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Recheck)
			},
		},
		{
			"broadcast flag updated",
			[]string{
				"mempool.broadcast",
				"false",
			},
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Broadcast)
			},
		},
		{
			"WAL path updated",
			[]string{
				"mempool.wal_dir",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.WalPath)
			},
		},
		{
			"size updated",
			[]string{
				"mempool.size",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.Size))
			},
		},
		{
			"max pending txs bytes updated",
			[]string{
				"mempool.max_pending_txs_bytes",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.MaxPendingTxsBytes))
			},
		},
		{
			"cache size updated",
			[]string{
				"mempool.cache_size",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.CacheSize))
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}
