package consensus

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	walm "github.com/gnolang/gno/tm2/pkg/bft/wal"
	"github.com/gnolang/gno/tm2/pkg/events"
)

// A valid block in non-canonical parts must get a nil prevote. Fast sync
// recomputes the header, so committing it would strand syncing nodes.
func TestStateProposalNonCanonicalPartSet(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		parts func(*types.Block) *types.PartSet
	}{
		{"padded", paddedPartSet},
		{"re-chunked", rechunkedPartSet},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(2)
			height, round := cs1.Height, cs1.Round
			vs2 := vss[1]

			proposalCh := subscribe(cs1.evsw, cstypes.EventCompleteProposal{})
			voteCh := subscribe(cs1.evsw, types.EventVote{})

			propBlock, _ := cs1.createProposalBlock()

			// make the second validator the proposer by incrementing round
			round++
			incrementRound(vss[1:]...)

			// Padding keeps Total and re-chunking keeps ByteSize, so the check under
			// test cannot be passing on a count.
			otherParts := tc.parts(propBlock)
			canonical := propBlock.MakePartSet(types.BlockPartSizeBytes)
			if otherParts.Header().Equals(canonical.Header()) {
				t.Fatal("the part set matches the canonical header; the test proves nothing")
			}
			if otherParts.Total() != canonical.Total() && otherParts.ByteSize() != canonical.ByteSize() {
				t.Fatal("the part set differs from the canonical one in both counts; the test proves nothing")
			}

			// The stream still decodes into the same block, which is what makes
			// the fabricated header reach a vote in the first place. Without this the
			// test could pass on a decode failure and never reach the check.
			var decoded types.Block
			if _, err := amino.UnmarshalSizedReader(otherParts.GetReader(), &decoded,
				cs1.state.ConsensusParams.Block.MaxDataBytes); err != nil {
				t.Fatal("the part set no longer decodes; the test proves nothing", err)
			}
			if !decoded.HashesTo(propBlock.Hash()) {
				t.Fatal("the part set decodes into a different block")
			}

			blockID := types.BlockID{Hash: propBlock.Hash(), PartsHeader: otherParts.Header()}
			proposal := types.NewProposal(vs2.Height, round, -1, blockID)
			if err := vs2.SignProposal(config.ChainID(), proposal); err != nil {
				t.Fatal("failed to sign the proposal", err)
			}

			// set the proposal block
			if err := cs1.SetProposalAndBlock(proposal, propBlock, otherParts, "some peer"); err != nil {
				t.Fatal(err)
			}

			// start the machine
			startFrom(cs1, height, round)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			defer ensureDrainedChannels(t, proposalCh, voteCh)

			// the node prevotes nil rather than signing the fabricated header
			ensurePrevote(voteCh, height, round)
			validatePrevote(cs1, round, vss[0], nil)

			if rs := cs1.GetRoundState(); rs.ProposalBlock != nil {
				t.Fatal("the node kept a block whose advertised chunking is not canonical")
			}

			// the second validator prevotes nil too, so the round concludes
			signAddVotes(cs1, types.PrevoteType, nil, types.PartSetHeader{}, vs2)
			ensurePrevote(voteCh, height, round)

			ensurePrecommit(voteCh, height, round)
			validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)

			// The complete-proposal event never fired: it sits past the check, and had
			// it fired on this unbuffered channel the prevote above would never have
			// been sent.
			ensureNoNewEvent(proposalCh, time.Second,
				"the node accepted a proposal whose PartSetHeader is not the canonical chunking of the block")
		})
	}
}

// A block whose parts carry a trailing field must get a nil prevote: the decode
// fails after filling the block, and committing it would store parts that
// LoadBlock cannot decode back.
func TestStateProposalUndecodablePartSet(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(2)
	height, round := cs1.Height, cs1.Round
	vs2 := vss[1]

	voteCh := subscribe(cs1.evsw, types.EventVote{})

	propBlock, _ := cs1.createProposalBlock()

	// make the second validator the proposer by incrementing round
	round++
	incrementRound(vss[1:]...)

	otherParts := trailingFieldPartSet(propBlock)

	blockID := types.BlockID{Hash: propBlock.Hash(), PartsHeader: otherParts.Header()}
	proposal := types.NewProposal(vs2.Height, round, -1, blockID)
	if err := vs2.SignProposal(config.ChainID(), proposal); err != nil {
		t.Fatal("failed to sign the proposal", err)
	}
	if err := cs1.SetProposalAndBlock(proposal, propBlock, otherParts, "some peer"); err != nil {
		t.Fatal(err)
	}

	startFrom(cs1, height, round)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()

	ensurePrevote(voteCh, height, round)
	validatePrevote(cs1, round, vss[0], nil)
	if rs := cs1.GetRoundState(); rs.ProposalBlock != nil {
		t.Fatal("the node kept a block whose parts failed to decode")
	}
}

// An honest block of several parts, every part delivered, must get a prevote:
// the check re-chunks at the part size the proposer chunked at.
func TestStateProposalOfSeveralParts(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(2)
	height, round := cs1.Height, cs1.Round
	vs2 := vss[1]

	voteCh := subscribe(cs1.evsw, types.EventVote{})

	// make the second validator the proposer by incrementing round
	round++
	incrementRound(vss[1:]...)

	// three 50 kB transactions, each within MaxTxBytes
	txs := []types.Tx{bytes.Repeat([]byte{0x01}, 50_000), bytes.Repeat([]byte{0x02}, 50_000), bytes.Repeat([]byte{0x03}, 50_000)}
	propBlock, propParts := cs1.state.MakeBlock(height, txs, types.NewCommit(types.BlockID{}, nil), vs2.PubKey().Address())
	if propParts.Total() < 2 {
		t.Fatal("the block fits one part; the test proves nothing")
	}

	proposal := types.NewProposal(height, round, -1, types.BlockID{Hash: propBlock.Hash(), PartsHeader: propParts.Header()})
	if err := vs2.SignProposal(config.ChainID(), proposal); err != nil {
		t.Fatal("failed to sign the proposal", err)
	}
	if err := cs1.SetProposalAndBlock(proposal, propBlock, propParts, "some peer"); err != nil {
		t.Fatal(err)
	}

	startFrom(cs1, height, round)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()

	ensurePrevote(voteCh, height, round)
	validatePrevote(cs1, round, vss[0], propBlock.Hash())
}

