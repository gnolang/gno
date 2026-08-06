package addrset

import (
	"strconv"
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

const (
	addr1 = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	addr2 = address("g1manfred47kzduec920z88wfr64ylksmdcedlf5")
	addr3 = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq")
)

func TestSetBasics(t *testing.T) {
	var set Set // the zero value is a usable empty set

	uassert.Equal(t, 0, set.Size())
	uassert.False(t, set.Has(addr1))
	uassert.False(t, set.Remove(addr1))

	uassert.True(t, set.Add(addr1), "expect newly added")
	uassert.False(t, set.Add(addr1), "expect duplicate add to report existing")
	uassert.True(t, set.Has(addr1))
	uassert.Equal(t, 1, set.Size())

	uassert.True(t, set.Add(addr2))
	uassert.True(t, set.Add(addr3))
	uassert.Equal(t, 3, set.Size())

	uassert.True(t, set.Remove(addr2), "expect removal of present address")
	uassert.False(t, set.Remove(addr2), "expect second removal to report absent")
	uassert.False(t, set.Has(addr2))
	uassert.Equal(t, 2, set.Size())
}

func TestSetIteration(t *testing.T) {
	var set Set
	set.Add(addr1)
	set.Add(addr2)
	set.Add(addr3)

	// Sorted (bech32-string) order: addr3 < addr1 < addr2.
	var got []address
	set.IterateByOffset(0, set.Size(), func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 3, len(got))
	uassert.True(t, got[0] == addr3 && got[1] == addr1 && got[2] == addr2, "expect sorted order")

	// Offset and count window.
	got = nil
	set.IterateByOffset(1, 1, func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 1, len(got))
	uassert.True(t, got[0] == addr1, "expect the middle element")

	// Early stop.
	count := 0
	set.IterateByOffset(0, set.Size(), func(a address) bool {
		count++
		return true
	})
	uassert.Equal(t, 1, count)

	// Reverse order.
	got = nil
	set.ReverseIterateByOffset(0, set.Size(), func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 3, len(got))
	uassert.True(t, got[0] == addr2 && got[2] == addr3, "expect reverse order")
}

func TestSetIterationLimits(t *testing.T) {
	var set Set
	set.Add(addr1)
	set.Add(addr2)
	set.Add(addr3)

	// Offset at/beyond the size visits nothing, in both directions.
	visits := 0
	count := func(a address) bool { visits++; return false }
	set.IterateByOffset(3, 10, count)
	set.IterateByOffset(99, 10, count)
	set.ReverseIterateByOffset(3, 10, count)
	uassert.Equal(t, 0, visits)

	// A count larger than the remainder visits just the remainder.
	set.IterateByOffset(1, 10, count)
	uassert.Equal(t, 2, visits)

	// The reverse offset counts from the END: offset 1, count 1 is the
	// second-to-last address in sorted order (addr1: addr3 < addr1 < addr2).
	var got []address
	set.ReverseIterateByOffset(1, 1, func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 1, len(got))
	uassert.True(t, got[0] == addr1, "expect reverse offset anchored at the end")
}

// TestSetAcrossLeafSplits exercises the wrapper over a multi-leaf tree
// (the backing splits past fanout 32): ordering, offset windows in both
// directions, and removal back down through merges.
func TestSetAcrossLeafSplits(t *testing.T) {
	const n = 100

	var (
		set   Set
		addrs []address
	)
	for i := 0; i < n; i++ {
		// Zero-padded so generation order == lexicographic order.
		s := strconv.Itoa(i)
		addrs = append(addrs, address("g1"+strings.Repeat("0", 38-len(s))+s))
	}
	for _, a := range addrs {
		urequire.True(t, set.Add(a), "expect newly added")
	}
	urequire.Equal(t, n, set.Size())

	// Full forward iteration is complete and sorted (keys were generated
	// in sorted order).
	i := 0
	set.IterateByOffset(0, n, func(a address) bool {
		urequire.True(t, a == addrs[i], "expect sorted order at "+strconv.Itoa(i))
		i++
		return false
	})
	urequire.Equal(t, n, i)

	// An offset window that crosses leaf boundaries.
	var got []address
	set.IterateByOffset(30, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 10, len(got))
	uassert.True(t, got[0] == addrs[30] && got[9] == addrs[39], "expect window [30,40)")

	// Reverse window anchored from the end.
	got = nil
	set.ReverseIterateByOffset(10, 5, func(a address) bool {
		got = append(got, a)
		return false
	})
	urequire.Equal(t, 5, len(got))
	uassert.True(t, got[0] == addrs[n-11] && got[4] == addrs[n-15], "expect reverse window from the end")

	// Remove every other address (drives leaf merges), then verify.
	for i := 0; i < n; i += 2 {
		urequire.True(t, set.Remove(addrs[i]), "expect removal")
	}
	urequire.Equal(t, n/2, set.Size())
	for i := 0; i < n; i++ {
		uassert.Equal(t, i%2 == 1, set.Has(addrs[i]))
	}
}

func TestReadonlySet(t *testing.T) {
	var set Set
	set.Add(addr1)

	view := set.Readonly()
	uassert.True(t, view.Has(addr1))
	uassert.Equal(t, 1, view.Size())

	// The view is a live handle, not a snapshot.
	set.Add(addr2)
	uassert.Equal(t, 2, view.Size())
	uassert.True(t, view.Has(addr2))

	// Iteration mirrors the set and reports early stop.
	stopped := view.IterateByOffset(0, view.Size(), func(a address) bool { return false })
	uassert.False(t, stopped)
	stopped = view.IterateByOffset(0, view.Size(), func(a address) bool { return true })
	uassert.True(t, stopped)

	stopped = view.ReverseIterateByOffset(0, view.Size(), func(a address) bool { return true })
	uassert.True(t, stopped)

	// An empty set's view iterates nothing and reports not-stopped.
	empty := NewReadonlySet(&Set{})
	uassert.Equal(t, 0, empty.Size())
	stopped = empty.IterateByOffset(0, 10, func(a address) bool { return true })
	uassert.False(t, stopped)
}
