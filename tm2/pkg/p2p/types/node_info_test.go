package types

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/versionset"
	"github.com/stretchr/testify/assert"
)

func TestNodeInfo_Validate(t *testing.T) {
	t.Parallel()

	t.Run("invalid peer ID", func(t *testing.T) {
		t.Parallel()

		info := &NodeInfo{
			PeerID: "", // zero
		}

		assert.ErrorIs(t, info.Validate(), ErrInvalidPeerID)
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
				fmt.Sprintf("\t"),
			},
			{
				"empty space version",
				fmt.Sprintf("  "),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					PeerID:  GenerateNodeKey().ID(),
					Version: testCase.version,
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
				fmt.Sprintf("\t"),
			},
			{
				"empty space moniker",
				fmt.Sprintf("  "),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					PeerID:  GenerateNodeKey().ID(),
					Moniker: testCase.moniker,
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
				fmt.Sprintf("\t"),
			},
			{
				"empty space RPC address",
				fmt.Sprintf("  "),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				info := &NodeInfo{
					PeerID:  GenerateNodeKey().ID(),
					Moniker: "valid moniker",
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
					PeerID:   GenerateNodeKey().ID(),
					Moniker:  "valid moniker",
					Channels: testCase.channels,
				}

				assert.ErrorIs(t, info.Validate(), testCase.expectedErr)
			})
		}
	})

	t.Run("valid node info", func(t *testing.T) {
		t.Parallel()

		info := &NodeInfo{
			PeerID:   GenerateNodeKey().ID(),
			Moniker:  "valid moniker",
			Channels: []byte{10, 20, 30},
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
