// PKGPATH: gno.land/r/test
package test

import (
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	testUser = testutils.TestAddress("test")
)

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		t, _ := memberstore.GetTier(memberstore.T1)
		return &memberstore.Member{
			InvitationPoints: t.InvitationPoints,
		}
	default:
		panic("unsupported tier: " + tier)
	}
}

func init() {
	// Load members for testing
	mstore := memberstore.Get()
	mstore.DeleteAll()
	mstore.SetTier(memberstore.T1)
	mstore.SetMember(memberstore.T1, testUser, memberByTier(memberstore.T1))

	// Set up the DAO implementation using proper constructor
	govDAO := impl.NewGovDAO()
	dao.UpdateImpl(cross, dao.UpdateRequest{
		DAO:         govDAO,
		AllowedDAOs: []string{"gno.land/r/test", "gno.land/r/gov/dao/v3/impl"},
	})
}

func main() {
	// Create an executor in a specific realm to test creation realm tracking
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/contract"))
	executor := dao.NewSimpleExecutor(func(realm) error { return nil }, "Test executor description")

	// Create a proposal request
	proposalRequest := dao.NewProposalRequest(
		"Test Proposal Title",
		"This is a test proposal description to verify StringifyProposal works correctly.",
		executor,
	)

	// Switch to user realm to create the proposal
	testing.SetOriginCaller(testUser)
	testing.SetRealm(testing.NewUserRealm(testUser))
	pid := dao.MustCreateProposal(cross, proposalRequest)

	// Get the proposal and test the core functionality
	prop := dao.MustGetProposal(cross, pid)

	// Test that executor string is captured correctly
	println("Executor string:", prop.ExecutorString())

	// Test that executor creation realm is captured correctly
	println("Executor creation realm:", prop.ExecutorCreationRealm())

	// Stringify
	println("----")
	println(impl.StringifyProposal(prop))
}

// Output:
// Executor string: Test executor description
// Executor creation realm: gno.land/r/template/contract
// ----
//
// ### Title: Test Proposal Title
//
// ### Proposed by: g1w3jhxazlta047h6lta047h6lta047h6lwmjv0n
//
// This is a test proposal description to verify StringifyProposal works correctly.
//
// This proposal contains the following metadata:
//
// Test executor description
//
// Executor created in: gno.land/r/template/contract