// A node that dropped a proposal block in the propose step must still commit
// it once +2/3 precommit it, as unpatched validators do, rather than wait at
// the height for good.
func TestStateCommitsDroppedNonCanonicalBlock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// height 2 gives the block a last commit whose hash can be left out
		height int64
		parts  func(*types.Block) *types.PartSet
	}{
		{"padded", 1, paddedPartSet},
		{"re-chunked", 1, rechunkedPartSet},
		{"trailing field", 1, trailingFieldPartSet},
		{"DataHash left out", 1, unfilledPartSet(func(b *types.Block) { b.DataHash = nil })},
		{"LastCommitHash left out", 2, unfilledPartSet(func(b *types.Block) { b.LastCommitHash = nil })},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]

			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			voteCh := subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())

			var height int64
			var round int
			var proposer *validatorStub
			var propBlock *types.Block
			if tc.height == 1 {
				// make the second validator the proposer by incrementing round
				height, round, proposer = cs1.Height, 1, vs2
				incrementRound(vss[1:]...)
				propBlock, _ = cs1.state.MakeBlock(height, []types.Tx{{0x00}}, types.NewCommit(types.BlockID{}, nil), cs1.privValidator.PubKey().Address())
			} else {
				startFrom(cs1, cs1.Height, 0)
				commitHeightOne(t, cs1, vss, newRoundCh, voteCh)
				height, round = 2, 0
				// the machine waits for a proposal and holds no lock
				cs1.mtx.Lock()
				for _, vs := range vss[1:] {
					if cs1.Validators.GetProposer().Address == vs.PubKey().Address() {
						proposer = vs
					}
				}
				propBlock, _ = cs1.state.MakeBlock(height, nil, cs1.LastCommit.MakeCommit(), cs1.privValidator.PubKey().Address())
				cs1.mtx.Unlock()
				if proposer == nil {
					t.Fatal("the node proposes at height 2 itself; the test proves nothing")
				}
			}
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()

			otherParts := tc.parts(propBlock)
			if otherParts.Header().Equals(propBlock.MakePartSet(types.BlockPartSizeBytes).Header()) {
				t.Fatal("the parts are the canonical chunking; the test proves nothing")
			}

			blockID := types.BlockID{Hash: propBlock.Hash(), PartsHeader: otherParts.Header()}
			proposal := types.NewProposal(height, round, -1, blockID)
			if err := proposer.SignProposal(config.ChainID(), proposal); err != nil {
				t.Fatal("failed to sign the proposal", err)
			}
			if err := cs1.SetProposalAndBlock(proposal, propBlock, otherParts, "some peer"); err != nil {
				t.Fatal(err)
			}

			if tc.height == 1 {
				startFrom(cs1, height, round)
				ensureNewRound(newRoundCh, height, round)
			}
			ensurePrevote(voteCh, height, round)
			validatePrevote(cs1, round, vss[0], nil)
			rs := cs1.GetRoundState()
			if rs.ProposalBlock != nil || !rs.ProposalBlockParts.IsComplete() {
				t.Fatal("the node should hold the complete parts and no block; the test proves nothing")
			}

			drainEvents(t, voteCh)
			// the other three precommit the header the node refused
			signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), otherParts.Header(), vs2, vs3, vs4)
			ensureNewRound(newRoundCh, height+1, 0)
			drainEvents(t, newRoundCh)
		})
	}
}

// A node that reaches the commit step before a dropped block's parts arrive
// must take the block +2/3 precommitted, rather than wait for good.
func TestStateCommitStepTakesNonCanonicalParts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		parts func(*types.Block) *types.PartSet
	}{
		{"padded", paddedPartSet},
		{"trailing field", trailingFieldPartSet},
		{"DataHash left out", unfilledPartSet(func(b *types.Block) { b.DataHash = nil })},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]
			height, round := cs1.Height, 1

			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			validBlockCh := subscribe(cs1.evsw, cstypes.EventNewValidBlock{})

			propBlock, _ := cs1.state.MakeBlock(height, []types.Tx{{0x00}}, types.NewCommit(types.BlockID{}, nil), vs2.PubKey().Address())
			otherParts := tc.parts(propBlock)

			startFrom(cs1, height, round)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()

			ensureNewRound(newRoundCh, height, round)
			signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), otherParts.Header(), vs2, vs3, vs4)
			ensureNewValidBlock(validBlockCh, height, round)
			if rs := cs1.GetRoundState(); rs.Step != cstypes.RoundStepCommit || rs.ProposalBlock != nil {
				t.Fatal("the node should be in the commit step holding no block; the test proves nothing")
			}

			// the parts arrive only now, in the commit step
			for i := 0; i < otherParts.Total(); i++ {
				cs1.AddProposalBlockPart(height, round, otherParts.GetPart(i), "some peer")
			}
			ensureNewRound(newRoundCh, height+1, 0)
			drainEvents(t, newRoundCh, validBlockCh)
		})
	}
}

// A block +2/3 committed whose parts fail to decode after filling it must not
// be kept when the state refuses it: the node waits rather than halting.
func TestStateCommitStepKeepsNoInvalidBlock(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(4)
	vs2, vs3, vs4 := vss[1], vss[2], vss[3]
	height, round := cs1.Height, 1

	newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
	validBlockCh := subscribe(cs1.evsw, cstypes.EventNewValidBlock{})

	propBlock, _ := cs1.state.MakeBlock(height, nil, types.NewCommit(types.BlockID{}, nil), vs2.PubKey().Address())
	propBlock.AppHash = []byte("other app hash")
	if cs1.state.ValidateBlock(propBlock) == nil {
		t.Fatal("the state accepts the block; the test proves nothing")
	}
	otherParts := trailingFieldPartSet(propBlock)

	startFrom(cs1, height, round)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()

	ensureNewRound(newRoundCh, height, round)
	signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), otherParts.Header(), vs2, vs3, vs4)
	ensureNewValidBlock(validBlockCh, height, round)
	drainEvents(t, validBlockCh)
	for i := 0; i < otherParts.Total(); i++ {
		cs1.AddProposalBlockPart(height, round, otherParts.GetPart(i), "some peer")
	}
	select {
	case <-cs1.done:
		t.Fatal("the consensus loop exited on the invalid block")
	case <-newRoundCh:
		t.Fatal("the node committed a block the state refuses")
	case <-time.After(time.Second):
	}
	drainEvents(t, newRoundCh)
}

