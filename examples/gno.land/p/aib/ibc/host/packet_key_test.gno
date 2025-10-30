package host_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/aib/ibc/host"
	"gno.land/p/nt/urequire"
)

// TestPacketCommitmentKey is primarily used to document the expected key output
// so that other implementations (such as the IBC Solidity) can replicate the
// same key output. But it is also useful to catch any changes in the keys.
func TestPacketCommitmentKey(t *testing.T) {
	actual := hex.EncodeToString(host.PacketCommitmentKey("channel-0", 1))
	urequire.Equal(t, "6368616e6e656c2d30010000000000000001", actual)
}

// TestPacketReceiptKey is primarily used to document the expected key output
// so that other implementations (such as the IBC Solidity) can replicate the
// same key output. But it is also useful to catch any changes in the keys.
func TestPacketReceiptKey(t *testing.T) {
	actual := hex.EncodeToString(host.PacketReceiptKey("channel-0", 1))
	urequire.Equal(t, "6368616e6e656c2d30020000000000000001", actual)
}

// TestPacketAcknowledgementKey is primarily used to document the expected key output
// so that other implementations (such as the IBC Solidity) can replicate the
// same key output. But it is also useful to catch any changes in the keys.
func TestPacketAcknowledgementKey(t *testing.T) {
	actual := hex.EncodeToString(host.PacketAcknowledgementKey("channel-0", 1))
	urequire.Equal(t, "6368616e6e656c2d30030000000000000001", actual)
}
