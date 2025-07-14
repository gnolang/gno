package governance

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	u256 "gno.land/p/gnoswap/uint256"
)

type GovernanceState struct {
	// Core state
	proposalCounter    uint64
	proposals          *avl.Tree // proposalId -> ProposalInfo
	archievedProposals *avl.Tree // proposalId -> ProposalInfo

	// voting state
	votes     *avl.Tree // voteKey(proposalId:voter) -> bool
	userVotes *avl.Tree // userVoteKey(voter:proposalId) -> voteWeight

	// tracker
	proposalIdByProposer *avl.Tree // proposer -> proposalId

	// config
	currentConfig  Config
	configVersions *avl.Tree // version -> config
	configCounter  uint64
}

// NewGovernanceState creates a new `GovernanceState` with initial config
func NewGovernanceState(initialConfig Config) *GovernanceState {
	versions := avl.NewTree()
	versions.Set(formatUint(uint64(1)), initialConfig)

	return &GovernanceState{
		proposalCounter:      0,
		proposals:            avl.NewTree(),
		archievedProposals:   avl.NewTree(),
		votes:                avl.NewTree(),
		userVotes:            avl.NewTree(),
		proposalIdByProposer: avl.NewTree(),
		currentConfig:        initialConfig,
		configVersions:       versions,
		configCounter:        1,
	}
}

/* Proposal management */

func (s *GovernanceState) WithNewProposal(proposer std.Address, proposal ProposalInfo) *GovernanceState {
	state := s.clone()
	state.proposalCounter++

	proposalId := state.proposalCounter
	state.proposals.Set(formatUint(proposalId), proposal)
	state.proposalIdByProposer.Set(proposal.Proposer.String(), proposalId)

	return state
}

func (s *GovernanceState) WithUpdatedProposal(proposalId uint64, updater func(ProposalInfo) ProposalInfo) *GovernanceState {
	proposal, exists := s.proposals.Get(formatUint(proposalId))
	if !exists {
		return s // no change if proposal doesn't exist
	}

	info, ok := proposal.(ProposalInfo)
	if !ok {
		panic(ufmt.Sprintf("proposal(%d) is not ProposalInfo", proposalId))
	}
	updated := updater(info)

	state := s.clone()
	state.proposals.Set(formatUint(proposalId), updated)

	return state
}

func (s *GovernanceState) WithArchivedProposal(proposalId uint64) *GovernanceState {
	id := formatUint(proposalId)
	proposal, exists := s.proposals.Get(id)
	if !exists {
		return s
	}

	state := s.clone()
	state.proposals.Remove(id)
	state.archievedProposals.Set(id, proposal)

	// remove from proposal mapping
	info, ok := proposal.(ProposalInfo)
	if !ok {
		panic(ufmt.Sprintf("archived proposal(%d) is not ProposalInfo", proposalId))
	}
	state.proposalIdByProposer.Remove(info.Proposer.String())

	return state
}

/* Vote management */

func (s *GovernanceState) WithVote(proposalId uint64, voter std.Address, vote bool, weight uint64) *GovernanceState {
	state := s.clone()

	// record the vote
	voteKey := createVoteKey(proposalId, voter.String())
	state.votes.Set(voteKey, vote)

	// record user vote with metadata
	userVoteKey := createUserVoteKey(voter, proposalId)
	voteData := voteWithWeight{
		Yes:         vote,
		Weight:      weight,
		VotedHeight: uint64(std.ChainHeight()),
		VotedAt:     uint64(time.Now().Unix()),
	}
	state.userVotes.Set(userVoteKey, voteData)

	// update proposal vote counts
	return state.WithUpdatedProposal(proposalId, func(p ProposalInfo) ProposalInfo {
		if vote {
			p.Yea = new(u256.Uint).Add(p.Yea, u256.NewUint(weight))
		} else {
			p.Nay = new(u256.Uint).Add(p.Nay, u256.NewUint(weight))
		}
		return p
	})
}

// configuration management

// WithConfig adds a new configuration version
func (s *GovernanceState) WithNewConfig(cfg Config) *GovernanceState {
	state := s.clone()
	state.configCounter++
	state.configVersions.Set(formatUint(state.configCounter), cfg)
	state.currentConfig = cfg
	return state
}

// Query methods (should readonly)

// GetProposal retrieves a proposal by ID
func (s *GovernanceState) GetProposal(proposalId uint64) (ProposalInfo, bool) {
	val, exists := s.proposals.Get(formatUint(proposalId))
	if !exists {
		return ProposalInfo{}, false
	}
	info, ok := val.(ProposalInfo)
	if !ok {
		panic(ufmt.Sprintf("proposal(%d) is not ProposalInfo", proposalId))
	}
	return info, true
}

func (s *GovernanceState) GetArchivedProposal(proposalId uint64) (ProposalInfo, bool) {
	val, exists := s.archievedProposals.Get(formatUint(proposalId))
	if !exists {
		return ProposalInfo{}, false
	}
	info, ok := val.(ProposalInfo)
	if !ok {
		panic(ufmt.Sprintf("archived proposal(%d) is not ProposalInfo", proposalId))
	}
	return info, true
}