// A node locked on a block's canonical parts must commit the same block once
// +2/3 precommit it under other parts, rather than halt.
func TestStateCommitsLockedBlockUnderOtherParts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// polka has the other three prevote the other parts before precommitting
		polka bool
	}{
		{"after a polka", true},
		{"precommits only", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]
			height, round := cs1.Height, cs1.Round

			timeoutWaitCh := subscribe(cs1.evsw, cstypes.EventTimeoutWait{})
			proposalCh := subscribe(cs1.evsw, cstypes.EventCompleteProposal{})
			voteCh := subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())
			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})

			startFrom(cs1, height, round)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()

			// the node proposes in round 0 and locks on its canonical parts
			ensureNewRound(newRoundCh, height, round)
			ensureNewProposal(proposalCh, height, round)
			ensurePrevote(voteCh, height, round)
			rs := cs1.GetRoundState()
			block, canonical := rs.ProposalBlock, rs.ProposalBlockParts.Header()
			signAddVotes(cs1, types.PrevoteType, block.Hash(), canonical, vs2, vs3, vs4)
			ensurePrecommit(voteCh, height, round)
			validatePrecommit(t, cs1, round, round, vss[0], block.Hash(), block.Hash())
			signAddVotes(cs1, types.PrecommitType, nil, types.PartSetHeader{}, vs2, vs3, vs4)
			ensureNewTimeout(timeoutWaitCh, height, round, cs1.config.Precommit(round).Nanoseconds())

			// round 1 re-proposes the same block, padded
			round++
			incrementRound(vs2, vs3, vs4)
			ensureNewRound(newRoundCh, height, round)
			drainEvents(t, timeoutWaitCh, proposalCh)

			padded := paddedPartSet(block)
			proposal := types.NewProposal(height, round, -1, types.BlockID{Hash: block.Hash(), PartsHeader: padded.Header()})
			if err := vs2.SignProposal(config.ChainID(), proposal); err != nil {
				t.Fatal("failed to sign the proposal", err)
			}
			if err := cs1.SetProposalAndBlock(proposal, block, padded, "some peer"); err != nil {
				t.Fatal(err)
			}
			ensurePrevote(voteCh, height, round)
			validatePrevote(cs1, round, vss[0], block.Hash())

			if tc.polka {
				// the other three prevote the padded header, which the node never signs
				signAddVotes(cs1, types.PrevoteType, block.Hash(), padded.Header(), vs2, vs3, vs4)
				ensurePrecommit(voteCh, height, round)
				validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)
			}
			drainEvents(t, voteCh)
			signAddVotes(cs1, types.PrecommitType, block.Hash(), padded.Header(), vs2, vs3, vs4)
			ensureNewRound(newRoundCh, height+1, 0)
			drainEvents(t, newRoundCh)
		})
	}
}

// A node holding a block's canonical parts must commit the same block once +2/3
// precommit it under other parts, rather than halt.
func TestStateCommitsHeldBlockUnderOtherParts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// polka has the other three prevote the other parts before precommitting
		polka bool
	}{
		{"after a polka", true},
		{"precommits only", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]
			height, round := cs1.Height, cs1.Round

			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			voteCh := subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())

			// make the second validator the proposer by incrementing round
			round++
			incrementRound(vss[1:]...)

			proposal, propBlock := decideProposal(cs1, vs2, height, round)
			if err := cs1.SetProposalAndBlock(proposal, propBlock, propBlock.MakePartSet(types.BlockPartSizeBytes), "some peer"); err != nil {
				t.Fatal(err)
			}

			startFrom(cs1, height, round)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()

			ensureNewRound(newRoundCh, height, round)
			ensurePrevote(voteCh, height, round)
			validatePrevote(cs1, round, vss[0], propBlock.Hash())

			padded := paddedPartSet(propBlock)
			if tc.polka {
				signAddVotes(cs1, types.PrevoteType, propBlock.Hash(), padded.Header(), vs2, vs3, vs4)
				ensurePrecommit(voteCh, height, round)
				validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)
			}
			drainEvents(t, voteCh)
			signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), padded.Header(), vs2, vs3, vs4)

			// the padded parts arrive only now
			for i := 0; i < padded.Total(); i++ {
				cs1.AddProposalBlockPart(height, round, padded.GetPart(i), "some peer")
			}
			ensureNewRound(newRoundCh, height+1, 0)
			drainEvents(t, newRoundCh)
		})
	}
}

// A node that meets a prevote majority for a block's padded parts before the
// canonical proposal of that block must not precommit the padded header.
func TestStatePolkaAheadOfProposalUnderOtherParts(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(4)
	vs2, vs3, vs4 := vss[1], vss[2], vss[3]
	height := cs1.Height

	newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
	voteCh := subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())

	round := 2
	proposal, propBlock := decideProposal(cs1, vs3, height, round)
	padded := paddedPartSet(propBlock)

	startFrom(cs1, height, 1)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()
	ensureNewRound(newRoundCh, height, 1)

	// +2/3 prevotes for round 2 arrive during round 1: the node skips ahead
	// without recording the padded parts as the valid block's.
	incrementRound(vss[1:]...)
	incrementRound(vss[1:]...)
	signAddVotes(cs1, types.PrevoteType, propBlock.Hash(), padded.Header(), vs2, vs3, vs4)
	ensureNewRound(newRoundCh, height, round)

	// the round's proposer then sends the same block canonically chunked
	if err := cs1.SetProposalAndBlock(proposal, propBlock, propBlock.MakePartSet(types.BlockPartSizeBytes), "some peer"); err != nil {
		t.Fatal(err)
	}
	ensurePrevote(voteCh, height, round)
	ensurePrecommit(voteCh, height, round)
	validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)
	drainEvents(t, voteCh)

	// the other three precommit the padded header, and its parts arrive
	signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), padded.Header(), vs2, vs3, vs4)
	for i := 0; i < padded.Total(); i++ {
		cs1.AddProposalBlockPart(height, round, padded.GetPart(i), "some peer")
	}
	ensureNewRound(newRoundCh, height+1, 0)
	drainEvents(t, newRoundCh)
}

