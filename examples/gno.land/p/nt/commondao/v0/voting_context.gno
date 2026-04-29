package commondao

import "errors"

// VotingContext contains current voting context, which includes a voting
// record for the votes that has been submitted for a proposal and also
// the list of DAO members. Context is readonly.
type VotingContext struct {
	VotingRecord VotingRecord
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

	if !r.IsSealed() {
		r, _ = r.Seal().(*VotingRecord)
		if r == nil {
			return VotingContext{}, errors.New("expected voting record to return a sealed copy")
		}
	}

	if !s.IsSealed() {
		s, _ = s.Seal().(MemberStorage)
		if s == nil {
			return VotingContext{}, errors.New("expected member storage to return a sealed copy")
		}
	}

	return VotingContext{
		VotingRecord: *r,
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
