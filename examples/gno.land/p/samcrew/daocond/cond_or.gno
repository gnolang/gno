package daocond

import (
	"math"
	"strings"
)

type orCond struct {
	conditions []Condition
}

// Creates a new condition from multiple conditions where any one must be true.
//
// Example: Or(MembersThreshold(0.6), RoleCount(2, "admin"))
// Requires either 60% member approval OR 2 admin votes.
func Or(conditions ...Condition) Condition {
	if len(conditions) < 2 {
		panic("at least two conditions are required")
	}
	return &orCond{conditions: conditions}
}

// Checks if any condition is true.
func (c *orCond) Eval(ballot Ballot) bool {
	for _, condition := range c.conditions {
		if condition.Eval(ballot) {
			return true
		}
	}
	return false
}

// Returns highest signal among all conditions between 0.0 and 1.0.
func (c *orCond) Signal(ballot Ballot) float64 {
	maxSignal := 0.0
	for _, condition := range c.conditions {
		maxSignal = math.Max(maxSignal, condition.Signal(ballot))
	}
	return maxSignal
}

// Displays the condition as text.
// Example output: "[ 60% Members OR 2 admin ]"
func (c *orCond) Render() string {
	renders := []string{}
	for _, condition := range c.conditions {
		renders = append(renders, condition.Render())
	}
	return "**[** " + strings.Join(renders, " **OR** ") + " **]**"
}

// Displays the condition with current vote counts.
// Example output: "[ 3/5 Members OR 1/2 admin ]"
func (c *orCond) RenderWithVotes(ballot Ballot) string {
	renders := []string{}
	for _, condition := range c.conditions {
		renders = append(renders, condition.RenderWithVotes(ballot))
	}
	return "**[** " + strings.Join(renders, " **OR** ") + " **]**"
}

var _ Condition = (*orCond)(nil)
