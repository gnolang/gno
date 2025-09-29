package types_test

import (
	"encoding/hex"
	"testing"

	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/urequire"
)

// TestCommitPacket is primarily used to document the expected commitment output
// so that other implementations (such as the IBC Solidity) can replicate the
// same commitment output. But it is also useful to catch any changes in the commitment.
func TestCommitPacket(t *testing.T) {
	var packet types.Packet
	testCases := []struct {
		name         string
		malleate     func()
		expectedHash string
	}{
		{
			"json packet",
			func() {}, // default is json packet
			"a096722aa6534040a0efbdae05765132a7b223ad306d6512f3734821bd046505",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transferData := `{"denom":"uatom","amount":"1000000","sender":"sender","receiver":"receiver","memo":"memo"}`
			packet = types.Packet{
				Sequence:          1,
				SourceClient:      "07-tendermint-0",
				DestinationClient: "07-tendermint-1",
				TimeoutTimestamp:  100,
				Payloads: []types.Payload{
					{
						SourcePort:      "transfer",
						DestinationPort: "transfer",
						Version:         "ics20-1",
						Encoding:        "application/json",
						Value:           []byte(transferData),
					},
				},
			}
			tc.malleate()

			commitment := types.CommitPacket(packet)

			urequire.Equal(t, tc.expectedHash, hex.EncodeToString(commitment))
			urequire.Equal(t, 32, len(commitment))
		})
	}
}
