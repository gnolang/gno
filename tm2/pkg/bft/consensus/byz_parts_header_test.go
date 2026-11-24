package consensus

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// proposalBlockWithTxs builds a proposal block carrying txs that passes ValidateBlock.
func proposalBlockWithTxs(t *testing.T, cs *ConsensusState, txs types.Txs) *types.Block {
	t.Helper()

	block, _ := cs.createProposalBlock()
	require.NotNil(t, block)
	block.Data = types.Data{Txs: txs}
	block.DataHash = nil
	block.NumTxs = int64(len(txs))
	block.TotalTxs = cs.state.LastBlockTotalTx + int64(len(txs))
	require.NotNil(t, block.Hash()) // fills DataHash and settles the header encoding
	require.NoError(t, cs.state.ValidateBlock(block))
	return block
}

// counterTxs returns n serial counter txs, so a block carrying them applies.
func counterTxs(n int) types.Txs {
	txs := make(types.Txs, n)
	for i := range txs {
		txs[i] = binary.BigEndian.AppendUint64(nil, uint64(i))
	}
	return txs
}

// queuedVote drains cs's internal queue and returns the last vote of voteType it signed for round.
func queuedVote(t *testing.T, cs *ConsensusState, voteType types.SignedMsgType, round int) *types.Vote {
	t.Helper()

	var vote *types.Vote
	for _, mi := range drainInternalQueue(cs) {
		if vm, ok := mi.Msg.(*VoteMessage); ok && vm.Vote.Round == round && vm.Vote.Type == voteType {
			vote = vm.Vote
		}
	}
	require.NotNil(t, vote)
	return vote
}

// equivocatingPartsHeader builds a block and two different chunkings of it.
func equivocatingPartsHeader(t *testing.T, cs *ConsensusState) (*types.Block, types.PartSetHeader, types.PartSetHeader) {
	t.Helper()

	// Make the block span several parts, so two chunkings give two headers.
	block := proposalBlockWithTxs(t, cs, types.Txs{types.Tx(bytes.Repeat([]byte("x"), 3*types.BlockPartSizeBytes/2))})

	fine := block.MakePartSet(types.BlockPartSizeBytes / 2).Header()
	// Three parts where the BlockPartSizeBytes cut gives two, so no rebuild reproduces it.
	coarse := block.MakePartSet(types.BlockPartSizeBytes * 5 / 8).Header()
	require.Equal(t, 3, coarse.Total)
	require.NotEqual(t, fine, coarse)
	require.Equal(t, 2, block.MakePartSet(types.BlockPartSizeBytes).Total())

	return block, fine, coarse
}

// requireDroppedFor asserts the block was dropped and its parts are fetched under committed.
func requireDroppedFor(t *testing.T, cs *ConsensusState, committed types.PartSetHeader) {
	t.Helper()

	require.Nil(t, cs.ProposalBlock, "a block whose cut cannot be rebuilt is dropped")
	require.Equal(t, committed, cs.ProposalBlockParts.Header())
}

// drainInternalQueue empties cs's internal queue and returns what it held.
func drainInternalQueue(cs *ConsensusState) []msgInfo {
	var msgs []msgInfo
	for len(cs.internalMsgQueue) > 0 {
		msgs = append(msgs, <-cs.internalMsgQueue)
	}
	return msgs
}

// addVotesNow signs a vote for (hash, header) from each of vss and adds it to cs at once, unlike the queued addVotes.
func addVotesNow(t *testing.T, cs *ConsensusState, voteType types.SignedMsgType, hash []byte, header types.PartSetHeader, vss ...*validatorStub) {
	t.Helper()

	for _, vote := range signVotes(voteType, hash, header, vss...) {
		_, err := cs.addVote(vote, "peer")
		require.NoError(t, err)
	}
}

// equalCut returns block cut into two equal parts: as many parts as the
// BlockPartSizeBytes cut, but another cut, so no rebuild reproduces it.
func equalCut(t *testing.T, block *types.Block) *types.PartSet {
	t.Helper()

	bz, err := amino.MarshalSized(block)
	require.NoError(t, err)
	parts := block.MakePartSet((len(bz) + 1) / 2)
	honest := block.MakePartSet(types.BlockPartSizeBytes)
	require.Equal(t, honest.Total(), parts.Total())
	require.NotEqual(t, honest.Header(), parts.Header())
	return parts
}

