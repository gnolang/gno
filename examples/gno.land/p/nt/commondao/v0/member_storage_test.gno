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
