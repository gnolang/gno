// Package hashmap provides a persistent string-keyed map whose storage-access
// cost is independent of the number of entries.
//
// Motivation: in the GnoVM, every struct, array, and map value is persisted as
// a separate realm object, and loading an object from the store carries a large
// flat gas cost plus a per-byte decode cost. An avl.Tree materializes its search
// path as one object per node, so a single lookup on an N-entry tree loads
// O(log N) objects; on state-heavy realms (token ledgers, registries) this makes
// transaction gas grow with state size (see gnolang/gno#5906).
//
// A hashmap.Map instead distributes entries over a fixed set of native Gno maps
// ("buckets"). A native map persists as a single object regardless of its entry
// count, so any lookup or update loads a constant number of objects — cost per
// operation is flat in N, at the price of key-ordered iteration (use avl.Tree
// when you need sorted keys or range queries; use hashmap.Map for pure key-value
// access such as balances or allowances).
//
// Two-level buckets: the buckets are not held in one flat array. A flat array of
// B bucket pointers is itself a single object that every operation must decode in
// full — at B=1024 that is ~190KB decoded on every Get/Set, fixed at any entry
// count. Instead the buckets live in a two-level directory: ≈√B pages of ≈√B
// buckets each. An operation decodes the directory and one page — each a ≈√B-slot
// array rather than a B-slot one — plus one leaf bucket, and pages are allocated lazily,
// so untouched pages cost nothing. This keeps the total bucket count high (small,
// cheap-to-decode leaves) without a large array to decode. Measured on a cold 1M
// entry grc20 transfer, this cut the ledger's decode+alloc cost by ~2M gas versus
// a flat array (see the ADR referenced above).
//
// Bucket sizing: the total bucket count is fixed at construction (a power of two)
// and split across the two levels — evenly when its exponent is even (4096 →
// 64×64), else the directory takes the extra bit (2048 → 64×32). Larger counts
// shrink each leaf bucket; because the array tax is now paid on the small
// per-level arrays, over-sizing is far less punishing than with a flat array.
// The default (4096) stays nearly flat from a few thousand up to ~1,000,000
// entries:
//
//	expected entries | recommended buckets
//	         < 1,000 |   256
//	    1k – 1,000k  |  4096 (default — flat across this whole range)
//	       > ~1,000k | 16384 (marginally cheaper once leaves dominate)
//
// Iteration order is deterministic (page, then bucket, then per-bucket insertion
// order) but NOT sorted by key, and is preserved across a store reload. Note the
// persisted image is insertion-order-dependent: two maps holding the same
// entries inserted in a different order serialize to different images (and
// iteration is not key-sorted, unlike avl.Tree). This is harmless for consensus
// (all nodes replay the same op order) but the state root is not a function of
// the key set alone — avoid layering Iterate-based consensus logic on it.
//
// Adversarial keys: bucket selection uses deterministic SHA-256, so keys can
// still be ground to land in the same bucket — though grinding SHA-256 preimages
// is far costlier than inverting a non-cryptographic hash. A party able to create
// many funded entries (e.g. via a faucet) can inflate one bucket, raising the
// per-op cost for keys that hash there — bounded by that bucket's size, not
// global. This is inherent to any public hash in a deterministic VM (no per-map
// secret is possible). Size the bucket count so expected entries-per-bucket stays
// small, which dilutes the effect; avl.Tree, whose cost is O(log N) regardless of
// key distribution, is the alternative when adversarial concentration is a
// concern.
package hashmap

import "crypto/sha256"

// DefaultBuckets is the total bucket count used by New. Good for maps up to
// roughly 1M entries; see the package documentation for sizing guidance.
const DefaultBuckets = 4096

// Map is a persistent string-keyed map with O(1) object loads per operation.
// Buckets are held in a two-level directory (dir of pages, each page an array of
// buckets); pages and buckets are allocated lazily on first write. The zero
// value is not usable; create one with New or NewWithBuckets.
type Map struct {
	dir      [][]map[string]any // dir[pageIdx][bucketIdx] -> bucket; pages lazily allocated
	dirMask  uint64
	pageMask uint64
	pageBits uint
	pageSize int
	size     int
}

// New returns an empty Map with DefaultBuckets buckets.
func New() *Map {
	return NewWithBuckets(DefaultBuckets)
}

// NewWithBuckets returns an empty Map with the given total number of buckets,
// which must be a power of two. The count is fixed for the lifetime of the map
// and split across the two directory levels: the low half of the hash bits
// selects the bucket within a page, the high half selects the page.
func NewWithBuckets(count int) *Map {
	if count <= 0 || count&(count-1) != 0 {
		panic("hashmap: bucket count must be a positive power of two")
	}
	// Split the count's index bits in half: the low pageBits select the bucket
	// within a page, the remaining high bits select the page.
	totalBits := 0
	for (1 << totalBits) < count {
		totalBits++
	}
	pageBits := totalBits / 2
	dirSize := 1 << (totalBits - pageBits)
	pageSize := 1 << pageBits
	return &Map{
		dir:      make([][]map[string]any, dirSize),
		dirMask:  uint64(dirSize - 1),
		pageMask: uint64(pageSize - 1),
		pageBits: uint(pageBits),
		pageSize: pageSize,
	}
}

// hash64 returns the first 64 bits of SHA-256 in little-endian order. Sum256 is
// a calibrated native binding, so it avoids the per-byte metered cost of an
// interpreted Gno hash loop on every Get, Set, Has, and Remove.
func hash64(s string) uint64 {
	h := sha256.Sum256([]byte(s))
	return uint64(h[0]) |
		uint64(h[1])<<8 |
		uint64(h[2])<<16 |
		uint64(h[3])<<24 |
		uint64(h[4])<<32 |
		uint64(h[5])<<40 |
		uint64(h[6])<<48 |
		uint64(h[7])<<56
}

// locate is the single source of truth for key placement; Get/Has/Remove (via
// bucketOf) and Set must all use it so reads and writes cannot diverge. It
// returns the page index and the in-page bucket index. The digest is fixed, so
// placement must be settled before any map is persisted — changing the hash or
// the split later relocates every key.
func (m *Map) locate(key string) (page, bucket int) {
	h := hash64(key)
	return int((h >> m.pageBits) & m.dirMask), int(h & m.pageMask)
}

func (m *Map) bucketOf(key string) map[string]any {
	di, pi := m.locate(key)
	page := m.dir[di]
	if page == nil {
		return nil
	}
	return page[pi]
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
	di, pi := m.locate(key)
	page := m.dir[di]
	if page == nil {
		page = make([]map[string]any, m.pageSize)
		m.dir[di] = page
	}
	b := page[pi]
	if b == nil {
		b = make(map[string]any)
		page[pi] = b
	}
	_, updated = b[key]
	b[key] = value
	if !updated {
		m.size++
	}
	return updated
}

// Remove deletes key from the map, returning the removed value and whether the
// key existed. Empty buckets and pages are kept allocated to avoid churn on the
// parent objects.
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
// whether iteration was stopped early. Order is deterministic — pages in index
// order, buckets in index order, entries in per-bucket insertion order — but NOT
// sorted by key.
func (m *Map) Iterate(cb func(key string, value any) bool) bool {
	for _, page := range m.dir {
		for _, b := range page {
			for k, v := range b {
				if cb(k, v) {
					return true
				}
			}
		}
	}
	return false
}
