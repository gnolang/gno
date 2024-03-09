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
			"DBBackend",
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
				"RootDir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RootDir)
			},
		},
		{
			"proxy app updated",
			[]string{
				"ProxyApp",
				"example proxy app",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProxyApp)
			},
		},
		{
			"moniker updated",
			[]string{
				"Moniker",
				"example moniker",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Moniker)
			},
		},
		{
			"fast sync mode updated",
			[]string{
				"FastSyncMode",
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
				"DBBackend",
				db.GoLevelDBBackend.String(),
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBBackend)
			},
		},
		{
			"db path updated",
			[]string{
				"DBPath",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.DBPath)
			},
		},
		{
			"genesis path updated",
			[]string{
				"Genesis",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Genesis)
			},
		},
		{
			"validator key updated",
			[]string{
				"PrivValidatorKey",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorKey)
			},
		},
		{
			"validator state file updated",
			[]string{
				"PrivValidatorState",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorState)
			},
		},
		{
			"validator listen addr updated",
			[]string{
				"PrivValidatorListenAddr",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.PrivValidatorListenAddr)
			},
		},
		{
			"node key path updated",
			[]string{
				"NodeKey",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.NodeKey)
			},
		},
		{
			"abci updated",
			[]string{
				"ABCI",
				config.LocalABCI,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ABCI)
			},
		},
		{
			"profiling listen address updated",
			[]string{
				"ProfListenAddress",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.ProfListenAddress)
			},
		},
		{
			"filter peers flag updated",
			[]string{
				"FilterPeers",
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
				"Consensus.RootDir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.RootDir)
			},
		},
		{
			"WAL path updated",
			[]string{
				"Consensus.WALPath",
				"example WAL path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.WALPath)
			},
		},
		{
			"propose timeout updated",
			[]string{
				"Consensus.TimeoutPropose",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPropose.String())
			},
		},
		{
			"propose timeout delta updated",
			[]string{
				"Consensus.TimeoutProposeDelta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutProposeDelta.String())
			},
		},
		{
			"prevote timeout updated",
			[]string{
				"Consensus.TimeoutPrevote",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevote.String())
			},
		},
		{
			"prevote timeout delta updated",
			[]string{
				"Consensus.TimeoutPrevoteDelta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrevoteDelta.String())
			},
		},
		{
			"precommit timeout updated",
			[]string{
				"Consensus.TimeoutPrecommit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommit.String())
			},
		},
		{
			"precommit timeout delta updated",
			[]string{
				"Consensus.TimeoutPrecommitDelta",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutPrecommitDelta.String())
			},
		},
		{
			"commit timeout updated",
			[]string{
				"Consensus.TimeoutCommit",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.TimeoutCommit.String())
			},
		},
		{
			"skip commit timeout toggle updated",
			[]string{
				"Consensus.SkipTimeoutCommit",
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
				"Consensus.CreateEmptyBlocks",
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
				"Consensus.CreateEmptyBlocksInterval",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.CreateEmptyBlocksInterval.String())
			},
		},
		{
			"peer gossip sleep duration updated",
			[]string{
				"Consensus.PeerGossipSleepDuration",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Consensus.PeerGossipSleepDuration.String())
			},
		},
		{
			"peer query majority sleep duration updated",
			[]string{
				"Consensus.PeerQueryMaj23SleepDuration",
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
				"TxEventStore.EventStoreType",
				file.EventStoreType,
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.TxEventStore.EventStoreType)
			},
		},
		{
			"event store params updated",
			[]string{
				"TxEventStore.Params",
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
				"P2P.RootDir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"P2P.ListenAddress",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ListenAddress)
			},
		},
		{
			"external address updated",
			[]string{
				"P2P.ExternalAddress",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.ExternalAddress)
			},
		},
		{
			"seeds updated",
			[]string{
				"P2P.Seeds",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.Seeds)
			},
		},
		{
			"persistent peers updated",
			[]string{
				"P2P.PersistentPeers",
				"nodeID@0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PersistentPeers)
			},
		},
		{
			"upnp toggle updated",
			[]string{
				"P2P.UPNP",
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
				"P2P.MaxNumInboundPeers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumInboundPeers))
			},
		},
		{
			"max outbound peers updated",
			[]string{
				"P2P.MaxNumOutboundPeers",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxNumOutboundPeers))
			},
		},
		{
			"flush throttle timeout updated",
			[]string{
				"P2P.FlushThrottleTimeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.FlushThrottleTimeout.String())
			},
		},
		{
			"max package payload size updated",
			[]string{
				"P2P.MaxPacketMsgPayloadSize",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.MaxPacketMsgPayloadSize))
			},
		},
		{
			"send rate updated",
			[]string{
				"P2P.SendRate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.SendRate))
			},
		},
		{
			"receive rate updated",
			[]string{
				"P2P.RecvRate",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.P2P.RecvRate))
			},
		},
		{
			"pex reactor toggle updated",
			[]string{
				"P2P.PexReactor",
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
				"P2P.SeedMode",
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
				"P2P.PrivatePeerIDs",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.PrivatePeerIDs)
			},
		},
		{
			"allow duplicate IPs updated",
			[]string{
				"P2P.AllowDuplicateIP",
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
				"P2P.HandshakeTimeout",
				"1s",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.P2P.HandshakeTimeout.String())
			},
		},
		{
			"dial timeout updated",
			[]string{
				"P2P.DialTimeout",
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
				"RPC.RootDir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.RootDir)
			},
		},
		{
			"listen address updated",
			[]string{
				"RPC.ListenAddress",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.ListenAddress)
			},
		},
		{
			"CORS Allowed Origins updated",
			[]string{
				"RPC.CORSAllowedOrigins",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedOrigins)
			},
		},
		{
			"CORS Allowed Methods updated",
			[]string{
				"RPC.CORSAllowedMethods",
				"POST,GET",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedMethods)
			},
		},
		{
			"CORS Allowed Headers updated",
			[]string{
				"RPC.CORSAllowedHeaders",
				"*",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, strings.SplitN(value, ",", -1), loadedCfg.RPC.CORSAllowedHeaders)
			},
		},
		{
			"GRPC listen address updated",
			[]string{
				"RPC.GRPCListenAddress",
				"0.0.0.0:0",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.GRPCListenAddress)
			},
		},
		{
			"GRPC max open connections updated",
			[]string{
				"RPC.GRPCMaxOpenConnections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.GRPCMaxOpenConnections))
			},
		},
		{
			"unsafe value updated",
			[]string{
				"RPC.Unsafe",
				"true",
			},
			func(loadedCfg *config.Config, value string) {
				boolVal, err := strconv.ParseBool(value)
				require.NoError(t, err)

				assert.Equal(t, boolVal, loadedCfg.RPC.Unsafe)
			},
		},
		{
			"RPC max open connections updated",
			[]string{
				"RPC.MaxOpenConnections",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxOpenConnections))
			},
		},
		{
			"tx commit broadcast timeout updated",
			[]string{
				"RPC.TimeoutBroadcastTxCommit",
				(time.Second * 10).String(),
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TimeoutBroadcastTxCommit.String())
			},
		},
		{
			"max body bytes updated",
			[]string{
				"RPC.MaxBodyBytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxBodyBytes))
			},
		},
		{
			"max header bytes updated",
			[]string{
				"RPC.MaxHeaderBytes",
				"10",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.RPC.MaxHeaderBytes))
			},
		},
		{
			"TLS cert file updated",
			[]string{
				"RPC.TLSCertFile",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.RPC.TLSCertFile)
			},
		},
		{
			"TLS key file updated",
			[]string{
				"RPC.TLSKeyFile",
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
				"Mempool.RootDir",
				"example root dir",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.RootDir)
			},
		},
		{
			"recheck flag updated",
			[]string{
				"Mempool.Recheck",
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
				"Mempool.Broadcast",
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
				"Mempool.WalPath",
				"example path",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, loadedCfg.Mempool.WalPath)
			},
		},
		{
			"size updated",
			[]string{
				"Mempool.Size",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.Size))
			},
		},
		{
			"max pending txs bytes updated",
			[]string{
				"Mempool.MaxPendingTxsBytes",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.MaxPendingTxsBytes))
			},
		},
		{
			"cache size updated",
			[]string{
				"Mempool.CacheSize",
				"100",
			},
			func(loadedCfg *config.Config, value string) {
				assert.Equal(t, value, fmt.Sprintf("%d", loadedCfg.Mempool.CacheSize))
			},
		},
	}

	verifySetTestTableCommon(t, testTable)
}