// TestByzProposerEquivocatingPartsHeaderAddVote covers an equivocated header arriving in addVote.
func TestByzProposerEquivocatingPartsHeaderAddVote(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	block, fine, coarse := equivocatingPartsHeader(t, cs)

	cs.ProposalBlock = block
	cs.ProposalBlockParts = types.NewPartSetFromHeader(fine)

	for _, vs := range vss[1:] {
		vote := signVote(vs, types.PrevoteType, block.Hash(), coarse)
		added, err := cs.addVote(vote, "peer")
		require.NoError(t, err)
		require.True(t, added)
	}

	requireDroppedFor(t, cs, coarse)
	require.Nil(t, cs.ValidBlock, "a dropped block must not become the ValidBlock")
	require.Equal(t, -1, cs.ValidRound, "a dropped block must not advance ValidRound")
}

// TestByzProposerEquivocatingPartsHeaderRebuild covers rebuilding the part set from the block.
func TestByzProposerEquivocatingPartsHeaderRebuild(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	block := proposalBlockWithTxs(t, cs, types.Txs{types.Tx(bytes.Repeat([]byte("x"), 3*types.BlockPartSizeBytes/2))})

	honest := block.MakePartSet(types.BlockPartSizeBytes).Header()
	fine := block.MakePartSet(types.BlockPartSizeBytes / 2).Header()

	cs.ProposalBlock = block
	cs.ProposalBlockParts = types.NewPartSetFromHeader(fine)

	for _, vs := range vss[1:] {
		vote := signVote(vs, types.PrevoteType, block.Hash(), honest)
		added, err := cs.addVote(vote, "peer")
		require.NoError(t, err)
		require.True(t, added)
	}

	require.NotNil(t, cs.ProposalBlock, "the honest header keeps the block")
	require.True(t, cs.ProposalBlockParts.IsComplete())
	require.True(t, cs.ProposalBlockParts.HasHeader(honest))
	require.NotNil(t, cs.ValidBlock, "the rebuilt pair becomes the ValidBlock")
	require.Equal(t, 0, cs.ValidRound)
	require.True(t, cs.ValidBlockParts.HasHeader(honest))
}

// TestByzProposerEquivocatingPartsHeaderEqualCut covers a node holding the
// honest cut, the only one addProposalBlockPart keeps, that sees +2/3 prevotes
// for another cut of the same block: it precommits nil rather than that cut.
func TestByzProposerEquivocatingPartsHeaderEqualCut(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	height := cs.Height
	block := proposalBlockWithTxs(t, cs, types.Txs{types.Tx(bytes.Repeat([]byte("x"), 3*types.BlockPartSizeBytes/2))})
	honest := block.MakePartSet(types.BlockPartSizeBytes)
	equal := equalCut(t, block).Header()

	// The Byzantine proposer sends this node the honest cut.
	idx, _ := cs.Validators.GetByAddress(cs.Validators.GetProposer().Address)
	proposal := types.NewProposal(height, 0, -1, types.BlockID{Hash: block.Hash(), PartsHeader: honest.Header()})
	require.NoError(t, vss[idx].SignProposal(cs.state.ChainID, proposal))
	require.NoError(t, cs.defaultSetProposal(proposal))
	for i := range honest.Total() {
		_, err := cs.addProposalBlockPart(&BlockPartMessage{height, 0, honest.GetPart(i)}, "peer")
		require.NoError(t, err)
	}
	require.NotNil(t, cs.ProposalBlock)

	// The other validators prevote the equal cut.
	addVotesNow(t, cs, types.PrevoteType, block.Hash(), equal, vss[1:]...)
	cs.enterPrecommit(height, 0)

	precommit := queuedVote(t, cs, types.PrecommitType, 0)
	require.True(t, precommit.BlockID.IsZero(), "a node precommits no cut but the honest one")
	require.Nil(t, cs.LockedBlock)
	requireDroppedFor(t, cs, equal)
}

