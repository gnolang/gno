package dao

import (
	"std"
	"testing"
)

func TestAddMember(t *testing.T) {
	d := New()
	addr := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")

	err := d.AddMember(addr)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	err = d.AddMember(addr)
	if err != ErrUserAlreadyIsMember {
		t.Errorf("Expected ErrUserAlreadyIsMember, got: %v", err)
	}
}

func TestGetMembers(t *testing.T) {
	d := New()
	addr := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")
	d.AddMember(addr)

	members := d.GetMembers()
	if len(members) != 1 || members[0] != addr {
		t.Errorf("Expected [addr1], got: %v", members)
	}
}

func TestDeleteMember(t *testing.T) {
	d := New()
	addr1 := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")
	addr2 := std.Address("g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa")

	d.AddMember(addr1)
	d.AddMember(addr2)

	err := d.DeleteMember(addr1)
	if err != nil {
		t.Errorf("Unexpected error deleting member: %v", err)
	}

	if d.IsMember(addr1) {
		t.Errorf("Expected addr1 to be removed")
	}

	err = d.DeleteMember(std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv"))
	if err != ErrUserIsNotMember {
		t.Errorf("Expected ErrUserIsNotMember, got: %v", err)
	}
}

func TestIsMember(t *testing.T) {
	d := New()
	addr := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")

	if d.IsMember(addr) {
		t.Errorf("Expected IsMember to return false for new DAO")
	}

	d.AddMember(addr)
	if !d.IsMember(addr) {
		t.Errorf("Expected IsMember to return true after adding member")
	}
}

func TestAsserUserIsMember(t *testing.T) {
	d := New()
	addr := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when user is not a member")
		}
	}()
	d.AsserUserIsMember(addr)
}

func TestAsserUserIsNotMember(t *testing.T) {
	d := New()
	addr := std.Address("g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv")
	d.AddMember(addr)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when user *is* already a member")
		}
	}()
	d.AsserUserIsNotMember(addr)
}
