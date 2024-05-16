package pos

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/r/sys/vals/types"
)

var stakeThreshold = std.Coin{"ugnot", int64(100000000)} // 100 GNOTs

const (
	minValidatorCount = 4   // the absolute minimum number of validators in the set
	maxValidatorCount = 150 // the practical maximum number of validators in the set
)

const (
	UserStakedEvent   = "UserStaked"   // emitted when a user stakes funds
	UserUnstakedEvent = "UserUnstaked" // emitted when a user unstakes funds
)

// PoS specifies the Proof of Stake validator set.
// In order to become part of the set, users need to stake funds (at least at the threshold).
// Upon leaving the set, users receive their staked funds back
type PoS struct {
	// validators holds the current validator set.
	// This slice can never practically grow more than ~150 elements,
	// due to Tendermint's quadratic network complexity
	validators              []types.Validator
	addressToValidatorIndex map[std.Address]int      // address -> index
	addressToStakedAmount   map[std.Address]std.Coin // address -> staked amount
}

// NewPoS creates a new empty Proof of Stake validator set
func NewPoS(opts ...Option) *PoS {
	// Create the empty set
	p := &PoS{
		validators:              make([]types.Validator, 0),
		addressToValidatorIndex: make(map[std.Address]int),
		addressToStakedAmount:   make(map[std.Address]std.Coin),
	}

	// Apply the options
	for _, opt := range opts {
		opt(p)
	}

	return p
}

// AddValidator adds a new validator to the validator set.
//
// Criteria for being added:
// - caller is a user account
// - caller is not already a validator
// - caller has staked funds >= the current staking threshold
// - the validator set is less than the maximum size
func (p *PoS) AddValidator(address, pubKey string, votingPower uint64) {
	// Check if the caller is a user account
	if !std.IsOriginCall() {
		panic("caller not user account")
	}

	// Check if the validator is already in the set
	if p.IsValidator(address) {
		panic("validator already exists")
	}

	// Check if the limit is reached
	if len(p.validators) <= maxValidatorCount {
		panic("maximum validator count reached")
	}

	var (
		caller       = std.GetOrigCaller()
		stakedAmount = std.GetOrigSend().AmountOf("ugnot")
	)

	// Check if the caller is supplying their own address
	if address != caller.String() {
		panic("validator address is not origin")
	}

	// Fetch the already staked amount
	addressStake, exists := p.addressToStakedAmount[caller]
	if !exists {
		addressStake = std.Coin{"ugnot", int64(0)}
	}

	// Save the staked amount
	addressStake.Amount += stakedAmount
	p.addressToStakedAmount[caller] = addressStake

	// Emit the event that the user staked funds
	std.Emit(
		UserStakedEvent,
		"address", address,
		"stake", ufmt.Sprintf("%d", stakedAmount.Amount),
	)

	// Check if the caller can become a validator
	if !p.canBecomeValidator(caller) {
		return
	}

	// Add the caller to the validator set
	v := types.Validator{
		Address:     caller,
		PubKey:      pubKey,      // TODO check public key?
		VotingPower: votingPower, // TODO validate voting power?
	}

	p.addressToValidatorIndex[caller] = len(p.validators)
	p.validators = append(p.validators, v)

	// Emit the validator set change
	std.Emit(
		types.ValidatorAddedEvent,
		"address", address,
	)
}

// canBecomeValidator checks if the address fulfills all criteria for
// becoming a validator:
// - is not already a validator
// - has staked >= the staking threshold
func (p *PoS) canBecomeValidator(address std.Address) bool {
	stake, _ := p.addressToStakedAmount[address]

	return !p.IsValidator(address.String()) && stake.IsAllGTE(stakeThreshold)
}

// RemoveValidator removes a validator from the validator set.
// Upon successful removal, the staked funds are returned to the user
//
// Criteria for being removed:
// - caller is a user account
// - caller is a validator
// - the validator set is more than the minimum size
func (p *PoS) RemoveValidator(address string) {
	// Check if the caller is a user account
	if !std.IsOriginCall() {
		panic("caller not user account")
	}

	// Check if this request came from a validator
	if !p.IsValidator(std.GetOrigCaller()) {
		panic("user is not validator")
	}

	// Check if the limit is reached
	if len(p.validators) == minValidatorCount {
		panic("minimum validator count reached")
	}

	var (
		caller       = std.GetOrigCaller()
		addressStake = p.addressToStakedAmount[caller]
	)

	// Check if the caller is supplying their own address
	if address != caller.String() {
		panic("validator address is not origin")
	}

	// Check if the validator is in the set
	index := addressToValidatorIndex[caller]

	// Remove the validator from the set
	p.validators = append(p.validators[:index], p.validators[index+1:]...)

	delete(p.addressToValidatorIndex, caller)
	delete(p.addressToStakedAmount, caller)

	// Return the stake
	returnStake(caller, addressStake)

	// Emit the validator set change
	std.Emit(
		types.ValidatorRemovedEvent,
		"address", address,
	)

	// Emit the unstake event
	std.Emit(
		UserUnstakedEvent,
		"address", address,
		"stake", ufmt.Sprintf("%d", addressStake.Amount),
	)
}

// returnStake returns the specified stake to the given address
func returnStake(address std.Address, amount std.Coin) {
	// Derive the current package address
	from := std.DerivePkgAddr(std.CurrentRealm())

	// Fetch the banker
	banker := std.GetBanker(std.BankerTypeRealmSend)

	// Return the staked funds
	banker.SendCoins(from, address, std.Coins{amount})
}

// IsValidator returns a flag indicating if the address
// is part of the staked validator set
func (p *PoS) IsValidator(address string) bool {
	_, exists := p.addressToValidatorIndex[std.Address(address)]

	return exists
}

// GetValidators returns the current staked validator set
func (p *PoS) GetValidators() []types.Validator {
	return p.validators
}
