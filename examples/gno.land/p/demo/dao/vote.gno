package dao

// NOTE:
// This voting pods will be removed in a future version of the
// p/demo/dao package. A DAO shouldn't have to comply with or define how the voting mechanism works internally;
// it should be viewed as an entity that makes decisions
//
// The extent of "votes being enforced" in this implementation is just in the context
// of types a DAO can use (import), and in the context of "Stats", where
// there is a notion of "Yay", "Nay" and "Abstain" votes.
const (
	VoteAddedEvent = "VoteAdded" // emitted when a vote was cast for a proposal

	VoteAddedIDKey     = "proposal-id"
	VoteAddedAuthorKey = "author"
	VoteAddedOptionKey = "option"
)

type VoteOption string

const (
	YesVote     VoteOption = "YES"
	NoVote      VoteOption = "NO"
	AbstainVote VoteOption = "ABSTAIN"
)

func (v VoteOption) String() string {
	return string(v)
}

// Stats encompasses the proposal voting stats
type Stats struct {
	YayVotes     uint64
	NayVotes     uint64
	AbstainVotes uint64

	TotalVotingPower uint64
}

// YayPercent returns the percentage (0-100) of the yay votes
// in relation to the total voting power
func (v Stats) YayPercent() uint64 {
	return v.YayVotes * 100 / v.TotalVotingPower
}

// NayPercent returns the percentage (0-100) of the nay votes
// in relation to the total voting power
func (v Stats) NayPercent() uint64 {
	return v.NayVotes * 100 / v.TotalVotingPower
}

// AbstainPercent returns the percentage (0-100) of the abstain votes
// in relation to the total voting power
func (v Stats) AbstainPercent() uint64 {
	return v.AbstainVotes * 100 / v.TotalVotingPower
}

// MissingVotes returns the summed voting power that has not
// participated in proposal voting yet
func (v Stats) MissingVotes() uint64 {
	return v.TotalVotingPower - (v.YayVotes + v.NayVotes + v.AbstainVotes)
}

// MissingVotesPercent returns the percentage (0-100) of the missing votes
// in relation to the total voting power
func (v Stats) MissingVotesPercent() uint64 {
	return v.MissingVotes() * 100 / v.TotalVotingPower
}
