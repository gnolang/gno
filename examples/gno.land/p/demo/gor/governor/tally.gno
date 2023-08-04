package governor

import (
	"gno.land/p/demo/avl"
	"std"
)

type Tally interface {
	Tally(propID string) (*TallyResult, error)
	CastVote(propID string, account std.Address, vp VoteType) error
}

type TallyResult struct {
	tr *avl.Tree // account -> voteType
}

func NewTallyResult() *TallyResult {
	return &TallyResult{
		tr: avl.NewTree(),
	}
}

func (trt *TallyResult) GetVotesByType(vt VoteType) uint64 {
	count, ok := trt.tr.Get(vt.String())
	if ok {
		return count.(uint64)
	}
	return 0
}