// A node that meets a polka ahead of the proposal must announce the parts
// header it then waits on, unless no whole block fills that header.
func TestStatePolkaAheadOfProposalAnnouncesParts(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		// header makes the polka's parts header from the block and the most
		// parts a whole block fills
		header    func(b *types.Block, maxParts int) types.PartSetHeader
		installed bool
	}{
		{"padded parts", func(b *types.Block, _ int) types.PartSetHeader { return paddedPartSet(b).Header() }, true},
		{"a whole block's parts", func(b *types.Block, n int) types.PartSetHeader { return types.PartSetHeader{Total: n, Hash: b.Hash()} }, true},
		{"one part more", func(b *types.Block, n int) types.PartSetHeader {
			return types.PartSetHeader{Total: n + 1, Hash: b.Hash()}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]
			height, round := cs1.Height, 2

			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			validBlockCh := subscribe(cs1.evsw, cstypes.EventNewValidBlock{})
			voteCh := subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())

			proposal, propBlock := decideProposal(cs1, vs3, height, round)
			canonical := propBlock.MakePartSet(types.BlockPartSizeBytes)
			header := tc.header(propBlock, fullBlockParts(cs1))

			startFrom(cs1, height, 1)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			ensureNewRound(newRoundCh, height, 1)

			// +2/3 prevotes for round 2 move the node there before the proposal
			incrementRound(vss[1:]...)
			incrementRound(vss[1:]...)
			signAddVotes(cs1, types.PrevoteType, propBlock.Hash(), header, vs2, vs3, vs4)
			ensureNewRound(newRoundCh, height, round)
			if err := cs1.SetProposalAndBlock(proposal, propBlock, canonical, "some peer"); err != nil {
				t.Fatal(err)
			}
			// the swap fires its event inside the prevote's handling, before the
			// prevote's own event
			if tc.installed {
				select {
				case ev := <-validBlockCh:
					if got := ev.(cstypes.EventNewValidBlock).BlockPartsHeader; !got.Equals(header) {
						t.Fatalf("the node announced %v, not the polka's %v", got, header)
					}
				case <-time.After(ensureTimeout):
					t.Fatal("the node waits on the polka's parts but announced no header")
				}
			}
			ensurePrevote(voteCh, height, round)
			ensurePrecommit(voteCh, height, round)
			validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)

			want := header
			if !tc.installed {
				want = canonical.Header()
			}
			if rs := cs1.GetRoundState(); !rs.ProposalBlockParts.HasHeader(want) {
				t.Fatalf("the node holds parts %v, want %v", rs.ProposalBlockParts.Header(), want)
			}
			drainEvents(t, newRoundCh, validBlockCh, voteCh)
		})
	}
}

// A prevote majority in the node's round installs a part set for its header
// only when a whole block fills it; otherwise the node keeps what it holds.
func TestStatePolkaPartsCeiling(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		extra int
		// proposed delivers the round's proposal before the polka
		proposed bool
	}{
		{"a whole block's parts", 0, true},
		{"one part more", 1, true},
		{"one part more, no proposal", 1, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			vs2, vs3, vs4 := vss[1], vss[2], vss[3]
			height, round := cs1.Height, 1

			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			var validBlockCh, voteCh <-chan events.Event
			if tc.proposed {
				validBlockCh = subscribe(cs1.evsw, cstypes.EventNewValidBlock{})
			} else {
				voteCh = subscribeToVoter(cs1, cs1.privValidator.PubKey().Address())
			}

			incrementRound(vss[1:]...)
			proposal, propBlock := decideProposal(cs1, vs2, height, round)
			canonical := propBlock.MakePartSet(types.BlockPartSizeBytes)
			header := types.PartSetHeader{Total: fullBlockParts(cs1) + tc.extra, Hash: propBlock.Hash()}
			if tc.proposed {
				if err := cs1.SetProposalAndBlock(proposal, propBlock, canonical, "some peer"); err != nil {
					t.Fatal(err)
				}
			}

			startFrom(cs1, height, round)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			ensureNewRound(newRoundCh, height, round)
			signAddVotes(cs1, types.PrevoteType, propBlock.Hash(), header, vs2, vs3, vs4)

			if !tc.proposed {
				// the node holds no part set, and goes on to precommit nil
				ensurePrevote(voteCh, height, round)
				ensurePrecommit(voteCh, height, round)
				validatePrecommit(t, cs1, round, -1, vss[0], nil, nil)
				if rs := cs1.GetRoundState(); rs.ProposalBlockParts != nil {
					t.Fatalf("a polka for %d parts installed a part set", header.Total)
				}
				drainEvents(t, newRoundCh, voteCh)
				return
			}

			want := header
			if tc.extra > 0 {
				want = canonical.Header()
			}
			select {
			case ev := <-validBlockCh:
				if got := ev.(cstypes.EventNewValidBlock).BlockPartsHeader; !got.Equals(want) {
					t.Fatalf("a polka for %d parts: the node holds parts %v, want %v", header.Total, got, want)
				}
			case <-time.After(ensureTimeout):
				t.Fatal("the node announced no parts header after the polka")
			}
			drainEvents(t, newRoundCh, validBlockCh)
		})
	}
}

// A node must commit a block +2/3 precommitted under a header of more parts
// than a whole block fills, as unpatched validators do, rather than halt.
func TestStateCommitsOverCeilingHeader(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(4)
	vs2, vs3, vs4 := vss[1], vss[2], vss[3]
	height, round := cs1.Height, 1

	newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
	validBlockCh := subscribe(cs1.evsw, cstypes.EventNewValidBlock{})

	// the block's own encoding padded with zeros to one part past the ceiling
	propBlock, _ := cs1.state.MakeBlock(height, []types.Tx{{0x00}}, types.NewCommit(types.BlockID{}, nil), vs2.PubKey().Address())
	maxParts := fullBlockParts(cs1)
	over := types.NewPartSetFromData(append(amino.MustMarshalSized(propBlock), make([]byte, maxParts*types.BlockPartSizeBytes)...), types.BlockPartSizeBytes)
	if over.Total() <= maxParts {
		t.Fatal("the parts fit under the ceiling; the test proves nothing")
	}

	startFrom(cs1, height, round)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()
	ensureNewRound(newRoundCh, height, round)
	incrementRound(vss[1:]...)
	signAddVotes(cs1, types.PrecommitType, propBlock.Hash(), over.Header(), vs2, vs3, vs4)
	ensureNewValidBlock(validBlockCh, height, round)
	for i := 0; i < over.Total(); i++ {
		cs1.AddProposalBlockPart(height, round, over.GetPart(i), "some peer")
	}
	ensureNewRound(newRoundCh, height+1, 0)
	drainEvents(t, newRoundCh, validBlockCh)
}

