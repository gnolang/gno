package daocond

import (
	"gno.land/p/nt/avl/v0"
)

type BallotAVL struct {
	tree *avl.Tree
}

func NewBallot() Ballot {
	return &BallotAVL{tree: avl.NewTree()}
}

// Implements Ballot interface
func (b *BallotAVL) Vote(voter string, vote Vote) {
	b.tree.Set(voter, vote)
}

// Implements Ballot interface
func (b *BallotAVL) Get(voter string) Vote {
	vote, ok := b.tree.Get(voter)
	if !ok {
		return VoteAbstain
	}
	return vote.(Vote)
}

// Implements Ballot interface
func (b *BallotAVL) Total() int {
	return b.tree.Size()
}

// Implements Ballot interface
func (b *BallotAVL) Iterate(fn func(voter string, vote Vote) bool) {
	b.tree.Iterate("", "", func(key string, value interface{}) bool {
		v, ok := value.(Vote)
		if !ok {
			return false
		}
		return fn(key, v)
	})
}
