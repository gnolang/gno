package snapshot_token

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/gor/checkpoints"
	"gno.land/p/demo/gor/counters"
	"gno.land/p/demo/ufmt"
	"std"
)

// TODO: might not need all these intermediate representation, since weight of member has been snapshotted
const zeroAddress = std.Address("")

// this represents snapshot tokens of account, 1:1
type SnapshotToken struct {
	accountCheckpoints *avl.Tree // address => checkpoints.History
	totalCheckpoints   *checkpoints.History
	nonces             *avl.Tree // address => counters.Counter
}

func NewSnapShotToken() *SnapshotToken {
	return &SnapshotToken{
		accountCheckpoints: avl.NewTree(),
		totalCheckpoints:   &checkpoints.History{},
		nonces:             avl.NewTree(),
	}
}

// Returns the current amount of voting power that `account` has.
func (v *SnapshotToken) GetAmount(account std.Address) uint64 {
	res, found := v.accountCheckpoints.Get(account.String())
	if !found {
		//panic("SnapshotToken: " + account.String() + " not exist!")
		return 0
	}
	history := res.(*checkpoints.History)
	return history.LatestValue()
}

// get voting power at specific height
func (v *SnapshotToken) GetPastAmount(account std.Address, blockNumber int64) uint64 {
	history, found := v.accountCheckpoints.Get(account.String())
	if !found {
		panic("SnapshotToken: " + account.String() + " not exsit!")
	}
	return history.(*checkpoints.History).GetAtBlock(blockNumber)
}

func (v *SnapshotToken) GetPastTotalSupply(blockNumber int64) uint64 {
	height := std.GetHeight()
	if blockNumber >= height {
		panic("SnapshotToken: block not yet mined")
	}
	return v.totalCheckpoints.GetAtBlock(blockNumber)
}

func (v *SnapshotToken) GetTotalSupply() uint64 {
	return v.totalCheckpoints.LatestValue()
}

// Transfers, mints, or burns voting units. To register a mint, `from` should be zero. To register a burn, `to`
// should be zero. Total supply of voting units will be adjusted with mints and burns.
func (v *SnapshotToken) TransferVotingUnits(from std.Address, to std.Address, amount uint64) error {
	if from == to {
		return ErrCannotTransferToSelf
	}
	if from != zeroAddress { // not mint
		// check from votes power
		fromVotes := v.GetAmount(from)
		if fromVotes < amount {
			return ErrInsufficientVotes
		}
	} else {
		v.totalCheckpoints.PushWithOp(add, amount) // if mint
	}
	if to == zeroAddress {
		v.totalCheckpoints.PushWithOp(subtract, amount)
	}
	v.moveDelegateVotes(from, to, amount)
	return nil
}

// Moves delegated votes from one delegatee to another.
func (v *SnapshotToken) moveDelegateVotes(from std.Address, to std.Address, amount uint64) {
	if from != to && amount > 0 {
		// NOTE: could delegate from zero address while minting
		// TODO: should not delegate to zero address? maybe we can burn it
		// checkIsValidAddress(to)

		// subtract
		if from != zeroAddress {
			var oldValue, newValue uint64
			res, exists := v.accountCheckpoints.Get(from.String())
			if !exists {
				history := &checkpoints.History{}
				oldValue, newValue = history.PushWithOp(subtract, amount)
				v.accountCheckpoints.Set(from.String(), history)
			} else {
				history := res.(*checkpoints.History)
				oldValue, newValue = history.PushWithOp(subtract, amount)
				v.accountCheckpoints.Set(from.String(), history)
			}
			emit(&DelegateVotesChangedEvent{from, oldValue, newValue})
		}
		// add
		if to != zeroAddress {
			var oldValue, newValue uint64
			res, exists := v.accountCheckpoints.Get(to.String())
			if !exists {
				history := &checkpoints.History{}
				oldValue, newValue = history.PushWithOp(add, amount)
				v.accountCheckpoints.Set(to.String(), history)
			} else {
				history := res.(*checkpoints.History)
				oldValue, newValue = history.PushWithOp(add, amount)
				v.accountCheckpoints.Set(to.String(), history)
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
func (v *SnapshotToken) useNonce(owner std.Address) uint64 {
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
func (v *SnapshotToken) getNonce(owner std.Address) uint64 {
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

func (v *SnapshotToken) RenderHome(account std.Address) string {
	totalSupply := v.GetTotalSupply()
	amount := v.GetAmount(account)
	str := ""
	str += ufmt.Sprintf("**Total supply**: %d\n", totalSupply)
	str += ufmt.Sprintf("**amount of %s is**: %d\n", account.String(), amount)
	return str
}

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}
