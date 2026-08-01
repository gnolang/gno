package addrset

import (
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
}
