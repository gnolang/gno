package consensus

// Internal test package: most reactor message types are unexported.

import (
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

func TestCodecParity_Consensus(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(ctypes.Package)
	cdc.RegisterPackage(Package)
	cdc.Seal()

	_ = time.Second // retain time import for timeoutInfo

	cases := []struct {
		name string
		v    any
	}{
		{"NewRoundStepMessage", &NewRoundStepMessage{
			Height: 100, Round: 0, Step: ctypes.RoundStepPropose,
			SecondsSinceStartTime: 3, LastCommitRound: -1,
		}},
		{"HasVoteMessage", &HasVoteMessage{
			Height: 100, Round: 0, Type: btypes.PrecommitType, Index: 2,
		}},
		{"VoteSetMaj23Message", &VoteSetMaj23Message{
			Height: 100, Round: 0, Type: btypes.PrecommitType,
			BlockID: btypes.BlockID{Hash: []byte{0x01, 0x02}},
		}},
		{"timeoutInfo", &timeoutInfo{
			Duration: 2 * time.Second, Height: 50, Round: 1,
			Step: ctypes.RoundStepPrevote,
		}},
		{"newRoundStepInfo", &newRoundStepInfo{
			HRS: ctypes.HRS{Height: 10, Round: 0, Step: ctypes.RoundStepPrevote},
		}},
		// msgInfo.Msg is a ConsensusMessage interface; leave nil to keep
		// this test self-contained (other cases exercise concrete message
		// types directly).
		{"msgInfo/nil-msg", &msgInfo{PeerID: p2pTypes.ID("peer-1")}},
	}

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
