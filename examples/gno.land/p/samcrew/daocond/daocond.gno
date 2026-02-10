package daocond

// We made conditions stateless to avoid complexity of handling events like (member added, member removed, role assigned, role removed, etc.)
// This model should work pretty good for small DAOs with few voters.
// In future, we should probably add a stateful model for large DAOs with many voters.

type Condition interface {
	Eval(ballot Ballot) bool
	Signal(ballot Ballot) float64

	Render() string
	RenderWithVotes(ballot Ballot) string
}

type Ballot interface {
	Vote(voter string, vote Vote)
	Get(voter string) Vote
	Total() int

	Iterate(fn func(voter string, vote Vote) bool)
}

type Vote string

const (
	VoteYes     = "yes"
	VoteNo      = "no"
	VoteAbstain = "abstain"
)
