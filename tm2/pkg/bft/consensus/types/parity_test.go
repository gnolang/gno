package cstypes_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/aminotest"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	btypes "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func TestCodecParity_ConsensusTypes(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterPackage(btypes.Package)
	cdc.RegisterPackage(ctypes.Package)
	cdc.Seal()

	stamp := time.Date(2026, time.April, 24, 12, 0, 0, 0, time.UTC)
	addr := crypto.AddressFromPreimage([]byte("proposer"))

	cases := []struct {
		name string
		v    any
	}{
		{"HRS", &ctypes.HRS{Height: 100, Round: 2, Step: ctypes.RoundStepPrecommit}},
		{"EventNewRound", &ctypes.EventNewRound{
			HRS:           ctypes.HRS{Height: 7, Round: 0, Step: ctypes.RoundStepPropose},
			Proposer:      btypes.Validator{Address: addr, VotingPower: 1},
			ProposerIndex: 0,
		}},
		{"EventNewRoundStep", &ctypes.EventNewRoundStep{
			HRS: ctypes.HRS{Height: 7, Round: 1, Step: ctypes.RoundStepPrevote},
		}},
		{"EventCompleteProposal", &ctypes.EventCompleteProposal{
			HRS:     ctypes.HRS{Height: 9, Round: 0, Step: ctypes.RoundStepPropose},
			BlockID: btypes.BlockID{Hash: []byte{0x01}},
		}},
		{"EventTimeoutPropose", &ctypes.EventTimeoutPropose{
			HRS: ctypes.HRS{Height: 1, Round: 0, Step: ctypes.RoundStepPropose},
		}},
		{"EventTimeoutWait", &ctypes.EventTimeoutWait{
			HRS: ctypes.HRS{Height: 1, Round: 0, Step: ctypes.RoundStepPrevoteWait},
		}},
	}
	_ = stamp

	for i, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%d/%s", i, c.name), func(t *testing.T) {
			t.Parallel()
			aminotest.AssertCodecParity(t, cdc, c.v)
		})
	}
}