// TestByzProposerEquivocatingPartsHeaderPolkaPastCeiling covers a polka naming
// more parts than a whole block fills: the node keeps the block it holds.
func TestByzProposerEquivocatingPartsHeaderPolkaPastCeiling(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	block := proposalBlockWithTxs(t, cs, types.Txs{types.Tx(bytes.Repeat([]byte("x"), 3*types.BlockPartSizeBytes/2))})
	honest := block.MakePartSet(types.BlockPartSizeBytes)
	cs.ProposalBlock, cs.ProposalBlockParts = block, honest

	header := types.PartSetHeader{Total: fullBlockParts(cs) + 1, Hash: block.Hash()}
	addVotesNow(t, cs, types.PrevoteType, block.Hash(), header, vss[1:]...)

	require.NotNil(t, cs.ProposalBlock, "a header no whole block fills drops nothing")
	require.True(t, cs.ProposalBlockParts.HasHeader(honest.Header()))
}

// TestByzProposerEquivocatingPartsHeaderEnterCommit covers a held block with a stale parts header.
func TestByzProposerEquivocatingPartsHeaderEnterCommit(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	block, fine, coarse := equivocatingPartsHeader(t, cs)

	cs.ProposalBlock = block
	cs.ProposalBlockParts = types.NewPartSetFromHeader(fine)

	for _, vs := range vss[1:] {
		vote := signVote(vs, types.PrecommitType, block.Hash(), coarse)
		added, err := cs.Votes.AddVote(vote, "peer")
		require.NoError(t, err)
		require.True(t, added)
	}

	cs.enterCommit(cs.Height, cs.Round)
	requireDroppedFor(t, cs, coarse)
}

// TestByzProposerEquivocatingPartsHeaderEnterCommitRebuild covers enterCommit rebuilding the parts.
func TestByzProposerEquivocatingPartsHeaderEnterCommitRebuild(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	// Serial counter txs, enough to span two parts, so the block applies.
	block := proposalBlockWithTxs(t, cs, counterTxs(10000))

	honest := block.MakePartSet(types.BlockPartSizeBytes).Header()
	fine := block.MakePartSet(types.BlockPartSizeBytes / 2).Header()
	require.NotEqual(t, fine, honest)

	cs.ProposalBlock = block
	cs.ProposalBlockParts = types.NewPartSetFromHeader(fine)

	for _, vs := range vss[1:] {
		vote := signVote(vs, types.PrecommitType, block.Hash(), honest)
		added, err := cs.Votes.AddVote(vote, "peer")
		require.NoError(t, err)
		require.True(t, added)
	}

	height := cs.Height
	cs.enterCommit(height, cs.Round)
	require.Equal(t, height, cs.blockStore.Height(), "the held block is committed from the rebuilt parts")
}

// TestByzProposerEquivocatingPartsHeaderRelock covers a later polka for the locked block under another chunking.
func TestByzProposerEquivocatingPartsHeaderRelock(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		late    bool // the round-1 polka arrives once the node is in round 2
		rebuilt bool
	}{
		{"relock in round 1, rebuilt", false, true},
		{"relock in round 1, dropped", false, false},
		{"round-1 polka seen in round 2, rebuilt", true, true},
		{"round-1 polka seen in round 2, dropped", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cs, vss := randConsensusState(4)
			height := cs.Height
			block := proposalBlockWithTxs(t, cs, types.Txs{types.Tx(bytes.Repeat([]byte("x"), 3*types.BlockPartSizeBytes/2))})

			honest, equal := block.MakePartSet(types.BlockPartSizeBytes), equalCut(t, block)
			// A lock on the honest cut meets a polka for a cut no rebuild reproduces.
			locked, polka := honest, equal.Header()
			if tc.rebuilt {
				// A lock on another cut, as relockOnOwnPrecommit restores, meets a polka for the honest one.
				locked, polka = equal, honest.Header()
			}

			lockOnParts(t, cs, vss, block, locked)

			// Round 1: polka for the same hash under the other chunking.
			incrementRound(vss[1:]...)
			if tc.late {
				cs.enterNewRound(height, 2)
			}
			addVotesNow(t, cs, types.PrevoteType, block.Hash(), polka, vss[1:]...)
			if !tc.late {
				require.Equal(t, 1, cs.Round)
				cs.enterPrecommit(height, 1)
			}

			// Round 2: the node prevotes the newest polka's pair, or nil once it dropped the lock.
			drainInternalQueue(cs)
			if !tc.late {
				cs.enterNewRound(height, 2)
			}
			cs.enterPrevote(height, 2)
			prevote := queuedVote(t, cs, types.PrevoteType, 2)
			if tc.rebuilt {
				require.True(t, cs.LockedBlockParts.IsComplete())
				require.Equal(t, types.BlockID{Hash: block.Hash(), PartsHeader: polka}, prevote.BlockID)
			} else {
				require.Nil(t, cs.LockedBlock)
				require.True(t, prevote.BlockID.IsZero(), "a dropped lock is not prevoted")
			}
		})
	}
}

