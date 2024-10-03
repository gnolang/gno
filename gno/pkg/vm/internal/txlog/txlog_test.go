package txlog

import (
	"fmt"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleWrap() {
	type User struct {
		ID   int
		Name string
	}
	m := GoMap[int, User](map[int]User{
		1: {ID: 1, Name: "alice"},
		2: {ID: 2, Name: "bob"},
	})

	// Wrap m in a transaction log.
	txl := Wrap(m)
	txl.Set(2, User{ID: 2, Name: "carl"})

	// m will still have bob, while txl will have carl.
	fmt.Println("m.Get(2):")
	fmt.Println(m.Get(2))
	fmt.Println("txl.Get(2):")
	fmt.Println(txl.Get(2))

	// after txl.Commit(), the transaction log will be committed to m.
	txl.Commit()
	fmt.Println("--- commit ---")
	fmt.Println("m.Get(2):")
	fmt.Println(m.Get(2))
	fmt.Println("txl.Get(2):")
	fmt.Println(txl.Get(2))

	// Output:
	// m.Get(2):
	// {2 bob} true
	// txl.Get(2):
	// {2 carl} true
	// --- commit ---
	// m.Get(2):
	// {2 carl} true
	// txl.Get(2):
	// {2 carl} true
}

func Test_txLog(t *testing.T) {
	t.Parallel()

	type Value = struct{}

	// Full "integration test" of the txLog + mapwrapper.
	source := GoMap[int, *Value](map[int]*Value{})

	// create 4 empty values (we'll just use the pointers)
	vs := [...]*Value{
		{},
		{},
		{},
		{},
	}
	source.Set(0, vs[0])
	source.Set(1, vs[1])
	source.Set(2, vs[2])

	{
		// Attempt getting, and deleting an item.
		v, ok := source.Get(0)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[0] == v, "pointer returned should be ==")

		source.Delete(0)
		v, ok = source.Get(0)
		assert.False(t, ok, "should be unsuccessful Get")
		assert.Nil(t, v, "pointer returned should be nil")

		verifyHashMapValues(t, source, map[int]*Value{1: vs[1], 2: vs[2]})
	}

	saved := maps.Clone(source)
	txm := Wrap(source).(*txLog[int, *Value])

	{
		// Attempt getting, deleting an item on a buffered map;
		// then creating a new one.
		v, ok := txm.Get(1)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[1] == v, "pointer returned should be ==")

		txm.Delete(1)
		v, ok = txm.Get(1)
		assert.False(t, ok, "should be unsuccessful Get")
		assert.Nil(t, v, "pointer returned should be nil")

		// Update an existing value to another value.
		txm.Set(2, vs[0])
		v, ok = txm.Get(2)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[0] == v, "pointer returned should be ==")

		// Add a new value
		txm.Set(3, vs[3])
		v, ok = txm.Get(3)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[3] == v, "pointer returned should be ==")

		// The original bufferedTxMap should still not know about the
		// new value, and the internal "source" map should still be the
		// same.
		v, ok = source.Get(3)
		assert.Nil(t, v)
		assert.False(t, ok)
		v, ok = source.Get(1)
		assert.True(t, vs[1] == v)
		assert.True(t, ok)
		assert.Equal(t, saved, source)
		assert.Equal(t, saved, txm.source)

		// double-check on the iterators.
		verifyHashMapValues(t, source, map[int]*Value{1: vs[1], 2: vs[0]})
		verifyHashMapValues(t, txm, map[int]*Value{2: vs[2], 3: vs[3]})
	}

	{
		// Using Commit() should cause txm's internal buffer to be cleared;
		// and for all changes to show up on the source map.
		txm.Commit()
		assert.Empty(t, txm.dirty)
		assert.Equal(t, source, txm.source)
		assert.NotEqual(t, saved, source)

		v, ok := source.Get(3)
		assert.True(t, vs[3] == v)
		assert.True(t, ok)
		v, ok = source.Get(1)
		assert.Nil(t, v)
		assert.False(t, ok)

		// double-check on the iterators.
		verifyHashMapValues(t, source, map[int]*Value{2: vs[0], 3: vs[3]})
		verifyHashMapValues(t, txm, map[int]*Value{2: vs[0], 3: vs[3]})
	}
}

func verifyHashMapValues(t *testing.T, m Map[int, *struct{}], expectedReadonly map[int]*struct{}) {
	t.Helper()

	expected := maps.Clone(expectedReadonly)
	m.Iterate()(func(k int, v *struct{}) bool {
		ev, eok := expected[k]
		_ = assert.True(t, eok, "mapping %d:%v should exist in expected map", k, v) &&
			assert.Equal(t, ev, v, "values should match")
		delete(expected, k)
		return true
	})
	assert.Empty(t, expected, "(some) expected values not found in the Map")
}