func (s *GovernanceState) GetVote(proposalId uint64, voter std.Address) (bool, bool) {
	key := createVoteKey(proposalId, voter.String())
	val, exists := s.votes.Get(key)
	if !exists {
		return false, false
	}
	return val.(bool), true
}

// GetUserVote retrieves detailed vote information
func (s *GovernanceState) GetUserVote(voter std.Address, proposalId uint64) (voteWithWeight, bool) {
	userVoteKey := createUserVoteKey(voter, proposalId)
	val, exists := s.userVotes.Get(userVoteKey)
	if !exists {
		return voteWithWeight{}, false
	}
	return val.(voteWithWeight), true
}

// GetProposalByProposer retrieves the active proposal ID for a proposer
func (s *GovernanceState) GetProposalByProposer(proposer std.Address) (uint64, bool) {
	val, exists := s.proposalIdByProposer.Get(proposer.String())
	if !exists {
		return 0, false
	}
	return val.(uint64), true
}

func (s *GovernanceState) GetConfig(version uint64) (Config, bool) {
	if version == 0 {
		return s.currentConfig, true
	}

	val, exists := s.configVersions.Get(formatUint(version))
	if !exists {
		return Config{}, false
	}
	cfg, ok := val.(Config)
	if !ok {
		panic(ufmt.Sprintf("config version(%d) is not Config", version))
	}
	return cfg, true
}

// GetProposalCounter returns the current proposal counter
func (s *GovernanceState) GetProposalCounter() uint64 {
	return s.proposalCounter
}

func (s *GovernanceState) GetAccumulatedVotes(proposalId uint64) (uint64, uint64) {
	proposal, exists := s.GetProposal(proposalId)
	if !exists {
		return 0, 0
	}
	return proposal.Yea.Uint64(), proposal.Nay.Uint64()
}

/* Iterators */

func (s *GovernanceState) IterateProposals(fn func(uint64, ProposalInfo) bool) {
	s.proposals.Iterate("", "", func(key string, value any) bool {
		proposalId := parseUint64(key)
		proposal, ok := value.(ProposalInfo)
		if !ok {
			panic(ufmt.Sprintf("proposal(%d) is not ProposalInfo", proposalId))
		}
		return fn(proposalId, proposal)
	})
}

// IterateUserVotes iterates over all votes by a specific user
func (s *GovernanceState) IterateUserVotes(voter std.Address, fn func(uint64, voteWithWeight) bool) {
	prefix := voter.String() + ":"
	s.userVotes.Iterate(prefix, prefix+"~", func(key string, val any) bool {
		// Extract proposal ID from key
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			return true
		}
		proposalId := parseUint64(parts[1])
		vote := val.(voteWithWeight)
		return fn(proposalId, vote)
	})
}

// Helper method to clone state
func (s *GovernanceState) clone() *GovernanceState {
	// Create new AVL trees by copying all data
	newProposals := avl.NewTree()
	s.proposals.Iterate("", "", func(key string, value any) bool {
		newProposals.Set(key, value)
		return false
	})

	newArchivedProposals := avl.NewTree()
	s.archievedProposals.Iterate("", "", func(key string, value any) bool {
		newArchivedProposals.Set(key, value)
		return false
	})

	newVotes := avl.NewTree()
	s.votes.Iterate("", "", func(key string, value any) bool {
		newVotes.Set(key, value)
		return false
	})

	newUserVotes := avl.NewTree()
	s.userVotes.Iterate("", "", func(key string, value any) bool {
		newUserVotes.Set(key, value)
		return false
	})

	newProposalIdByProposer := avl.NewTree()
	s.proposalIdByProposer.Iterate("", "", func(key string, value any) bool {
		newProposalIdByProposer.Set(key, value)
		return false
	})

	newConfigVersions := avl.NewTree()
	s.configVersions.Iterate("", "", func(key string, value any) bool {
		newConfigVersions.Set(key, value)
		return false
	})

	return &GovernanceState{
		proposalCounter:      s.proposalCounter,
		proposals:            newProposals,
		archievedProposals:   newArchivedProposals,
		votes:                newVotes,
		userVotes:            newUserVotes,
		proposalIdByProposer: newProposalIdByProposer,
		currentConfig:        s.currentConfig,
		configVersions:       newConfigVersions,
		configCounter:        s.configCounter,
	}
}

// Validation methods

// CanPropose checks if an address can create a proposal
func (s *GovernanceState) CanPropose(proposer std.Address) bool {
	_, hasActiveProposal := s.GetProposalByProposer(proposer)
	return !hasActiveProposal
}

// HasVoted checks if a voter has already voted on a proposal
func (s *GovernanceState) HasVoted(proposalId uint64, voter std.Address) bool {
	_, exists := s.GetVote(proposalId, voter)
	return exists
}
