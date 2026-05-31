package memberstorage_test

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/p/gnoland/boards/exts/memberstorage"
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

	storage := memberstorage.NewWithGrouping()
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
		setup func() memberstorage.MemberStorage
	}{
		{
			name:  "found",
			found: true,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
		{
			name:  "not found",
			found: false,
			setup: func() memberstorage.MemberStorage { return memberstorage.New() },
		},
		{
			name:  "found in base storage with grouping",
			found: true,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

				g, _ := s.Grouping().Add("foo")
				g.Members().Add("g125t352u4pmdrr57emc4pe04y40sknr5ztng5mt")
				return s
			},
		},
		{
			name:  "not found with grouping",
			found: false,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
				g, _ := s.Grouping().Add("foo")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
		{
			name:  "not found with empty grouping",
			found: false,
			setup: func() memberstorage.MemberStorage { return memberstorage.NewWithGrouping() },
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
		setup func() memberstorage.MemberStorage
		want  int
	}{
		{
			name: "empty",
			setup: func() memberstorage.MemberStorage {
				return memberstorage.New()
			},
		},
		{
			name: "empty with grouping",
			setup: func() memberstorage.MemberStorage {
				return memberstorage.NewWithGrouping()
			},
		},
		{
			name: "one member",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: 1,
		},
		{
			name: "multiple members",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			want: 2,
		},
		{
			name: "one grouped member",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
				g, _ := s.Grouping().Add("A")
				g.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			want: 1,
		},
		{
			name: "multiple grouped members",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()

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
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
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
			got := memberstorage.GetTotalMemberStorageSize(storage)

			// Assert
			uassert.Equal(t, tc.want, got)
		})
	}
}

func TestMemberStorageAdd(t *testing.T) {
	cases := []struct {
		name   string
		member address
		want   bool
		setup  func() memberstorage.MemberStorage
	}{
		{
			name:   "success",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			want:   true,
			setup: func() memberstorage.MemberStorage {
				return memberstorage.New()
			},
		},
		{
			name:   "existing member",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			want:   false,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(cur realm, t *testing.T) {
			// Arrange
			var got bool
			storage := tc.setup()

			// Act
			got = storage.Add(tc.member)

			// Assert
			urequire.True(t, got == tc.want, "expect member to be added")
			urequire.True(t, storage.Has(tc.member), "expect member to be found")
		})
	}
}

func TestMemberStorageRemove(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name   string
		member address
		want   bool
		setup  func() memberstorage.MemberStorage
	}{
		{
			name:   "success",
			member: member,
			want:   true,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(member)
				return s
			},
		},
		{
			name:   "member not found",
			member: member,
			want:   false,
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				return s
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(cur realm, t *testing.T) {
			// Arrange
			var got bool
			storage := tc.setup()

			// Act
			got = storage.Remove(tc.member)

			// Assert
			urequire.True(t, got == tc.want, "expect member to be removed")
			urequire.False(t, storage.Has(tc.member), "expect member to be missing")
		})
	}
}

func TestExistsInMemberStorage(t *testing.T) {
	member := address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	cases := []struct {
		name  string
		setup func() memberstorage.MemberStorage
		want  bool
	}{
		{
			name: "empty storage",
			setup: func() memberstorage.MemberStorage {
				return memberstorage.New()
			},
			want: false,
		},
		{
			name: "exists",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(member)
				return s
			},
			want: true,
		},
		{
			name: "not found",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			want: false,
		},
		{
			name: "exists in grouping storage",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
				g, _ := s.Grouping().Add("foo")
				g.Members().Add(member)
				return s
			},
			want: true,
		},
		{
			name: "not found in grouping storage",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
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
			got := memberstorage.ExistsInMemberStorage(storage, member)

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
	mixed := func() memberstorage.MemberStorage {
		s := memberstorage.NewWithGrouping()
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
		setup       func() memberstorage.MemberStorage
		offset      int
		count       int
		stopAfter   int // when > 0, fn returns true on the Nth visit (1-indexed)
		wantStopped bool
		want        []address
	}{
		{
			name:  "empty storage without grouping",
			setup: func() memberstorage.MemberStorage { return memberstorage.New() },
			count: 10,
		},
		{
			name:  "empty storage with grouping",
			setup: func() memberstorage.MemberStorage { return memberstorage.NewWithGrouping() },
			count: 10,
		},
		{
			name: "ungrouped only - full walk",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			count: 10,
			want:  []address{u1, u2},
		},
		{
			name: "ungrouped only - offset skips first",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
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
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			count: 1,
			want:  []address{u1},
		},
		{
			name: "ungrouped only - offset beyond size",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(u1)
				return s
			},
			offset: 10,
			count:  10,
		},
		{
			name: "ungrouped only - negative offset treated as zero",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.New()
				s.Add(u1)
				s.Add(u2)
				return s
			},
			offset: -1,
			count:  10,
			want:   []address{u1, u2},
		},
		{
			name: "grouped only - full walk in group order",
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
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
			name:   "mixed - negative offset treated as zero",
			setup:  mixed,
			offset: -5,
			count:  10,
			want:   []address{u1, u2, a1, a2, b1},
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
			setup: func() memberstorage.MemberStorage {
				s := memberstorage.NewWithGrouping()
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
			stopped := memberstorage.IterateMemberStorage(storage, tc.offset, tc.count, fn)

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
		storage := memberstorage.NewWithGrouping()
		urequire.True(t, storage.Grouping() != nil, "expect a grouping to be returned")
	})

	t.Run("without grouping", func(t *testing.T) {
		storage := memberstorage.New()
		urequire.True(t, storage.Grouping() == nil, "expect a grouping to be nil")
	})
}

func TestIsCanonicalMemberStorage(t *testing.T) {
	cases := []struct {
		name  string
		setup func() memberstorage.MemberStorage
		want  bool
	}{
		{
			name:  "canonical without grouping",
			setup: func() memberstorage.MemberStorage { return memberstorage.New() },
			want:  true,
		},
		{
			name:  "canonical with grouping",
			setup: func() memberstorage.MemberStorage { return memberstorage.NewWithGrouping() },
			want:  true,
		},
		{
			name:  "non canonical",
			setup: func() memberstorage.MemberStorage { return stubMemberStorage{} },
			want:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uassert.Equal(t, tc.want, memberstorage.IsCanonicalMemberStorage(tc.setup()))
		})
	}
}

type stubMemberStorage struct{}

func (stubMemberStorage) Size() int                              { return 0 }
func (stubMemberStorage) Has(address) bool                       { return false }
func (stubMemberStorage) Add(address) bool                       { return false }
func (stubMemberStorage) Remove(address) bool                    { return false }
func (stubMemberStorage) Grouping() memberstorage.MemberGrouping { return nil }

func (stubMemberStorage) IterateByOffset(offset, count int, fn memberstorage.MemberIterFn) bool {
	return false
}
