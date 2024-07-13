package simpledao

import (
	"testing"

	"std"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/urequire"
	"gno.land/p/gov/dao"
)

// generateMembers generates dummy govdao members
func generateMembers(t *testing.T, count int) []dao.Member {
	t.Helper()

	members := make([]dao.Member, 0, count)

	for i := 0; i < count; i++ {
		members = append(members, dao.Member{
			Address:     testutils.TestAddress(ufmt.Sprintf("member %d", i)),
			Name:        "test member",
			VotingPower: 10,
		})
	}

	return members
}

func TestMembStore_GetMember(t *testing.T) {
	t.Parallel()

	t.Run("member not found", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		_, err := m.GetMember(testutils.TestAddress("random"))
		uassert.ErrorIs(t, err, errMissingMember)
	})

	t.Run("valid member fetched", func(t *testing.T) {
		t.Parallel()

		// Create a non-empty store
		members := generateMembers(t, 1)

		m := NewMembStore(WithInitialMembers(members))

		_, err := m.GetMember(members[0].Address)
		uassert.NoError(t, err)
	})
}

func TestMembStore_GetMembers(t *testing.T) {
	t.Parallel()

	t.Run("no members", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		members := m.GetMembers()
		uassert.Equal(t, 0, len(members))
	})

	t.Run("valid members fetched", func(t *testing.T) {
		t.Parallel()

		// Create a non-empty store
		members := generateMembers(t, 50)

		m := NewMembStore(WithInitialMembers(members))

		fetchedMembers := m.GetMembers()

		urequire.Equal(t, len(members), len(fetchedMembers))

		for _, fetchedMember := range fetchedMembers {
			for _, member := range members {
				if member.Address != fetchedMember.Address {
					continue
				}

				uassert.Equal(t, member.VotingPower, fetchedMember.VotingPower)
				uassert.Equal(t, member.Name, fetchedMember.Name)
			}
		}
	})
}

func TestMembStore_IsMember(t *testing.T) {
	t.Parallel()

	t.Run("non-existing member", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		uassert.False(t, m.IsMember(testutils.TestAddress("random")))
	})

	t.Run("existing member", func(t *testing.T) {
		t.Parallel()

		// Create a non-empty store
		members := generateMembers(t, 50)

		m := NewMembStore(WithInitialMembers(members))

		for _, member := range members {
			uassert.True(t, m.IsMember(member.Address))
		}
	})
}

func TestMembStore_AddMember(t *testing.T) {
	t.Parallel()

	t.Run("caller not govdao", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		// Attempt to add a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.AddMember(member), errNotGovDAO)
	})

	t.Run("member already exists", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithInitialMembers(members))

		// Attempt to add a member
		uassert.ErrorIs(t, m.AddMember(members[0]), errAlreadyMember)
	})

	t.Run("new member added", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create an empty store
		members := generateMembers(t, 1)
		m := NewMembStore()

		// Attempt to add a member
		urequire.NoError(t, m.AddMember(members[0]))

		// Make sure the member is added
		uassert.True(t, m.IsMember(members[0].Address))
	})
}

func TestMembStore_RemoveMember(t *testing.T) {
	t.Parallel()

	t.Run("caller not govdao", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		// Attempt to add a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.RemoveMember(member.Address), errNotGovDAO)
	})

	t.Run("member does not exist", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create an empty store
		m := NewMembStore()

		// Attempt to add a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.RemoveMember(member.Address), errMissingMember)
	})

	t.Run("member removed", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithInitialMembers(members))

		// Attempt to remove a member
		urequire.NoError(t, m.RemoveMember(members[0].Address))

		// Make sure the member is removed
		uassert.False(t, m.IsMember(members[0].Address))
	})
}

func TestMembStore_UpdateMember(t *testing.T) {
	t.Parallel()

	t.Run("caller not govdao", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		// Attempt to update a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.UpdateMember(member.Address, member), errNotGovDAO)
	})

	t.Run("non-existing member", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create an empty store
		members := generateMembers(t, 1)
		m := NewMembStore()

		// Attempt to update a member
		uassert.ErrorIs(t, m.UpdateMember(members[0].Address, members[0]), errMissingMember)
	})

	t.Run("overwrite member attempt", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 2)
		m := NewMembStore(WithInitialMembers(members))

		// Attempt to update a member
		uassert.ErrorIs(t, m.UpdateMember(members[0].Address, members[1]), errInvalidAddressUpdate)
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		// Execute as the /r/gov/dao caller
		r := std.NewCodeRealm(daoPkgPath)
		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithInitialMembers(members))

		name := "new name!"
		members[0].Name = name

		// Attempt to update a member
		uassert.NoError(t, m.UpdateMember(members[0].Address, members[0]))
		uassert.Equal(t, name, m.GetMembers()[0].Name)
	})
}
