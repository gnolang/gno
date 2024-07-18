package govdao

import (
	"errors"
	"std"

	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

const (
	daoPkgPath = "gno.land/r/gov/dao/v2"

	errNoChangesProposed = "no set changes proposed"
)

var errNotGovDAO = errors.New("only r/gov/dao can be the caller")

// NewGovDAOExecutor creates the govdao wrapped callback executor
func NewGovDAOExecutor(cb func() error) dao.Executor {
	if cb == nil {
		panic(errNoChangesProposed)
	}

	return executor.NewCallbackExecutor(
		cb,
		std.CurrentRealm().PkgPath(),
	)
}

// NewMemberPropExecutor returns the GOVDAO member change executor
func NewMemberPropExecutor(changesFn func() []dao.Member) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
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

	return NewGovDAOExecutor(callback)
}

func NewDAOImplExecutor(changeFn func() dao.DAO) dao.Executor {
	if changeFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		setDAOImpl(changeFn())

		return nil
	}

	return NewGovDAOExecutor(callback)
}

func NewMembStoreImplExecutor(changeFn func() dao.MemberStore) dao.Executor {
	if changeFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		setMembStoreImpl(changeFn())

		return nil
	}

	return NewGovDAOExecutor(callback)
}

// setDAOImpl sets a new DAO implementation
func setDAOImpl(impl dao.DAO) {
	d = impl
}

// setMembStoreImpl sets a new dao.MembStore implementation
func setMembStoreImpl(impl dao.MemberStore) {
	members = impl
}

// assertCalledByGovdao asserts that the calling Realm is /r/gov/dao
func assertCalledByGovdao() {
	caller := std.CurrentRealm().PkgPath()

	if caller != daoPkgPath {
		panic(errNotGovDAO)
	}
}
