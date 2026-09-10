package dao

import (
	"chain/runtime/unsafe"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/urequire/v0"
)

// cur is a zero-value realm used as a placeholder when forwarding to
// uassert/urequire dispatch helpers that gained an `rlm realm` param.
// These tests pass `func()` callbacks (no crossing inside the callback),
// so rlm is ignored — a nil realm here is safe.
var cur realm

const (
	v0 = "gno.land/r/gov/dao/v0/impl"
	v1 = "gno.land/r/gov/dao/v1/impl"
	v2 = "gno.land/r/gov/dao/v2/impl"
	v3 = "gno.land/r/gov/dao/v3/impl"
)

const invalid = "gno.land/r/invalid/dao"

var alice = testutils.TestAddress("alice")

func TestProxy_Functions(cur realm, t *testing.T) {
	// initialize tests
	UpdateImpl(cross(cur), UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{v0},
	})

	// invalid package cannot add a new dao in charge
	testing.SetRealm(testing.NewCodeRealm(invalid))
	urequire.AbortsWithMessage(t, cur, "permission denied for prev realm: gno.land/r/invalid/dao", func() {
		UpdateImpl(cross(cur), UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// dao in charge can add a new dao
	testing.SetRealm(testing.NewCodeRealm(v0))
	urequire.NotPanics(t, cur, func() {
		UpdateImpl(cross(cur), UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// v0 that is in charge adds v2 in charge
	testing.SetRealm(testing.NewCodeRealm(v0))
	urequire.NotPanics(t, cur, func() {
		UpdateImpl(cross(cur), UpdateRequest{
			DAO:         &dummyDao{},
			AllowedDAOs: []string{v0, v2},
		})
	})

	// v0 can still do updates
	testing.SetRealm(testing.NewCodeRealm(v0))
	urequire.NotPanics(t, cur, func() {
		UpdateImpl(cross(cur), UpdateRequest{
			AllowedDAOs: []string{v1},
		})
	})

	// not after removing himself from allowedDAOs list
	testing.SetRealm(testing.NewCodeRealm(v0))
	urequire.AbortsWithMessage(t, cur, "permission denied for prev realm: gno.land/r/gov/dao/v0/impl", func() {
		UpdateImpl(cross(cur), UpdateRequest{
			AllowedDAOs: []string{v0},
		})
	})

	var pid ProposalID
	testing.SetRealm(testing.NewUserRealm(alice))
	urequire.NotPanics(t, cur, func() {
		e := NewSimpleExecutor(0, cur,
			func(realm) error {
				return nil
			},
			"",
		)
		pid = MustCreateProposal(cross(cur), NewProposalRequest("Proposal Title", "Description", e))
	})

	p, err := GetProposal(1000)
	if p != nil || err == nil {
		panic("proposal should not exist and should return an error")
	}
	p = MustGetProposal(pid)
	urequire.Equal(t, "Proposal Title", p.Title())
	urequire.Equal(t, p.Author().String(), alice.String())

	// need to switch the context back to v1
	testing.SetRealm(testing.NewCodeRealm(v1))
	urequire.Equal(
		t,
		"Render: gno.land/r/gov/dao/test",
		Render(cross(cur), "test"),
	)

	// reset state
	testing.SetRealm(testing.NewCodeRealm(v1))
	UpdateImpl(cross(cur), UpdateRequest{
		DAO: &dummyDao{},
	})
	// UpdateImpl deliberately ignores an empty AllowedDAOs, so it can no
	// longer be used to restore the bootstrap-open state — that is the point
	// of the guard. Reset the package var directly instead; we are in the
	// same package.
	allowedDAOs = nil
}

type dummyDao struct{}

func (dd *dummyDao) PreCreateProposal(_ int, rlm realm, r ProposalRequest) (address, error) {
	return unsafe.OriginCaller(), nil
}

func (dd *dummyDao) PostCreateProposal(_ int, rlm realm, r ProposalRequest, pid ProposalID) {
}

func (dd *dummyDao) VoteOnProposal(_ int, rlm realm, r VoteRequest) error {
	return nil
}

func (dd *dummyDao) PreExecuteProposal(_ int, rlm realm, pid ProposalID) (bool, error) {
	return true, nil
}

func (dd *dummyDao) ExecuteProposal(_ int, rlm realm, pid ProposalID, e Executor) error {
	return nil
}

func (dd *dummyDao) Render(cur realm, pkgpath string, path string) string {
	return "Render: " + pkgpath + "/" + path
}
