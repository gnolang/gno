package daocond

import (
	"strings"
)

type andCond struct {
	conditions []Condition
}

// Creates a new condition from multiple conditions
// where all must be true to pass.
//
// Example: And(MembersThreshold(0.6), RoleCount(2, "admin"))
// Requires both 60% member approval AND 2 admin votes.
func And(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("at least two conditions are required")
	}
	return &andCond{conditions: conditions}
}

// Checks if all conditions are true.
func (c *andCond) Eval(ballot Ballot) bool {
	for _, condition := range c.conditions {
		if !condition.Eval(ballot) {
			return false
		}
	}
	return true
}

// Returns average signal across all conditions between 0.0 and 1.0.
func (c *andCond) Signal(ballot Ballot) float64 {
	totalSignal := 0.0
	for _, condition := range c.conditions {
		totalSignal += condition.Signal(ballot)
	}
	return totalSignal / float64(len(c.conditions))
}

// Displays the condition as text.
// Example output: "[ 60% Members AND 2 admin ]"
func (c *andCond) Render() string {
	renders := []string{}
	for _, condition := range c.conditions {
		renders = append(renders, condition.Render())
	}
	return "**[** " + strings.Join(renders, " **AND** ") + " **]**"
}

// Displays the condition with current vote counts.
// Example output: "[ 3/5 Members AND 1/2 admin ]"
func (c *andCond) RenderWithVotes(ballot Ballot) string {
	renders := []string{}
	for _, condition := range c.conditions {
		renders = append(renders, condition.RenderWithVotes(ballot))
	}
	return "**[** " + strings.Join(renders, " **AND** ") + " **]**"
}

var _ Condition = (*andCond)(nil)
