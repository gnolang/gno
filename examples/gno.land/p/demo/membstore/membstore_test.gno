package membstore

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

		_, err := m.Member(testutils.TestAddress("random"))
		uassert.ErrorIs(t, err, ErrMissingMember)
	})

	t.Run("valid member fetched", func(t *testing.T) {
		t.Parallel()

		// Create a non-empty store
		members := generateMembers(t, 1)

		m := NewMembStore(WithInitialMembers(members))

		_, err := m.Member(members[0].Address)
		uassert.NoError(t, err)
	})
}

func TestMembStore_GetMembers(t *testing.T) {
	t.Parallel()

	t.Run("no members", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		members := m.Members(0, 10)
		uassert.Equal(t, 0, len(members))
	})

	t.Run("proper pagination", func(t *testing.T) {
		t.Parallel()

		var (
			numMembers = maxRequestMembers * 2
			halfRange  = numMembers / 2

			members = generateMembers(t, numMembers)
			m       = NewMembStore(WithInitialMembers(members))

			verifyMembersPresent = func(members, fetchedMembers []dao.Member) {
				for _, fetchedMember := range fetchedMembers {
					for _, member := range members {
						if member.Address != fetchedMember.Address {
							continue
						}

						uassert.Equal(t, member.VotingPower, fetchedMember.VotingPower)
					}
				}
			}
		)

		urequire.Equal(t, numMembers, m.Size())

		fetchedMembers := m.Members(0, uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedMembers))

		// Verify the members
		verifyMembersPresent(members, fetchedMembers)

		// Fetch the other half
		fetchedMembers = m.Members(uint64(halfRange), uint64(halfRange))
		urequire.Equal(t, halfRange, len(fetchedMembers))

		// Verify the members
		verifyMembersPresent(members, fetchedMembers)
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
		m := NewMembStore(WithDAOPkgPath("gno.land/r/gov/dao"))

		// Attempt to add a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.AddMember(member), ErrNotGovDAO)
	})

	t.Run("member already exists", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath), WithInitialMembers(members))

		// Attempt to add a member
		uassert.ErrorIs(t, m.AddMember(members[0]), ErrAlreadyMember)
	})

	t.Run("new member added", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create an empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath))

		// Attempt to add a member
		urequire.NoError(t, m.AddMember(members[0]))

		// Make sure the member is added
		uassert.True(t, m.IsMember(members[0].Address))
	})
}

func TestMembStore_Size(t *testing.T) {
	t.Parallel()

	t.Run("empty govdao", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore()

		uassert.Equal(t, 0, m.Size())
	})

	t.Run("non-empty govdao", func(t *testing.T) {
		t.Parallel()

		// Create a non-empty store
		members := generateMembers(t, 50)
		m := NewMembStore(WithInitialMembers(members))

		uassert.Equal(t, len(members), m.Size())
	})
}

func TestMembStore_UpdateMember(t *testing.T) {
	t.Parallel()

	t.Run("caller not govdao", func(t *testing.T) {
		t.Parallel()

		// Create an empty store
		m := NewMembStore(WithDAOPkgPath("gno.land/r/gov/dao"))

		// Attempt to update a member
		member := generateMembers(t, 1)[0]
		uassert.ErrorIs(t, m.UpdateMember(member.Address, member), ErrNotGovDAO)
	})

	t.Run("non-existing member", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create an empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath))

		// Attempt to update a member
		uassert.ErrorIs(t, m.UpdateMember(members[0].Address, members[0]), ErrMissingMember)
	})

	t.Run("overwrite member attempt", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 2)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath), WithInitialMembers(members))

		// Attempt to update a member
		uassert.ErrorIs(t, m.UpdateMember(members[0].Address, members[1]), ErrInvalidAddressUpdate)
	})

	t.Run("successful update", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath), WithInitialMembers(members))

		oldVotingPower := m.totalVotingPower
		urequire.Equal(t, members[0].VotingPower, oldVotingPower)

		votingPower := uint64(300)
		members[0].VotingPower = votingPower

		// Attempt to update a member
		uassert.NoError(t, m.UpdateMember(members[0].Address, members[0]))
		uassert.Equal(t, votingPower, m.Members(0, 10)[0].VotingPower)
		urequire.Equal(t, votingPower, m.totalVotingPower)
	})

	t.Run("member removed", func(t *testing.T) {
		t.Parallel()

		var (
			// Execute as the /r/gov/dao caller
			daoPkgPath = "gno.land/r/gov/dao"
			r          = std.NewCodeRealm(daoPkgPath)
		)

		std.TestSetRealm(r)

		// Create a non-empty store
		members := generateMembers(t, 1)
		m := NewMembStore(WithDAOPkgPath(daoPkgPath), WithInitialMembers(members))

		votingPower := uint64(0)
		members[0].VotingPower = votingPower

		// Attempt to update a member
		uassert.NoError(t, m.UpdateMember(members[0].Address, members[0]))

		// Make sure the member was removed
		uassert.False(t, m.IsMember(members[0].Address))
	})
}
