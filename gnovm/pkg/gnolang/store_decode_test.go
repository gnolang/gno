package gnolang

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/require"
)

// TestObjectDecodeMatchesReflect checks that the decode loadObjectSafe uses
// produces the same object as reflection decoding, for every object kind a
// realm persists, and that both re-encode to the stored bytes.
func TestObjectDecodeMatchesReflect(t *testing.T) {
	_, _, objs := persistObjectLoadRealm(t)
	kinds := map[string]int{}
	for _, o := range objs {
		kinds[o.kind]++
		var viaReflect, viaGen Object
		amino.MustUnmarshal(o.bz, &viaReflect)
		require.NoError(t, amino.UnmarshalAnyBinary2(o.bz, &viaGen))
		normalizeDecodedCap(viaReflect)
		normalizeDecodedCap(viaGen)
		require.Equal(t, viaReflect, viaGen, "%s %v", o.kind, o.oid)
		require.Equal(t, o.bz, amino.MustMarshalAny(viaGen), "%s %v", o.kind, o.oid)
	}
	t.Logf("objects by kind: %v", kinds)
	for _, k := range []string{"Block", "FuncValue", "HeapItemValue", "PackageValue", "StructValue"} {
		require.NotZero(t, kinds[k], "realm persisted no %s; kinds: %v", k, kinds)
	}
}
