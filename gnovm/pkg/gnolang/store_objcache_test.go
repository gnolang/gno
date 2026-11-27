package gnolang

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// encodedStruct returns the stored amino bytes of a persisted StructValue.
func encodedStruct(i int, s string) []byte {
	sv := &StructValue{
		ObjectInfo: ObjectInfo{
			ID:       ObjectID{PkgID: PkgIDFromPkgPath("gno.land/r/test/objcache"), NewTime: uint64(i + 1)},
			RefCount: 1,
		},
		Fields: []TypedValue{
			{T: StringType, V: StringValue(s)},
			{T: IntType},
		},
	}
	sv.Fields[1].SetInt(int64(i))
	return amino.MustMarshalAny(copyValueWithRefs(sv))
}

// A transaction owns what the cache returns: mutating it, as a transaction
// that later fails would, must not reach the cached entry.
func TestObjectDecodeCacheCopyIsolation(t *testing.T) {
	c := newObjectDecodeCache(1 << 20)
	bz := encodedStruct(1, "orig")

	first := c.decode(bz).(*StructValue) // miss: inserts a pristine copy
	c.rc.Wait()
	first.Fields[0].V = StringValue("dirty")
	first.SetIsDirty(true, 1)

	second := c.decode(bz).(*StructValue) // hit
	third := c.decode(bz).(*StructValue)  // hit
	require.Equal(t, uint64(2), c.hits.Load())
	require.NotSame(t, second, third, "every hit must return a fresh copy")

	fresh := decodeObject(bz)
	for _, got := range []Object{second, third} {
		require.Equal(t, amino.MustMarshalAny(fresh), amino.MustMarshalAny(got))
		require.True(t, reflect.DeepEqual(fresh, got), "hit differs from a fresh decode")
	}
	second.Fields[0].V = StringValue("dirty again")
	require.Equal(t, StringValue("orig"), c.decode(bz).(*StructValue).Fields[0].V)
}

// Entries are keyed by content: rewritten bytes never hit the old entry.
func TestObjectDecodeCacheContentKeyed(t *testing.T) {
	c := newObjectDecodeCache(1 << 20)
	before, after := encodedStruct(1, "v1"), encodedStruct(1, "v2")
	require.False(t, bytes.Equal(before, after))
	c.decode(before)
	c.rc.Wait()
	got := c.decode(after).(*StructValue)
	require.Equal(t, uint64(0), c.hits.Load())
	require.Equal(t, StringValue("v2"), got.Fields[0].V)
}

// The cache is bounded by the encoded size of its entries, whatever is fed
// into it.
func TestObjectDecodeCacheBounded(t *testing.T) {
	const maxCost = 16 << 10
	c := newObjectDecodeCache(maxCost)
	var all [][]byte
	var total int
	for i := range 2000 {
		bz := encodedStruct(i, fmt.Sprintf("value-%d", i))
		all = append(all, bz)
		total += len(bz)
		c.decode(bz)
	}
	c.rc.Wait()
	require.Greater(t, total, 10*maxCost, "the test must overflow the bound")
	var held int
	for _, bz := range all {
		sum := sha256.Sum256(bz)
		if _, ok := c.rc.Get(sum[:]); ok {
			held += len(bz)
		}
	}
	require.LessOrEqual(t, held, maxCost)
	require.Greater(t, held, 0)
	t.Logf("fed %d bytes, held %d (bound %d)", total, held, maxCost)
}
