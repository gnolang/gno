package commondao_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/nt/commondao/v0"
)

func TestMemberStorageWithGrouping(t *testing.T) {
	// Prepare
	tiers := []struct {
		Name    string
		Weight  int
		Members []address
	}{
		{
			Name:   "Tier 1",
			Weight: 3,
			Members: []address{
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
		},
		{
			Name:   "Tier 2",
			Weight: 2,
			Members: []address{
				"g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
			},
		},
	}

	storage := commondao.NewMemberStorageWithGrouping()
	for _, tier := range tiers {
		mg, err := storage.Grouping().Add(tier.Name)
		urequire.NoError(t, err, "expect no error adding tier")

		mg.SetMeta(tier.Weight)

		for _, addr := range tier.Members {
			ok := mg.Members().Add(addr)
			urequire.True(t, ok, "expect member to be added")
		}
	}

	// Assert
	for i := 0; i < len(tiers); i++ {
		tier := tiers[i]
		mg, found := storage.Grouping().Get(tier.Name)
		urequire.True(t, found, "expect member group to be found")

		v := mg.GetMeta()
		urequire.True(t, v != nil, "expect meta to be not nil")

		weight, ok := v.(int)
		urequire.True(t, ok, "expect group metadata to be an integer")
		uassert.Equal(t, tier.Weight, weight, "expect group weight to match")

		var i int
		mg.Members().IterateByOffset(0, len(tier.Members), func(addr address) bool {
			uassert.Equal(t, tier.Members[i], addr, "expect tier member to match")

			i++
			return false
		})

		uassert.Equal(t, len(tier.Members), i, "expect all tier members to be iterated")
	}
}

func TestMemberStorageHas(t *testing.T) {
	cases := []struct {
		name  string
		found bool
		setup func() commondao.MemberStorage
	}{
		{
			name:  "found",
			found: true,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
		{
			name:  "not found",
			found: false,
			setup: func() commondao.MemberStorage { return commondao.NewMemberStorage() },
		},
		{
			name:  "found in base storage with grouping",
			found: true,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

				g, _ := s.Grouping().Add("foo")
				g.Members().Add("g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt")
				return s
			},
		},
		{
			name:  "not found with grouping",
			found: false,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				g, _ := s.Grouping().Add("foo")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
		{
			name:  "not found with empty grouping",
			found: false,
			setup: func() commondao.MemberStorage { return commondao.NewMemberStorageWithGrouping() },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			storage := tc.setup()

			// Act
			found := storage.Has("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

			// Assert
			urequire.Equal(t, tc.found, found)
		})
	}
}

func TestMemberStorageTotalSize(t *testing.T) {
	cases := []struct {
		name  string
		setup func() commondao.MemberStorage
		want  int
	}{
		{
			name: "empty",
			setup: func() commondao.MemberStorage {
				return commondao.NewMemberStorage()
			},
		},
		{
			name: "empty with grouping",
			setup: func() commondao.MemberStorage {
				return commondao.NewMemberStorageWithGrouping()
			},
		},
		{
			name: "not empty",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			want: 2,
		},
		{
			name: "not empty with grouping",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")

				g, _ := s.Grouping().Add("A")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				g.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

				g, _ = s.Grouping().Add("B")
				g.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj") // Add a member that exists in other group
				return s
			},
			want: 4,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			storage := tc.setup()

			// Act
			got := storage.TotalSize()

			// Assert
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestMemberStorageAdd(t *testing.T) {
	cases := []struct {
		name     string
		member   address
		want     bool
		setup    func() commondao.MemberStorage
		panicMsg string
	}{
		{
			name:   "success",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			want:   true,
			setup: func() commondao.MemberStorage {
				return commondao.NewMemberStorage()
			},
		},
		{
			name:   "existing member",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			want:   false,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
		{
			name:   "sealed storage",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				return s.Seal().(commondao.MemberStorage)
			},
			panicMsg: "type is sealed so it can't be modified",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var got bool
			storage := tc.setup()

			// Act
			fn := func() {
				got = storage.Add(tc.member)
			}

			// Assert
			if tc.panicMsg != "" {
				urequire.PanicsWithMessage(t, tc.panicMsg, fn)
				urequire.False(t, storage.Has(tc.member), "expect member to be missing")
			} else {
				urequire.NotPanics(t, fn)
				urequire.True(t, got == tc.want, "expect member to be added")
				urequire.True(t, storage.Has(tc.member), "expect member to be found")
			}
		})
	}
}

func TestMemberStorageRemove(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name     string
		member   address
		want     bool
		setup    func() commondao.MemberStorage
		panicMsg string
	}{
		{
			name:   "success",
			member: member,
			want:   true,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(member)
				return s
			},
		},
		{
			name:   "member not found",
			member: member,
			want:   false,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				return s
			},
		},
		{
			name:   "sealed storage",
			member: member,
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(member)
				return s.Seal().(commondao.MemberStorage)
			},
			panicMsg: "type is sealed so it can't be modified",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var got bool
			storage := tc.setup()

			// Act
			fn := func() {
				got = storage.Remove(tc.member)
			}

			// Assert
			if tc.panicMsg != "" {
				urequire.PanicsWithMessage(t, tc.panicMsg, fn)
				urequire.True(t, storage.Has(tc.member), "expect member to be found")
			} else {
				urequire.NotPanics(t, fn)
				urequire.True(t, got == tc.want, "expect member to be removed")
				urequire.False(t, storage.Has(tc.member), "expect member to be missing")
			}
		})
	}
}

func TestMemberStorageGrouping(t *testing.T) {
	t.Run("with grouping", func(t *testing.T) {
		storage := commondao.NewMemberStorageWithGrouping()
		urequire.True(t, storage.Grouping() != nil, "expect a grouping to be returned")
		uassert.False(t, storage.Grouping().IsSealed(), "expect a non sealed grouping")
	})

	t.Run("without grouping", func(t *testing.T) {
		storage := commondao.NewMemberStorage()
		urequire.True(t, storage.Grouping() == nil, "expect a grouping to be nil")
	})

	t.Run("sealed storage", func(t *testing.T) {
		storage := commondao.NewMemberStorageWithGrouping()
		storage = storage.Seal().(commondao.MemberStorage)
		urequire.True(t, storage.Grouping() != nil, "expect a grouping to be returned")
		uassert.True(t, storage.Grouping().IsSealed(), "expect a sealed grouping")
	})
}
