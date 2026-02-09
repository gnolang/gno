package main

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
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
	isRaw    bool
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

			if testCase.isRaw {
				args = append(args, "--raw")
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

func escapeNewline(value []byte) string {
	return strings.ReplaceAll(string(value), "\n", "")
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
			false,
		},
		{
			"root dir fetched, raw",
			"home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RootDir, escapeNewline(value))
			},
			true,
		},
		{
			"proxy app fetched",
			"proxy_app",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProxyApp, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"proxy app fetched, raw",
			"proxy_app",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProxyApp, escapeNewline(value))
			},
			true,
		},
		{
			"moniker fetched",
			"moniker",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Moniker, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"moniker fetched, raw",
			"moniker",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Moniker, escapeNewline(value))
			},
			true,
		},
		{
			"fast sync mode fetched",
			"fast_sync",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.FastSyncMode, unmarshalJSONCommon[bool](t, value))
			},
			false,
		},
		{
			"db backend fetched",
			"db_backend",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBBackend, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"db backend fetched, raw",
			"db_backend",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBBackend, escapeNewline(value))
			},
			true,
		},
		{
			"db path fetched",
			"db_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBPath, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"db path fetched, raw",
			"db_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.DBPath, escapeNewline(value))
			},
			true,
		},
		{
			"validator sign state fetched",
			"consensus.priv_validator.sign_state",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.PrivValidator.SignState, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"validator sign state fetched, raw",
			"consensus.priv_validator.sign_state",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.PrivValidator.SignState, escapeNewline(value))
			},
			true,
		},
		{
			"validator local signer fetched",
			"consensus.priv_validator.local_signer",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.PrivValidator.LocalSigner, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"validator local_signer fetched, raw",
			"consensus.priv_validator.local_signer",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.PrivValidator.LocalSigner, escapeNewline(value))
			},
			true,
		},
		{
			"node key path fetched",
			"node_key_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.NodeKey, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"node key path fetched, raw",
			"node_key_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.NodeKey, escapeNewline(value))
			},
			true,
		},
		{
			"abci fetched",
			"abci",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ABCI, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"abci fetched, raw",
			"abci",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ABCI, escapeNewline(value))
			},
			true,
		},
		{
			"profiling listen address fetched",
			"prof_laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProfListenAddress, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"profiling listen address fetched, raw",
			"prof_laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.ProfListenAddress, escapeNewline(value))
			},
			true,
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
			false,
		},
		{
			"root dir updated, raw",
			"consensus.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.RootDir, escapeNewline(value))
			},
			true,
		},
		{
			"WAL path updated",
			"consensus.wal_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.WALPath, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"WAL path updated, raw",
			"consensus.wal_file",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Consensus.WALPath, escapeNewline(value))
			},
			true,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
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
			false,
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_Events(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"event store type",
			"tx_event_store.event_store_type",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.TxEventStore.EventStoreType,
					unmarshalJSONCommon[string](t, value),
				)
			},
			false,
		},
		{
			"event store type, raw",
			"tx_event_store.event_store_type",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.TxEventStore.EventStoreType,
					escapeNewline(value),
				)
			},
			true,
		},
		{
			"event store params",
			"tx_event_store.event_store_params",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(
					t,
					loadedCfg.TxEventStore.Params,
					unmarshalJSONCommon[types.EventStoreParams](t, value),
				)
			},
			false,
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_P2P(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir",
			"p2p.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.RootDir, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"root dir, raw",
			"p2p.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.RootDir, escapeNewline(value))
			},
			true,
		},
		{
			"listen address",
			"p2p.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ListenAddress, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"listen address, raw",
			"p2p.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ListenAddress, escapeNewline(value))
			},
			true,
		},
		{
			"external address",
			"p2p.external_address",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ExternalAddress, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"external address, raw",
			"p2p.external_address",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.ExternalAddress, escapeNewline(value))
			},
			true,
		},
		{
			"seeds",
			"p2p.seeds",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.Seeds, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"seeds, raw",
			"p2p.seeds",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.Seeds, escapeNewline(value))
			},
			true,
		},
		{
			"persistent peers",
			"p2p.persistent_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PersistentPeers, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"persistent peers, raw",
			"p2p.persistent_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PersistentPeers, escapeNewline(value))
			},
			true,
		},
		{
			"max inbound peers",
			"p2p.max_num_inbound_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxNumInboundPeers, unmarshalJSONCommon[uint64](t, value))
			},
			false,
		},
		{
			"max outbound peers",
			"p2p.max_num_outbound_peers",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxNumOutboundPeers, unmarshalJSONCommon[uint64](t, value))
			},
			false,
		},
		{
			"flush throttle timeout",
			"p2p.flush_throttle_timeout",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.FlushThrottleTimeout, unmarshalJSONCommon[time.Duration](t, value))
			},
			false,
		},
		{
			"max package payload size",
			"p2p.max_packet_msg_payload_size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.MaxPacketMsgPayloadSize, unmarshalJSONCommon[int](t, value))
			},
			false,
		},
		{
			"send rate",
			"p2p.send_rate",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.SendRate, unmarshalJSONCommon[int64](t, value))
			},
			false,
		},
		{
			"receive rate",
			"p2p.recv_rate",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.RecvRate, unmarshalJSONCommon[int64](t, value))
			},
			false,
		},
		{
			"pex reactor toggle",
			"p2p.pex",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PeerExchange, unmarshalJSONCommon[bool](t, value))
			},
			false,
		},
		{
			"private peer IDs",
			"p2p.private_peer_ids",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PrivatePeerIDs, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"private peer IDs, raw",
			"p2p.private_peer_ids",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.P2P.PrivatePeerIDs, escapeNewline(value))
			},
			true,
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_RPC(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"listen address",
			"rpc.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.ListenAddress, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"listen address, raw",
			"rpc.laddr",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.RPC.ListenAddress, escapeNewline(value))
			},
			true,
		},
	}

	verifyGetTestTableCommon(t, testTable)
}

func TestConfig_Get_Mempool(t *testing.T) {
	t.Parallel()

	testTable := []testGetCase{
		{
			"root dir",
			"mempool.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.RootDir, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"root dir, raw",
			"mempool.home",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.RootDir, escapeNewline(value))
			},
			true,
		},
		{
			"recheck flag",
			"mempool.recheck",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Recheck, unmarshalJSONCommon[bool](t, value))
			},
			false,
		},
		{
			"broadcast flag",
			"mempool.broadcast",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Broadcast, unmarshalJSONCommon[bool](t, value))
			},
			false,
		},
		{
			"WAL path",
			"mempool.wal_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.WalPath, unmarshalJSONCommon[string](t, value))
			},
			false,
		},
		{
			"WAL path, raw",
			"mempool.wal_dir",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.WalPath, escapeNewline(value))
			},
			true,
		},
		{
			"size",
			"mempool.size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.Size, unmarshalJSONCommon[int](t, value))
			},
			false,
		},
		{
			"max pending txs bytes",
			"mempool.max_pending_txs_bytes",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.MaxPendingTxsBytes, unmarshalJSONCommon[int64](t, value))
			},
			false,
		},
		{
			"cache size",
			"mempool.cache_size",
			func(loadedCfg *config.Config, value []byte) {
				assert.Equal(t, loadedCfg.Mempool.CacheSize, unmarshalJSONCommon[int](t, value))
			},
			false,
		},
	}

	verifyGetTestTableCommon(t, testTable)
}
