// Package txlog is an internal package containing data structures that can
// function as "transaction logs" on top of a hash map (or other key/value
// data type implementing [Map]).
//
// A transaction log keeps track of the write operations performed in a
// transaction
package txlog

// Map is a generic interface to a key/value map, like Go's builtin map.
type Map[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
	Delete(K)
	Iterate() func(yield func(K, V) bool)
}

// MapCommitter is a Map which also implements a Commit() method, which writes
// to the underlying (parent) [Map].
type MapCommitter[K comparable, V any] interface {
	Map[K, V]

	// Commit writes the logged operations to the underlying map.
	// After calling commit, the underlying tx log is cleared and the
	// MapCommitter may be reused.
	Commit()
}

// GoMap is a simple implementation of [Map], which wraps the operations of
// Go's map builtin to implement [Map].
type GoMap[K comparable, V any] map[K]V

// Get implements [Map].
func (m GoMap[K, V]) Get(k K) (V, bool) {
	v, ok := m[k]
	return v, ok
}

// Set implements [Map].
func (m GoMap[K, V]) Set(k K, v V) {
	m[k] = v
}

// Delete implements [Map].
func (m GoMap[K, V]) Delete(k K) {
	delete(m, k)
}

// Iterate implements [Map].
func (m GoMap[K, V]) Iterate() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Wrap wraps the map m into a data structure to keep a transaction log.
// To write data to m, use MapCommitter.Commit.
func Wrap[K comparable, V any](m Map[K, V]) MapCommitter[K, V] {
	return &txLog[K, V]{source: m, dirty: make(map[K]deletable[V])}
}

type txLog[K comparable, V any] struct {
	source Map[K, V]
	dirty  map[K]deletable[V]
}

func (b *txLog[K, V]) Commit() {
	for k, v := range b.dirty {
		if v.deleted {
			b.source.Delete(k)
		} else {
			b.source.Set(k, v.v)
		}
	}
	b.dirty = make(map[K]deletable[V])
}

func (b txLog[K, V]) Get(k K) (V, bool) {
	if bufValue, ok := b.dirty[k]; ok {
		if bufValue.deleted {
			var zeroV V
			return zeroV, false
		}
		return bufValue.v, true
	}

	return b.source.Get(k)
}

func (b txLog[K, V]) Set(k K, v V) {
	b.dirty[k] = deletable[V]{v: v}
}

func (b txLog[K, V]) Delete(k K) {
	b.dirty[k] = deletable[V]{deleted: true}
}

func (b txLog[K, V]) Iterate() func(yield func(K, V) bool) {
	return func(yield func(K, V) bool) {
		b.source.Iterate()(func(k K, v V) bool {
			if dirty, ok := b.dirty[k]; ok {
				if dirty.deleted {
					return true
				}
				return yield(k, dirty.v)
			}

			// not in dirty
			return yield(k, v)
		})
		// yield for new values
		for k, v := range b.dirty {
			if v.deleted {
				continue
			}
			_, ok := b.source.Get(k)
			if ok {
				continue
			}
			if !yield(k, v.v) {
				break
			}
		}
	}
}

type deletable[V any] struct {
	v       V
	deleted bool
}