func Test_bufferedTxMap(t *testing.T) {
	t.Parallel()

	type Value struct{}

	// Full "integration test" of the bufferedTxMap.
	var m bufferedTxMap[int, *Value]
	m.init()

	vs := [...]*Value{
		{},
		{},
		{},
		{},
	}
	m.Set(0, vs[0])
	m.Set(1, vs[1])
	m.Set(2, vs[2])

	{
		// Attempt getting, and deleting an item.
		v, ok := m.Get(0)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[0] == v, "pointer returned should be ==")

		m.Delete(0)
		v, ok = m.Get(0)
		assert.False(t, ok, "should be unsuccessful Get")
		assert.Nil(t, v, "pointer returned should be nil")
	}

	saved := maps.Clone(m.source)
	bm := m.buffered()

	{
		// Attempt getting, deleting an item on a buffered map;
		// then creating a new one.
		v, ok := bm.Get(1)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[1] == v, "pointer returned should be ==")

		bm.Delete(1)
		v, ok = bm.Get(1)
		assert.False(t, ok, "should be unsuccessful Get")
		assert.Nil(t, v, "pointer returned should be nil")

		bm.Set(3, vs[3])
		v, ok = bm.Get(3)
		assert.True(t, ok, "should be successful Get")
		assert.True(t, vs[3] == v, "pointer returned should be ==")

		// The original bufferedTxMap should still not know about the
		// new value, and the internal "source" map should still be the
		// same.
		v, ok = m.Get(3)
		assert.Nil(t, v)
		assert.False(t, ok)
		v, ok = m.Get(1)
		assert.True(t, vs[1] == v)
		assert.True(t, ok)
		assert.Equal(t, saved, m.source)
		assert.Equal(t, saved, bm.source)
	}

	{
		// Using write() should cause bm's internal buffer to be cleared;
		// and for all changes to show up on the source map.
		bm.write()
		assert.Empty(t, bm.dirty)
		assert.Equal(t, m.source, bm.source)
		assert.NotEqual(t, saved, m.source)

		v, ok := m.Get(3)
		assert.True(t, vs[3] == v)
		assert.True(t, ok)
		v, ok = m.Get(1)
		assert.Nil(t, v)
		assert.False(t, ok)
	}
}

func Test_bufferedTxMap_initErr(t *testing.T) {
	t.Parallel()

	var b bufferedTxMap[bool, bool]
	b.init()

	assert.PanicsWithValue(t, "cannot init with a dirty buffer", func() {
		buf := b.buffered()
		buf.init()
	})
}

func Test_bufferedTxMap_bufferedErr(t *testing.T) {
	t.Parallel()

	var b bufferedTxMap[bool, bool]
	b.init()
	buf := b.buffered()

	assert.PanicsWithValue(t, "cannot stack multiple bufferedTxMap", func() {
		buf.buffered()
	})
}

// bufferedTxMap is a wrapper around the map type, supporting regular Get, Set
// and Delete operations. Additionally, it can create a "buffered" version of
// itself, which will keep track of all write (set and delete) operations to the
// map; so that they can all be atomically committed when calling "write".
type bufferedTxMap[K comparable, V any] struct {
	source map[K]V
	dirty  map[K]deletable[V]
}

// init should be called when creating the bufferedTxMap, in a non-buffered
// context.
func (b *bufferedTxMap[K, V]) init() {
	if b.dirty != nil {
		panic("cannot init with a dirty buffer")
	}
	b.source = make(map[K]V)
}

// buffered creates a copy of b, which has a usable dirty map.
func (b bufferedTxMap[K, V]) buffered() bufferedTxMap[K, V] {
	if b.dirty != nil {
		panic("cannot stack multiple bufferedTxMap")
	}
	return bufferedTxMap[K, V]{
		source: b.source,
		dirty:  make(map[K]deletable[V]),
	}
}

// write commits the data in dirty to the map in source.
func (b *bufferedTxMap[K, V]) write() {
	for k, v := range b.dirty {
		if v.deleted {
			delete(b.source, k)
		} else {
			b.source[k] = v.v
		}
	}
	b.dirty = make(map[K]deletable[V])
}

func (b bufferedTxMap[K, V]) Get(k K) (V, bool) {
	if b.dirty != nil {
		if bufValue, ok := b.dirty[k]; ok {
			if bufValue.deleted {
				var zeroV V
				return zeroV, false
			}
			return bufValue.v, true
		}
	}
	v, ok := b.source[k]
	return v, ok
}

func (b bufferedTxMap[K, V]) Set(k K, v V) {
	if b.dirty == nil {
		b.source[k] = v
		return
	}
	b.dirty[k] = deletable[V]{v: v}
}

func (b bufferedTxMap[K, V]) Delete(k K) {
	if b.dirty == nil {
		delete(b.source, k)
		return
	}
	b.dirty[k] = deletable[V]{deleted: true}
}

