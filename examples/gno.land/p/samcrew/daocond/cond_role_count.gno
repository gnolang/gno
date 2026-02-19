package daocond

import (
	"errors"
	"math"

	"gno.land/p/nt/ufmt"
)

type roleCountCond struct {
	hasRoleFn func(memberId string, role string) bool
	count     uint64
	role      string
}

// Creates a condition requiring a specific number of role holders to vote yes.
//
// Example: RoleCount(2, "admin", hasRoleFn)
// Requires 2 admin votes.
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

// Checks if enough role holders voted yes to meet the count.
func (c *roleCountCond) Eval(ballot Ballot) bool {
	return c.totalVote(ballot, VoteYes) >= c.count
}

// Returns progress toward meeting the required count between 0.0 and 1.0.
func (c *roleCountCond) Signal(ballot Ballot) float64 {
	return math.Min(float64(c.totalVote(ballot, VoteYes))/float64(c.count), 1)
}

// Displays the condition as text.
// Example output: "2 admin"
func (c *roleCountCond) Render() string {
	return ufmt.Sprintf("%d %s", c.count, c.role)
}

// Displays the condition with current vote counts.
// Example output: "2 admin must vote yes\n\nYes: 1/2"
func (c *roleCountCond) RenderWithVotes(ballot Ballot) string {
	s := ""
	s += ufmt.Sprintf("%d members with role %s must vote yes\n\n", c.count, c.role)
	s += ufmt.Sprintf("Yes: %d/%d\n\n", c.totalVote(ballot, VoteYes), c.count)
	s += ufmt.Sprintf("No: %d/%d\n\n", c.totalVote(ballot, VoteNo), c.count)
	s += ufmt.Sprintf("Abstain: %d/%d\n\n", c.totalVote(ballot, VoteAbstain), c.count)
	return s
}

var _ Condition = (*roleCountCond)(nil)

func (c *roleCountCond) totalVote(ballot Ballot, vote Vote) uint64 {
	total := uint64(0)
	ballot.Iterate(func(voter string, v Vote) bool {
		if v == vote && c.hasRoleFn(voter, c.role) {
			total += 1
		}
		return false
	})
	return total
}
