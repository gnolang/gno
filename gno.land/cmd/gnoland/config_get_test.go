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
			"RootDir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.RootDir, value)
			},
		},
		{
			"proxy app fetched",
			"ProxyApp",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ProxyApp, value)
			},
		},
		{
			"moniker fetched",
			"Moniker",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.Moniker, value)
			},
		},
		{
			"fast sync mode fetched",
			"FastSyncMode",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, loadedCfg.FastSyncMode, boolVal)
			},
		},
		{
			"db backend fetched",
			"DBBackend",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.DBBackend, value)
			},
		},
		{
			"db path fetched",
			"DBPath",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.DBPath, value)
			},
		},
		{
			"genesis path fetched",
			"Genesis",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.Genesis, value)
			},
		},
		{
			"validator key fetched",
			"PrivValidatorKey",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorKey, value)
			},
		},
		{
			"validator state file fetched",
			"PrivValidatorState",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorState, value)
			},
		},
		{
			"validator listen addr fetched",
			"PrivValidatorListenAddr",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.PrivValidatorListenAddr, value)
			},
		},
		{
			"node key path fetched",
			"NodeKey",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.NodeKey, value)
			},
		},
		{
			"abci fetched",
			"ABCI",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ABCI, value)
			},
		},
		{
			"profiling listen address fetched",
			"ProfListenAddress",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, loadedCfg.ProfListenAddress, value)
			},
		},
		{
			"filter peers flag fetched",
			"FilterPeers",
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
			"Consensus.RootDir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.RootDir)
			},
		},
		{
			"WAL path updated",
			"Consensus.WALPath",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.WALPath)
			},
		},
		{
			"propose timeout updated",
			"Consensus.TimeoutPropose",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPropose.String())
			},
		},
		{
			"propose timeout delta updated",
			"Consensus.TimeoutProposeDelta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutProposeDelta.String())
			},
		},
		{
			"prevote timeout updated",
			"Consensus.TimeoutPrevote",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevote.String())
			},
		},
		{
			"prevote timeout delta updated",
			"Consensus.TimeoutPrevoteDelta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevoteDelta.String())
			},
		},
		{
			"precommit timeout updated",
			"Consensus.TimeoutPrecommit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommit.String())
			},
		},
		{
			"precommit timeout delta updated",
			"Consensus.TimeoutPrecommitDelta",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommitDelta.String())
			},
		},
		{
			"commit timeout updated",
			"Consensus.TimeoutCommit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutCommit.String())
			},
		},
		{
			"skip commit timeout toggle updated",
			"Consensus.SkipTimeoutCommit",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.Consensus.SkipTimeoutCommit)
			},
		},
		{
			"create empty blocks toggle updated",
			"Consensus.CreateEmptyBlocks",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)
				assert.Equal(t, boolVal, loadedCfg.Consensus.CreateEmptyBlocks)
			},
		},
		{
			"create empty blocks interval updated",
			"Consensus.CreateEmptyBlocksInterval",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.CreateEmptyBlocksInterval.String())
			},
		},
		{
			"peer gossip sleep duration updated",
			"Consensus.PeerGossipSleepDuration",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerGossipSleepDuration.String())
			},
		},
		{
			"peer query majority sleep duration updated",
			"Consensus.PeerQueryMaj23SleepDuration",
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
			"TxEventStore.EventStoreType",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.TxEventStore.EventStoreType)
			},
		},
		{
			"event store params updated",
			"TxEventStore.Params",
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
			"P2P.RootDir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.RootDir)
			},
		},
		{
			"listen address updated",
			"P2P.ListenAddress",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ListenAddress)
			},
		},
		{
			"external address updated",
			"P2P.ExternalAddress",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ExternalAddress)
			},
		},
		{
			"seeds updated",
			"P2P.Seeds",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.Seeds)
			},
		},
		{
			"persistent peers updated",
			"P2P.PersistentPeers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PersistentPeers)
			},
		},
		{
			"upnp toggle updated",
			"P2P.UPNP",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.UPNP)
			},
		},
		{
			"max inbound peers updated",
			"P2P.MaxNumInboundPeers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumInboundPeers))
			},
		},
		{
			"max outbound peers updated",
			"P2P.MaxNumOutboundPeers",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumOutboundPeers))
			},
		},
		{
			"flush throttle timeout updated",
			"P2P.FlushThrottleTimeout",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.FlushThrottleTimeout.String())
			},
		},
		{
			"max package payload size updated",
			"P2P.MaxPacketMsgPayloadSize",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxPacketMsgPayloadSize))
			},
		},
		{
			"send rate updated",
			"P2P.SendRate",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.SendRate))
			},
		},
		{
			"receive rate updated",
			"P2P.RecvRate",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.RecvRate))
			},
		},
		{
			"pex reactor toggle updated",
			"P2P.PexReactor",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.PexReactor)
			},
		},
		{
			"seed mode updated",
			"P2P.SeedMode",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.SeedMode)
			},
		},
		{
			"private peer IDs updated",
			"P2P.PrivatePeerIDs",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PrivatePeerIDs)
			},
		},
		{
			"allow duplicate IPs updated",
			"P2P.AllowDuplicateIP",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.P2P.AllowDuplicateIP)
			},
		},
		{
			"handshake timeout updated",
			"P2P.HandshakeTimeout",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.HandshakeTimeout.String())
			},
		},
		{
			"dial timeout updated",
			"P2P.DialTimeout",
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
			"RPC.RootDir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.RootDir)
			},
		},
		{
			"listen address updated",
			"RPC.ListenAddress",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.ListenAddress)
			},
		},
		{
			"CORS Allowed Origins updated",
			"RPC.CORSAllowedOrigins",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedOrigins))
			},
		},
		{
			"CORS Allowed Methods updated",
			"RPC.CORSAllowedMethods",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedMethods))
			},
		},
		{
			"CORS Allowed Headers updated",
			"RPC.CORSAllowedHeaders",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%v", loadedCfg.RPC.CORSAllowedHeaders))
			},
		},
		{
			"GRPC listen address updated",
			"RPC.GRPCListenAddress",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.GRPCListenAddress)
			},
		},
		{
			"GRPC max open connections updated",
			"RPC.GRPCMaxOpenConnections",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.GRPCMaxOpenConnections))
			},
		},
		{
			"unsafe value updated",
			"RPC.Unsafe",
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.RPC.Unsafe)
			},
		},
		{
			"RPC max open connections updated",
			"RPC.MaxOpenConnections",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxOpenConnections))
			},
		},
		{
			"tx commit broadcast timeout updated",
			"RPC.TimeoutBroadcastTxCommit",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TimeoutBroadcastTxCommit.String())
			},
		},
		{
			"max body bytes updated",
			"RPC.MaxBodyBytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxBodyBytes))
			},
		},
		{
			"max header bytes updated",
			"RPC.MaxHeaderBytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxHeaderBytes))
			},
		},
		{
			"TLS cert file updated",
			"RPC.TLSCertFile",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSCertFile)
			},
		},
		{
			"TLS key file updated",
			"RPC.TLSKeyFile",
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
			"Mempool.RootDir",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.RootDir)
			},
		},
		{
			"recheck flag updated",
			"Mempool.Recheck",
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Recheck)
			},
		},
		{
			"broadcast flag updated",
			"Mempool.Broadcast",
			func(loadedCfg *config.Config, value string) {
				boolVar, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVar, loadedCfg.Mempool.Broadcast)
			},
		},
		{
			"WAL path updated",
			"Mempool.WalPath",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.WalPath)
			},
		},
		{
			"size updated",
			"Mempool.Size",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.Size))
			},
		},
		{
			"max pending txs bytes updated",
			"Mempool.MaxPendingTxsBytes",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.MaxPendingTxsBytes))
			},
		},
		{
			"cache size updated",
			"Mempool.CacheSize",
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.CacheSize))
			},
		},
	}

	verifyGetTestTableCommon(t, testTable)
}
