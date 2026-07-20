package multisig

import (
	"strconv"
	"testing"
)

func TestSetupInitializesCorrectly(cur realm, t *testing.T) {
	owner1 := cur.Previous().Address()
	owner2 := address("g1dummyaddressfortestpurposesonlyxxxxxx1")
	owner3 := address("g1dummyaddressfortestpurposesonlyxxxxxx2")

	Setup(cur, owner1, owner2, owner3, 2)

	if !initialized {
		t.Fatal("expected initialized to be true")
	}
	if threshold != 2 {
		t.Fatalf("expected threshold 2, got %d", threshold)
	}
	if ownerCount != 3 {
		t.Fatalf("expected 3 owners, got %d", ownerCount)
	}
}

func TestSetupTwiceFails(cur realm, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on second setup")
		}
	}()
	Setup(cur, cur.Previous().Address(), address("g1dummyaddressfortestpurposesonlyxxxxxx1"), address("g1dummyaddressfortestpurposesonlyxxxxxx2"), 2)
}

func TestProposeCreatesProposal(cur realm, t *testing.T) {
	id := Propose(cur, "test proposal", address("g1dummyaddressfortestpurposesonlyxxxxxx1"), 1000)

	raw := proposals.Get(strconv.Itoa(id))
	if raw == nil {
		t.Fatal("proposal was not stored")
	}
	p := raw.(*Proposal)
	if p.executed {
		t.Fatal("new proposal should not be executed")
	}
	if p.approveCnt != 0 {
		t.Fatalf("expected 0 approvals, got %d", p.approveCnt)
	}
}

func TestExecuteWithoutApprovalsFails(cur realm, t *testing.T) {
	id := Propose(cur, "underfunded approval test", address("g1dummyaddressfortestpurposesonlyxxxxxx1"), 500)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic executing with 0/2 approvals")
		}
	}()
	Execute(cur, id)
}

func TestApproveIncrementsCount(cur realm, t *testing.T) {
	id := Propose(cur, "approval count test", address("g1dummyaddressfortestpurposesonlyxxxxxx1"), 250)
	Approve(cur, id)

	raw := proposals.Get(strconv.Itoa(id))
	p := raw.(*Proposal)
	if p.approveCnt != 1 {
		t.Fatalf("expected 1 approval, got %d", p.approveCnt)
	}
}

func TestDoubleApproveFails(cur realm, t *testing.T) {
	id := Propose(cur, "double approve test", address("g1dummyaddressfortestpurposesonlyxxxxxx1"), 250)
	Approve(cur, id)

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on double approve from same owner")
		}
	}()
	Approve(cur, id)
}

func TestApproveNonExistentProposalFails(cur realm, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic approving non-existent proposal")
		}
	}()
	Approve(cur, 9999)
}

func TestExecuteNonExistentProposalFails(cur realm, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic executing non-existent proposal")
		}
	}()
	Execute(cur, 9999)
}
