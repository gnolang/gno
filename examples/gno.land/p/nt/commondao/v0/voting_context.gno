package commondao

import "gno.land/p/moul/addrset"

// VotingContext contains current voting context, which includes a voting
// record with the votes that have been submitted for a proposal and also
// the proposal's electorate (the council snapshot taken at creation).
type VotingContext struct {
	VotingRecord ReadonlyVotingRecord
	Electorate   *addrset.ReadonlySet
}

// NewVotingContext creates a new voting context.
func NewVotingContext(r ReadonlyVotingRecord, electorate *addrset.ReadonlySet) VotingContext {
	return VotingContext{
		VotingRecord: r,
		Electorate:   electorate,
	}
}
