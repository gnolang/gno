package stake

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/avl"
)

// TODO : implement a Grc721 adapter

type StakeOptions struct {
	BondingLockDuration   int64 // time wait for a stake done
	UnbondingLockDuration int64 // time wait for a unbond done
}

func DefaultOptions() *StakeOptions {
	return &StakeOptions{
		UnbondingLockDuration: 21,
		BondingLockDuration:   0,
	}
}

// TODO: param getters

type Stake struct {
	GovToken                    // token staking
	delegations       *avl.Tree // delegatorAddr => DelegatePair
	delegates         *avl.Tree // addr => Delegate, to keep a map of Delegates, i.e, Validators
	unbondDelegations *avl.Tree // delegatorAddr =>  UnbondDelegatePair
}

func NewStake(token GovToken) *Stake {
	stake := &Stake{
		GovToken:          token,
		delegations:       avl.NewTree(),
		delegates:         avl.NewTree(),
		unbondDelegations: avl.NewTree(),
	}
	return stake
}

func (s *Stake) getTotalSupply() uint64 {
	return s.GovToken.TotalSupply()
}

func (s *Stake) SetDelegates(ds *avl.Tree) {
	s.delegates = ds
}
