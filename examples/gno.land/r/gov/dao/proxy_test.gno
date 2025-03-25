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
	UpdateImpl(UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{v3},
	})

	// invalid package cannot add a new dao in charge
	testing.SetRealm(std.NewCodeRealm(invalid))
	urequire.PanicsWithMessage(t, "permission denied for prev realm: gno.land/r/invalid/dao", func() {
		UpdateImpl(UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// dao in charge can add a new dao
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// v3 that is in charge adds v5 in charge
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(UpdateRequest{
			DAO:         &dummyDao{},
			AllowedDAOs: []string{v3, v5},
		})
	})

	// v3 can still do updates
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(UpdateRequest{
			AllowedDAOs: []string{v4},
		})
	})

	// not after removing himself from allowedDAOs list
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.PanicsWithMessage(t, "permission denied for prev realm: gno.land/r/gov/dao/v3/impl", func() {
		UpdateImpl(UpdateRequest{
			AllowedDAOs: []string{v3},
		})
	})

	var pid ProposalID
	testing.SetRealm(std.NewUserRealm(alice))
	urequire.NotPanics(t, func() {
		pid = MustCreateProposal(
			NewProposalRequestBuilder("Proposal Title").
				WithDescription("Description").
				WithCallback(func(m Metadata) error {
					return nil
				}).Build())
	})

	p := MustGetProposal(1000)
	if p != nil {
		panic("proposal must be nil")
	}

	p = MustGetProposal(pid)
	urequire.Equal(t, "Proposal Title", p.Title())
	urequire.Equal(t, p.Author().ID(), "user:g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm")

	urequire.Equal(
		t,
		"Render: test",
		Render("test"),
	)

	// reset state
	testing.SetRealm(std.NewCodeRealm(v4))
	UpdateImpl(UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{},
	})
}

type dummyDao struct {
}

func (dd *dummyDao) PreCreateProposal(r ProposalRequest) (Author, error) {
	return &dummyAuthor{std.OriginCaller()}, nil
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

func (dd *dummyDao) Render(i, x string) string {
	return "Render: " + i
}

type dummyAuthor struct {
	caller std.Address
}

func (da *dummyAuthor) ID() string {
	return "user:" + string(da.caller)
}

func (da *dummyAuthor) String() string {
	return da.ID()
}
