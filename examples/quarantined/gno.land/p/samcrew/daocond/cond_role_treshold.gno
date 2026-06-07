package daocond

import (
	"errors"
	"math"

	"gno.land/p/nt/ufmt/v0"
)

type roleThresholdCond struct {
	hasRoleFn        func(memberId string, role string) bool
	usersRoleCountFn func(role string) uint32
	threshold        float64
	role             string
}

func RoleThreshold(threshold float64, role string, hasRoleFn func(memberId string, role string) bool, usersRoleCountFn func(role string) uint32) Condition {
	if threshold <= 0 || threshold > 1 {
		panic(errors.New("invalid threshold"))
	}
	if hasRoleFn == nil {
		panic(errors.New("nil hasRoleFn"))
	}
	if usersRoleCountFn == nil {
		panic(errors.New("nil usersRoleCountFn"))
	}
	return &roleThresholdCond{
		threshold:        threshold,
		hasRoleFn:        hasRoleFn,
		usersRoleCountFn: usersRoleCountFn,
		role:             role,
	}
}

// Eval implements Condition.
func (c *roleThresholdCond) Eval(ballot Ballot) bool {
	return c.yesRatio(ballot) >= c.threshold
}

// Signal implements Condition.
func (c *roleThresholdCond) Signal(ballot Ballot) float64 {
	return math.Min(c.yesRatio(ballot)/c.threshold, 1)
}

// Render implements Condition.
func (c *roleThresholdCond) Render() string {
	return ufmt.Sprintf("%g%% of %s members", c.threshold*100, c.role)
}

// RenderWithVotes implements Condition.
func (c *roleThresholdCond) RenderWithVotes(ballot Ballot) string {
	s := ""
	s += ufmt.Sprintf("To meet the condition, %g%% of %s members with role %s must vote yes\n\n", c.threshold*100, c.role)
	s += ufmt.Sprintf("Yes: %d/%d = %g%%\n\n", c.totalYes(ballot), c.usersRoleCountFn(c.role), c.yesRatio(ballot)*100)
	return s
}

var _ Condition = (*roleThresholdCond)(nil)

func (c *roleThresholdCond) yesRatio(ballot Ballot) float64 {
	return float64(c.totalYes(ballot)) / float64(c.usersRoleCountFn(c.role))
}

func (c *roleThresholdCond) totalYes(ballot Ballot) uint32 {
	totalYes := uint32(0)
	ballot.Iterate(func(voter string, vote Vote) bool {
		if vote == VoteYes && c.hasRoleFn(voter, c.role) {
			totalYes += 1
		}
		return false
	})
	return totalYes
}
