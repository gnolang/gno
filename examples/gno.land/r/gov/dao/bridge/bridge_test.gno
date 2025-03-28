package bridge

import (
	"std"
	"testing"

	"gno.land/p/demo/dao"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
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

func TestBridge_LoadGovDAO(t *testing.T) {
	t.Run("invalid initializer path", func(t *testing.T) {
		testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/init")) // invalid loader

		// Attempt to set a new DAO implementation
		uassert.PanicsWithMessage(t, "unauthorized", func() {
			LoadGovDAO(&mockDAO{})
		})
	})

	t.Run("valid loader", func(t *testing.T) {
		var (
			initializer = "gno.land/r/gov/dao/init"
			proposalID  = uint64(10)
			mockDAO     = &mockDAO{
				proposeFn: func(_ dao.ProposalRequest) uint64 {
					return proposalID
				},
			}
		)

		testing.SetRealm(std.NewCodeRealm(initializer))

		// Attempt to set a new DAO implementation
		uassert.NotPanics(t, func() {
			LoadGovDAO(mockDAO)
		})

		uassert.Equal(
			t,
			mockDAO.Propose(dao.ProposalRequest{}),
			GovDAO().Propose(dao.ProposalRequest{}),
		)
	})
}

func TestBridge_SetDAO(t *testing.T) {
	t.Run("invalid owner", func(t *testing.T) {
		// Attempt to set a new DAO implementation
		uassert.PanicsWithMessage(t, ownable.ErrUnauthorized.Error(), func() {
			SetGovDAO(&mockDAO{})
		})
	})

	t.Run("valid owner", func(t *testing.T) {
		var (
			addr = testutils.TestAddress("owner")

			proposalID = uint64(10)
			mockDAO    = &mockDAO{
				proposeFn: func(_ dao.ProposalRequest) uint64 {
					return proposalID
				},
			}
		)

		testing.SetOriginCaller(addr)

		Ownable = ownable.NewWithAddress(addr)

		urequire.NotPanics(t, func() {
			SetGovDAO(mockDAO)
		})

		uassert.Equal(
			t,
			mockDAO.Propose(dao.ProposalRequest{}),
			GovDAO().Propose(dao.ProposalRequest{}),
		)
	})
}