// TestByzProposerEquivocatingPartsHeaderCommitKeepsHeldParts covers a commit for parts held under another header than the lock's.
func TestByzProposerEquivocatingPartsHeaderCommitKeepsHeldParts(t *testing.T) {
	t.Parallel()

	cs, vss := randConsensusState(4)
	height := cs.Height
	// Serial counter txs, so the block applies once committed.
	block := proposalBlockWithTxs(t, cs, counterTxs(10000))
	honest := block.MakePartSet(types.BlockPartSizeBytes)
	held := block.MakePartSet(types.BlockPartSizeBytes * 5 / 8)
	require.NotEqual(t, honest.Header(), held.Header(), "the held chunking is not the honest cut")

	lockOnParts(t, cs, vss, block, honest)

	// Round 1: the node holds every part of a cut it cannot rebuild, then +2/3 precommit that cut.
	incrementRound(vss[1:]...)
	cs.enterNewRound(height, 1)
	cs.ProposalBlock, cs.ProposalBlockParts = block, held
	addVotesNow(t, cs, types.PrecommitType, block.Hash(), held.Header(), vss[1:]...)
	require.Equal(t, height+1, cs.Height, "a node holding the committed block and its parts commits it")
}

// lockOnParts locks cs on block with parts through a round-0 polka.
func lockOnParts(t *testing.T, cs *ConsensusState, vss []*validatorStub, block *types.Block, parts *types.PartSet) {
	t.Helper()
	cs.ProposalBlock, cs.ProposalBlockParts = block, parts
	addVotesNow(t, cs, types.PrevoteType, block.Hash(), parts.Header(), vss[1:]...)
	cs.enterPrecommit(cs.Height, 0)
	require.True(t, cs.LockedBlockParts.HasHeader(parts.Header()))
}

// A node locked on the committed block, holding part of a later round's set
// for the same header, commits the block rather than saving that partial set.
func TestByzProposerEquivocatingPartsHeaderCommitPartialHonestSet(t *testing.T) {
	cs, vss := randConsensusState(4)
	height := cs.Height
	block := proposalBlockWithTxs(t, cs, counterTxs(10000))
	honest := block.MakePartSet(types.BlockPartSizeBytes)
	require.Greater(t, honest.Total(), 1)
	lockOnParts(t, cs, vss, block, honest)

	incrementRound(vss[1:]...)
	cs.enterNewRound(height, 1)
	partial := types.NewPartSetFromHeader(honest.Header())
	_, err := partial.AddPart(honest.GetPart(0))
	require.NoError(t, err)
	cs.ProposalBlock, cs.ProposalBlockParts = nil, partial
	require.NotPanics(t, func() {
		addVotesNow(t, cs, types.PrecommitType, block.Hash(), honest.Header(), vss[1:]...)
	})
	require.Equal(t, height+1, cs.Height, "a node locked on the committed block commits it")
}

// A node locked on a cut the block cannot rebuild commits from its locked parts
// when a later round commits that cut, since no peer gossips a lock's parts.
func TestByzProposerEquivocatingPartsHeaderCommitLockedUnrebuildableCut(t *testing.T) {
	for _, tc := range []struct {
		name  string
		polka bool // a round-1 polka on the cut sets up an empty part set to fetch
		held  int  // parts of that set received before the commit
	}{
		{"no round-1 part set", false, 0},
		{"round-1 part set empty", true, 0},
		{"round-1 part set partial", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cs, vss := randConsensusState(4)
			height := cs.Height
			block := proposalBlockWithTxs(t, cs, counterTxs(10000))
			cut := block.MakePartSet(types.BlockPartSizeBytes * 5 / 8)
			require.Greater(t, cut.Total(), 2)
			require.NotEqual(t, block.MakePartSet(types.BlockPartSizeBytes).Header(), cut.Header(), "the locked cut is not the honest cut")
			lockOnParts(t, cs, vss, block, cut)

			incrementRound(vss[1:]...)
			cs.enterNewRound(height, 1)
			if tc.polka {
				addVotesNow(t, cs, types.PrevoteType, block.Hash(), cut.Header(), vss[1:]...)
				require.True(t, cs.ProposalBlockParts.HasHeader(cut.Header()))
				for i := 0; i < tc.held; i++ {
					_, err := cs.ProposalBlockParts.AddPart(cut.GetPart(i))
					require.NoError(t, err)
				}
			}
			addVotesNow(t, cs, types.PrecommitType, block.Hash(), cut.Header(), vss[1:]...)
			require.Equal(t, height+1, cs.Height, "a node locked on the committed cut commits it")
		})
	}
}

