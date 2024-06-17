package poa

import (
	"std"

	"gno.land/r/sys/vals/types"
)

const (
	UserVotedAddEvent    = "UserVotedAdd"    // emitted when someone votes to add a validator
	UserVotedRemoveEvent = "UserVotedRemove" // emitted when someone votes to remove a validator
)

var errCallerNotValidator = "caller is not validator"

// PoA specifies the Proof of Authority validator set.
// In order to become part of the set, users to be voted into
// the validator set by the majority of the existing set
type PoA struct {
	// validators holds the current validator set.
	// This slice can never practically grow more than ~150 elements,
	// due to Tendermint's quadratic network complexity
	validators              []*types.Validator
	addressToValidatorIndex map[std.Address]int // address -> index
	votes                   *votingSystem       // system that keeps track of votes

	totalVotingPower uint64 // cached value for quick lookups
	majorityPower    uint64 // cached value for quick lookups
}

// NewPoA creates a new empty Proof of Authority validator set
func NewPoA(opts ...Option) *PoA {
	// Create the empty set
	p := &PoA{
		validators:              make([]*types.Validator, 0),
		addressToValidatorIndex: make(map[std.Address]int),
		votes:                   newVotingSystem(),
	}

	// Apply the options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// AddValidator adds a vote to add a new validator to the validator set.
//
// Criteria for being added:
// - caller is a user account
// - caller is a validator
// - the proposed address is not a validator
// - a 2/3+ majority of the set have voted to add
func (p *PoA) AddValidator(address std.Address, pubKey string) *types.Validator {
	caller := std.GetOrigCaller()

	// Validate that the operation is a valid call
	p.validateAdd(caller, address)

	// Cast the vote
	voterPower := p.validators[p.addressToValidatorIndex[caller]].VotingPower
	p.votes.castVote(caller, address, voterPower, true)

	// Emit the vote event
	std.Emit(UserVotedAddEvent)

	// Calculate the votes
	addCount, _ := p.votes.getTally(address)

	// Check if there is a majority
	// to apply the action
	if addCount <= p.majorityPower {
		// No super majority to add
		return nil
	}

	// Execute the decision
	return p.addValidator(address, pubKey)
}

// validateAdd validates a validator add call
func (p *PoA) validateAdd(caller, address std.Address) {
	// Check if the caller is a user account
	if !std.PrevRealm().IsUser() {
		panic(types.ErrCallerNotUserAccount)
	}

	// Check if the caller is in the set
	if !p.IsValidator(caller) {
		panic(types.ErrValidatorMissing)
	}

	// Check if the validator is already in the set
	if p.IsValidator(address) {
		panic(types.ErrValidatorExists)
	}
}

// addValidator adds the given validator to the PoA validator set
func (p *PoA) addValidator(address std.Address, pubKey string) *types.Validator {
	v := &types.Validator{
		Address:     address,
		PubKey:      pubKey, // TODO: in the future, verify the public key
		VotingPower: 1,      // in this PoA system, all new validators have the same voting power
	}

	// Add the validator to the set
	p.addressToValidatorIndex[v.Address] = len(p.validators)
	p.validators = append(p.validators, v)

	// Update the total voting power
	p.totalVotingPower += v.VotingPower
	p.majorityPower = (2 * p.totalVotingPower) / 3

	// Emit the validator set change
	std.Emit(types.ValidatorAddedEvent)

	// Remove the candidate from the voting system
	p.votes.resetCandidate(v.Address)

	return v
}

// RemoveValidator adds a vote to remove a validator from the validator set.
//
// Criteria for being added:
// - caller is a user account
// - caller is a validator
// - the proposed address is a validator
// - a 2/3+ majority of the set have voted to remove
func (p *PoA) RemoveValidator(address std.Address) *types.Validator {
	caller := std.GetOrigCaller()

	// Validate that the operation is a valid call
	p.validateRemove(caller, address)

	// Cast the vote
	voterPower := p.validators[p.addressToValidatorIndex[caller]].VotingPower
	p.votes.castVote(caller, address, voterPower, false)

	// Emit the vote event
	std.Emit(UserVotedRemoveEvent)

	// Calculate the votes
	_, removeCount := p.votes.getTally(address)

	// Check if there is a majority
	// to apply the action
	if removeCount <= p.majorityPower {
		// No super majority to remove
		return nil
	}

	// Execute the decision
	return p.removeValidator(address)
}

// validateRemove validates a validator remove call
func (p *PoA) validateRemove(caller, address std.Address) {
	// Check if the caller is a user account
	if !std.PrevRealm().IsUser() {
		panic(types.ErrCallerNotUserAccount)
	}

	// Check if this request came from a validator
	if !p.IsValidator(caller) {
		panic(errCallerNotValidator)
	}

	// Check if the address is a validator
	if !p.IsValidator(address) {
		panic(types.ErrValidatorMissing)
	}
}

// removeValidator removes the given address from the PoA validator set
func (p *PoA) removeValidator(address std.Address) *types.Validator {
	// Fetch the validator index
	index := p.addressToValidatorIndex[address]

	// Remove the validator from the set
	validator := p.validators[index]
	p.validators = append(p.validators[:index], p.validators[index+1:]...)

	delete(p.addressToValidatorIndex, address)

	// Update the total voting power
	p.totalVotingPower -= validator.VotingPower
	p.majorityPower = (2 * p.totalVotingPower) / 3

	// Emit the validator set change
	std.Emit(types.ValidatorRemovedEvent)

	// Remove the candidate from the voting system
	p.votes.resetCandidate(address)

	return validator
}

// IsValidator returns a flag indicating if the address
// is part of the staked validator set
func (p *PoA) IsValidator(address std.Address) bool {
	_, exists := p.addressToValidatorIndex[address]

	return exists
}

// GetValidators returns the current staked validator set
func (p *PoA) GetValidators() []*types.Validator {
	return p.validators
}
