package addrset

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestSet(t *testing.T) {
	addr1 := address("addr1")
	addr2 := address("addr2")
	addr3 := address("addr3")

	tests := []struct {
		name    string
		actions func(s *Set)
		size    int
		has     map[address]bool
		addrs   []address // for iteration checks
	}{
		{
			name:    "empty set",
			actions: func(s *Set) {},
			size:    0,
			has:     map[address]bool{addr1: false},
		},
		{
			name: "single address",
			actions: func(s *Set) {
				s.Add(addr1)
			},
			size: 1,
			has: map[address]bool{
				addr1: true,
				addr2: false,
			},
			addrs: []address{addr1},
		},
		{
			name: "multiple addresses",
			actions: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Add(addr3)
			},
			size: 3,
			has: map[address]bool{
				addr1: true,
				addr2: true,
				addr3: true,
			},
			addrs: []address{addr1, addr2, addr3},
		},
		{
			name: "remove address",
			actions: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Remove(addr1)
			},
			size: 1,
			has: map[address]bool{
				addr1: false,
				addr2: true,
			},
			addrs: []address{addr2},
		},
		{
			name: "duplicate adds",
			actions: func(s *Set) {
				uassert.True(t, s.Add(addr1))     // first add returns true
				uassert.False(t, s.Add(addr1))    // second add returns false
				uassert.True(t, s.Remove(addr1))  // remove existing returns true
				uassert.False(t, s.Remove(addr1)) // remove non-existing returns false
			},
			size: 0,
			has: map[address]bool{
				addr1: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var set Set

			// Execute test actions
			tt.actions(&set)

			// Check size
			uassert.Equal(t, tt.size, set.Size())

			// Check existence
			for addr, expected := range tt.has {
				uassert.Equal(t, expected, set.Has(addr))
			}

			// Check iteration if addresses are specified
			if tt.addrs != nil {
				collected := []address{}
				set.IterateByOffset(0, 10, func(addr address) bool {
					collected = append(collected, addr)
					return false
				})

				// Check length
				uassert.Equal(t, len(tt.addrs), len(collected))

				// Check each address
				for i, addr := range tt.addrs {
					uassert.Equal(t, addr, collected[i])
				}
			}
		})
	}
}

func TestSetIterationLimits(t *testing.T) {
	tests := []struct {
		name     string
		addrs    []address
		offset   int
		limit    int
		expected int
	}{
		{
			name:     "zero offset full list",
			addrs:    []address{"a1", "a2", "a3"},
			offset:   0,
			limit:    10,
			expected: 3,
		},
		{
			name:     "offset with limit",
			addrs:    []address{"a1", "a2", "a3", "a4"},
			offset:   1,
			limit:    2,
			expected: 2,
		},
		{
			name:     "offset beyond size",
			addrs:    []address{"a1", "a2"},
			offset:   3,
			limit:    1,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var set Set
			for _, addr := range tt.addrs {
				set.Add(addr)
			}

			// Test forward iteration
			count := 0
			set.IterateByOffset(tt.offset, tt.limit, func(addr address) bool {
				count++
				return false
			})
			uassert.Equal(t, tt.expected, count)

			// Test reverse iteration
			count = 0
			set.ReverseIterateByOffset(tt.offset, tt.limit, func(addr address) bool {
				count++
				return false
			})
			uassert.Equal(t, tt.expected, count)
		})
	}
}
