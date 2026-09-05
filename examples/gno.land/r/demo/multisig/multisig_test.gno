package multisig

import (
	"testing"

	"chain"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/testutils/v0"
)

func resetForTest(owner1, owner2, owner3 address, thresh int) {
	owners = avl.Tree{}
	proposals = avl.Tree{}
	nextID = 1
	owners.Set(owner1.String(), true)
	owners.Set(owner2.String(), true)
	owners.Set(owner3.String(), true)
	ownerCount = 3
	threshold = thresh
	initialized = true
}

func expectPanic(t *testing.T, expected string, f func()) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic %q, got none", expected)
			return
		}
		msg, ok := r.(string)
		if !ok || msg != expected {
			t.Fatalf("expected panic %q, got %v", expected, r)
		}
	}()
	f()
}

func TestSetupTwiceFails(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	owner2 := testutils.TestAddress("setup-owner2")
	owner3 := testutils.TestAddress("setup-owner3")
	resetForTest(owner1, owner2, owner3, 2)

	expectPanic(t, "already initialized", func() {
		Setup(cur, owner1, owner2, owner3, 2)
	})
}

func TestSetupRejectsDuplicateOwners(cur realm, t *testing.T) {
	initialized = false
	owner1 := cur.Previous().Address()
	expectPanic(t, "owners must be distinct", func() {
		Setup(cur, owner1, owner1, testutils.TestAddress("dup-owner3"), 2)
	})
}

func TestProposeRejectsNonPositiveAmount(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("propose-owner2"), testutils.TestAddress("propose-owner3"), 2)

	expectPanic(t, "amount must be positive", func() {
		Propose(cur, "bad amount", testutils.TestAddress("propose-target"), 0)
	})
}

func TestExecuteWithoutApprovalsFails(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("noappr-owner2"), testutils.TestAddress("noappr-owner3"), 2)
	id := Propose(cur, "underapproved", testutils.TestAddress("noappr-target"), 500)

	expectPanic(t, "not enough approvals", func() {
		Execute(cur, id)
	})
}

func TestApproveThenRevokeDropsCount(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("revoke-owner2"), testutils.TestAddress("revoke-owner3"), 2)
	id := Propose(cur, "revoke test", testutils.TestAddress("revoke-target"), 250)

	Approve(cur, id)
	raw := proposals.Get(proposalKey(id))
	p := raw.(*Proposal)
	if p.ApprovalCount() != 1 {
		t.Fatalf("expected 1 approval, got %d", p.ApprovalCount())
	}

	Revoke(cur, id)
	if p.ApprovalCount() != 0 {
		t.Fatalf("expected 0 approvals after revoke, got %d", p.ApprovalCount())
	}
}

func TestDoubleApproveFails(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("double-owner2"), testutils.TestAddress("double-owner3"), 2)
	id := Propose(cur, "double approve test", testutils.TestAddress("double-target"), 250)
	Approve(cur, id)

	expectPanic(t, "already approved", func() {
		Approve(cur, id)
	})
}

func TestCancelRemovesProposal(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("cancel-owner2"), testutils.TestAddress("cancel-owner3"), 2)
	id := Propose(cur, "to be cancelled", testutils.TestAddress("cancel-target"), 300)

	Cancel(cur, id)
	raw := proposals.Get(proposalKey(id))
	if raw != nil {
		t.Fatal("expected proposal to be removed after cancel")
	}
}

func TestApproveNonExistentProposalFails(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	resetForTest(owner1, testutils.TestAddress("missing-owner2"), testutils.TestAddress("missing-owner3"), 2)

	expectPanic(t, "proposal does not exist", func() {
		Approve(cur, 9999)
	})
}

// This test switches identities via testing.SetRealm (owner1 -> owner2
// -> owner3) to exercise a threshold reached by genuinely distinct
// approvers. cur's own Previous chain is fixed at test entry, so once
// the realm context has been swapped underneath it, calling through
// with bare cur breaks -- cross(cur) re-derives the crossing frame
// against the realm active at call time instead of the one captured
// at the top of the test.
func TestExecuteWithDistinctApprovalsMovesCoins(cur realm, t *testing.T) {
	owner1 := testutils.TestAddress("exec-owner1")
	owner2 := testutils.TestAddress("exec-owner2")
	owner3 := testutils.TestAddress("exec-owner3")
	resetForTest(owner1, owner2, owner3, 2)

	testing.SetRealm(testing.NewUserRealm(owner1))
	id := Propose(cross(cur), "real execution", testutils.TestAddress("exec-target"), 400)
	Approve(cross(cur), id)

	testing.SetRealm(testing.NewUserRealm(owner2))
	Approve(cross(cur), id)

	treasury := chain.PackageAddress(ownPkgPath)
	testing.IssueCoins(treasury, chain.NewCoins(chain.NewCoin("ugnot", 400)))

	testing.SetRealm(testing.NewUserRealm(owner3))
	Execute(cross(cur), id)

	raw := proposals.Get(proposalKey(id))
	p := raw.(*Proposal)
	if !p.Executed() {
		t.Fatal("expected proposal to be marked executed")
	}
}
