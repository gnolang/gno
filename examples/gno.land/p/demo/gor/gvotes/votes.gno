package gvotes

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/governance/checkpoints"
	"gno.land/p/demo/governance/counters"
	"gno.land/p/demo/ufmt"
	"std"
)

const zeroAddress = std.Address("")

// this represents snapshot tokens of account, 1:1
type snapshotVotingPower struct {
	votingPowerCheckpoints *avl.MutTree // address => checkpoints.History
	totalCheckpoints       *checkpoints.History
	nonces                 *avl.MutTree // address => counters.Counter
}

func NewSnapShotVotingPower() *snapshotVotingPower {
	return &snapshotVotingPower{
		delegateCheckpoints: avl.NewMutTree(),
		totalCheckpoints:    &checkpoints.History{},
		nonces:              avl.NewMutTree(),
	}
}

// Returns the current amount of voting power that `account` has.
func (v *snapshotVotingPower) GetVotingPower(account std.Address) uint64 {
	res, found := v.votingPowerCheckpoints.Get(account.String())
	if !found {
		panic("snapshotVotingPower: " + account.String() + " not exist!")
	}
	history := res.(*checkpoints.History)
	return history.Latest()
}

// get voting power at specific height
func (v *snapshotVotingPower) GetPastVotingPower(account std.Address, blockNumber int64) uint64 {
	history, found := v.delegateCheckpoints.Get(account.String())
	if !found {
		panic("snapshotVotingPower: " + account.String() + " not exsit!")
	}
	return history.(*checkpoints.History).GetAtBlock(blockNumber)
}

func (v *snapshotVotingPower) GetPastTotalSupply(blockNumber int64) uint64 {
	height := std.GetHeight()
	if blockNumber >= height {
		panic("snapshotVotingPower: block not yet mined")
	}
	return v.totalCheckpoints.GetAtBlock(blockNumber)
}

func (v *snapshotVotingPower) getTotalSupply() uint64 {
	return v.totalCheckpoints.Latest()
}

// Transfers, mints, or burns voting units. To register a mint, `from` should be zero. To register a burn, `to`
// should be zero. Total supply of voting units will be adjusted with mints and burns.
func (v *snapshotVotingPower) TransferVotingUnits(from std.Address, to std.Address, amount uint64) error {
	if from == to {
		return ErrCannotTransferToSelf
	}
	if from != zeroAddress { // not mint
		// check from votes power
		fromVotes := v.GetVotes(from)
		if fromVotes < amount {
			return ErrInsufficientVotes
		}
	} else {
		v.totalCheckpoints.PushWithOp(add, amount) // if mint
	}
	if to == zeroAddress {
		v.totalCheckpoints.PushWithOp(subtract, amount)
	}
	v.moveDelegateVotes(v.delegates(from), v.delegates(to), amount)
	return nil
}

// Moves delegated votes from one delegatee to another.
func (v *snapshotVotingPower) moveDelegateVotes(from std.Address, to std.Address, amount uint64) {
	if from != to && amount > 0 {
		// NOTE: could delegate from zero address while minting
		// TODO: should not delegate to zero address? maybe we can burn it
		// checkIsValidAddress(to)

		// subtract
		if from != zeroAddress {
			var oldValue, newValue uint64
			res, exists := v.delegateCheckpoints.Get(from.String())
			if !exists {
				history := &checkpoints.History{}
				oldValue, newValue = history.PushWithOp(subtract, amount)
				v.delegateCheckpoints.Set(from.String(), history)
			} else {
				history := res.(*checkpoints.History)
				oldValue, newValue = history.PushWithOp(subtract, amount)
				v.delegateCheckpoints.Set(from.String(), history)
			}
			emit(&DelegateVotesChangedEvent{from, oldValue, newValue})
		}
		// add
		if to != zeroAddress {
			var oldValue, newValue uint64
			res, exists := v.delegateCheckpoints.Get(to.String())
			if !exists {
				history := &checkpoints.History{}
				oldValue, newValue = history.PushWithOp(add, amount)
				v.delegateCheckpoints.Set(to.String(), history)
			} else {
				history := res.(*checkpoints.History)
				oldValue, newValue = history.PushWithOp(add, amount)
				v.delegateCheckpoints.Set(to.String(), history)
			}
			emit(&DelegateVotesChangedEvent{to, oldValue, newValue})
		}
	}
}

func add(a uint64, b uint64) uint64 {
	return a + b
}

func subtract(a uint64, b uint64) uint64 {
	return a - b
}

// Consumes a nonce.
// Returns the current value and increments nonce.
func (v *snapshotVotingPower) useNonce(owner std.Address) uint64 {
	res, found := v.nonces.Get(owner.String())
	if !found {
		return 0 // default 0
	}
	nonce := res.(counters.Counter)
	current := nonce.Current()
	nonce.Increment()
	v.nonces.Set(owner.String(), nonce)
	return current
}

// Returns an address nonce.
func (v *snapshotVotingPower) getNonce(owner std.Address) uint64 {
	res, found := v.nonces.Get(owner.String())
	if !found {
		return 0 // default 0
	}
	nonce := res.(counters.Counter)
	return nonce.Current()
}

// Must return the voting units held by an account.
// func getVotingUnits(account std.Address) uint64 {
// 	// TODO: should return balances[account]
// }

func checkIsValidAddress(addr std.Address) {
	if addr.String() == "" {
		panic("invalid address")
	}
	return
}

func (v *snapshotVotingPower) RenderHome(account std.Address) string {
	totalSupply := v.getTotalSupply()
	votes := v.GetVotes(account)
	delegatee := v.delegates(account)
	str := ""
	str += ufmt.Sprintf("**Total supply**: %d\n", totalSupply)
	str += ufmt.Sprintf("**votes of %s is**: %d\n", account.String(), votes)
	str += ufmt.Sprintf("**delegatee of %s is**: %s\n", account.String(), delegatee.String())
	return str
}

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}
