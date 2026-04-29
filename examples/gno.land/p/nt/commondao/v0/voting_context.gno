package commondao

import "errors"

// VotingContext contains current voting context, which includes a voting
// record for the votes that has been submitted for a proposal and also
// the list of DAO members. Context is readonly.
type VotingContext struct {
	VotingRecord ReadonlyVotingRecord
	Members      MemberStorage
}

// NewVotingContext creates a new voting context.
func NewVotingContext(r *VotingRecord, s MemberStorage) (VotingContext, error) {
	if r == nil {
		return VotingContext{}, errors.New("voting record is required")
	}

	if s == nil {
		return VotingContext{}, errors.New("member storage is required")
	}

	if !s.IsSealed() {
		return VotingContext{}, errors.New("member storage is not sealed")
	}

	return VotingContext{
		VotingRecord: r.Readonly(),
		Members:      s,
	}, nil
}

// MustNewVotingContext creates a new voting context or panics on error.
func MustNewVotingContext(r *VotingRecord, s MemberStorage) VotingContext {
	ctx, err := NewVotingContext(r, s)
	if err != nil {
		panic(err)
	}
	return ctx
}
