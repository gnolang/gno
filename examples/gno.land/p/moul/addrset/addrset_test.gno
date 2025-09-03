package addrset

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestSet(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	addr3 := std.Address("addr3")

	tests := []struct {
		name    string
		actions func(s *Set)
		size    int
		has     map[std.Address]bool
		addrs   []std.Address // for iteration checks
	}{
		{
			name:    "empty set",
			actions: func(s *Set) {},
			size:    0,
			has:     map[std.Address]bool{addr1: false},
		},
		{
			name: "single address",
			actions: func(s *Set) {
				s.Add(addr1)
			},
			size: 1,
			has: map[std.Address]bool{
				addr1: true,
				addr2: false,
			},
			addrs: []std.Address{addr1},
		},
		{
			name: "multiple addresses",
			actions: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Add(addr3)
			},
			size: 3,
			has: map[std.Address]bool{
				addr1: true,
				addr2: true,
				addr3: true,
			},
			addrs: []std.Address{addr1, addr2, addr3},
		},
		{
			name: "remove address",
			actions: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Remove(addr1)
			},
			size: 1,
			has: map[std.Address]bool{
				addr1: false,
				addr2: true,
			},
			addrs: []std.Address{addr2},
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
			has: map[std.Address]bool{
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
				collected := []std.Address{}
				set.IterateByOffset(0, 10, func(addr std.Address) bool {
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
		addrs    []std.Address
		offset   int
		limit    int
		expected int
	}{
		{
			name:     "zero offset full list",
			addrs:    []std.Address{"a1", "a2", "a3"},
			offset:   0,
			limit:    10,
			expected: 3,
		},
		{
			name:     "offset with limit",
			addrs:    []std.Address{"a1", "a2", "a3", "a4"},
			offset:   1,
			limit:    2,
			expected: 2,
		},
		{
			name:     "offset beyond size",
			addrs:    []std.Address{"a1", "a2"},
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
			set.IterateByOffset(tt.offset, tt.limit, func(addr std.Address) bool {
				count++
				return false
			})
			uassert.Equal(t, tt.expected, count)

			// Test reverse iteration
			count = 0
			set.ReverseIterateByOffset(tt.offset, tt.limit, func(addr std.Address) bool {
				count++
				return false
			})
			uassert.Equal(t, tt.expected, count)
		})
	}
}

func TestNew(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	addr3 := std.Address("addr3")

	tests := []struct {
		name     string
		addrs    []std.Address
		expected int
	}{
		{
			name:     "empty constructor",
			addrs:    []std.Address{},
			expected: 0,
		},
		{
			name:     "single address",
			addrs:    []std.Address{addr1},
			expected: 1,
		},
		{
			name:     "multiple addresses",
			addrs:    []std.Address{addr1, addr2, addr3},
			expected: 3,
		},
		{
			name:     "duplicate addresses",
			addrs:    []std.Address{addr1, addr1, addr2},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := New(tt.addrs...)
			uassert.Equal(t, tt.expected, set.Size())
		})
	}
}

func TestToSlice(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	addr3 := std.Address("addr3")
	addr4 := std.Address("addr4")

	tests := []struct {
		name     string
		addrs    []std.Address
		offset   int
		limit    int
		expected []std.Address
	}{
		{
			name:     "empty set",
			addrs:    []std.Address{},
			offset:   0,
			limit:    10,
			expected: []std.Address{},
		},
		{
			name:     "full slice",
			addrs:    []std.Address{addr1, addr2, addr3},
			offset:   0,
			limit:    10,
			expected: []std.Address{addr1, addr2, addr3},
		},
		{
			name:     "partial slice with offset",
			addrs:    []std.Address{addr1, addr2, addr3, addr4},
			offset:   1,
			limit:    2,
			expected: []std.Address{addr2, addr3},
		},
		{
			name:     "offset beyond size",
			addrs:    []std.Address{addr1, addr2},
			offset:   3,
			limit:    1,
			expected: []std.Address{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := New(tt.addrs...)
			slice := set.ToSlice(tt.offset, tt.limit)
			uassert.Equal(t, len(tt.expected), len(slice))
			for i, addr := range tt.expected {
				uassert.Equal(t, addr, slice[i])
			}
		})
	}
}

func TestTree(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")

	set := New(addr1, addr2)
	tree := set.Tree()

	// Verify the tree interface works as expected
	uassert.Equal(t, 2, tree.Size())
	uassert.True(t, tree.Has(string(addr1)))
	uassert.True(t, tree.Has(string(addr2)))
	uassert.False(t, tree.Has("nonexistent"))
}

func TestIterationEarlyReturn(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	addr3 := std.Address("addr3")
	addr4 := std.Address("addr4")

	set := New(addr1, addr2, addr3, addr4)

	// Test forward iteration with early return
	count := 0
	set.IterateByOffset(0, 10, func(addr std.Address) bool {
		count++
		return count == 2 // Stop after second item
	})
	uassert.Equal(t, 2, count)

	// Test reverse iteration with early return
	count = 0
	set.ReverseIterateByOffset(0, 10, func(addr std.Address) bool {
		count++
		return count == 2 // Stop after second item
	})
	uassert.Equal(t, 2, count)
}

