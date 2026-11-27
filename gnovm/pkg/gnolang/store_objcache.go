package gnolang

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// Decoded-object cache.
//
// Every transaction starts with an empty object cache, so it re-reads and
// re-decodes every object it touches, even when the previous transaction just
// decoded the very same bytes. Amino decode is the dominant cost of an object
// load. This cache keeps a pristine decoded copy of recently loaded objects,
// keyed by the SHA-256 of their stored amino bytes, and serves a fresh copy on
// a hit instead of decoding again.
//
// Consensus rule: the state machine must not be able to tell a hit from a
// miss. loadObjectSafe consults this cache only after the backend read (which
// charges I/O gas) and after the decode gas charge, and everything after the
// decode (allocator charge, SetHash, fillPackage, fillTypesOfValue) runs on
// the returned copy exactly as on a freshly decoded object. Gas, allocation
// and results are therefore independent of the cache contents, which may
// differ freely between nodes, restarts, CheckTx and queries. This is the
// pattern GetTypeSafe already uses for types (aminoCache).
//
// Hazards and how they are closed:
//   - A transaction never sees a cached object, only a copy, so a mutation by
//     a transaction that later fails cannot reach the cache.
//   - Entries are keyed by content. An object a realm modifies is stored
//     under new bytes, so every later load misses the old entry; no
//     invalidation is needed, and stale entries are simply never hit again.
//   - Entries are inserted from a copy taken immediately after decode, before
//     anything touches the decoded object.
//   - The cache is bounded (objectDecodeCacheMaxCost, objectDecodeCacheCost) and
//     evicts with ristretto's policy: TinyLFU admission, sampled LFU eviction.
//     Eviction is unobservable by the rule above, so it may differ by node.
//
// Built with -tags debugAssert, every hit is checked against a fresh decode of
// the same bytes and panics on any divergence (see verifyCachedObject).

// objectDecodeCacheMaxCost bounds the cache; see objectDecodeCacheCost.
const objectDecodeCacheMaxCost = 64 << 20

var globalObjectDecodeCache = sync.OnceValue(func() *objectDecodeCache {
	return newObjectDecodeCache(objectDecodeCacheMaxCost)
})

// objectDecodeCacheDisabled turns the cache off; for tests and benchmarks.
var objectDecodeCacheDisabled atomic.Bool

type objectDecodeCache struct {
	rc     *ristretto.Cache[[]byte, Object]
	hits   atomic.Uint64
	misses atomic.Uint64
}

func newObjectDecodeCache(maxCost int64) *objectDecodeCache {
	rc, err := ristretto.NewCache(&ristretto.Config[[]byte, Object]{
		NumCounters:        max(maxCost/64, 1000), // ~10x the expected entry count at ~600 B/object
		MaxCost:            maxCost,
		BufferItems:        64,
		IgnoreInternalCost: true,
	})
	if err != nil {
		panic(err)
	}
	return &objectDecodeCache{rc: rc}
}

// decode returns the object encoded by bz: a copy of the cached pristine
// object on a hit, a fresh amino decode on a miss. The returned object is
// owned by the caller.
func (c *objectDecodeCache) decode(bz []byte) Object {
	if c == nil || objectDecodeCacheDisabled.Load() {
		return decodeObject(bz)
	}
	sum := sha256.Sum256(bz)
	if cached, ok := c.rc.Get(sum[:]); ok {
		c.hits.Add(1)
		oo := copyValueWithRefs(cached).(Object)
		if debugAssert {
			verifyCachedObject(bz, oo)
		}
		return oo
	}
	c.misses.Add(1)
	oo := decodeObject(bz)
	c.rc.Set(sum[:], copyValueWithRefs(oo).(Object), objectDecodeCacheCost(bz, oo))
	return oo
}

// objectDecodeCacheCost is the larger of the encoded size and the allocator's
// estimate of the decoded object, so that an object whose encoding is much
// smaller than its decoded form (say, a long list of zero values) is charged
// for the memory it takes.
func objectDecodeCacheCost(bz []byte, oo Object) int64 {
	return max(int64(len(bz)), oo.GetShallowSize()+internalRefSize(oo))
}

func decodeObject(bz []byte) Object {
	var oo Object
	amino.MustUnmarshal(bz, &oo)
	// Must precede GetShallowSize: it charges by capacity.
	normalizeDecodedCap(oo)
	return oo
}

// verifyCachedObject panics unless oo, served from the cache, is
// indistinguishable from a fresh decode of bz in everything the VM charges or
// persists: its re-encoding, shallow size, internal ref size, ObjectInfo and
// full Go structure.
func verifyCachedObject(bz []byte, oo Object) {
	fresh := decodeObject(bz)
	if a, b := amino.MustMarshalAny(oo), amino.MustMarshalAny(fresh); !bytes.Equal(a, b) {
		panic(fmt.Sprintf("object decode cache: re-encoding diverges for %v:\ncached %X\nfresh  %X",
			oo.GetObjectID(), a, b))
	}
	if a, b := oo.GetShallowSize(), fresh.GetShallowSize(); a != b {
		panic(fmt.Sprintf("object decode cache: shallow size %d != %d for %v", a, b, oo.GetObjectID()))
	}
	if a, b := internalRefSize(oo), internalRefSize(fresh); a != b {
		panic(fmt.Sprintf("object decode cache: ref size %d != %d for %v", a, b, oo.GetObjectID()))
	}
	if a, b := *oo.GetObjectInfo(), *fresh.GetObjectInfo(); !reflect.DeepEqual(a, b) {
		panic(fmt.Sprintf("object decode cache: object info %v != %v", &a, &b))
	}
	if !reflect.DeepEqual(oo, fresh) {
		panic(fmt.Sprintf("object decode cache: cached copy of %v differs from a fresh decode",
			oo.GetObjectID()))
	}
}

// SetObjectDecodeCacheEnabled turns the process-wide decoded-object cache on
// or off and returns the previous setting. It exists for tests and
// benchmarks: by construction the cache has no effect on gas or state.
func SetObjectDecodeCacheEnabled(enabled bool) (prev bool) {
	return !objectDecodeCacheDisabled.Swap(!enabled)
}

// ObjectDecodeCacheStats returns the process-wide cache hit and miss counts.
func ObjectDecodeCacheStats() (hits, misses uint64) {
	c := globalObjectDecodeCache()
	return c.hits.Load(), c.misses.Load()
}

// ClearObjectDecodeCache empties the process-wide decoded-object cache.
func ClearObjectDecodeCache() {
	c := globalObjectDecodeCache()
	c.rc.Clear()
}
