package gvotes

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/governance/checkpoints"
	std "github.com/gnolang/gno/stdlibs/stdshim"
)

const zeroAddress = std.Address("")

// this is only used to track total voting power, by moveVotingPower,
// trigger by mint or tranfer ahead
// it's independant with XToken, implement free
type Votes struct {
	totalCheckpoints *checkpoints.History
}

func NewVotes() *Votes {
	return &Votes{
		totalCheckpoints: &checkpoints.History{},
	}
}

func (v *Votes) GetPastTotalSupply(blockNumber int64) uint64 {
	height := std.GetHeight()
	if blockNumber >= height {
		panic("Votes: block not yet mined")
	}
	return v.totalCheckpoints.GetAtBlock(blockNumber)
}

func (v *Votes) getTotalSupply() uint64 {
	return v.totalCheckpoints.Latest()
}

// tracking total voting power
// only when minting or burning
func (v *Votes) UpdateTotalVotingPower(from std.Address, to std.Address, amount uint64) error {
	if from == to {
		return ErrCannotTransferToSelf
	}
	if from == zeroAddress { // not mint
		v.totalCheckpoints.PushWithOp(add, amount) // if mint
		return nil
	}
	if to == zeroAddress {
		v.totalCheckpoints.PushWithOp(subtract, amount)
	}
	return nil
}

func add(a uint64, b uint64) uint64 {
	return a + b
}

func subtract(a uint64, b uint64) uint64 {
	return a - b
}

func checkIsValidAddress(addr std.Address) {
	if addr.String() == "" {
		panic("invalid address")
	}
	return
}

func emit(event interface{}) {
	// TODO: should we do something there?
	// noop
}
