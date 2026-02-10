package mgroup

import (
	"testing"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ownable"
	"gno.land/p/nt/uassert"
)

func TestManagedGroup(t *testing.T) {
	u1 := testing.NewCodeRealm("gno.land/r/test/test/u1")
	u2 := testing.NewCodeRealm("gno.land/r/test/test/u2")
	u3 := testing.NewCodeRealm("gno.land/r/test/test/u3")
	u4 := testing.NewCodeRealm("gno.land/r/test/test/u4")

	// testUsingPreviousRealm sets the realm to "gno.land/r/test/test"
	// and uses the SetRealm in the test as the previous caller
	testUsingPreviousRealm := func(fn func()) {
		testing.SetRealm(testing.NewCodeRealm("gno.land/r/test/test"))
		fn()
	}

	t.Run("AddBackupOwner", func(t *testing.T) {
		g := New(u1.Address())

		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			if err := g.AddBackupOwner(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			// ensure invalid address is caught
			var badAddr address
			if err := g.AddBackupOwner(badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
		})

		testing.SetRealm(u3)
		testUsingPreviousRealm(func() {
			// an address that is not the owner should not be allowed to add a backup owner
			if err := g.AddBackupOwner(u4.Address()); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		})

	})
	t.Run("RemoveBackupOwner", func(t *testing.T) {
		g := New(u1.Address())

		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			g.AddBackupOwner(u2.Address())
			if err := g.RemoveBackupOwner(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			// running this twice should not error.
			if err := g.RemoveBackupOwner(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			var badAddr address
			if err := g.RemoveBackupOwner(badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
			if err := g.RemoveBackupOwner(u1.Address()); err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		})

		testing.SetRealm(u3)
		testUsingPreviousRealm(func() {
			g.AddBackupOwner(u2.Address())
			if err := g.RemoveBackupOwner(u2.Address()); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		})
	})

	t.Run("ClaimOwnership", func(t *testing.T) {
		g := New(u1.Address())

		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			g.AddBackupOwner(u2.Address())
		})
		testing.SetRealm(u2)
		testUsingPreviousRealm(func() {
			if err := g.ClaimOwnership(); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.Owner() != u2.Address() {
				t.Errorf("expected %v, got %v", u2, g.Owner())
			}
			if !g.IsMember(u2.Address()) {
				t.Errorf("expected %v to be a member", u2)
			}
			if err := g.ClaimOwnership(); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		})
		testing.SetRealm(u3)
		testUsingPreviousRealm(func() {
			if err := g.ClaimOwnership(); err != ErrNotMember {
				t.Errorf("expected %v, got %v", ErrNotMember.Error(), err.Error())
			}
		})
	})

	t.Run("AddMember", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			if err := g.AddMember(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if !g.IsMember(u2.Address()) {
				t.Errorf("expected %v to be a member", u2)
			}
			var badAddr address
			if err := g.AddMember(badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress, err.Error())
			}
		})
		testing.SetRealm(u3)
		testUsingPreviousRealm(func() {
			if err := g.AddMember(u4.Address()); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		})
	})

	t.Run("RemoveMember", func(t *testing.T) {
		g := New(u1.Address())

		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			g.AddMember(u2.Address())
			if err := g.RemoveMember(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.IsMember(u2.Address()) {
				t.Errorf("expected %v to not be a member", u2)
			}
			// running this twice should not error
			if err := g.RemoveMember(u2.Address()); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			var badAddr address
			if err := g.RemoveMember(badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
			if err := g.RemoveMember(u1.Address()); err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		})

		testing.SetRealm(u2)
		testUsingPreviousRealm(func() {
			if err := g.RemoveMember(u3.Address()); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized.Error(), err.Error())
			}
		})
	})

	t.Run("MemberCount", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			uassert.Equal(t, 1, g.MemberCount())
			g.AddMember(u2.Address())
			uassert.Equal(t, 2, g.MemberCount())
			g.AddMember(u3.Address())
			uassert.Equal(t, 3, g.MemberCount())
			g.RemoveMember(u2.Address())
			uassert.Equal(t, 2, g.MemberCount())
		})
	})

	t.Run("BackupOwnerCount", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			uassert.Equal(t, 1, g.BackupOwnerCount())
			g.AddBackupOwner(u2.Address())
			uassert.Equal(t, 2, g.BackupOwnerCount())
			g.AddBackupOwner(u3.Address())
			uassert.Equal(t, 3, g.BackupOwnerCount())
			g.RemoveBackupOwner(u2.Address())
			uassert.Equal(t, 2, g.BackupOwnerCount())
		})
	})

	t.Run("IsMember", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			if !g.IsMember(u1.Address()) {
				t.Errorf("expected %v to be a member", u1)
			}
			if g.IsMember(u2.Address()) {
				t.Errorf("expected %v to not be a member", u2)
			}
			g.AddMember(u2.Address())
			if !g.IsMember(u2.Address()) {
				t.Errorf("expected %v to be a member", u2)
			}
		})
	})
	t.Run("IsBackupOwner", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			if !g.IsBackupOwner(u1.Address()) {
				t.Errorf("expected %v to be a backup owner", u1)
			}
			if g.IsBackupOwner(u2.Address()) {
				t.Errorf("expected %v to not be a backup owner", u2)
			}
			g.AddBackupOwner(u2.Address())
			if !g.IsBackupOwner(u2.Address()) {
				t.Errorf("expected %v to be a backup owner", u2)
			}
		})
	})

	t.Run("Owner", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			if g.Owner() != u1.Address() {
				t.Errorf("expected %v, got %v", u1, g.Owner())
			}
			g.AddBackupOwner(u2.Address())
			if g.Owner() != u1.Address() {
				t.Errorf("expected %v, got %v", u1, g.Owner())
			}
		})
		testing.SetRealm(u2)
		testUsingPreviousRealm(func() {
			g.ClaimOwnership()
			if g.Owner() != u2.Address() {
				t.Errorf("expected %v, got %v", u2, g.Owner())
			}
		})
	})
	t.Run("BackupOwners", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			g.AddBackupOwner(u2.Address())
			g.AddBackupOwner(u3.Address())
			owners := g.BackupOwners()
			if len(owners) != 3 {
				t.Errorf("expected 3, got %v", len(owners))
			}
			if owners[0] != u1.Address().String() {
				t.Errorf("expected %v, got %v", u1, owners[0])
			}
			if owners[1] != u2.Address().String() {
				t.Errorf("expected %v, got %v", u2, owners[1])
			}
			if owners[2] != u3.Address().String() {
				t.Errorf("expected %v, got %v", u3, owners[1])
			}
		})
	})
	t.Run("Members", func(t *testing.T) {
		g := New(u1.Address())
		testing.SetRealm(u1)
		testUsingPreviousRealm(func() {
			g.AddMember(u2.Address())
			g.AddMember(u3.Address())
			members := g.Members()
			if len(members) != 3 {
				t.Errorf("expected 3, got %v", len(members))
			}
			if members[0] != u1.Address().String() {
				t.Errorf("expected %v, got %v", u1, members[0])
			}
			if members[1] != u2.Address().String() {
				t.Errorf("expected %v, got %v", u2, members[1])
			}
			if members[2] != u3.Address().String() {
				t.Errorf("expected %v, got %v", u3, members[1])
			}
		})
	})
}

func TestSliceWithOffset(t *testing.T) {
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
