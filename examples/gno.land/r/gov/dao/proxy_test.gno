package dao

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/urequire"
)

const (
	v3 = "gno.land/r/gov/dao/v3/impl"
	v4 = "gno.land/r/gov/dao/v4/impl"
	v5 = "gno.land/r/gov/dao/v5/impl"
	v6 = "gno.land/r/gov/dao/v6/impl"
)

const invalid = "gno.land/r/invalid/dao"

var alice = testutils.TestAddress("alice")

func TestProxy_Functions(cur realm, t *testing.T) {
	// initialize tests
	UpdateImpl(cross, UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{v3},
	})

	// invalid package cannot add a new dao in charge
	testing.SetRealm(std.NewCodeRealm(invalid))
	urequire.AbortsWithMessage(t, "permission denied for prev realm: gno.land/r/invalid/dao", func() {
		UpdateImpl(cross, UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// dao in charge can add a new dao
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(cross, UpdateRequest{
			DAO: &dummyDao{},
		})
	})

	// v3 that is in charge adds v5 in charge
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(cross, UpdateRequest{
			DAO:         &dummyDao{},
			AllowedDAOs: []string{v3, v5},
		})
	})

	// v3 can still do updates
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.NotPanics(t, func() {
		UpdateImpl(cross, UpdateRequest{
			AllowedDAOs: []string{v4},
		})
	})

	// not after removing himself from allowedDAOs list
	testing.SetRealm(std.NewCodeRealm(v3))
	urequire.AbortsWithMessage(t, "permission denied for prev realm: gno.land/r/gov/dao/v3/impl", func() {
		UpdateImpl(cross, UpdateRequest{
			AllowedDAOs: []string{v3},
		})
	})

	var pid ProposalID
	testing.SetRealm(std.NewUserRealm(alice))
	urequire.NotPanics(t, func() {
		e := NewSimpleExecutor(
			func(realm) error {
				return nil
			},
			"",
		)
		pid = MustCreateProposal(cross, NewProposalRequest("Proposal Title", "Description", e))
	})

	p := MustGetProposal(cross, 1000)
	if p != nil {
		panic("proposal must be nil")
	}
	p = MustGetProposal(cross, pid)
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
	UpdateImpl(cross, UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{},
	})
}

// TestDirectAllowedDAOsModification tests that the allowedDAOs slice 
// cannot be modified directly from external code, preventing the 
// proposal spoofing vulnerability that existed before PR #4427
func TestDirectAllowedDAOsModification(t *testing.T) {
	// Initialize with known state
	testing.SetRealm(std.NewCodeRealm(v3))
	UpdateImpl(cross, UpdateRequest{
		DAO:         &dummyDao{},
		AllowedDAOs: []string{v3},
	})

	// Verify initial state
	allowed := AllowedDAOs()
	urequire.Equal(t, 1, len(allowed))
	urequire.Equal(t, v3, allowed[0])
	
	// Test that we cannot directly modify the returned slice
	// (this tests the defensive copy behavior)
	allowed[0] = "malicious.package"
	newAllowed := AllowedDAOs()
	urequire.Equal(t, v3, newAllowed[0], "allowedDAOs should not be modifiable through returned slice")
	
	// Test that appending to returned slice doesn't affect internal state
	allowed = append(allowed, "another.malicious.package")
	newAllowed = AllowedDAOs()
	urequire.Equal(t, 1, len(newAllowed), "allowedDAOs length should not change")
	
	// This test simulates what the removed NewAllowImplRequest function
	// would have allowed - direct manipulation of allowedDAOs without
	// proper permission checks through UpdateImpl
	
	// Before PR #4427, external code could potentially call something like:
	// NewAllowImplRequest("malicious.package") 
	// which would append to allowedDAOs without permission validation
	
	// Now, the only way to modify allowedDAOs is through UpdateImpl,
	// which requires being in the allowedDAOs list already
	
	testing.SetRealm(std.NewCodeRealm(invalid))
	urequire.AbortsWithMessage(t, "permission denied for prev realm: gno.land/r/invalid/dao", func() {
		UpdateImpl(cross, UpdateRequest{
			AllowedDAOs: []string{invalid, v3},
		})
	})
	
	// Verify state unchanged after failed attempt
	finalAllowed := AllowedDAOs()
	urequire.Equal(t, 1, len(finalAllowed))
	urequire.Equal(t, v3, finalAllowed[0])
}

type dummyDao struct{}

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
