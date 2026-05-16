package mgroup

import (
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

// Address-only test principals. We use NewUserRealm(addr) + a crossing
// closure to drive each call, because SetRealm(NewUserRealm) updates
// the per-tx OriginCaller — which the closure picks up via
// callingCurOrOrigin when there's no surrounding crossing frame. The
// underlying ownable check is addr-based, so it doesn't care whether
// the principal is conceptually an EOA or a realm.
var (
	u1 = testutils.TestAddress("u1")
	u2 = testutils.TestAddress("u2")
	u3 = testutils.TestAddress("u3")
	u4 = testutils.TestAddress("u4")
)

func TestManagedGroup_AddBackupOwner(cur realm, t *testing.T) {
	g := New(u1)

	testing.SetRealm(testing.NewUserRealm(u1))
	var err error
	func(cur realm) { err = g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	var badAddr address
	func(cur realm) { err = g.AddBackupOwner(0, cur, badAddr) }(cross2(cur))
	if err != ErrInvalidAddress {
		t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
	}

	// not the owner → unauthorized
	testing.SetRealm(testing.NewUserRealm(u3))
	func(cur realm) { err = g.AddBackupOwner(0, cur, u4) }(cross2(cur))
	if err != ownable.ErrUnauthorized {
		t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
	}
}

func TestManagedGroup_RemoveBackupOwner(cur realm, t *testing.T) {
	g := New(u1)

	testing.SetRealm(testing.NewUserRealm(u1))
	var err error
	func(cur realm) { err = g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	func(cur realm) { err = g.RemoveBackupOwner(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	// idempotent
	func(cur realm) { err = g.RemoveBackupOwner(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	var badAddr address
	func(cur realm) { err = g.RemoveBackupOwner(0, cur, badAddr) }(cross2(cur))
	if err != ErrInvalidAddress {
		t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
	}
	func(cur realm) { err = g.RemoveBackupOwner(0, cur, u1) }(cross2(cur))
	if err != ErrCannotRemoveOwner {
		t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
	}

	func(cur realm) { err = g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	testing.SetRealm(testing.NewUserRealm(u3))
	func(cur realm) { err = g.RemoveBackupOwner(0, cur, u2) }(cross2(cur))
	if err != ownable.ErrUnauthorized {
		t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
	}
}

func TestManagedGroup_ClaimOwnership(cur realm, t *testing.T) {
	g := New(u1)

	testing.SetRealm(testing.NewUserRealm(u1))
	var err error
	func(cur realm) { err = g.AddBackupOwner(0, cur, u2) }(cross2(cur))

	testing.SetRealm(testing.NewUserRealm(u2))
	func(cur realm) { err = g.ClaimOwnership(0, cur) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	if g.Owner() != u2 {
		t.Errorf("expected owner %s, got %s", u2.String(), g.Owner().String())
	}
	if !g.IsMember(u2) {
		t.Errorf("expected %s to be a member", u2.String())
	}
	// already owner — no-op success.
	func(cur realm) { err = g.ClaimOwnership(0, cur) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	// u3 is not a backup owner → fails.
	testing.SetRealm(testing.NewUserRealm(u3))
	func(cur realm) { err = g.ClaimOwnership(0, cur) }(cross2(cur))
	if err != ErrNotMember {
		t.Errorf("expected %v, got %v", ErrNotMember.Error(), err.Error())
	}
}

func TestManagedGroup_AddMember(cur realm, t *testing.T) {
	g := New(u1)

	testing.SetRealm(testing.NewUserRealm(u1))
	var err error
	func(cur realm) { err = g.AddMember(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	if !g.IsMember(u2) {
		t.Errorf("expected %s to be a member", u2.String())
	}
	var badAddr address
	func(cur realm) { err = g.AddMember(0, cur, badAddr) }(cross2(cur))
	if err != ErrInvalidAddress {
		t.Errorf("expected %v, got %v", ErrInvalidAddress, err.Error())
	}
	testing.SetRealm(testing.NewUserRealm(u3))
	func(cur realm) { err = g.AddMember(0, cur, u4) }(cross2(cur))
	if err != ownable.ErrUnauthorized {
		t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
	}
}

func TestManagedGroup_RemoveMember(cur realm, t *testing.T) {
	g := New(u1)

	testing.SetRealm(testing.NewUserRealm(u1))
	var err error
	func(cur realm) { err = g.AddMember(0, cur, u2) }(cross2(cur))
	func(cur realm) { err = g.RemoveMember(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	if g.IsMember(u2) {
		t.Errorf("expected %s to not be a member", u2.String())
	}
	// idempotent
	func(cur realm) { err = g.RemoveMember(0, cur, u2) }(cross2(cur))
	if err != nil {
		t.Errorf("expected nil, got %v", err.Error())
	}
	var badAddr address
	func(cur realm) { err = g.RemoveMember(0, cur, badAddr) }(cross2(cur))
	if err != ErrInvalidAddress {
		t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
	}
	func(cur realm) { err = g.RemoveMember(0, cur, u1) }(cross2(cur))
	if err != ErrCannotRemoveOwner {
		t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
	}

	testing.SetRealm(testing.NewUserRealm(u2))
	func(cur realm) { err = g.RemoveMember(0, cur, u3) }(cross2(cur))
	if err != ownable.ErrUnauthorized {
		t.Errorf("expected %v, got %v", ownable.ErrUnauthorized.Error(), err.Error())
	}
}

func TestManagedGroup_MemberCount(cur realm, t *testing.T) {
	g := New(u1)
	uassert.Equal(t, 1, g.MemberCount())
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddMember(0, cur, u2) }(cross2(cur))
	uassert.Equal(t, 2, g.MemberCount())
	func(cur realm) { g.AddMember(0, cur, u3) }(cross2(cur))
	uassert.Equal(t, 3, g.MemberCount())
	func(cur realm) { g.RemoveMember(0, cur, u2) }(cross2(cur))
	uassert.Equal(t, 2, g.MemberCount())
}

func TestManagedGroup_BackupOwnerCount(cur realm, t *testing.T) {
	g := New(u1)
	uassert.Equal(t, 1, g.BackupOwnerCount())
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	uassert.Equal(t, 2, g.BackupOwnerCount())
	func(cur realm) { g.AddBackupOwner(0, cur, u3) }(cross2(cur))
	uassert.Equal(t, 3, g.BackupOwnerCount())
	func(cur realm) { g.RemoveBackupOwner(0, cur, u2) }(cross2(cur))
	uassert.Equal(t, 2, g.BackupOwnerCount())
}

func TestManagedGroup_IsMember(cur realm, t *testing.T) {
	g := New(u1)
	if !g.IsMember(u1) {
		t.Errorf("expected %s to be a member", u1.String())
	}
	if g.IsMember(u2) {
		t.Errorf("expected %s to not be a member", u2.String())
	}
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddMember(0, cur, u2) }(cross2(cur))
	if !g.IsMember(u2) {
		t.Errorf("expected %s to be a member", u2.String())
	}
}

func TestManagedGroup_IsBackupOwner(cur realm, t *testing.T) {
	g := New(u1)
	if !g.IsBackupOwner(u1) {
		t.Errorf("expected %s to be a backup owner", u1.String())
	}
	if g.IsBackupOwner(u2) {
		t.Errorf("expected %s to not be a backup owner", u2.String())
	}
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	if !g.IsBackupOwner(u2) {
		t.Errorf("expected %s to be a backup owner", u2.String())
	}
}

func TestManagedGroup_Owner(cur realm, t *testing.T) {
	g := New(u1)
	if g.Owner() != u1 {
		t.Errorf("expected %s, got %s", u1.String(), g.Owner().String())
	}
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	if g.Owner() != u1 {
		t.Errorf("expected %s, got %s", u1.String(), g.Owner().String())
	}
	testing.SetRealm(testing.NewUserRealm(u2))
	func(cur realm) { g.ClaimOwnership(0, cur) }(cross2(cur))
	if g.Owner() != u2 {
		t.Errorf("expected %s, got %s", u2.String(), g.Owner().String())
	}
}

func TestManagedGroup_BackupOwners(cur realm, t *testing.T) {
	g := New(u1)
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddBackupOwner(0, cur, u2) }(cross2(cur))
	func(cur realm) { g.AddBackupOwner(0, cur, u3) }(cross2(cur))
	owners := g.BackupOwners()
	if len(owners) != 3 {
		t.Errorf("expected 3, got %v", len(owners))
	}
	// AVL iteration is by key (address string); assert set membership.
	set := map[string]bool{owners[0]: true, owners[1]: true, owners[2]: true}
	for _, addr := range []address{u1, u2, u3} {
		if !set[addr.String()] {
			t.Errorf("expected %s in backup owners", addr.String())
		}
	}
}

func TestManagedGroup_Members(cur realm, t *testing.T) {
	g := New(u1)
	testing.SetRealm(testing.NewUserRealm(u1))
	func(cur realm) { g.AddMember(0, cur, u2) }(cross2(cur))
	func(cur realm) { g.AddMember(0, cur, u3) }(cross2(cur))
	members := g.Members()
	if len(members) != 3 {
		t.Errorf("expected 3, got %v", len(members))
	}
	set := map[string]bool{members[0]: true, members[1]: true, members[2]: true}
	for _, addr := range []address{u1, u2, u3} {
		if !set[addr.String()] {
			t.Errorf("expected %s in members", addr.String())
		}
	}
}

func TestSliceWithOffset(cur realm, t *testing.T) {
	testTable := []struct {
		name          string
		slice         []string
		offset        int
		count         int
		expected      []string
		expectedCount int
	}{
		{
			name:          "empty",
			slice:         []string{},
			offset:        0,
			count:         0,
			expected:      []string{},
			expectedCount: 0,
		},
		{
			name:          "single",
			slice:         []string{"a"},
			offset:        0,
			count:         1,
			expected:      []string{"a"},
			expectedCount: 1,
		},
		{
			name:          "single offset",
			slice:         []string{"a"},
			offset:        1,
			count:         1,
			expected:      []string{},
			expectedCount: 0,
		},
		{
			name:          "multiple",
			slice:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			offset:        0,
			count:         10,
			expected:      []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			expectedCount: 10,
		},
		{
			name:          "multiple offset",
			slice:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			offset:        5,
			count:         5,
			expected:      []string{"f", "g", "h", "i", "j"},
			expectedCount: 5,
		},
		{
			name:          "multiple offset end",
			slice:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			offset:        10,
			count:         5,
			expected:      []string{},
			expectedCount: 0,
		},
		{
			name:          "multiple offset past end",
			slice:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			offset:        11,
			count:         5,
			expected:      []string{},
			expectedCount: 0,
		},
		{
			name:          "multiple offset count past end",
			slice:         []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
			offset:        5,
			count:         20,
			expected:      []string{"f", "g", "h", "i", "j"},
			expectedCount: 5,
		},
	}
	for _, test := range testTable {
		t.Run(test.name, func(t *testing.T) {
			tree := avl.NewTree()
			for _, s := range test.slice {
				tree.Set(s, struct{}{})
			}
			slice := sliceWithOffset(tree, test.offset, test.count)
			if len(slice) != test.expectedCount {
				t.Errorf("expected %v, got %v", test.expectedCount, len(slice))
			}
		})
	}
}
