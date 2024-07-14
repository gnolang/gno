package govdao

import (
	"std"

	"gno.land/p/demo/context"
	"gno.land/p/gov/dao"
	"gno.land/p/gov/proposal"
	pproposal "gno.land/p/gov/proposal"
)

const daoPkgPath = "gno.land/r/gov/dao"

const (
	errNoChangesProposed = "no set changes proposed"
	errNotGovDAO         = "caller not govdao executor"
)

// NewMemberPropExecutor returns the GOVDAO member change executor
func NewMemberPropExecutor(changesFn func() []dao.Member) pproposal.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		// Make sure the GovDAO executor runs the member changes
		assertGovDAOCaller()

		errs := &pproposal.CombinedErrors{}
		for _, member := range changesFn() {
			switch {
			case !members.IsMember(member.Address):
				// Addition request
				err := members.AddMember(member)

				errs.Add(err)
			case member.VotingPower == 0:
				// Remove request
				err := members.UpdateMember(member.Address, dao.Member{
					Address:     member.Address,
					VotingPower: 0, // 0 indicated removal
				})

				errs.Add(err)
			default:
				// Update request
				err := members.UpdateMember(member.Address, member)

				errs.Add(err)
			}
		}

		// Check if there were any execution errors
		if errs.Size() == 0 {
			return nil
		}

		return errs
	}

	return pproposal.NewExecutor(callback)
}

// SetDAOImpl sets a new DAO implementation
func SetDAOImpl(ctx context.Context, impl dao.DAO) {
	proposal.AssertContextApprovedByGovDAO(ctx)

	d = impl
}

// SetMembStoreImpl sets a new dao.MembStore implementation
func SetMembStoreImpl(ctx context.Context, impl dao.MembStore) {
	proposal.AssertContextApprovedByGovDAO(ctx)

	members = impl
}

// SetPropStoreImpl sets a new dao.PropStore implementation
func SetPropStoreImpl(ctx context.Context, impl dao.PropStore) {
	proposal.AssertContextApprovedByGovDAO(ctx)

	proposals = impl
}

// assertGovDAOCaller verifies the caller is the GovDAO executor
func assertGovDAOCaller() {
	if std.CurrentRealm().PkgPath() != daoPkgPath {
		panic(errNotGovDAO)
	}
}
