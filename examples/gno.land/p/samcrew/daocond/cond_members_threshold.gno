package daocond

import (
	"errors"
	"math"

	"gno.land/p/nt/ufmt/v0"
)

type membersThresholdCond struct {
	isMemberFn     func(memberId string) bool
	membersCountFn func() uint64
	threshold      float64
}

func MembersThreshold(threshold float64, isMemberFn func(memberId string) bool, membersCountFn func() uint64) Condition {
	if threshold <= 0 || threshold > 1 {
		panic(errors.New("invalid threshold"))
	}
	if isMemberFn == nil {
		panic(errors.New("nil isMemberFn"))
	}
	if membersCountFn == nil {
		panic(errors.New("nil membersCountFn"))
	}
	return &membersThresholdCond{
		threshold:      threshold,
		isMemberFn:     isMemberFn,
		membersCountFn: membersCountFn,
	}
}

// Eval implements Condition.
func (c *membersThresholdCond) Eval(ballot Ballot) bool {
	return c.yesRatio(ballot) >= c.threshold
}

// Signal implements Condition.
func (c *membersThresholdCond) Signal(ballot Ballot) float64 {
	return math.Min(c.yesRatio(ballot)/c.threshold, 1)
}

// Render implements Condition.
func (c *membersThresholdCond) Render() string {
	return ufmt.Sprintf("%g%% of members", c.threshold*100)
}

// RenderWithVotes implements Condition.
func (c *membersThresholdCond) RenderWithVotes(ballot Ballot) string {
	s := ""
	s += ufmt.Sprintf("to meet the condition, %g%% of members must vote yes\n\n", c.threshold*100)
	s += ufmt.Sprintf("Yes: %d/%d = %g%%\n\n", c.totalYes(ballot), c.membersCountFn(), c.yesRatio(ballot)*100)
	return s
}

var _ Condition = (*membersThresholdCond)(nil)

func (c *membersThresholdCond) yesRatio(ballot Ballot) float64 {
	return float64(c.totalYes(ballot)) / float64(c.membersCountFn())
}

func (c *membersThresholdCond) totalYes(ballot Ballot) uint64 {
	totalYes := uint64(0)
	ballot.Iterate(func(voter string, vote Vote) bool {
		if vote == VoteYes && c.isMemberFn(voter) {
			totalYes += 1
		}
		return false
	})
	return totalYes
}
