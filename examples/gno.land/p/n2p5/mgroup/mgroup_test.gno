package mgroup

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestManagedGroup(t *testing.T) {
	t.Parallel()
	testing.SetRealm(std.NewCodeRealm("gno.land/r/test/test"))

	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")
	u3 := testutils.TestAddress("u3")

	t.Run("AddBackupOwner", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		// happy path
		{
			testing.SetOriginCaller(u1)
			err := g.AddBackupOwner(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// ensure checking for authorized caller
		{
			testing.SetOriginCaller(u2)
			err := g.AddBackupOwner(u3)
			if err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ErrNotBackupOwner.Error(), err.Error())
			}
		}
		// ensure invalid address is caught
		{
			testing.SetOriginCaller(u1)
			var badAddr std.Address
			err := g.AddBackupOwner(badAddr)
			if err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
		}
	})
	t.Run("RemoveBackupOwner", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		// happy path
		{
			testing.SetOriginCaller(u1)
			g.AddBackupOwner(u2)
			err := g.RemoveBackupOwner(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// running this twice should not error.
		{
			testing.SetOriginCaller(u1)
			err := g.RemoveBackupOwner(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// ensure checking for authorized caller
		{
			testing.SetOriginCaller(u2)
			err := g.RemoveBackupOwner(u3)
			if err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ErrNotBackupOwner.Error(), err.Error())
			}
		}
		{
			testing.SetOriginCaller(u1)
			var badAddr std.Address
			err := g.RemoveBackupOwner(badAddr)
			if err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
		}
		{
			testing.SetOriginCaller(u1)
			err := g.RemoveBackupOwner(u1)
			if err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		}
	})
	t.Run("ClaimOwnership", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		// add backup owner
		{
			testing.SetOriginCaller(u1)
			err := g.AddBackupOwner(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// happy path
		{
			testing.SetOriginCaller(u2)
			err := g.ClaimOwnership()
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.Owner() != u2 {
				t.Errorf("expected %v, got %v", u2, g.Owner())
			}
			if !g.IsMember(u2) {
				t.Errorf("expected %v to be a member", u2)
			}
		}
		// running this twice should not error.
		{
			testing.SetOriginCaller(u2)
			err := g.ClaimOwnership()
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// ensure checking for authorized caller
		{
			testing.SetOriginCaller(u3)
			err := g.ClaimOwnership()
			if err != ErrNotMember {
				t.Errorf("expected %v, got %v", ErrNotMember.Error(), err.Error())
			}
		}
	})
	t.Run("AddMember", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		// happy path
		{
			testing.SetOriginCaller(u1)
			err := g.AddMember(u2)
			uassert.NoError(t, err)
			if !g.IsMember(u2) {
				t.Errorf("expected %v to be a member", u2)
			}
		}
		// ensure checking for authorized caller
		{
			testing.SetOriginCaller(u2)
			err := g.AddMember(u3)
			uassert.ErrorIs(t, err, ownable.ErrUnauthorized)
		}
		// ensure invalid address is caught
		{
			testing.SetOriginCaller(u1)
			var badAddr std.Address
			err := g.AddMember(badAddr)
			uassert.ErrorContains(t, err, "address is invalid")
		}
	})
	t.Run("RemoveMember", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		// happy path
		{
			testing.SetOriginCaller(u1)
			err := g.AddMember(u2)
			uassert.NoError(t, err)
			err = g.RemoveMember(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.IsMember(u2) {
				t.Errorf("expected %v to not be a member", u2)
			}
		}
		// running this twice should not error.
		{
			testing.SetOriginCaller(u1)
			err := g.RemoveMember(u2)
			if err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}
		// ensure checking for authorized caller
		{
			testing.SetOriginCaller(u2)
			err := g.RemoveMember(u3)
			if err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized.Error(), err.Error())
			}
		}
		// ensure invalid address is caught
		{
			testing.SetOriginCaller(u1)
			var badAddr std.Address
			err := g.RemoveMember(badAddr)
			if err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
		}
		// ensure owner cannot be removed
		{
			testing.SetOriginCaller(u1)
			err := g.RemoveMember(u1)
			if err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		}
	})
	t.Run("MemberCount", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		testing.SetOriginCaller(u1)

		uassert.Equal(t, 1, g.MemberCount())
		err := g.AddMember(u2)
		uassert.NoError(t, err)
		uassert.Equal(t, 2, g.MemberCount())
		err = g.AddMember(u3)
		uassert.NoError(t, err)
		uassert.Equal(t, 3, g.MemberCount())
		g.RemoveMember(u2)
		uassert.Equal(t, 2, g.MemberCount())
	})
	t.Run("BackupOwnerCount", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		testing.SetOriginCaller(u1)

		uassert.Equal(t, 1, g.BackupOwnerCount())
		g.AddBackupOwner(u2)
		uassert.Equal(t, 2, g.BackupOwnerCount())
		g.AddBackupOwner(u3)
		uassert.Equal(t, 3, g.BackupOwnerCount())
		g.RemoveBackupOwner(u2)
		uassert.Equal(t, 2, g.BackupOwnerCount())
	})
	t.Run("IsMember", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		testing.SetOriginCaller(u1)

		if !g.IsMember(u1) {
			t.Errorf("expected %v to be a member", u1)
		}
		if g.IsMember(u2) {
			t.Errorf("expected %v to not be a member", u2)
		}
		err := g.AddMember(u2)
		uassert.NoError(t, err)
		if !g.IsMember(u2) {
			t.Errorf("expected %v to be a member", u2)
		}
	})
	t.Run("IsBackupOwner", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		testing.SetOriginCaller(u1)

		if !g.IsBackupOwner(u1) {
			t.Errorf("expected %v to be a backup owner", u1)
		}
		if g.IsBackupOwner(u2) {
			t.Errorf("expected %v to not be a backup owner", u2)
		}
		g.AddBackupOwner(u2)
		if !g.IsBackupOwner(u2) {
			t.Errorf("expected %v to be a backup owner", u2)
		}
	})
	t.Run("Owner", func(t *testing.T) {
		t.Parallel()
		g := New(u1)
		testing.SetOriginCaller(u1)

		if g.Owner() != u1 {
			t.Errorf("expected %v, got %v", u1, g.Owner())
		}
		g.AddBackupOwner(u2)
		if g.Owner() != u1 {
			t.Errorf("expected %v, got %v", u1, g.Owner())
		}
		testing.SetOriginCaller(u2)
		g.ClaimOwnership()
		if g.Owner() != u2 {
			t.Errorf("expected %v, got %v", u2, g.Owner())
		}
	})
	t.Run("BackupOwners", func(t *testing.T) {
		t.Parallel()
		testing.SetOriginCaller(u1)
		g := New(u1)
		g.AddBackupOwner(u2)
		g.AddBackupOwner(u3)
		owners := g.BackupOwners()
		if len(owners) != 3 {
			t.Errorf("expected 2, got %v", len(owners))
		}
		if owners[0] != u1.String() {
			t.Errorf("expected %v, got %v", u2, owners[0])
		}
		if owners[1] != u3.String() {
			t.Errorf("expected %v, got %v", u3, owners[1])
		}
		if owners[2] != u2.String() {
			t.Errorf("expected %v, got %v", u3, owners[1])
		}
	})
	t.Run("Members", func(t *testing.T) {
		t.Parallel()
		testing.SetOriginCaller(u1)
		g := New(u1)
		err := g.AddMember(u2)
		uassert.NoError(t, err)
		err = g.AddMember(u3)
		uassert.NoError(t, err)
		members := g.Members()
		if len(members) != 3 {
			t.Errorf("expected 2, got %v", len(members))
		}
		if members[0] != u1.String() {
			t.Errorf("expected %v, got %v", u2, members[0])
		}
		if members[1] != u3.String() {
			t.Errorf("expected %v, got %v", u3, members[1])
		}
		if members[2] != u2.String() {
			t.Errorf("expected %v, got %v", u3, members[1])
		}
	})
}

func TestSliceWithOffset(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
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