// fullBlockParts returns the parts a block of exactly cs1's Block.MaxDataBytes
// chunks to.
func fullBlockParts(cs1 *ConsensusState) int {
	return types.NewPartSetFromData(make([]byte, cs1.state.ConsensusParams.Block.MaxDataBytes), types.BlockPartSizeBytes).Total()
}

// The part ceiling must admit the parts a block of exactly Block.MaxDataBytes
// chunks to, and refuse one more with no part set installed: a ceiling one
// short halts every full block.
func TestStateProposalPartsCeiling(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		extra int
		err   error
	}{
		{0, nil},
		{1, ErrInvalidProposalPartSetHeader},
	} {
		cs1, vss := randConsensusState(1)
		maxDataBytes := cs1.state.ConsensusParams.Block.MaxDataBytes
		if maxDataBytes%types.BlockPartSizeBytes == 0 {
			t.Fatal("MaxDataBytes is a whole number of parts; the test cannot tell a ceiling from a floor")
		}

		propBlock, _ := cs1.createProposalBlock()
		header := types.PartSetHeader{Total: fullBlockParts(cs1) + tc.extra, Hash: propBlock.Hash()}
		proposal := types.NewProposal(cs1.Height, cs1.Round, -1, types.BlockID{Hash: propBlock.Hash(), PartsHeader: header})
		if err := vss[0].SignProposal(config.ChainID(), proposal); err != nil {
			t.Fatal("failed to sign the proposal", err)
		}

		if err := cs1.defaultSetProposal(proposal); !errors.Is(err, tc.err) {
			t.Fatalf("%d parts for a %d-byte ceiling: got %v, want %v", header.Total, maxDataBytes, err, tc.err)
		}
		if tc.err != nil && (cs1.Proposal != nil || cs1.ProposalBlockParts != nil) {
			t.Fatalf("the refused proposal of %d parts was installed", header.Total)
		}
	}
}

// A validator an earlier binary locked on a block under non-canonical parts,
// restarted onto this one mid-height, replays its own precommit for a block the
// checks drop: it must come back locked on it, and prevote it in the next round
// rather than another proposal.
func TestStateReplayKeepsOwnPrecommitLock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name  string
		parts func(*types.Block) *types.PartSet
	}{
		{"padded", paddedPartSet},
		{"trailing field", trailingFieldPartSet},
		{"DataHash left out", unfilledPartSet(func(b *types.Block) { b.DataHash = nil })},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			caller, others := vss[0], vss[1:]
			height := cs1.Height
			blockID, msgs := proposedRoundOne(t, cs1, vss, tc.parts)
			msgs = append(msgs,
				msgInfo{&VoteMessage{signVote(caller, types.PrevoteType, blockID.Hash, blockID.PartsHeader)}, ""},
				msgInfo{&VoteMessage{signVote(caller, types.PrecommitType, blockID.Hash, blockID.PartsHeader)}, ""},
			)
			restartFromWAL(t, cs1, msgs)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			voteCh := subscribeToVoter(cs1, caller.PubKey().Address())

			rs := cs1.GetRoundState()
			if rs.LockedRound != 1 || !rs.LockedBlock.HashesTo(blockID.Hash) || !rs.LockedBlockParts.HasHeader(blockID.PartsHeader) {
				t.Fatalf("the node replayed its own precommit for the block and holds no lock on it: locked round %d", rs.LockedRound)
			}

			// the others precommit nil in round 1, and one of them proposes
			// another block in round 2
			signAddVotes(cs1, types.PrecommitType, nil, types.PartSetHeader{}, others...)
			ensureNewRound(newRoundCh, height, 2)
			incrementRound(others...)
			prevotesLockOverOtherProposal(t, cs1, caller, others, voteCh, blockID)
			drainEvents(t, newRoundCh, voteCh)
		})
	}
}

// An earlier binary takes up to MaxBlockPartsCount parts, so it can lock on a
// block cut into more parts than the ceiling admits. Restarted onto this one
// mid-height, the validator must come back locked on it too, whether the
// proposal or a polka ahead of it brought the parts.
func TestStateReplayKeepsOwnPrecommitLockPastCeiling(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		polkaFirst bool
	}{
		{"proposal", false},
		{"polka ahead of any proposal", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			caller, others := vss[0], vss[1:]
			height := cs1.Height
			ceiling := fullBlockParts(cs1)
			// the block's own bytes, cut small enough to need parts past the ceiling
			pastCeiling := func(block *types.Block) *types.PartSet {
				bz := amino.MustMarshalSized(block)
				return types.NewPartSetFromData(bz, len(bz)/(ceiling+1))
			}
			blockID, msgs := proposedRoundOne(t, cs1, vss, pastCeiling)
			if blockID.PartsHeader.Total <= ceiling {
				t.Fatal("the parts fit under the ceiling; the test proves nothing")
			}
			own := func(type_ types.SignedMsgType) msgInfo {
				return msgInfo{&VoteMessage{signVote(caller, type_, blockID.Hash, blockID.PartsHeader)}, ""}
			}

			var wal []walm.WALMessage
			if tc.polkaFirst {
				// msgs[0] is the proposal, which never arrives: the prevotes come
				// first, and the timeouts move the earlier binary to its votes
				parts, prevotes := msgs[1:len(msgs)-len(others)], msgs[len(msgs)-len(others):]
				for _, mi := range prevotes {
					wal = append(wal, mi)
				}
				for _, mi := range parts {
					wal = append(wal, mi)
				}
				wal = append(wal,
					timeoutInfo{Height: height, Round: 1, Step: cstypes.RoundStepPropose},
					own(types.PrevoteType),
					timeoutInfo{Height: height, Round: 1, Step: cstypes.RoundStepPrevoteWait},
					own(types.PrecommitType),
				)
			} else {
				for _, mi := range msgs {
					wal = append(wal, mi)
				}
				wal = append(wal, own(types.PrevoteType), own(types.PrecommitType))
			}
			restartFromWAL(t, cs1, wal)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			voteCh := subscribeToVoter(cs1, caller.PubKey().Address())

			rs := cs1.GetRoundState()
			if rs.LockedRound != 1 || !rs.LockedBlock.HashesTo(blockID.Hash) || !rs.LockedBlockParts.HasHeader(blockID.PartsHeader) {
				t.Fatalf("the node replayed its own precommit for the block and holds no lock on it: locked round %d", rs.LockedRound)
			}

			signAddVotes(cs1, types.PrecommitType, nil, types.PartSetHeader{}, others...)
			ensureNewRound(newRoundCh, height, 2)
			incrementRound(others...)
			prevotesLockOverOtherProposal(t, cs1, caller, others, voteCh, blockID)
			drainEvents(t, newRoundCh, voteCh)
		})
	}
}

