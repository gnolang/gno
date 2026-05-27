package mgroup

import (
	"testing"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ownable/v0"
	"gno.land/p/nt/testutils/v0"
	"gno.land/p/nt/uassert/v0"
)

func TestManagedGroup(cur realm, t *testing.T) {
	u1 := testutils.TestAddress("u1")
	u2 := testutils.TestAddress("u2")
	u3 := testutils.TestAddress("u3")
	u4 := testutils.TestAddress("u4")

	t.Run("AddBackupOwner", func(cur realm, t *testing.T) {
		g := New(u1)

		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			if err := g.AddBackupOwner(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			// ensure invalid address is caught
			var badAddr address
			if err := g.AddBackupOwner(0, cur, badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
		}(cross(cur))

		testing.SetRealm(testing.NewUserRealm(u3))
		func(cur realm) {
			// an address that is not the owner should not be allowed to add a backup owner
			if err := g.AddBackupOwner(0, cur, u4); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		}(cross(cur))

	})
	t.Run("RemoveBackupOwner", func(cur realm, t *testing.T) {
		g := New(u1)

		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			g.AddBackupOwner(0, cur, u2)
			if err := g.RemoveBackupOwner(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			// running this twice should not error.
			if err := g.RemoveBackupOwner(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			var badAddr address
			if err := g.RemoveBackupOwner(0, cur, badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
			if err := g.RemoveBackupOwner(0, cur, u1); err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		}(cross(cur))

		testing.SetRealm(testing.NewUserRealm(u3))
		func(cur realm) {
			g.AddBackupOwner(0, cur, u2)
			if err := g.RemoveBackupOwner(0, cur, u2); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		}(cross(cur))
	})

	t.Run("ClaimOwnership", func(cur realm, t *testing.T) {
		g := New(u1)

		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			g.AddBackupOwner(0, cur, u2)
		}(cross(cur))
		testing.SetRealm(testing.NewUserRealm(u2))
		func(cur realm) {
			if err := g.ClaimOwnership(0, cur); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.Owner() != u2 {
				t.Errorf("expected %v, got %v", u2, g.Owner())
			}
			if !g.IsMember(u2) {
				t.Errorf("expected %v to be a member", u2)
			}
			if err := g.ClaimOwnership(0, cur); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
		}(cross(cur))
		testing.SetRealm(testing.NewUserRealm(u3))
		func(cur realm) {
			if err := g.ClaimOwnership(0, cur); err != ErrNotMember {
				t.Errorf("expected %v, got %v", ErrNotMember.Error(), err.Error())
			}
		}(cross(cur))
	})

	t.Run("AddMember", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			if err := g.AddMember(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if !g.IsMember(u2) {
				t.Errorf("expected %v to be a member", u2)
			}
			var badAddr address
			if err := g.AddMember(0, cur, badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress, err.Error())
			}
		}(cross(cur))
		testing.SetRealm(testing.NewUserRealm(u3))
		func(cur realm) {
			if err := g.AddMember(0, cur, u4); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized, err.Error())
			}
		}(cross(cur))
	})

	t.Run("RemoveMember", func(cur realm, t *testing.T) {
		g := New(u1)

		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			g.AddMember(0, cur, u2)
			if err := g.RemoveMember(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			if g.IsMember(u2) {
				t.Errorf("expected %v to not be a member", u2)
			}
			// running this twice should not error
			if err := g.RemoveMember(0, cur, u2); err != nil {
				t.Errorf("expected nil, got %v", err.Error())
			}
			var badAddr address
			if err := g.RemoveMember(0, cur, badAddr); err != ErrInvalidAddress {
				t.Errorf("expected %v, got %v", ErrInvalidAddress.Error(), err.Error())
			}
			if err := g.RemoveMember(0, cur, u1); err != ErrCannotRemoveOwner {
				t.Errorf("expected %v, got %v", ErrCannotRemoveOwner.Error(), err.Error())
			}
		}(cross(cur))

		testing.SetRealm(testing.NewUserRealm(u2))
		func(cur realm) {
			if err := g.RemoveMember(0, cur, u3); err != ownable.ErrUnauthorized {
				t.Errorf("expected %v, got %v", ownable.ErrUnauthorized.Error(), err.Error())
			}
		}(cross(cur))
	})

	t.Run("MemberCount", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			uassert.Equal(t, 1, g.MemberCount())
			g.AddMember(0, cur, u2)
			uassert.Equal(t, 2, g.MemberCount())
			g.AddMember(0, cur, u3)
			uassert.Equal(t, 3, g.MemberCount())
			g.RemoveMember(0, cur, u2)
			uassert.Equal(t, 2, g.MemberCount())
		}(cross(cur))
	})

	t.Run("BackupOwnerCount", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			uassert.Equal(t, 1, g.BackupOwnerCount())
			g.AddBackupOwner(0, cur, u2)
			uassert.Equal(t, 2, g.BackupOwnerCount())
			g.AddBackupOwner(0, cur, u3)
			uassert.Equal(t, 3, g.BackupOwnerCount())
			g.RemoveBackupOwner(0, cur, u2)
			uassert.Equal(t, 2, g.BackupOwnerCount())
		}(cross(cur))
	})

	t.Run("IsMember", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			if !g.IsMember(u1) {
				t.Errorf("expected %v to be a member", u1)
			}
			if g.IsMember(u2) {
				t.Errorf("expected %v to not be a member", u2)
			}
			g.AddMember(0, cur, u2)
			if !g.IsMember(u2) {
				t.Errorf("expected %v to be a member", u2)
			}
		}(cross(cur))
	})
	t.Run("IsBackupOwner", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			if !g.IsBackupOwner(u1) {
				t.Errorf("expected %v to be a backup owner", u1)
			}
			if g.IsBackupOwner(u2) {
				t.Errorf("expected %v to not be a backup owner", u2)
			}
			g.AddBackupOwner(0, cur, u2)
			if !g.IsBackupOwner(u2) {
				t.Errorf("expected %v to be a backup owner", u2)
			}
		}(cross(cur))
	})

	t.Run("Owner", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			if g.Owner() != u1 {
				t.Errorf("expected %v, got %v", u1, g.Owner())
			}
			g.AddBackupOwner(0, cur, u2)
			if g.Owner() != u1 {
				t.Errorf("expected %v, got %v", u1, g.Owner())
			}
		}(cross(cur))
		testing.SetRealm(testing.NewUserRealm(u2))
		func(cur realm) {
			g.ClaimOwnership(0, cur)
			if g.Owner() != u2 {
				t.Errorf("expected %v, got %v", u2, g.Owner())
			}
		}(cross(cur))
	})
	t.Run("BackupOwners", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			g.AddBackupOwner(0, cur, u2)
			g.AddBackupOwner(0, cur, u3)
			owners := g.BackupOwners()
			if len(owners) != 3 {
				t.Errorf("expected 3, got %v", len(owners))
			}
			// BackupOwners iterates AVL by address-string key. TestAddress
			// does not preserve u1<u2<u3 order, so assert set membership.
			set := map[string]bool{owners[0]: true, owners[1]: true, owners[2]: true}
			for _, addr := range []address{u1, u2, u3} {
				if !set[addr.String()] {
					t.Errorf("expected %v in backup owners", addr)
				}
			}
		}(cross(cur))
	})
	t.Run("Members", func(cur realm, t *testing.T) {
		g := New(u1)
		testing.SetRealm(testing.NewUserRealm(u1))
		func(cur realm) {
			g.AddMember(0, cur, u2)
			g.AddMember(0, cur, u3)
			members := g.Members()
			if len(members) != 3 {
				t.Errorf("expected 3, got %v", len(members))
			}
			set := map[string]bool{members[0]: true, members[1]: true, members[2]: true}
			for _, addr := range []address{u1, u2, u3} {
				if !set[addr.String()] {
					t.Errorf("expected %v in members", addr)
				}
			}
		}(cross(cur))
	})
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
