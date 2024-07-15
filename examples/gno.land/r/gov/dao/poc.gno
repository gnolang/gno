package govdao

import (
	"std"

	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

const daoPkgPath = "gno.land/r/gov/dao"

const (
	errNoChangesProposed = "no set changes proposed"
	errNotGovDAO         = "caller not govdao executor"
)

// NewMemberPropExecutor returns the GOVDAO member change executor
func NewMemberPropExecutor(changesFn func() []dao.Member) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		// Make sure the GovDAO executor runs the member changes
		assertGovDAOCaller()

		errs := &executor.CombinedError{}
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

	return executor.NewCallbackExecutor(callback)
}

func NewDAOImplExecutor(changeFn func() dao.DAO) dao.Executor {
	if changeFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		assertGovDAOCaller()

		setDAOImpl(changeFn())

		return nil
	}

	return executor.NewCallbackExecutor(callback)
}

func NewMembStoreImplExecutor(changeFn func() dao.MemberStore) dao.Executor {
	if changeFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		assertGovDAOCaller()

		setMembStoreImpl(changeFn())

		return nil
	}

	return executor.NewCallbackExecutor(callback)
}

func NewPropStoreImplExecutor(changeFn func() dao.PropStore) dao.Executor {
	if changeFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		assertGovDAOCaller()

		setPropStoreImpl(changeFn())

		return nil
	}

	return executor.NewCallbackExecutor(callback)
}

// setDAOImpl sets a new DAO implementation
func setDAOImpl(impl dao.DAO) {
	d = impl
}

// setMembStoreImpl sets a new dao.MembStore implementation
func setMembStoreImpl(impl dao.MemberStore) {
	members = impl
}

// setPropStoreImpl sets a new dao.PropStore implementation
func setPropStoreImpl(impl dao.PropStore) {
	proposals = impl
}

// assertGovDAOCaller verifies the caller is the GovDAO executor
func assertGovDAOCaller() {
	if std.CurrentRealm().PkgPath() != daoPkgPath {
		panic(errNotGovDAO)
	}
}