// A lock restored from the replay must survive a polka for its block that
// arrives ahead of the round's proposal, with the part set still empty, and
// commit the block once the others precommit it, dropping the refused parts
// with the height.
func TestStateReplayedLockRelocksOnPolkaAheadOfProposal(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(4)
	caller, others := vss[0], vss[1:]
	height := cs1.Height
	blockID, msgs := proposedRoundOne(t, cs1, vss, paddedPartSet)
	msgs = append(msgs,
		msgInfo{&VoteMessage{signVote(caller, types.PrevoteType, blockID.Hash, blockID.PartsHeader)}, ""},
		msgInfo{&VoteMessage{signVote(caller, types.PrecommitType, blockID.Hash, blockID.PartsHeader)}, ""},
	)
	restartFromWAL(t, cs1, msgs)
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()
	newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
	voteCh := subscribeToVoter(cs1, caller.PubKey().Address())

	signAddVotes(cs1, types.PrecommitType, nil, types.PartSetHeader{}, others...)
	ensureNewRound(newRoundCh, height, 2)
	incrementRound(others...)
	otherProposer(t, cs1, others)
	// the others prevote the block again before any proposal or part arrives
	signAddVotes(cs1, types.PrevoteType, blockID.Hash, blockID.PartsHeader, others...)

	ensurePrevote(voteCh, height, 2)
	ensurePrecommit(voteCh, height, 2)
	rs := cs1.GetRoundState()
	if rs.ProposalBlockParts.IsComplete() {
		t.Fatal("the round-2 part set is complete; the test proves nothing")
	}
	own := rs.Votes.Precommits(2).GetByAddress(caller.PubKey().Address())
	if own == nil || !own.BlockID.Equals(blockID) || rs.LockedRound != 2 || !rs.LockedBlock.HashesTo(blockID.Hash) {
		t.Fatalf("the node did not relock on its block: locked round %d", rs.LockedRound)
	}

	signAddVotes(cs1, types.PrecommitType, blockID.Hash, blockID.PartsHeader, others...)
	ensureNewRound(newRoundCh, height+1, 0)
	drainEvents(t, newRoundCh, voteCh)
	cs1.mtx.RLock()
	defer cs1.mtx.RUnlock()
	if cs1.droppedParts != nil {
		t.Fatal("the refused parts outlive their height")
	}
}

// The replay locks only on the node's own precommit, while no later polka
// unlocked it: a patched node that refused the block live, with nil votes of
// its own and another validator's precommit for the block in its WAL, comes
// back holding no lock, and so does a node with no validator key, or one whose
// own precommit names another block's hash, or is handled after the next
// round's polka for another block, for nil or for the same block under other
// parts.
func TestStateReplayLocksOnlyOnOwnPrecommit(t *testing.T) {
	t.Parallel()

	refusedLive := func(vss []*validatorStub, blockID types.BlockID) []msgInfo {
		caller, other := vss[0], vss[1]
		return []msgInfo{
			{&VoteMessage{signVote(other, types.PrecommitType, blockID.Hash, blockID.PartsHeader)}, "some peer"},
			{&VoteMessage{signVote(caller, types.PrevoteType, nil, types.PartSetHeader{})}, ""},
			{&VoteMessage{signVote(caller, types.PrecommitType, nil, types.PartSetHeader{})}, ""},
		}
	}
	afterPolka := func(polka types.BlockID) func([]*validatorStub, types.BlockID) []msgInfo {
		return func(vss []*validatorStub, blockID types.BlockID) []msgInfo {
			return lateOwnPrecommit(vss, blockID, polka, polka, polka)
		}
	}

	for _, tc := range []struct {
		name    string
		keyless bool
		wal     func([]*validatorStub, types.BlockID) []msgInfo
	}{
		{"own nil votes", false, refusedLive},
		{"no validator key", true, refusedLive},
		{"own precommit after a polka for another block", false, afterPolka(otherBlockID)},
		{"own precommit after a polka for nil", false, afterPolka(types.BlockID{})},
		{"own precommit after a polka for the block under other parts", false, func(vss []*validatorStub, blockID types.BlockID) []msgInfo {
			polka := types.BlockID{Hash: blockID.Hash, PartsHeader: otherBlockID.PartsHeader}
			return lateOwnPrecommit(vss, blockID, polka, polka, polka)
		}},
		{"own precommit for another block under the refused parts' header", false, func(vss []*validatorStub, blockID types.BlockID) []msgInfo {
			named := types.BlockID{Hash: otherBlockID.Hash, PartsHeader: blockID.PartsHeader}
			return []msgInfo{
				{&VoteMessage{signVote(vss[0], types.PrevoteType, named.Hash, named.PartsHeader)}, ""},
				{&VoteMessage{signVote(vss[0], types.PrecommitType, named.Hash, named.PartsHeader)}, ""},
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			blockID, msgs := proposedRoundOne(t, cs1, vss, paddedPartSet)
			msgs = append(msgs, tc.wal(vss, blockID)...)
			if tc.keyless {
				cs1.privValidator = nil
			}
			restartFromWAL(t, cs1, msgs)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()

			rs := cs1.GetRoundState()
			if rs.Votes.Precommits(1).GetByAddress(vss[0].PubKey().Address()) == nil {
				t.Fatal("the replay did not restore the node's own precommit; the test proves nothing")
			}
			if rs.LockedRound != -1 || rs.LockedBlock != nil {
				t.Fatalf("the node comes back locked at round %d", rs.LockedRound)
			}
		})
	}
}

// A node whose own precommit is handled only after the next round's prevotes
// moved it on, with no polka among them or a polka for its own block, must
// still come back locked on its block and prevote it in that round.
func TestStateReplayKeepsLockThroughRoundSkip(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		round2 func(blockID types.BlockID) []types.BlockID
	}{
		{"no polka", func(types.BlockID) []types.BlockID {
			return []types.BlockID{{}, {}, otherBlockID}
		}},
		{"a polka for its own block", func(blockID types.BlockID) []types.BlockID {
			return []types.BlockID{blockID, blockID, blockID}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs1, vss := randConsensusState(4)
			caller, others := vss[0], vss[1:]
			blockID, msgs := proposedRoundOne(t, cs1, vss, paddedPartSet)
			msgs = append(msgs, lateOwnPrecommit(vss, blockID, tc.round2(blockID)...)...)
			restartFromWAL(t, cs1, msgs)
			defer func() {
				cs1.Stop()
				cs1.Wait()
			}()
			newRoundCh := subscribe(cs1.evsw, cstypes.EventNewRound{})
			voteCh := subscribeToVoter(cs1, caller.PubKey().Address())

			rs := cs1.GetRoundState()
			if rs.Round != 2 {
				t.Fatalf("the replay ends in round %d, not after the skip; the test proves nothing", rs.Round)
			}
			if rs.LockedRound != 1 || !rs.LockedBlock.HashesTo(blockID.Hash) || !rs.LockedBlockParts.HasHeader(blockID.PartsHeader) {
				t.Fatalf("the node replayed its own precommit for the block and holds no lock on it: locked round %d", rs.LockedRound)
			}

			prevotesLockOverOtherProposal(t, cs1, caller, others, voteCh, blockID)
			drainEvents(t, newRoundCh, voteCh)
		})
	}
}

