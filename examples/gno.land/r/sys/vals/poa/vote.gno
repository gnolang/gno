package poa

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

type (
	tallyMap     map[std.Address]tally
	haveVotedMap map[string]struct{}
)

type tally struct {
	toAdd    uint64
	toRemove uint64
}

type votingSystem struct {
	tallyMap     tallyMap
	haveVotedMap haveVotedMap
}

// newVotingSystem creates a new PoA voting system
func newVotingSystem() *votingSystem {
	return &votingSystem{
		tallyMap:     make(tallyMap),
		haveVotedMap: make(haveVotedMap),
	}
}

// castVote casts a new vote to the given candidate
func (v *votingSystem) castVote(voter, candidate std.Address, toAdd bool) {
	// Construct a unique proposal key (voterAddr_candidateAddr)
	votedKey := ufmt.Sprintf(
		"%s_%s",
		voter.String(),
		candidate.String(),
	)

	if _, voted := v.haveVotedMap[votedKey]; voted {
		// Vote already cast, noop
		return
	}

	// Fetch the tally
	t, exists := v.tallyMap[candidate]
	if !exists {
		t = tally{
			toAdd:    0,
			toRemove: 0,
		}
	}

	// Update the tally
	if toAdd {
		t.toAdd++
	} else {
		t.toRemove++
	}

	// Save the tally
	v.tallyMap[candidate] = t
	v.haveVotedMap[votedKey] = struct{}{}
}

// getTally returns the current voting tally for a candidate
func (v *votingSystem) getTally(candidate std.Address) (uint64, uint64) {
	t, exists := v.tallyMap[candidate]
	if !exists {
		return 0, 0
	}

	return t.toAdd, t.toRemove
}

// reset removes all candidates from the voting system
func (v *votingSystem) reset() {
	v.tallyMap = make(tallyMap)
	v.haveVotedMap = make(haveVotedMap)
}

// resetCandidate removes a candidate from the voting system
func (v *votingSystem) resetCandidate(candidate std.Address) {
	// Drop the tally entries
	delete(v.tallyMap, candidate)

	// Drop the tracked votes
	candidateStr := candidate.String()
	for key := range v.haveVotedMap {
		if !strings.Contains(key, candidateStr) {
			continue
		}

		delete(v.haveVotedMap, key)
	}
}
