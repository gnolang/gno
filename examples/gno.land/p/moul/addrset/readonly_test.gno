package addrset

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestReadonlySet(t *testing.T) {
	addr1 := address("addr1")
	addr2 := address("addr2")
	addr3 := address("addr3")

	var set Set
	set.Add(addr1)
	set.Add(addr2)

	ro := set.Readonly()
	uassert.Equal(t, 2, ro.Size())
	uassert.True(t, ro.Has(addr1))
	uassert.True(t, ro.Has(addr2))
	uassert.False(t, ro.Has(addr3))

	// NewReadonlySet is equivalent to Set.Readonly.
	ro2 := NewReadonlySet(&set)
	uassert.Equal(t, 2, ro2.Size())
	uassert.True(t, ro2.Has(addr1))

	// The view is a live handle, not a snapshot.
	set.Add(addr3)
	uassert.Equal(t, 3, ro.Size())
	uassert.True(t, ro.Has(addr3))
	set.Remove(addr1)
	uassert.False(t, ro.Has(addr1))
}

func TestReadonlySetIterate(t *testing.T) {
	var set Set
	set.Add(address("addr2"))
	set.Add(address("addr1"))
	set.Add(address("addr3"))
	ro := set.Readonly()

	var got []address
	stopped := ro.Iterate(0, 10, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.False(t, stopped)
	uassert.Equal(t, 3, len(got))
	uassert.Equal(t, address("addr1"), got[0]) // sorted order
	uassert.Equal(t, address("addr2"), got[1])
	uassert.Equal(t, address("addr3"), got[2])

	// offset/count window.
	got = nil
	ro.Iterate(1, 1, func(a address) bool {
		got = append(got, a)
		return false
	})
	uassert.Equal(t, 1, len(got))
	uassert.Equal(t, address("addr2"), got[0])

	// Early stop is reported.
	count := 0
	stopped = ro.Iterate(0, 10, func(a address) bool {
		count++
		return true
	})
	uassert.True(t, stopped)
	uassert.Equal(t, 1, count)

	// Empty set: callback never runs, not stopped.
	var empty Set
	stopped = empty.Readonly().Iterate(0, 10, func(a address) bool { return true })
	uassert.False(t, stopped)
}
