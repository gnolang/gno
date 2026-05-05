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

func TestGetTotalMemberStorageSize(t *testing.T) {
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
			name: "one member",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: 1,
		},
		{
			name: "multiple members",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			want: 2,
		},
		{
			name: "one grouped member",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				g, _ := s.Grouping().Add("A")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			want: 1,
		},
		{
			name: "multiple grouped members",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()

				g, _ := s.Grouping().Add("A")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				g.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")

				g, _ = s.Grouping().Add("B")
				g.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: 3,
		},
		{
			name: "multiple members with grouping",
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
			got := commondao.GetTotalMemberStorageSize(storage)

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

func TestExistsInMemberStorage(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name  string
		setup func() commondao.MemberStorage
		want  bool
	}{
		{
			name: "empty storage",
			setup: func() commondao.MemberStorage {
				return commondao.NewMemberStorage()
			},
			want: false,
		},
		{
			name: "exists",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(member)
				return s
			},
			want: true,
		},
		{
			name: "not found",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: false,
		},
		{
			name: "exists in grouping storage",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				g, _ := s.Grouping().Add("foo")
				g.Members().Add(member)
				return s
			},
			want: true,
		},
		{
			name: "not found in grouping storage",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				g, _ := s.Grouping().Add("foo")
				g.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			storage := tc.setup()

			// Act
			got := commondao.ExistsInMemberStorage(storage, member)

			// Assert
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestIterateMemberStorage(t *testing.T) {
	// Sorted address order: u1 < u2 < a1 < a2 < b1.
	var (
		u1 address = "g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt"
		u2 address = "g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq"
		a1 address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
		a2 address = "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"
		b1 address = "g1vh7krmmzfua5xjmkatvmx09z37w34lsvd2mxa5"
	)

	// mixed builds a storage with ungrouped {u1, u2}, group A {a1, a2}, group B {b1}.
	mixed := func() commondao.MemberStorage {
		s := commondao.NewMemberStorageWithGrouping()
		s.Add(u1)
		s.Add(u2)
		ga, _ := s.Grouping().Add("A")
		ga.Members().Add(a1)
		ga.Members().Add(a2)
		gb, _ := s.Grouping().Add("B")
		gb.Members().Add(b1)
		return s
	}

	cases := []struct {
		name        string
		setup       func() commondao.MemberStorage
		offset      int
		count       int
		stopAfter   int // when > 0, fn returns true on the Nth visit (1-indexed)
		wantStopped bool
		want        []address
	}{
		{
			name:  "empty storage without grouping",
			setup: func() commondao.MemberStorage { return commondao.NewMemberStorage() },
			count: 10,
		},
		{
			name:  "empty storage with grouping",
			setup: func() commondao.MemberStorage { return commondao.NewMemberStorageWithGrouping() },
			count: 10,
		},
		{
			name: "ungrouped only - full walk",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			count: 10,
			want:  []address{u1, u2},
		},
		{
			name: "ungrouped only - offset skips first",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			offset: 1,
			count:  10,
			want:   []address{u2},
		},
		{
			name: "ungrouped only - count limits walk",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			count: 1,
			want:  []address{u1},
		},
		{
			name: "ungrouped only - offset beyond size",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorage()
				s.Add(u1)
				return s
			},
			offset: 10,
			count:  10,
		},
		{
			name: "grouped only - full walk in group order",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				ga, _ := s.Grouping().Add("A")
				ga.Members().Add(a1)
				ga.Members().Add(a2)
				gb, _ := s.Grouping().Add("B")
				gb.Members().Add(b1)
				return s
			},
			count: 10,
			want:  []address{a1, a2, b1},
		},
		{
			name:  "mixed - ungrouped first then groups in order",
			setup: mixed,
			count: 10,
			want:  []address{u1, u2, a1, a2, b1},
		},
		{
			name:   "mixed - offset spans into groups",
			setup:  mixed,
			offset: 2,
			count:  10,
			want:   []address{a1, a2, b1},
		},
		{
			name:   "mixed - offset spans ungrouped and first group",
			setup:  mixed,
			offset: 4,
			count:  10,
			want:   []address{b1},
		},
		{
			name:   "mixed - offset beyond total",
			setup:  mixed,
			offset: 100,
			count:  10,
		},
		{
			name:  "mixed - count limits within ungrouped",
			setup: mixed,
			count: 1,
			want:  []address{u1},
		},
		{
			name:  "mixed - count limits across boundary",
			setup: mixed,
			count: 3,
			want:  []address{u1, u2, a1},
		},
		{
			name:        "mixed - fn stops in ungrouped",
			setup:       mixed,
			count:       10,
			stopAfter:   1,
			wantStopped: true,
			want:        []address{u1},
		},
		{
			name:        "mixed - fn stops in group",
			setup:       mixed,
			count:       10,
			stopAfter:   3,
			wantStopped: true,
			want:        []address{u1, u2, a1},
		},
		{
			name:  "mixed - count zero",
			setup: mixed,
		},
		{
			name:  "mixed - count negative",
			setup: mixed,
			count: -5,
		},
		{
			name: "duplicate across groups visited twice",
			setup: func() commondao.MemberStorage {
				s := commondao.NewMemberStorageWithGrouping()
				ga, _ := s.Grouping().Add("A")
				ga.Members().Add(a1)
				gb, _ := s.Grouping().Add("B")
				gb.Members().Add(a1)
				return s
			},
			count: 10,
			want:  []address{a1, a1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			storage := tc.setup()

			var visited []address
			fn := func(member address) bool {
				visited = append(visited, member)
				return tc.stopAfter > 0 && len(visited) == tc.stopAfter
			}

			// Act
			stopped := commondao.IterateMemberStorage(storage, tc.offset, tc.count, fn)

			// Assert
			uassert.Equal(t, tc.wantStopped, stopped, "expect stopped to match")
			urequire.Equal(t, len(tc.want), len(visited), "expect visited count to match")
			for i, want := range tc.want {
				uassert.Equal(t, want, visited[i], "expect member at visit index to match")
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
