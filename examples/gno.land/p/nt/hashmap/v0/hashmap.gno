// Package hashmap provides a persistent string-keyed map whose storage-access
// cost is independent of the number of entries.
//
// Motivation: in the GnoVM, every struct, array, and map value is persisted as
// a separate realm object, and loading an object from the store carries a
// large flat gas cost. An avl.Tree materializes its search path as one object
// per node, so a single lookup on an N-entry tree loads O(log N) objects; on
// state-heavy realms (token ledgers, registries) this makes transaction gas
// grow with state size (see gnolang/gno#5906).
//
// A hashmap.Map instead distributes entries over a fixed array of native Gno
// maps ("buckets"). A native map persists as a single object regardless of its
// entry count, so any lookup or update loads a constant number of objects:
// the Map struct, the bucket array, and one bucket. Cost per operation is
// therefore flat in N, at the price of key-ordered iteration, which this
// structure does not provide. Use avl.Tree when you need sorted keys or range
// queries; use hashmap.Map for pure key-value access such as balances or
// allowances.
//
// Bucket sizing: buckets is fixed at construction (a power of two). Every
// operation loads the whole bucket-pointer array once, then decodes the one or
// two buckets it touches. This is a balance, not "more is better": a larger
// bucket count shrinks each bucket but enlarges the array that every op must
// decode. Measured (see the ADR referenced above), the default 1024 stays
// nearly flat and is cheapest from a few thousand up to ~1,000,000 entries;
// 4096 was consistently *worse* at every size because the bigger array tax
// outweighed the smaller buckets. So:
//
//	expected entries | recommended buckets
//	         < 1,000 |   256
//	    1k – 1,000k  |  1024 (default — flat across this whole range)
//	       > ~1,000k |   4096 (only once per-bucket decode dominates the array)
//
// Iteration order is deterministic (bucket index, then per-bucket insertion
// order) but NOT sorted by key.
//
// Adversarial keys: the bucket hash (FNV-1a) uses a fixed seed, so keys can be
// ground to land in the same bucket. A party able to create many funded
// entries (e.g. via a faucet) can inflate one bucket, raising the per-op cost
// for keys that hash there — bounded by that bucket's size, not global. This
// is inherent to any fixed hash in a deterministic VM (no per-map secret is
// possible). Size the bucket count so expected entries-per-bucket stays small,
// which dilutes the effect; avl.Tree, whose cost is O(log N) regardless of key
// distribution, is the alternative when adversarial concentration is a concern.
package hashmap

// DefaultBuckets is the bucket count used by New. Good for maps up to roughly
// 100k entries; see the package documentation for sizing guidance.
const DefaultBuckets = 1024

// Map is a persistent string-keyed map with O(1) object loads per operation.
// The zero value is not usable; create one with New or NewWithBuckets.
type Map struct {
	buckets []map[string]any // len is a power of two; entries lazily allocated
	size    int
}

// New returns an empty Map with DefaultBuckets buckets.
func New() *Map {
	return NewWithBuckets(DefaultBuckets)
}

// NewWithBuckets returns an empty Map with the given number of buckets.
// The count is fixed for the lifetime of the map and must be a power of two.
func NewWithBuckets(count int) *Map {
	if count <= 0 || count&(count-1) != 0 {
		panic("hashmap: bucket count must be a positive power of two")
	}
	return &Map{buckets: make([]map[string]any, count)}
}

// fnv1a is the 64-bit FNV-1a hash. Pure Gno, deterministic across nodes.
func fnv1a(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// bucketIndex is the single source of truth for key placement; Get/Has/Remove
// (via bucketOf) and Set must all use it so reads and writes cannot diverge.
func (m *Map) bucketIndex(key string) int {
	return int(fnv1a(key) & uint64(len(m.buckets)-1))
}

func (m *Map) bucketOf(key string) map[string]any {
	return m.buckets[m.bucketIndex(key)]
}

// Get returns the value stored under key, or nil if the key is absent.
// (Matches avl.Tree/bptree.BPTree; use Has to distinguish an absent key from a
// key stored with a nil value.)
func (m *Map) Get(key string) any {
	b := m.bucketOf(key)
	if b == nil {
		return nil
	}
	return b[key]
}

// Has reports whether key exists in the map.
func (m *Map) Has(key string) bool {
	b := m.bucketOf(key)
	if b == nil {
		return false
	}
	_, exists := b[key]
	return exists
}

// Set stores value under key, and reports whether the key already existed.
func (m *Map) Set(key string, value any) (updated bool) {
	i := m.bucketIndex(key)
	b := m.buckets[i]
	if b == nil {
		b = make(map[string]any)
		m.buckets[i] = b
	}
	_, updated = b[key]
	b[key] = value
	if !updated {
		m.size++
	}
	return updated
}

// Remove deletes key from the map, returning the removed value and whether the
// key existed. Empty buckets are kept allocated to avoid churn on the bucket
// array object.
func (m *Map) Remove(key string) (value any, removed bool) {
	b := m.bucketOf(key)
	if b == nil {
		return nil, false
	}
	value, removed = b[key]
	if removed {
		delete(b, key)
		m.size--
	}
	return value, removed
}

// Size returns the number of entries in the map.
func (m *Map) Size() int {
	return m.size
}

// Iterate calls cb for every entry until cb returns true (stop), and reports
// whether iteration was stopped early. Order is deterministic — buckets in
// index order, entries in per-bucket insertion order — but NOT sorted by key.
func (m *Map) Iterate(cb func(key string, value any) bool) bool {
	for _, b := range m.buckets {
		for k, v := range b {
			if cb(k, v) {
				return true
			}
		}
	}
	return false
}
