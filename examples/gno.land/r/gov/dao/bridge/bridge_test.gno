package bridge

import (
	"testing"

	"std"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestBridge_DAO(t *testing.T) {
	var (
		proposalID = uint64(10)
		mockDAO    = &mockDAO{
			proposeFn: func(_ dao.ProposalRequest) uint64 {
				return proposalID
			},
		}
	)

	b.dao = mockDAO

	uassert.Equal(t, proposalID, GovDAO().Propose(dao.ProposalRequest{}))
}

func TestBridge_AllowNewGovDAOVersion(t *testing.T) {
	var (
		addr             = testutils.TestAddress("owner")
		newGovDAOPkgPath = "gno.land/r/gov/dao/v3"
	)

	std.TestSetOrigCaller(addr)
	b.Ownable = ownable.NewWithAddress(addr)

	t.Run("invalid pkgpath", func(t *testing.T) {
		// Attempt to set a new DAO implementation
		uassert.PanicsWithMessage(t, ErrInvalidGovDAOPkgPath.Error(), func() {
			AllowNewGovDAOVersion("invalidpkgpath")
		})
	})

	t.Run("valid path", func(t *testing.T) {
		// Attempt to set a new DAO implementation
		uassert.NotPanics(t, func() {
			AllowNewGovDAOVersion(newGovDAOPkgPath)
		})
	})
}

func TestBridge_RegisterNewDAOImpl(t *testing.T) {
	var (
		addr             = testutils.TestAddress("owner")
		newGovDAOPkgPath = "gno.land/r/gov/dao/v3"
	)

	std.TestSetOrigCaller(addr)
	b.Ownable = ownable.NewWithAddress(addr)
	AllowNewGovDAOVersion(newGovDAOPkgPath)

	t.Run("invalid pkgpath", func(t *testing.T) {
		std.TestSetRealm(std.NewCodeRealm("gno.land/r/invalid/path"))

		uassert.PanicsWithMessage(t, ErrNotAllowlisted.Error(), func() {
			RegisterNewDAOImpl(&mockDAO{})
		})
	})

	t.Run("valid pkgpath", func(t *testing.T) {
		std.TestSetRealm(std.NewCodeRealm(newGovDAOPkgPath))

		uassert.NotPanics(t, func() {
			RegisterNewDAOImpl(&mockDAO{})
		})
	})
}
