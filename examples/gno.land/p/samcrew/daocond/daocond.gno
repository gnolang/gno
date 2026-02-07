package daocond

// Package daocond provides a stateless condition system for DAO voting.
//
// This model replaces complex event handlers. It works well for
// small DAOs with few voters, with plans to add stateful models for
// larger DAOs in the future.

// Represents a voting requirement.
type Condition interface {
	// Checks if the condition is satisfied by the current votes.
	Eval(ballot Ballot) bool
	// Returns progress toward satisfying the condition (0.0 to 1.0).
	Signal(ballot Ballot) float64

	// Displays the condition as human-readable text.
	// Example output: "[ 60% Members OR 2 admin ]"
	Render() string
	// Displays the condition with current vote counts.
	// Example output: "[ 3/5 Members OR 1/2 admin ]"
	RenderWithVotes(ballot Ballot) string
}

// Represents a collection of votes that can be evaluated against conditions.
type Ballot interface {
	// Records a vote from a specific voter.
	Vote(voter string, vote Vote)

	// Retrieves the vote from a specific voter.
	Get(voter string) Vote
	// Returns the total number of votes cast.
	Total() int

	// Iterate over all votes, calling fn for each voter/vote pair.
	// If fn returns false, iteration stops.
	// Useful for counting specific vote types or finding particular voters.
	Iterate(fn func(voter string, vote Vote) bool)
}

// Represents a voting decision.
type Vote string

// Available vote options.
const (
	VoteYes     = "yes"
	VoteNo      = "no"
	VoteAbstain = "abstain"
)
