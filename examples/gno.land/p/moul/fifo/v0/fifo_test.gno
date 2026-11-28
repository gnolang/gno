package fifo

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
)

func TestNew(t *testing.T) {
	l := New(5)
	uassert.Equal(t, 5, l.MaxSize())
	uassert.Equal(t, 0, l.Size())
}

func TestAppend(t *testing.T) {
	l := New(3)

	// Test adding within capacity
	l.Append(1)
	l.Append(2)
	uassert.Equal(t, 2, l.Size())
	uassert.Equal(t, 1, l.Get(0))
	uassert.Equal(t, 2, l.Get(1))

	// Test overflow behavior
	l.Append(3)
	l.Append(4)
	uassert.Equal(t, 3, l.Size())
	uassert.Equal(t, 2, l.Get(0))
	uassert.Equal(t, 3, l.Get(1))
	uassert.Equal(t, 4, l.Get(2))
}

func TestPrepend(t *testing.T) {
	l := New(3)

	// Test adding within capacity
	l.Prepend(1)
	l.Prepend(2)
	uassert.Equal(t, 2, l.Size())
	uassert.Equal(t, 2, l.Get(0))
	uassert.Equal(t, 1, l.Get(1))

	// Test overflow behavior
	l.Prepend(3)
	l.Prepend(4)
	uassert.Equal(t, 3, l.Size())
	uassert.Equal(t, 4, l.Get(0))
	uassert.Equal(t, 3, l.Get(1))
	uassert.Equal(t, 2, l.Get(2))
}

func TestGet(t *testing.T) {
	l := New(3)
	l.Append(1)
	l.Append(2)
	l.Append(3)

	// Test valid indices
	uassert.Equal(t, 1, l.Get(0))
	uassert.Equal(t, 2, l.Get(1))
	uassert.Equal(t, 3, l.Get(2))

	// Test invalid indices
	uassert.True(t, l.Get(-1) == nil)
	uassert.True(t, l.Get(3) == nil)
}

func TestEntries(t *testing.T) {
	l := New(3)
	l.Append(1)
	l.Append(2)
	l.Append(3)

	entries := l.Entries()
	uassert.Equal(t, 3, len(entries))
	uassert.Equal(t, 1, entries[0])
	uassert.Equal(t, 2, entries[1])
	uassert.Equal(t, 3, entries[2])
}

func TestLatest(t *testing.T) {
	l := New(5)

	// Test empty list
	uassert.True(t, l.Latest() == nil)

	// Test single entry
	l.Append(1)
	uassert.Equal(t, 1, l.Latest())

	// Test multiple entries
	l.Append(2)
	l.Append(3)
	uassert.Equal(t, 3, l.Latest())

	// Test after overflow
	l.Append(4)
	l.Append(5)
	l.Append(6)
	uassert.Equal(t, 6, l.Latest())
}

func TestIterator(t *testing.T) {
	l := New(3)
	l.Append(1)
	l.Append(2)
	l.Append(3)

	iter := l.Iterator()
	uassert.Equal(t, 1, iter())
	uassert.Equal(t, 2, iter())
	uassert.Equal(t, 3, iter())
	uassert.True(t, iter() == nil)
}

func TestMixedOperations(t *testing.T) {
	l := New(3)

	// Mix of append and prepend operations
	l.Append(1)  // [1]
	l.Prepend(2) // [2,1]
	l.Append(3)  // [2,1,3]
	l.Prepend(4) // [4,2,1]

	entries := l.Entries()
	uassert.Equal(t, 3, len(entries))
	uassert.Equal(t, 4, entries[0])
	uassert.Equal(t, 2, entries[1])
	uassert.Equal(t, 1, entries[2])
}

func TestEmptyList(t *testing.T) {
	l := New(3)

	// Test operations on empty list
	uassert.Equal(t, 0, l.Size())
	uassert.True(t, l.Get(0) == nil)
	uassert.Equal(t, 0, len(l.Entries()))
	uassert.True(t, l.Latest() == nil)

	iter := l.Iterator()
	uassert.True(t, iter() == nil)
}

