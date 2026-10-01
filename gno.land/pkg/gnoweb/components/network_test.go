package components

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkKindValid(t *testing.T) {
	t.Parallel()

	assert.True(t, NetworkMainnet.Valid())
	assert.True(t, NetworkTestnet.Valid())
	assert.True(t, NetworkLocal.Valid())
	assert.False(t, NetworkKind("").Valid())
	assert.False(t, NetworkKind("Mainnet").Valid(), "the check is case-sensitive")
	assert.False(t, NetworkKind("prod").Valid())
}

func TestHeaderNetworkChipText(t *testing.T) {
	t.Parallel()

	testnet := HeaderData{ChainId: "pearl-1", NetworkKind: NetworkTestnet}
	assert.Contains(t, testnet.NetworkChipTitle(), "not mainnet")
	assert.Contains(t, testnet.NetworkChipTitle(), "pearl-1")
}
