package types

import (
	"net"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeInfo_Validate(t *testing.T) {
	t.Parallel()

	generateNetAddress := func() *NetAddress {
		var (
			key     = GenerateNodeKey()
			address = "127.0.0.1:8080"
		)

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		require.NoError(t, err)

		addr, err := NewNetAddress(key.ID(), tcpAddr)
		require.NoError(t, err)

		return addr
	}

	t.Run("invalid peer ID", func(t *testing.T) {
		t.Parallel()

		info := &NodeInfo{
			NetAddress: &NetAddress{
				ID: "", // zero
			},
		}

		assert.ErrorIs(t, info.Validate(), crypto.ErrZeroID)
	})

	t.Run("invalid version", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name    string
			version string
		}{
			{
				"non-ascii version",
				"¢§µ",
			},
			{
				"empty tab version",
				"\t",
			},
			{
				"empty space version",
				"  ",
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					NetAddress: generateNetAddress(),
					Version:    testCase.version,
				}

				assert.ErrorIs(t, info.Validate(), ErrInvalidVersion)
			})
		}
	})

	t.Run("invalid moniker", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name    string
			moniker string
		}{
			{
				"empty moniker",
				"",
			},
			{
				"non-ascii moniker",
				"¢§µ",
			},
			{
				"empty tab moniker",
				"\t",
			},
			{
				"empty space moniker",
				"  ",
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					NetAddress: generateNetAddress(),
					Moniker:    testCase.moniker,
				}

				assert.ErrorIs(t, info.Validate(), ErrInvalidMoniker)
			})
		}
	})

	t.Run("invalid RPC Address", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name       string
			rpcAddress string
		}{
			{
				"non-ascii moniker",
				"¢§µ",
			},
			{
				"empty tab RPC address",
				"\t",
			},
			{
				"empty space RPC address",
				"  ",
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					NetAddress: generateNetAddress(),
					Moniker:    "valid moniker",
					Other: NodeInfoOther{
						RPCAddress: testCase.rpcAddress,
					},
				}

				assert.ErrorIs(t, info.Validate(), ErrInvalidRPCAddress)
			})
		}
	})

	t.Run("invalid channels", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name        string
			channels    []byte
			expectedErr error
		}{
			{
				"too many channels",
				make([]byte, maxNumChannels+1),
				ErrExcessiveChannels,
			},
			{
				"duplicate channels",
				[]byte{
					byte(10),
					byte(20),
					byte(10),
				},
				ErrDuplicateChannels,
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					NetAddress: generateNetAddress(),
					Moniker:    "valid moniker",
					Channels:   testCase.channels,
				}

				assert.ErrorIs(t, info.Validate(), testCase.expectedErr)
			})
		}
	})

	t.Run("valid node info", func(t *testing.T) {
		t.Parallel()

		info := &NodeInfo{
			NetAddress: generateNetAddress(),
			Moniker:    "valid moniker",
			Channels:   []byte{10, 20, 30},
			Other: NodeInfoOther{
				RPCAddress: "0.0.0.0:26657",
			},
		}

		assert.NoError(t, info.Validate())
	})
}

func TestNodeInfo_CompatibleWith(t *testing.T) {
	t.Parallel()

	t.Run("incompatible version sets", func(t *testing.T) {
		t.Parallel()

		var (
			name = "Block"

			infoOne = &NodeInfo{
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: "badversion",
					},
				},
			}

			infoTwo = &NodeInfo{
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: "v0.0.0",
					},
				},
			}
		)

		assert.Error(t, infoTwo.CompatibleWith(*infoOne))
	})

	t.Run("incompatible networks", func(t *testing.T) {
		t.Parallel()

		var (
			name    = "Block"
			version = "v0.0.0"

			infoOne = &NodeInfo{
				Network: "+wrong",
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
			}

			infoTwo = &NodeInfo{
				Network: "gno",
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
			}
		)

		assert.ErrorIs(t, infoTwo.CompatibleWith(*infoOne), ErrIncompatibleNetworks)
	})

	t.Run("no common channels", func(t *testing.T) {
		t.Parallel()

		var (
			name    = "Block"
			version = "v0.0.0"
			network = "gno"

			infoOne = &NodeInfo{
				Network: network,
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
				Channels: []byte{10},
			}

			infoTwo = &NodeInfo{
				Network: network,
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
				Channels: []byte{20},
			}
		)

		assert.ErrorIs(t, infoTwo.CompatibleWith(*infoOne), ErrNoCommonChannels)
	})

	t.Run("fully compatible node infos", func(t *testing.T) {
		t.Parallel()

		var (
			name     = "Block"
			version  = "v0.0.0"
			network  = "gno"
			channels = []byte{10, 20, 30}

			infoOne = &NodeInfo{
				Network: network,
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
				Channels: channels,
			}

			infoTwo = &NodeInfo{
				Network: network,
				VersionSet: []versionset.VersionInfo{
					{
						Name:    name,
						Version: version,
					},
				},
				Channels: channels[1:],
			}
		)

		assert.NoError(t, infoTwo.CompatibleWith(*infoOne))
	})
}