// A node locked on another cut keeps the parts it already received for a
// committed header it cannot rebuild, rather than fetching them again.
func TestByzProposerEquivocatingPartsHeaderCommitKeepsPartialParts(t *testing.T) {
	cs, vss := randConsensusState(4)
	height := cs.Height
	block, _, coarse := equivocatingPartsHeader(t, cs)
	lockOnParts(t, cs, vss, block, block.MakePartSet(types.BlockPartSizeBytes))

	incrementRound(vss[1:]...)
	cs.enterNewRound(height, 1)
	coarseParts := block.MakePartSet(types.BlockPartSizeBytes * 5 / 8)
	require.Equal(t, coarse, coarseParts.Header())
	partial := types.NewPartSetFromHeader(coarse)
	for i := 0; i < coarse.Total-1; i++ {
		_, err := partial.AddPart(coarseParts.GetPart(i))
		require.NoError(t, err)
	}
	cs.ProposalBlock, cs.ProposalBlockParts = nil, partial
	held := partial.Count()
	for _, vs := range vss[1:] {
		_, err := cs.Votes.AddVote(signVote(vs, types.PrecommitType, block.Hash(), coarse), "peer")
		require.NoError(t, err)
	}
	cs.enterCommit(height, 1)
	require.True(t, cs.ProposalBlockParts.HasHeader(coarse))
	require.EqualValues(t, held, cs.ProposalBlockParts.Count(), "parts already received for the committed header are kept")
}

// A node whose lock moved to the honest cut proposes, on its next turn, the cut
// it prevotes, and accepts its own proposal.
func TestByzProposerEquivocatingPartsHeaderProposesLockedCut(t *testing.T) {
	cs, vss := randConsensusState(4)
	height := cs.Height
	block, _, _ := equivocatingPartsHeader(t, cs)
	honest := block.MakePartSet(types.BlockPartSizeBytes).Header()
	equal := equalCut(t, block)
	lockOnParts(t, cs, vss, block, equal)
	require.True(t, cs.ValidBlockParts.HasHeader(equal.Header()))

	// Round 1: a polka on the honest cut moves the lock to it.
	incrementRound(vss[1:]...)
	cs.enterNewRound(height, 1)
	addVotesNow(t, cs, types.PrevoteType, block.Hash(), honest, vss[1:]...)
	cs.enterPrecommit(height, 1)
	require.True(t, cs.LockedBlockParts.HasHeader(honest))

	// The next round this node proposes.
	addr := vss[0].PubKey().Address()
	round := -1
	for r := 2; r < 20 && round == -1; r++ {
		v := cs.Validators.Copy()
		v.IncrementProposerPriority(r - cs.Round)
		if v.GetProposer().Address == addr {
			round = r
		}
	}
	require.NotEqual(t, -1, round)
	drainInternalQueue(cs)
	cs.enterNewRound(height, round)
	cs.enterPrevote(height, round)
	var proposal *types.Proposal
	var parts []*BlockPartMessage
	for _, mi := range drainInternalQueue(cs) {
		switch m := mi.Msg.(type) {
		case *ProposalMessage:
			proposal = m.Proposal
		case *BlockPartMessage:
			parts = append(parts, m)
		}
	}
	require.NotNil(t, proposal)
	require.Equal(t, types.BlockID{Hash: block.Hash(), PartsHeader: honest}, proposal.BlockID, "the proposal carries the locked cut")

	require.NoError(t, cs.defaultSetProposal(proposal))
	for _, p := range parts {
		_, err := cs.addProposalBlockPart(p, "")
		require.NoError(t, err)
	}
	require.True(t, cs.ProposalBlock.HashesTo(block.Hash()), "the node accepts the proposal it signed")
}