// A node relocked in round 3 before its own round-2 precommit is handled must
// keep the round-3 lock: the earlier precommit never moves a lock back.
func TestStateReplayKeepsLaterLock(t *testing.T) {
	t.Parallel()

	cs1, vss := randConsensusState(4)
	caller, others := vss[0], vss[1:]
	blockID, msgs := proposedRoundOne(t, cs1, vss, paddedPartSet)
	vote := func(vs *validatorStub, type_ types.SignedMsgType, id types.BlockID) msgInfo {
		mi := msgInfo{Msg: &VoteMessage{signVote(vs, type_, id.Hash, id.PartsHeader)}}
		if vs != caller {
			mi.PeerID = "some peer"
		}
		return mi
	}
	msgs = append(msgs, vote(caller, types.PrevoteType, blockID), vote(caller, types.PrecommitType, blockID))

	// rounds 2 and 3: the others prevote the block and precommit nil, and the
	// node's own round-2 precommit is handled only after round 3's votes
	var late msgInfo
	for round := 2; round <= 3; round++ {
		incrementRound(vss...)
		for _, vs := range others {
			msgs = append(msgs, vote(vs, types.PrevoteType, blockID))
		}
		if round == 2 {
			msgs = append(msgs, vote(caller, types.PrevoteType, blockID))
		}
		for _, vs := range others {
			msgs = append(msgs, vote(vs, types.PrecommitType, types.BlockID{}))
		}
		if round == 2 {
			late = vote(caller, types.PrecommitType, blockID)
		}
	}
	restartFromWAL(t, cs1, append(msgs, late))
	defer func() {
		cs1.Stop()
		cs1.Wait()
	}()

	rs := cs1.GetRoundState()
	if rs.Votes.Precommits(2).GetByAddress(caller.PubKey().Address()) == nil {
		t.Fatal("the replay did not restore the node's own round-2 precommit; the test proves nothing")
	}
	if rs.LockedRound != 3 || !rs.LockedBlock.HashesTo(blockID.Hash) {
		t.Fatalf("the node comes back locked at round %d, not 3", rs.LockedRound)
	}
}

// prevotesLockOverOtherProposal has one of others, already in round 2, propose
// another block in round 2, and checks that caller, cs1's own validator,
// prevotes the block it is locked on, blockID, instead.
func prevotesLockOverOtherProposal(t *testing.T, cs1 *ConsensusState, caller *validatorStub, others []*validatorStub, voteCh <-chan events.Event, blockID types.BlockID) {
	t.Helper()

	height := cs1.GetRoundState().Height
	proposer := otherProposer(t, cs1, others)
	otherBlock, otherParts := cs1.state.MakeBlock(height, []types.Tx{{0x01}}, types.NewCommit(types.BlockID{}, nil), proposer.PubKey().Address())
	proposal := types.NewProposal(height, 2, -1, types.BlockID{Hash: otherBlock.Hash(), PartsHeader: otherParts.Header()})
	if err := proposer.SignProposal(config.ChainID(), proposal); err != nil {
		t.Fatal("failed to sign the proposal", err)
	}
	if err := cs1.SetProposalAndBlock(proposal, otherBlock, otherParts, "some peer"); err != nil {
		t.Fatal(err)
	}
	ensurePrevote(voteCh, height, 2)
	validatePrevote(cs1, 2, caller, blockID.Hash)
}

// otherBlockID names a block no test proposes, for votes that are for neither
// the proposed block nor nil.
var otherBlockID = types.BlockID{
	Hash:        bytes.Repeat([]byte{0xAB}, 32),
	PartsHeader: types.PartSetHeader{Total: 1, Hash: bytes.Repeat([]byte{0xAB}, 32)},
}

// lateOwnPrecommit returns WAL entries in which the node prevotes blockID in
// round 1, the others then prevote in round 2, each for the block ID at its
// index with the zero ID meaning nil, and only then the node's own round-1
// precommit for blockID is handled. It moves the others to round 2.
func lateOwnPrecommit(vss []*validatorStub, blockID types.BlockID, round2 ...types.BlockID) []msgInfo {
	caller, others := vss[0], vss[1:]
	msgs := []msgInfo{{&VoteMessage{signVote(caller, types.PrevoteType, blockID.Hash, blockID.PartsHeader)}, ""}}
	late := msgInfo{&VoteMessage{signVote(caller, types.PrecommitType, blockID.Hash, blockID.PartsHeader)}, ""}

	incrementRound(others...)
	for i, id := range round2 {
		msgs = append(msgs, msgInfo{&VoteMessage{signVote(others[i], types.PrevoteType, id.Hash, id.PartsHeader)}, "some peer"})
	}
	return append(msgs, late)
}

