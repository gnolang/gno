package daocond

import (
	"errors"
	"math"

	"gno.land/p/nt/ufmt/v0"
)

type roleCountCond struct {
	hasRoleFn func(memberId string, role string) bool
	count     uint64
	role      string
}

func RoleCount(count uint64, role string, hasRoleFn func(memberId string, role string) bool) Condition {
	if count == 0 {
		panic(errors.New("count must be greater than 0"))
	}
	if role == "" {
		panic(errors.New("role must not be empty"))
	}
	if hasRoleFn == nil {
		panic(errors.New("nil hasRoleFn"))
	}
	return &roleCountCond{
		count:     count,
		hasRoleFn: hasRoleFn,
		role:      role,
	}
}

// Eval implements Condition.
func (c *roleCountCond) Eval(ballot Ballot) bool {
	return c.totalYes(ballot) >= c.count
}

// Signal implements Condition.
func (c *roleCountCond) Signal(ballot Ballot) float64 {
	return math.Min(float64(c.totalYes(ballot))/float64(c.count), 1)
}

// Render implements Condition.
func (c *roleCountCond) Render() string {
	return ufmt.Sprintf("%d %s", c.count, c.role)
}

// RenderWithVotes implements Condition.
func (c *roleCountCond) RenderWithVotes(ballot Ballot) string {
	s := ""
	s += ufmt.Sprintf("to meet the condition, %d members with role %s must vote yes\n\n", c.count, c.role)
	s += ufmt.Sprintf("Yes: %d/%d\n\n", c.totalYes(ballot), c.count)
	return s
}

var _ Condition = (*roleCountCond)(nil)

func (c *roleCountCond) totalYes(ballot Ballot) uint64 {
	totalYes := uint64(0)
	ballot.Iterate(func(voter string, vote Vote) bool {
		if vote == VoteYes && c.hasRoleFn(voter, c.role) {
			totalYes += 1
		}
		return false
	})
	return totalYes
}
