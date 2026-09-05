// PKGPATH: gno.land/r/test
package test

import (
	"strings"
	"testing"

	"gno.land/p/nt/testutils/v0"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

var (
	testUser = testutils.TestAddress("test")
)

// StringifyProposal carries its own copy of the disclosure expression, separate
// from the one in render.gno. It has no production caller today, but it is
// exported, and an untested copy of a security-relevant expression is free to
// drift from the tested one. This executor pins it: CreationRealm is reached
// through the public interface, so it can return anything, and the run of two
// backticks closes a narrower fence.
type hostileExec struct{}

func (e *hostileExec) Execute(cur realm) error { return nil }

func (e *hostileExec) String() string { return "" }

// Padded on both ends as well as fenced: without the TrimSpace, InlineCode
// pads the fence to protect the leading and trailing spaces, and the expected
// span below no longer matches. That makes one payload cover both the escaping
// and the trim.
func (e *hostileExec) CreationRealm() string { return "  gno.land/r/evil`` **INJECTED**  " }

func memberByTier(tier string) *memberstore.Member {
	switch tier {
	case memberstore.T1:
		t, _ := memberstore.GetTier(memberstore.T1)
		return memberstore.NewMember(t.InvitationPoints)
	default:
		panic("unsupported tier: " + tier)
	}
}

func init(cur realm) {
	// Load members for testing
	mstore := memberstore.Get(0, cur)
	mstore.DeleteAll()
	mstore.SetTier(memberstore.T1)
	mstore.SetMember(memberstore.T1, testUser, memberByTier(memberstore.T1))

	// Set up the DAO implementation using proper constructor
	govDAO := impl.NewGovDAO()
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(govDAO, []string{"gno.land/r/test", "gno.land/r/gov/dao/v3/impl"}))
}

func main(cur realm) {
	// Create an executor in a specific realm to test creation realm tracking
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/template/contract"))
	executor := dao.NewSimpleExecutor(0, cur, func(realm) error { return nil }, "Test executor description")

	// Create a proposal request
	proposalRequest := dao.NewProposalRequest(
		"Test Proposal Title",
		"This is a test proposal description to verify StringifyProposal works correctly.",
		executor,
	)

	// Switch to user realm to create the proposal
	testing.SetOriginCaller(testUser)
	testing.SetRealm(testing.NewUserRealm(testUser))
	pid := dao.MustCreateProposal(cross(cur), proposalRequest)

	// Get the proposal and test the core functionality
	prop := dao.MustGetProposal(pid)

	// Test that executor string is captured correctly
	println("Executor string:", prop.ExecutorString())

	// Test that executor creation realm is captured correctly
	println("Executor creation realm:", prop.ExecutorCreationRealm())

	// Stringify
	println("----")
	println(impl.StringifyProposal(prop))

	// The same disclosure, through a hostile executor.
	hpid := dao.MustCreateProposal(cross(cur), dao.NewProposalRequest(
		"Hostile", "creation realm attacks the code span", &hostileExec{}))
	hout := impl.StringifyProposal(dao.MustGetProposal(hpid))

	println("stringify widens the fence too:",
		strings.Contains(hout, "```gno.land/r/evil`` **INJECTED**```"))
	println("and no bold escapes it:", !strings.Contains(hout, "\n**INJECTED**"))
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
// Executor created in: `gno.land/r/template/contract`
//
// stringify widens the fence too: true
// and no bold escapes it: true
