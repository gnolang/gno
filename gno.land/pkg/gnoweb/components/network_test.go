package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkKindFromChainID(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		chainID string
		want    NetworkKind
	}{
		{"mainnet", "gnoland-1", NetworkMainnet},
		// The mainnet series increments on a chain restart.
		{"future mainnet", "gnoland-2", NetworkMainnet},
		{"future mainnet two digits", "gnoland-12", NetworkMainnet},
		// gnoland1 is the betanet, a different chain from gnoland-1.
		{"betanet is not mainnet", "gnoland1", NetworkTestnet},
		{"suffixed lookalike", "gnoland-1-fork", NetworkTestnet},
		{"zero is not a chain", "gnoland-0", NetworkTestnet},
		{"no number", "gnoland-", NetworkTestnet},
		{"pearl", "pearl-1", NetworkTestnet},
		{"staging", "staging", NetworkTestnet},
		{"cli default", "dev", NetworkTestnet},
		{"unknown", "some-new-chain", NetworkTestnet},
		{"empty", "", NetworkTestnet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, NetworkKindFromChainID(tc.chainID))
		})
	}
}

func TestNetworkKindValid(t *testing.T) {
	t.Parallel()

	assert.True(t, NetworkMainnet.Valid())
	assert.True(t, NetworkTestnet.Valid())
	assert.False(t, NetworkKind("").Valid())
	assert.False(t, NetworkKind("Mainnet").Valid(), "the check is case-sensitive")
	assert.False(t, NetworkKind("prod").Valid())
}

func TestHeaderNetworkChipText(t *testing.T) {
	t.Parallel()

	mainnet := HeaderData{ChainId: "gnoland-1", NetworkKind: NetworkMainnet}
	assert.Contains(t, mainnet.NetworkChipTitle(), "mainnet")
	assert.Contains(t, mainnet.NetworkChipTitle(), "gnoland-1")

	testnet := HeaderData{ChainId: "pearl-1", NetworkKind: NetworkTestnet}
	assert.Contains(t, testnet.NetworkChipTitle(), "not mainnet")
	assert.Contains(t, testnet.NetworkChipTitle(), "pearl-1")

	// The sr-only prefix must not repeat the chain-id: it is read immediately
	// before the chip's visible text.
	assert.NotContains(t, testnet.NetworkChipLabel(), "pearl-1")
	assert.Contains(t, testnet.NetworkChipLabel(), "not mainnet")
	assert.NotContains(t, mainnet.NetworkChipLabel(), "gnoland-1")
}
