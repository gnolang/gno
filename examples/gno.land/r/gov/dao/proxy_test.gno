package dao

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

const v3 = "gno.land/r/gov/dao/v3/impl"
const v4 = "gno.land/r/gov/dao/v4/impl"
const v5 = "gno.land/r/gov/dao/v5/impl"
const v6 = "gno.land/r/gov/dao/v6/impl"

const invalid = "gno.land/r/invalid/dao"

var (
	alice = testutils.TestAddress("alice")
)

func TestProxy_Functions(t *testing.T) {
	// initialize tests
	cross(UpdateImpl)(UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{v3},
	})

	// invalid package cannot add a new dao in charge
	testing.SetRealm(std.NewCodeRealm(invalid))
	urequire.AbortsWithMessage(t, "permission denied for prev realm: gno.land/r/invalid/dao", func() {
		cross(UpdateImpl)(UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// dao in charge can add a new dao
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		cross(UpdateImpl)(UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// v3 that is in charge adds v5 in charge
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		cross(UpdateImpl)(UpdateRequest{
			DAO:         &dummyDao{},
			AllowedDAOs: []string{v3, v5},
		})
	})

	// v3 can still do updates
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		cross(UpdateImpl)(UpdateRequest{
			AllowedDAOs: []string{v4},
		})
	})

	// not after removing himself from allowedDAOs list
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.AbortsWithMessage(t, "permission denied for prev realm: gno.land/r/gov/dao/v3/impl", func() {
		cross(UpdateImpl)(UpdateRequest{
			AllowedDAOs: []string{v3},
		})
	})

	var pid ProposalID
	testing.SetRealm(std.NewUserRealm(alice))
	urequire.NotPanics(t, func() {
		e := NewSimpleExecutor(
			func() error {
				return nil
			},
			"",
		)
		pid = cross(MustCreateProposal)(
			NewProposalRequest("Proposal Title", "Description", e),
		)
	})

	p := cross(MustGetProposal)(1000)
	if p != nil {
		panic("proposal must be nil")
	}
	p = cross(MustGetProposal)(pid)
	urequire.Equal(t, "Proposal Title", p.Title())
	urequire.Equal(t, p.Author().String(), alice.String())

	// need to switch the context back to v4
	testing.SetRealm(std.NewCodeRealm(v4))
	urequire.Equal(
		t,
		"Render: gno.land/r/gov/dao/test",
		Render("test"),
	)

	// reset state
	testing.SetRealm(std.NewCodeRealm(v4))
	cross(UpdateImpl)(UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{},
	})
}

type dummyDao struct {
}

func (dd *dummyDao) PreCreateProposal(r ProposalRequest) (std.Address, error) {
	return std.OriginCaller(), nil
}

func (dd *dummyDao) PostCreateProposal(r ProposalRequest, pid ProposalID) {

}

func (dd *dummyDao) VoteOnProposal(r VoteRequest) error {
	return nil
}

func (dd *dummyDao) PreGetProposal(pid ProposalID) error {
	return nil
}

func (dd *dummyDao) PostGetProposal(pid ProposalID, p *Proposal) error {
	return nil
}

func (dd *dummyDao) PreExecuteProposal(pid ProposalID) (bool, error) {
	return true, nil
}

func (dd *dummyDao) Render(pkgpath string, path string) string {
	return "Render: " + pkgpath + "/" + path
}