func TestEdgeCases(t *testing.T) {
	// Test zero-size list
	l := New(0)
	uassert.Equal(t, 0, l.MaxSize())
	l.Append(1) // Should be no-op
	uassert.Equal(t, 0, l.Size())

	// Test single-element list
	l = New(1)
	l.Append(1)
	l.Append(2) // Should replace 1
	uassert.Equal(t, 1, l.Size())
	uassert.Equal(t, 2, l.Latest())

	// Test rapid append/prepend alternation
	l = New(3)
	l.Append(1)  // [1]
	l.Prepend(2) // [2,1]
	l.Append(3)  // [2,1,3]
	l.Prepend(4) // [4,2,1]
	l.Append(5)  // [2,1,5]
	uassert.Equal(t, 3, l.Size())
	entries := l.Entries()
	uassert.Equal(t, 2, entries[0])
	uassert.Equal(t, 1, entries[1])
	uassert.Equal(t, 5, entries[2])

	// Test nil values
	l = New(2)
	l.Append(nil)
	l.Prepend(nil)
	uassert.Equal(t, 2, l.Size())
	uassert.True(t, l.Get(0) == nil)
	uassert.True(t, l.Get(1) == nil)

	// Test index bounds
	l = New(3)
	l.Append(1)
	uassert.True(t, l.Get(-1) == nil)
	uassert.True(t, l.Get(1) == nil)

	// Test iterator exhaustion
	l = New(2)
	l.Append(1)
	l.Append(2)
	iter := l.Iterator()
	uassert.Equal(t, 1, iter())
	uassert.Equal(t, 2, iter())
	uassert.True(t, iter() == nil)
	uassert.True(t, iter() == nil)

	// Test prepend on full list
	l = New(2)
	l.Append(1)
	l.Append(2)  // [1,2]
	l.Prepend(3) // [3,1]
	uassert.Equal(t, 2, l.Size())
	entries = l.Entries()
	uassert.Equal(t, 3, entries[0])
	uassert.Equal(t, 1, entries[1])
}

func TestSetMaxSize(t *testing.T) {
	l := New(5)

	// Fill the list
	l.Append(1)
	l.Append(2)
	l.Append(3)
	l.Append(4)
	l.Append(5)

	// Test increasing maxSize
	l.SetMaxSize(7)
	uassert.Equal(t, 7, l.MaxSize())
	uassert.Equal(t, 5, l.Size())

	// Test reducing maxSize
	l.SetMaxSize(3)
	uassert.Equal(t, 3, l.Size())
	entries := l.Entries()
	uassert.Equal(t, 3, entries[0])
	uassert.Equal(t, 4, entries[1])
	uassert.Equal(t, 5, entries[2])

	// Test setting to zero
	l.SetMaxSize(0)
	uassert.Equal(t, 0, l.Size())
	uassert.True(t, l.head == nil)
	uassert.True(t, l.tail == nil)

	// Test negative maxSize
	l.SetMaxSize(-1)
	uassert.Equal(t, 0, l.MaxSize())

	// Test setting back to positive
	l.SetMaxSize(2)
	l.Append(1)
	l.Append(2)
	l.Append(3)
	uassert.Equal(t, 2, l.Size())
	entries = l.Entries()
	uassert.Equal(t, 2, entries[0])
	uassert.Equal(t, 3, entries[1])
}

func TestDelete(t *testing.T) {
	l := New(5)

	// Test delete on empty list
	uassert.False(t, l.Delete(0))
	uassert.False(t, l.Delete(-1))

	// Fill list
	l.Append(1)
	l.Append(2)
	l.Append(3)
	l.Append(4)

	// Test invalid indices
	uassert.False(t, l.Delete(-1))
	uassert.False(t, l.Delete(4))

	// Test deleting from middle
	uassert.True(t, l.Delete(1))
	uassert.Equal(t, 3, l.Size())
	entries := l.Entries()
	uassert.Equal(t, 1, entries[0])
	uassert.Equal(t, 3, entries[1])
	uassert.Equal(t, 4, entries[2])

	// Test deleting from head
	uassert.True(t, l.Delete(0))
	uassert.Equal(t, 2, l.Size())
	entries = l.Entries()
	uassert.Equal(t, 3, entries[0])
	uassert.Equal(t, 4, entries[1])

	// Test deleting from tail
	uassert.True(t, l.Delete(1))
	uassert.Equal(t, 1, l.Size())
	uassert.Equal(t, 3, l.Latest())

	// Test deleting last element
	uassert.True(t, l.Delete(0))
	uassert.Equal(t, 0, l.Size())
	uassert.True(t, l.head == nil)
	uassert.True(t, l.tail == nil)
}
