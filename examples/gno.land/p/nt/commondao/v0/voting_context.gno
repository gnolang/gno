package commondao

import "gno.land/p/moul/addrset"

// VotingContext contains current voting context, which includes a voting
// record with the votes that has been submitted for a proposal and also
// the proposal's electorate (the council snapshot taken at creation).
type VotingContext struct {
	VotingRecord ReadonlyVotingRecord
	Council      *addrset.ReadonlySet
}

// NewVotingContext creates a new voting context.
func NewVotingContext(r ReadonlyVotingRecord, council *addrset.ReadonlySet) VotingContext {
	return VotingContext{
		VotingRecord: r,
		Council:      council,
	}
}