func TestEmptySetIteration(t *testing.T) {
	var set Set

	// Test forward iteration on empty set
	count := 0
	set.IterateByOffset(0, 10, func(addr std.Address) bool {
		count++
		return false
	})
	uassert.Equal(t, 0, count)

	// Test reverse iteration on empty set
	count = 0
	set.ReverseIterateByOffset(0, 10, func(addr std.Address) bool {
		count++
		return false
	})
	uassert.Equal(t, 0, count)

	// Test ToSlice on empty set
	slice := set.ToSlice(0, 10)
	uassert.Equal(t, 0, len(slice))
}

func TestReadonlySet(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	addr3 := std.Address("addr3")

	tests := []struct {
		name  string
		setup func(s *Set)
		check func(t *testing.T, rs ReadonlySet)
	}{
		{
			name:  "empty readonly set",
			setup: func(s *Set) {},
			check: func(t *testing.T, rs ReadonlySet) {
				uassert.Equal(t, 0, rs.Size())
				uassert.False(t, rs.Has(addr1))
			},
		},
		{
			name: "readonly set with addresses",
			setup: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
			},
			check: func(t *testing.T, rs ReadonlySet) {
				uassert.Equal(t, 2, rs.Size())
				uassert.True(t, rs.Has(addr1))
				uassert.True(t, rs.Has(addr2))
				uassert.False(t, rs.Has(addr3))
			},
		},
		{
			name: "readonly set iteration",
			setup: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Add(addr3)
			},
			check: func(t *testing.T, rs ReadonlySet) {
				// Test forward iteration
				collected := []std.Address{}
				rs.IterateByOffset(0, 10, func(addr std.Address) bool {
					collected = append(collected, addr)
					return false
				})
				uassert.Equal(t, 3, len(collected))
				uassert.Equal(t, addr1, collected[0])
				uassert.Equal(t, addr2, collected[1])
				uassert.Equal(t, addr3, collected[2])

				// Test reverse iteration
				collected = []std.Address{}
				rs.ReverseIterateByOffset(0, 10, func(addr std.Address) bool {
					collected = append(collected, addr)
					return false
				})
				uassert.Equal(t, 3, len(collected))
				uassert.Equal(t, addr3, collected[0])
				uassert.Equal(t, addr2, collected[1])
				uassert.Equal(t, addr1, collected[2])
			},
		},
		{
			name: "readonly set toslice",
			setup: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
				s.Add(addr3)
			},
			check: func(t *testing.T, rs ReadonlySet) {
				// Test full slice
				slice := rs.ToSlice(0, 10)
				uassert.Equal(t, 3, len(slice))
				uassert.Equal(t, addr1, slice[0])
				uassert.Equal(t, addr2, slice[1])
				uassert.Equal(t, addr3, slice[2])

				// Test partial slice
				slice = rs.ToSlice(1, 2)
				uassert.Equal(t, 2, len(slice))
				uassert.Equal(t, addr2, slice[0])
				uassert.Equal(t, addr3, slice[1])
			},
		},
		{
			name: "readonly set tree",
			setup: func(s *Set) {
				s.Add(addr1)
				s.Add(addr2)
			},
			check: func(t *testing.T, rs ReadonlySet) {
				tree := rs.Tree()
				uassert.Equal(t, 2, tree.Size())
				uassert.True(t, tree.Has(string(addr1)))
				uassert.True(t, tree.Has(string(addr2)))
				uassert.False(t, tree.Has(string(addr3)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var set Set
			tt.setup(&set)
			readonly := set.Readonly()
			tt.check(t, readonly)
		})
	}
}

func TestReadonlySetReflectsChanges(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")

	var set Set
	readonly := set.Readonly()

	// Initial state
	uassert.Equal(t, 0, readonly.Size())
	uassert.False(t, readonly.Has(addr1))

	// Modify original set
	set.Add(addr1)
	uassert.Equal(t, 1, readonly.Size())
	uassert.True(t, readonly.Has(addr1))

	// Add another address
	set.Add(addr2)
	uassert.Equal(t, 2, readonly.Size())
	uassert.True(t, readonly.Has(addr2))

	// Remove an address
	set.Remove(addr1)
	uassert.Equal(t, 1, readonly.Size())
	uassert.False(t, readonly.Has(addr1))
	uassert.True(t, readonly.Has(addr2))
}

func TestReadWriteSetInterface(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	var _ ReadWriteSet = &Set{}
	var s Set
	uassert.True(t, s.Add(addr1))
	uassert.True(t, s.Has(addr1))
	uassert.True(t, s.Remove(addr1))
	uassert.False(t, s.Has(addr1))
	uassert.False(t, s.Remove(addr2))
}

func TestReadonlySetTreeWrap(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	set := New(addr1, addr2)
	readonly := set.Readonly()
	tree := readonly.Tree()
	uassert.True(t, tree.Has(string(addr1)))
	uassert.True(t, tree.Has(string(addr2)))
}

func TestIterationNegativeOffsetAndLimit(t *testing.T) {
	addr1 := std.Address("addr1")
	addr2 := std.Address("addr2")
	set := New(addr1, addr2)
	count := 0
	set.IterateByOffset(-1, 10, func(addr std.Address) bool {
		count++
		return false
	})
	uassert.Equal(t, 2, count)
	count = 0
	set.IterateByOffset(0, -1, func(addr std.Address) bool {
		count++
		return false
	})
	uassert.Equal(t, 0, count)
}

func TestInterfaceAssertions(t *testing.T) {
	var _ ReadWriteSet = &Set{}
	var _ ReadonlySet = &Set{}
	var _ ReadonlySet = &readonlySet{}
}
