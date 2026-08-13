package dao

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

// allowedDAOs is the sole authorization for UpdateImpl, SafeExecutor.Execute,
// treasury.Send and treasury.SetTokenKeys, and InAllowedDAOs() fails OPEN when
// it is empty — the bootstrap window that lets the genesis MsgRun seed the
// member set before lockdown.
//
// These tests pin that the transition empty -> non-empty is one-way: once the
// DAO is locked down, UpdateImpl cannot put it back into the fail-open state.
// Both paths below reopened the gate before the guard, because
// NewUpdateRequest copies nil into a NON-nil empty slice and the old test was
// `r.AllowedDAOs != nil`.
//
// These call the real UpdateImpl through a code realm rather than replaying
// its logic, so reverting the guard makes them fail.
func TestUpdateImplIgnoresEmptyAllowedDAOs(cur realm, t *testing.T) {
	savedDAOs, savedDAO := allowedDAOs, dao
	defer func() { allowedDAOs, dao = savedDAOs, savedDAO }()

	lock := func() {
		allowedDAOs = nil // reopen so the next UpdateImpl is permitted
		testing.SetRealm(testing.NewCodeRealm(v3))
		UpdateImpl(cross(cur), UpdateRequest{DAO: &dummyDao{}, AllowedDAOs: []string{v3}})
		uassert.False(t, InAllowedDAOs(invalid), "precondition: locked down")
	}

	// Path 1: NewUpdateRequest(d, nil) — "swap the implementation, leave
	// permissions alone", and the form v3/loader uses.
	lock()
	testing.SetRealm(testing.NewCodeRealm(v3))
	UpdateImpl(cross(cur), NewUpdateRequest(&dummyDao{}, nil))
	uassert.False(t, InAllowedDAOs(invalid),
		"a nil AllowedDAOs must not reopen the permission gate")

	// Path 2: an explicitly empty slice via the struct literal.
	lock()
	testing.SetRealm(testing.NewCodeRealm(v3))
	UpdateImpl(cross(cur), UpdateRequest{DAO: &dummyDao{}, AllowedDAOs: []string{}})
	uassert.False(t, InAllowedDAOs(invalid),
		"an empty AllowedDAOs must not reopen the permission gate")

	// A legitimate extension of the list must still apply.
	lock()
	testing.SetRealm(testing.NewCodeRealm(v3))
	UpdateImpl(cross(cur), NewUpdateRequest(&dummyDao{}, []string{v3, v4}))
	uassert.True(t, InAllowedDAOs(v4), "a non-empty AllowedDAOs must still be stored")
	uassert.False(t, InAllowedDAOs(invalid), "and must not admit anyone else")
}

// The bootstrap window itself must survive the guard: with no allowlist
// configured yet, any caller is allowed so the genesis MsgRun can seed the
// member set and then lock down.
func TestBootstrapWindowStillOpen(t *testing.T) {
	saved := allowedDAOs
	defer func() { allowedDAOs = saved }()

	allowedDAOs = nil
	uassert.True(t, InAllowedDAOs("gno.land/r/gov/dao/v3/loader"),
		"an unset allowlist must stay open for genesis bootstrap")
}

// len(AllowedDAOs) != 0 is not sufficient on its own. InAllowedDAOs compares by
// exact string and a user realm's PkgPath() is "", so a single "" entry admits
// any caller whose previous frame is a user realm — the same fail-open outcome
// the guard exists to prevent. NewUpgradeDaoImplRequest passes its realmPkg
// argument straight into the list, so an empty one reaches here.
func TestUpdateImplRejectsBlankAllowedDAOEntry(cur realm, t *testing.T) {
	savedDAOs, savedDAO := allowedDAOs, dao
	defer func() { allowedDAOs, dao = savedDAOs, savedDAO }()

	for _, blank := range []string{"", "   "} {
		allowedDAOs = nil
		testing.SetRealm(testing.NewCodeRealm(v3))
		UpdateImpl(cross(cur), UpdateRequest{DAO: &dummyDao{}, AllowedDAOs: []string{v3}})

		testing.SetRealm(testing.NewCodeRealm(v3))
		urequire.AbortsWithMessage(t, cur,
			"AllowedDAOs entries must be realm paths; got an empty one",
			func() {
				UpdateImpl(cross(cur), UpdateRequest{
					DAO:         &dummyDao{},
					AllowedDAOs: []string{v3, blank},
				})
			})

		uassert.False(t, InAllowedDAOs(""),
			"a blank entry must never make it into the allowlist")
		uassert.True(t, InAllowedDAOs(v3),
			"the rejected request must leave the previous allowlist intact")
	}
}

// A padded entry passes a plain non-blank test but is useless: entries are
// stored exactly as given and InAllowedDAOs compares whole strings, so
// " gno.land/r/x " matches no caller. The list is still non-empty, so the
// bootstrap window is closed. A proposal that padded every entry would lock
// the DAO out of its own allowlist with no way back.
func TestUpdateImplRejectsPaddedAllowedDAOEntry(cur realm, t *testing.T) {
	savedDAOs, savedDAO := allowedDAOs, dao
	defer func() { allowedDAOs, dao = savedDAOs, savedDAO }()

	for _, padded := range []string{" " + v4, v4 + " ", "\t" + v4} {
		allowedDAOs = nil
		testing.SetRealm(testing.NewCodeRealm(v3))
		UpdateImpl(cross(cur), UpdateRequest{DAO: &dummyDao{}, AllowedDAOs: []string{v3}})

		testing.SetRealm(testing.NewCodeRealm(v3))
		urequire.AbortsWithMessage(t, cur,
			"AllowedDAOs entries must not have leading or trailing spaces; entry 1",
			func() {
				UpdateImpl(cross(cur), UpdateRequest{
					DAO:         &dummyDao{},
					AllowedDAOs: []string{v3, padded},
				})
			})

		uassert.False(t, InAllowedDAOs(padded),
			"a padded entry must never make it into the allowlist")
		uassert.True(t, InAllowedDAOs(v3),
			"the rejected request must leave the previous allowlist intact")
	}
}