func Benchmark_txLogRead(b *testing.B) {
	const maxValues = (1 << 10) * 9 // must be multiple of 9

	var (
		baseMap = make(map[int]int)        // all values filled
		wrapped = GoMap[int, int](baseMap) // wrapper around baseMap
		stack1  = Wrap(wrapped)            // n+1, n+4, n+7 values filled (n%9 == 0)
		stack2  = Wrap(stack1)             // n'th values filled (n%9 == 0)
	)

	for i := 0; i < maxValues; i++ {
		baseMap[i] = i
		switch i % 9 {
		case 1, 4, 7:
			stack1.Set(i, i+1_000_000)
		case 0:
			stack2.Set(i, i+10_000_000)
		}
	}

	var v int
	var ok bool
	_, _ = v, ok

	// through closure, so func calls have to go through "indirection".
	runbench := func(b *testing.B, src Map[int, int]) { //nolint:thelper
		for i := 0; i < b.N; i++ {
			v, ok = src.Get(i % maxValues)
		}
	}

	b.Run("stack2", func(b *testing.B) { runbench(b, stack2) })
	b.Run("stack1", func(b *testing.B) { runbench(b, stack1) })
	b.Run("wrapped", func(b *testing.B) { runbench(b, wrapped) })
	b.Run("baseline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v, ok = baseMap[i%maxValues]
		}
	})
}

func Benchmark_txLogWrite(b *testing.B) {
	// after this amount of values, the maps are re-initialized.
	// you can tweak this to see how the benchmarks behave on a variety of
	// values.
	// NOTE: setting this too high will skew the benchmark in favour those which
	// have a smaller N, as those with a higher N have to allocate more in a
	// single map.
	const maxValues = 1 << 15 // 32768

	var v int
	var ok bool
	_, _ = v, ok

	b.Run("stack1", func(b *testing.B) {
		src := GoMap[int, int](make(map[int]int))
		st := Wrap(src)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			st.Set(k, i)
			// we use this assignment to prevent the compiler from optimizing
			// out code, especially in the baseline case.
			v, ok = st.Get(k)

			if k == maxValues-1 {
				st = Wrap(src)
			}
		}
	})
	b.Run("wrapped", func(b *testing.B) {
		src := GoMap[int, int](make(map[int]int))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			src.Set(k, i)
			// we use this assignment to prevent the compiler from optimizing
			// out code, especially in the baseline case.
			v, ok = src.Get(k)

			if k == maxValues-1 {
				src = GoMap[int, int](make(map[int]int))
			}
		}
	})
	b.Run("baseline", func(b *testing.B) {
		// this serves to have a baseline value in the benchmark results
		// for when we just use a map directly.
		m := make(map[int]int)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			m[k] = i
			v, ok = m[k]

			if k == maxValues-1 {
				m = make(map[int]int)
			}
		}
	})
}

func Benchmark_bufferedTxMapRead(b *testing.B) {
	const maxValues = (1 << 10) * 9 // must be multiple of 9

	var (
		baseMap = make(map[int]int) // all values filled
		wrapped = bufferedTxMap[int, int]{source: baseMap}
		stack1  = wrapped.buffered() // n, n+1, n+4, n+7 values filled (n%9 == 0)
		// this test doesn't have stack2 as bufferedTxMap
		// does not support stacking
	)

	for i := 0; i < maxValues; i++ {
		baseMap[i] = i
		switch i % 9 {
		case 0, 1, 4, 7:
			stack1.Set(i, i+1_000_000)
		}
	}

	var v int
	var ok bool
	_, _ = v, ok

	b.Run("stack1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// use assignment to avoid the compiler optimizing out the loops
			v, ok = stack1.Get(i % maxValues)
		}
	})
	b.Run("wrapped", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v, ok = wrapped.Get(i % maxValues)
		}
	})
	b.Run("baseline", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			v, ok = baseMap[i%maxValues]
		}
	})
}

func Benchmark_bufferedTxMapWrite(b *testing.B) {
	// after this amount of values, the maps are re-initialized.
	// you can tweak this to see how the benchmarks behave on a variety of
	// values.
	// NOTE: setting this too high will skew the benchmark in favour those which
	// have a smaller N, as those with a higher N have to allocate more in a
	// single map.
	const maxValues = 1 << 15 // 32768

	var v int
	var ok bool
	_, _ = v, ok

	b.Run("buffered", func(b *testing.B) {
		var orig bufferedTxMap[int, int]
		orig.init()
		txm := orig.buffered()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			txm.Set(k, i)
			// we use this assignment to prevent the compiler from optimizing
			// out code, especially in the baseline case.
			v, ok = txm.Get(k)

			if k == maxValues-1 {
				txm = orig.buffered()
			}
		}
	})
	b.Run("unbuffered", func(b *testing.B) {
		var txm bufferedTxMap[int, int]
		txm.init()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			txm.Set(k, i)
			v, ok = txm.Get(k)

			if k == maxValues-1 {
				txm.init()
			}
		}
	})
	b.Run("baseline", func(b *testing.B) {
		// this serves to have a baseline value in the benchmark results
		// for when we just use a map directly.
		m := make(map[int]int)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			k := i % maxValues

			m[k] = i
			v, ok = m[k]

			if k == maxValues-1 {
				m = make(map[int]int)
			}
		}
	})
}