// proposedRoundOne moves vss to round 1 of cs1's height and returns the block ID and the WAL
// entries of a round in which the second validator proposed a block under the
// part set parts makes, and the other three prevoted it.
func proposedRoundOne(t *testing.T, cs1 *ConsensusState, vss []*validatorStub, parts func(*types.Block) *types.PartSet) (types.BlockID, []msgInfo) {
	t.Helper()

	incrementRound(vss...)
	height, round, proposer := cs1.Height, 1, vss[1]
	// the node's own stub signs the votes its WAL replays
	vss[0].Height = height
	block, _ := cs1.state.MakeBlock(height, []types.Tx{{0x00}}, types.NewCommit(types.BlockID{}, nil), proposer.PubKey().Address())
	blockParts := parts(block)
	blockID := types.BlockID{Hash: block.Hash(), PartsHeader: blockParts.Header()}
	proposal := types.NewProposal(height, round, -1, blockID)
	if err := proposer.SignProposal(config.ChainID(), proposal); err != nil {
		t.Fatal("failed to sign the proposal", err)
	}

	msgs := []msgInfo{{&ProposalMessage{proposal}, "some peer"}}
	for i := 0; i < blockParts.Total(); i++ {
		msgs = append(msgs, msgInfo{&BlockPartMessage{Height: height, Round: round, Part: blockParts.GetPart(i)}, "some peer"})
	}
	for _, vs := range vss[1:] {
		msgs = append(msgs, msgInfo{&VoteMessage{signVote(vs, types.PrevoteType, blockID.Hash, blockID.PartsHeader)}, "some peer"})
	}
	return blockID, msgs
}

// restartFromWAL starts cs1 in round 1 of its height from a WAL holding msgs,
// as a node restarted mid-height catches up, and returns once the replay ends.
func restartFromWAL[M walm.WALMessage](t *testing.T, cs1 *ConsensusState, msgs []M) {
	t.Helper()

	wal := makeTempWAL(t, 1<<20)
	if err := wal.WriteMetaSync(walm.MetaMessage{Height: cs1.Height}); err != nil {
		t.Fatal(err)
	}
	for _, mi := range msgs {
		if err := wal.Write(mi); err != nil {
			t.Fatal(err)
		}
	}
	if err := wal.FlushAndSync(); err != nil {
		t.Fatal(err)
	}

	cs1.enterNewRound(cs1.Height, 1)
	cs1.wal = wal
	cs1.doWALCatchup = true
	if err := cs1.Start(); err != nil {
		t.Fatal(err)
	}
}

// otherProposer returns the validator of others proposing cs1's round, and
// fails the test when cs1 proposes it itself.
func otherProposer(t *testing.T, cs1 *ConsensusState, others []*validatorStub) *validatorStub {
	t.Helper()

	cs1.mtx.Lock()
	defer cs1.mtx.Unlock()
	for _, vs := range others {
		if cs1.Validators.GetProposer().Address == vs.PubKey().Address() {
			return vs
		}
	}
	t.Fatal("the node proposes this round itself; the test proves nothing")
	return nil
}

// paddedPartSet chunks block with one byte of trailing padding: the same block,
// under a part-set header its canonical chunking does not reproduce.
func paddedPartSet(block *types.Block) *types.PartSet {
	return types.NewPartSetFromData(append(amino.MustMarshalSized(block), 0x00), types.BlockPartSizeBytes)
}

// rechunkedPartSet cuts block's own encoding into 64-byte parts: the same
// bytes, so the same ByteSize, under a header the canonical chunking does not make.
func rechunkedPartSet(block *types.Block) *types.PartSet {
	return types.NewPartSetFromData(amino.MustMarshalSized(block), 64)
}

// trailingFieldPartSet chunks block's encoding followed by an empty unknown
// field 4, inside the length prefix: the decode fills the block, then fails.
func trailingFieldPartSet(block *types.Block) *types.PartSet {
	bz := append(amino.MustMarshal(block), 0x22, 0x00)
	return types.NewPartSetFromData(append(binary.AppendUvarint(nil, uint64(len(bz))), bz...), types.BlockPartSizeBytes)
}

// unfilledPartSet returns a part set maker chunking block with leaveOut applied
// to a copy: a header hash Hash() refills once the block is decoded.
func unfilledPartSet(leaveOut func(*types.Block)) func(*types.Block) *types.PartSet {
	return func(block *types.Block) *types.PartSet {
		unfilled := amino.DeepCopy(block).(*types.Block)
		leaveOut(unfilled)
		return unfilled.MakePartSet(types.BlockPartSizeBytes)
	}
}

// commitHeightOne takes cs1, the proposer at height 1 round 0, through a full
// round with the other validators, and returns once it enters height 2.
func commitHeightOne(t *testing.T, cs1 *ConsensusState, vss []*validatorStub, newRoundCh, voteCh <-chan events.Event) {
	t.Helper()

	ensureNewRound(newRoundCh, 1, 0)
	ensurePrevote(voteCh, 1, 0)
	rs := cs1.GetRoundState()
	hash, header := rs.ProposalBlock.Hash(), rs.ProposalBlockParts.Header()
	signAddVotes(cs1, types.PrevoteType, hash, header, vss[1:]...)
	ensurePrecommit(voteCh, 1, 0)
	signAddVotes(cs1, types.PrecommitType, hash, header, vss[1:]...)
	ensureNewRound(newRoundCh, 2, 0)
	incrementHeight(vss[1:]...)
}

// drainEvents reads chs until the test ends, so an event no assertion waits
// for cannot block the state machine.
func drainEvents(t *testing.T, chs ...<-chan events.Event) {
	t.Helper()

	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	for _, ch := range chs {
		go func() {
			for {
				select {
				case <-ch:
				case <-done:
					return
				}
			}
		}()
	}
}
