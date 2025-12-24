package daocond

import (
	"gno.land/p/nt/avl"
)

type BallotAVL struct {
	tree *avl.Tree
}

// Creates a new ballot implementation using AVL tree.
func NewBallot() Ballot {
	return &BallotAVL{tree: avl.NewTree()}
}

// Records a vote from a specific voter, overwriting any previous vote.
func (b *BallotAVL) Vote(voter string, vote Vote) {
	b.tree.Set(voter, vote)
}

// Gets the vote from a specific voter, returns VoteAbstain if no vote recorded.
func (b *BallotAVL) Get(voter string) Vote {
	vote, ok := b.tree.Get(voter)
	if !ok {
		return VoteAbstain
	}
	return vote.(Vote)
}

// Returns the total number of votes cast.
func (b *BallotAVL) Total() int {
	return b.tree.Size()
}

// Iterates over all votes, calling fn for each voter-vote pair.
func (b *BallotAVL) Iterate(fn func(voter string, vote Vote) bool) {
	b.tree.Iterate("", "", func(key string, value interface{}) bool {
		v, ok := value.(Vote)
		if !ok {
			return false
		}
		return fn(key, v)
	})
}
