package storage_test

import (
	"testing"

	"gno.land/p/nt/commondao"
	"gno.land/p/nt/urequire"

	"gno.land/p/nt/commondao/exts/storage"
)

func TestMemberStorageGroupingMemberAdd(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() commondao.MemberStorage
		group  string
		member address
	}{
		{
			name: "empty group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				s.Grouping().Add("foo")
				return s
			},
			group:  "foo",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name: "group with members",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				group.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
				return s
			},
			group:  "foo",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name: "multiple groups with members",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				group, _ = s.Grouping().Add("bar")
				group.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
				return s
			},
			group:  "bar",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := tt.setup()
			group, _ := s.Grouping().Get(tt.group)

			// Act
			group.Members().Add(tt.member)

			// Assert
			urequire.True(t, s.Has(tt.member), "expect member to also be added to parent storage")
		})
	}
}

func TestMemberStorageGroupingMemberRemove(t *testing.T) {
	tests := []struct {
		name   string
		setup  func() commondao.MemberStorage
		group  string
		member address
	}{
		{
			name: "one group one member",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			group:  "foo",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name: "one group multiple members",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
				return s
			},
			group:  "foo",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
		{
			name: "multiple groups multiple members",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group, _ = s.Grouping().Add("bar")
				group.Members().Add("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
				return s
			},
			group:  "foo",
			member: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := tt.setup()
			group, _ := s.Grouping().Get(tt.group)

			// Act
			group.Members().Remove(tt.member)

			// Assert
			urequire.False(t, s.Has(tt.member), "expect member to also be removed from parent storage")
		})
	}
}

func TestMemberStorageSize(t *testing.T) {
	tests := []struct {
		name  string
		setup func() commondao.MemberStorage
		size  int
	}{
		{
			name: "empty",
			setup: func() commondao.MemberStorage {
				return storage.NewMemberStorage()
			},
			size: 0,
		},
		{
			name: "member without group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			size: 1,
		},
		{
			name: "multiple members without group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			size: 2,
		},
		{
			name: "member in group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				return s
			},
			size: 1,
		},
		{
			name: "multiple members in group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			size: 2,
		},
		{
			name: "multiple members in different groups",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group, _ = s.Grouping().Add("bar")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			size: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			s := tt.setup()

			// Act
			size := s.Size()

			// Assert
			urequire.Equal(t, tt.size, size)
		})
	}
}

func TestMemberStorageIterateByOffset(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() commondao.MemberStorage
		members []address
	}{
		{
			name: "empty",
			setup: func() commondao.MemberStorage {
				return storage.NewMemberStorage()
			},
		},
		{
			name: "without group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				s.Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				s.Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			members: []address{
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
		},
		{
			name: "single group",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			members: []address{
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
		},
		{
			name: "multiple groups",
			setup: func() commondao.MemberStorage {
				s := storage.NewMemberStorage()
				group, _ := s.Grouping().Add("foo")
				group.Members().Add("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
				group, _ = s.Grouping().Add("bar")
				group.Members().Add("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
				return s
			},
			members: []address{
				"g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq",
				"g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			var i int
			s := tt.setup()
			members := make([]address, s.Size())

			// Act
			s.IterateByOffset(0, s.Size(), func(addr address) bool {
				members[i] = addr
				i++
				return false
			})

			// Assert
			urequire.Equal(t, len(tt.members), len(members), "expect iterated members to match")

			for i, member := range members {
				urequire.Equal(t, tt.members[i], member, "expect member to match")
			}
		})
	}
}
