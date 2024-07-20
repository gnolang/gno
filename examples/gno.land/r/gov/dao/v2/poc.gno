package govdao

import (
	"errors"
	"std"

	"gno.land/p/demo/membstore"
	"gno.land/p/gov/dao"
	"gno.land/p/gov/executor"
)

var errNoChangesProposed = errors.New("no set changes proposed")

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
func NewMemberPropExecutor(changesFn func() []membstore.Member) dao.Executor {
	if changesFn == nil {
		panic(errNoChangesProposed)
	}

	callback := func() error {
		errs := &executor.CombinedError{}
		cbMembers := changesFn()

		for _, member := range cbMembers {
			switch {
			case !members.IsMember(member.Address):
				// Addition request
				err := members.AddMember(member)

				errs.Add(err)
			case member.VotingPower == 0:
				// Remove request
				err := members.UpdateMember(member.Address, membstore.Member{
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

func NewMembStoreImplExecutor(changeFn func() membstore.MemberStore) dao.Executor {
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
	if impl == nil {
		panic("invalid member store")
	}

	d = impl
}

// setMembStoreImpl sets a new dao.MembStore implementation
func setMembStoreImpl(impl membstore.MemberStore) {
	if impl == nil {
		panic("invalid member store")
	}

	members = impl
}
